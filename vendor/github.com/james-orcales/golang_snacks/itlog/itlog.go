package itlog

import (
	"errors"
	"io"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/james-orcales/golang_snacks/invariant"
)

const (
	DefaultLoggerBufferCapacity = DefaultEventBufferCapacity / 2
	DefaultEventBufferCapacity  = HeaderCapacity + 1 + ContextCapacity + len("\n")
	HeaderCapacity              = TimestampCapacity + 1 + LevelCapacity + 1 + MessageCapacity

	TimestampCapacity = len(time.RFC3339) - len("07:00") // hardcoded to always be in UTC
	LevelCapacity     = LevelMaxWordLength
	// You can get a 10ns/op improvement if you reduce this down to 50
	// charaters but it's not worth it for such a short message window.
	MessageCapacity = 80
	// This should cover most cases. Note that the buffers can still grow
	// when the need arises, in which case, they don't get returned to the
	// pool but are instead left for the garbage collector.
	ContextCapacity = 300

	LevelMaxWordLength = 3
	LevelDebug         = -100
	LevelInfo          = 0
	LevelWarn          = 100
	LevelError         = 200
	LevelDisabled      = 500

	// Note: This implementation assumes that these separators are
	// characters and thus have a length of 1.
	ComponentDelimiter   = '|'
	KeyValDelimiter      = '='
	Quote                = '"'
	EmptyIndicatorString = "__EMPTY__"
)

var (
	EmptyIndicatorBytes = []byte(EmptyIndicatorString)
	TickCallback        = func() time.Time {
		return time.Now().UTC()
	}
)

// === Encoding ===
//
// - backslash (0x5C)    -> `\\`
// - quote     (0x22)    -> `\"`
// - newline   (0x0A)    -> `\n`
// - nullbyte  (0x00)    -> `\0`
//
// Every null byte in the log is escaped, ensuring that any
// unescaped null byte appearing in the logs can be safely interpreted as data
// corruption.
//
// Notes:
// Other non-readable characters remain unchanged.
// UTF is not handled.
//
// The resulting encoding is **lossless** — it can be decoded back to the original data.
func appendEscaped(dst, src []byte) []byte {
	invariant.Always(dst != nil, "appendEscaped must receive a non-nil slice pointer")
	invariant.Always(src != nil, "appendEscaped caller handles empty string argument")

	before := len(dst)
	defer func() {
		after := len(dst)
		invariant.Sometimes(after-before == len(src), "String to encode was appended as is")
		invariant.Sometimes(after-before > len(src), "String to encode contained escaped characters")
	}()
	for _, ch := range src {
		switch ch {
		case '\\':
			invariant.Sometimes(true, "String to encode contains escaped bytes")
			dst = append(dst, '\\', '\\')
		case '"':
			invariant.Sometimes(true, "String to encode contains Quote")
			dst = append(dst, '\\', Quote)
		case '\n':
			invariant.Sometimes(true, "String to encode contains raw newline")
			dst = append(dst, '\\', 'n')
		case 0:
			invariant.Sometimes(true, "String to encode contains raw null byte")
			dst = append(dst, '\\', '0')
		default:
			dst = append(dst, ch)
		}
	}
	return dst
}

func ValidateKey(key []byte) error {
	if len(key) == 0 {
		return errors.New("Key is empty")
	}
	period := 0
	underscore := 0
	for _, ch := range key {
		switch {
		case 'a' <= ch && ch <= 'z':
		case 'A' <= ch && ch <= 'Z':
		case '0' <= ch && ch <= '9':
		case ch == '.':
			period++
		case ch == '_':
			underscore++
		default:
			return errors.New("Log context key must contain alphanumeric, periods, and underscores only")
		}
	}
	if period == len(key) {
		return errors.New("Log context must not only contain periods")
	} else if underscore == len(key) {
		return errors.New("Log context must not only contain underscores")
	} else if period+underscore == len(key) {
		return errors.New("Log context must not only contain periods and underscores")
	}
	return nil
}

func (lgr *Logger) Clone() *Logger {
	if lgr == nil {
		invariant.Unreachable("Logger.Clone callers don't propagate nil loggers")
		return nil
	}

	invariant.Always(cap(lgr.Buffer) >= DefaultLoggerBufferCapacity, "All loggers have at least DefaultLoggerBufferCapacity")
	invariant.Sometimes(len(lgr.Buffer) == 0, "Logger has no inheritable context")
	invariant.Sometimes(len(lgr.Buffer) > 0, "Logger has inheritable context")

	dst := New(lgr.Writer, lgr.Level)
	// Assume that the inherited buffer was already processed by appendEscaped
	dst.Buffer = append(dst.Buffer, lgr.Buffer...)

	return dst
}

func New(writer io.Writer, level int) *Logger {
	if level >= LevelDisabled {
		invariant.Sometimes(true, "Logger is disabled completely")
		return nil
	}
	if writer == nil {
		invariant.Sometimes(true, "log Writer is nil")
		return nil
	}
	return &Logger{
		Writer: writer,
		Buffer: make([]byte, 0, DefaultLoggerBufferCapacity),
		Level:  level,
	}
}

func (lgr *Logger) Debug() *Event {
	if lgr == nil {
		invariant.Sometimes(true, "Logger.Debug Logger is nil")
		return nil
	} else if lgr.Level > LevelDebug {
		invariant.Sometimes(true, "Debug level and below is disabled")
		return nil
	}
	invariant.Sometimes(true, "Create debug log")
	return lgr.newEvent("DBG")
}

func (lgr *Logger) Info() *Event {
	if lgr == nil {
		invariant.Sometimes(true, "Logger.Info Logger is nil")
		return nil
	} else if lgr.Level > LevelInfo {
		invariant.Sometimes(true, "Info level and below is disabled")
		return nil
	}
	invariant.Sometimes(true, "Create info log")
	return lgr.newEvent("INF")
}

func (lgr *Logger) Warn() *Event {
	if lgr == nil {
		invariant.Sometimes(true, "Logger.Warn Logger is nil")
		return nil
	} else if lgr.Level > LevelWarn {
		invariant.Sometimes(true, "Warn level and below is disabled")
		return nil
	}
	invariant.Sometimes(true, "Create warn log")
	return lgr.newEvent("WRN")
}

// The errs parameter is mainly convenience. One benefit of it however, the
// error is ensured to be the first key value pair in the context if the parent
// logger doesn't have a context to inherit.
func (lgr *Logger) Error(errs ...error) *Event {
	if lgr == nil {
		invariant.Sometimes(true, "Logger.Error Logger is nil")
		return nil
	} else if lgr.Level > LevelError {
		invariant.Sometimes(true, "Logger.Error error level and below is disabled")
		return nil
	}
	ev := lgr.newEvent("ERR")
	switch len(errs) {
	case 0:
		invariant.Sometimes(true, "Logger.Error has zero arguments")
		noop()
	case 1:
		invariant.Sometimes(true, "Logger.Error has one argument")
		ev = ev.Err(errs[0])
	default:
		invariant.Sometimes(true, "Logger.Error has multiple arguments")
		ev = ev.Errs(errs...)
	}
	return ev
}

func (lgr *Logger) WithData(key, val []byte) *Logger {
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithData Logger is nil")
		return nil
	}
	// These are invalid but we want this logger to be fault tolerant
	if len(key) == 0 {
		invariant.Sometimes(true, "Logger.WithData key is empty")
		key = EmptyIndicatorBytes
	}
	if len(val) == 0 {
		invariant.Sometimes(true, "Logger.WithData val is empty")
		val = EmptyIndicatorBytes
	}
	invariant.XAlwaysNil(func() any { return ValidateKey(key) }, "Log context key is valid")

	lgr.Buffer = append(lgr.Buffer, key...)
	lgr.Buffer = append(lgr.Buffer, KeyValDelimiter)
	lgr.Buffer = append(lgr.Buffer, val...)
	lgr.Buffer = append(lgr.Buffer, ComponentDelimiter)

	invariant.Always(lgr.Buffer[0] != ComponentDelimiter, "Logger's context is appended AFTER ComponentDelimiter")
	return lgr
}

// With* functions create a deep copy of logger and appends context to the Buffer.
func (lgr *Logger) WithStr(key, val string) *Logger {
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithStr Logger is nil")
		return nil
	}
	// These are invalid but we want this logger to be fault tolerant
	if key == "" {
		invariant.Sometimes(true, "Logger.WithStr key is empty")
		key = EmptyIndicatorString
	}
	if val == "" {
		invariant.Sometimes(true, "Logger.WithStr val is empty")
		val = EmptyIndicatorString
	}
	invariant.XAlwaysNil(func() any { return ValidateKey(stringToBytesUnsafe(key)) }, "Log context key is valid")

	lgr.Buffer = append(lgr.Buffer, stringToBytesUnsafe(key)...)
	lgr.Buffer = append(lgr.Buffer, KeyValDelimiter)
	lgr.Buffer = append(lgr.Buffer, Quote)
	lgr.Buffer = appendEscaped(lgr.Buffer, stringToBytesUnsafe(val))
	lgr.Buffer = append(lgr.Buffer, Quote)
	lgr.Buffer = append(lgr.Buffer, ComponentDelimiter)

	invariant.Always(lgr.Buffer[0] != ComponentDelimiter, "Logger's context is appended AFTER ComponentDelimiter")
	return lgr
}

func (lgr *Logger) WithErr(key string, val error) *Logger {
	invariant.Sometimes(key == "", "Logger.WithErr is empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithErr Logger is nil")
		return nil
	}
	if val != nil {
		lgr = lgr.WithStr(key, val.Error())
	} else {
		invariant.Sometimes(true, "Logger.WithErr got nil error")
	}
	return lgr
}

func (lgr *Logger) WithInt(key string, val int) *Logger {
	invariant.Sometimes(key == "", "Logger.WithInt is empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithInt Logger is nil")
		return nil
	}
	return lgr.WithInt64(key, int64(val))
}

func (lgr *Logger) WithInt8(key string, val int8) *Logger {
	invariant.Sometimes(key == "", "Logger.WithInt8 is empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithInt8 Logger is nil")
		return nil
	}
	return lgr.WithInt64(key, int64(val))
}

func (lgr *Logger) WithInt16(key string, val int16) *Logger {
	invariant.Sometimes(key == "", "Logger.WithInt16 empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithInt16 Logger is nil")
		return nil
	}
	return lgr.WithInt64(key, int64(val))
}

func (lgr *Logger) WithInt32(key string, val int32) *Logger {
	invariant.Sometimes(key == "", "Logger.WithInt32 empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithInt32 Logger is nil")
		return nil
	}
	return lgr.WithInt64(key, int64(val))
}

func (lgr *Logger) WithInt64(key string, val int64) *Logger {
	invariant.Sometimes(key == "", "Logger.WithInt64 empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithInt64 Logger is nil")
		return nil
	}

	array := [64]byte{}
	buf := array[:0]
	buf = strconv.AppendInt(buf, val, 10)
	return lgr.WithData(stringToBytesUnsafe(key), buf)
}

func (lgr *Logger) WithUint(key string, val uint) *Logger {
	invariant.Sometimes(key == "", "Event.Uint is empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithUint Logger is nil")
		return nil
	}
	return lgr.WithInt64(key, int64(val))
}

func (lgr *Logger) WithUint8(key string, val uint8) *Logger {
	invariant.Sometimes(key == "", "Event.Uint8 key is empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithUint8 Logger is nil")
		return nil
	}
	return lgr.WithInt64(key, int64(val))
}

func (lgr *Logger) WithUint16(key string, val uint16) *Logger {
	invariant.Sometimes(key == "", "Event.Uint16 key is empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithUint16 Logger is nil")
		return nil
	}
	return lgr.WithInt64(key, int64(val))
}

func (lgr *Logger) WithUint32(key string, val uint32) *Logger {
	invariant.Sometimes(key == "", "Event.Uint32 key is empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithUint32 Logger is nil")
		return nil
	}
	return lgr.WithInt64(key, int64(val))
}

func (lgr *Logger) WithUint64(key string, val uint64) *Logger {
	invariant.Sometimes(key == "", "Event.Uint64 key is empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithUint64 Logger is nil")
		return nil
	}
	array := [64]byte{}
	buf := array[:0]
	buf = strconv.AppendUint(buf, val, 10)
	return lgr.WithData(stringToBytesUnsafe(key), buf)
}

func (lgr *Logger) WithBool(key string, cond bool) *Logger {
	invariant.Sometimes(key == "", "Event.WithBool key is empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithBool Logger is nil")
		return nil
	}

	val := []byte{'f', 'a', 'l', 's', 'e'}
	if cond {
		invariant.Sometimes(cond, "Logger.WithBool cond is true")
		val = []byte{'t', 'r', 'u', 'e'}
	}

	return lgr.WithData(stringToBytesUnsafe(key), val)
}

func (lgr *Logger) WithFloat32(key string, val float32) *Logger {
	invariant.Sometimes(key == "", "Event.WithFloat32 key is empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithFloat32 Logger is nil")
		return nil
	}
	// overcompensate
	array := [32]byte{}
	buf := array[:0]
	buf = strconv.AppendFloat(buf, float64(val), 'e', -1, 32)
	return lgr.WithData(stringToBytesUnsafe(key), buf)
}

func (lgr *Logger) WithFloat64(key string, val float64) *Logger {
	invariant.Sometimes(key == "", "Event.WithFloat64 key is empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithFloat64 Logger is nil")
		return nil
	}
	// overcompensate
	array := [64]byte{}
	buf := array[:0]
	buf = strconv.AppendFloat(buf, float64(val), 'e', -1, 64)
	return lgr.WithData(stringToBytesUnsafe(key), buf)
}

func (lgr *Logger) WithTime(key string, t time.Time) *Logger {
	invariant.Sometimes(key == "", "Logger.WithTime key is empty")
	if lgr == nil {
		invariant.Sometimes(true, "Logger.WithTime Logger is nil")
		return nil
	}
	array := [TimestampCapacity]byte{}
	buf := array[:0]
	buf = appendTime(buf, t)
	return lgr.WithData(stringToBytesUnsafe(key), buf)
}

// Convenience functions to help guide your message. With these prefixes, you'd
// want to start with verbs in the present-progressive form.
// Usage:
//
//	lgr.Info().Begin("fetching zip")
//	lgr.Info().Begin("extracting zip")
func (ev *Event) Begin(msg string) {
	if ev == nil {
		invariant.Sometimes(true, "Event.Begin Event is nil")
		return
	}
	invariant.Always(msg != "", "Empty Event.Begin verb")
	ev.Msg("begin " + msg)
}

// Don't use these like tracing logs. Instead of deferring unconditionally,
// manually write this as the very last statement of a function to indicate
// success. Otherwise, you probably have some WARN/ERROR log outputted
// beforehand, making it redundant.
func (ev *Event) Done(msg string) {
	if ev == nil {
		invariant.Sometimes(true, "Event.Done Event is nil")
		return
	}
	invariant.Always(msg != "", "Empty Event.Done verb")
	ev.Msg("done  " + msg)
}

func (ev *Event) Data(key, val []byte) *Event {
	if ev == nil {
		invariant.Sometimes(true, "Event.Data event is nil")
		return nil
	}
	// These are invalid but we want this logger to be fault tolerant
	if len(key) == 0 {
		invariant.Sometimes(true, "Event.Data key is empty")
		key = EmptyIndicatorBytes
	}
	if len(val) == 0 {
		invariant.Sometimes(true, "Event.Data val is empty")
		val = EmptyIndicatorBytes
	}
	invariant.XAlwaysNil(func() any { return ValidateKey(key) }, "Log context key is valid")

	ev.Buffer = append(ev.Buffer, key...)
	ev.Buffer = append(ev.Buffer, KeyValDelimiter)
	ev.Buffer = append(ev.Buffer, val...)
	ev.Buffer = append(ev.Buffer, ComponentDelimiter)

	return ev
}

// String context values are wrapped in quotes. Quote characters inside the
// value are left unescaped. Parsing requires at least two quote delimiters;
// everything between them is taken literally.
func (ev *Event) Str(key, val string) *Event {
	if ev == nil {
		invariant.Sometimes(true, "Event.Str event is nil")
		return nil
	}
	if key == "" {
		invariant.Sometimes(true, "Event.Str string is empty")
		key = EmptyIndicatorString
	}
	if val == "" {
		invariant.Sometimes(true, "Event.Str val is empty")
		val = EmptyIndicatorString
	}
	invariant.XAlwaysNil(func() any { return ValidateKey(stringToBytesUnsafe(key)) }, "Log context key is valid")

	ev.Buffer = append(ev.Buffer, stringToBytesUnsafe(key)...)
	ev.Buffer = append(ev.Buffer, KeyValDelimiter)
	ev.Buffer = append(ev.Buffer, Quote)
	ev.Buffer = appendEscaped(ev.Buffer, stringToBytesUnsafe(val))
	ev.Buffer = append(ev.Buffer, Quote)
	ev.Buffer = append(ev.Buffer, ComponentDelimiter)

	return ev
}

func (ev *Event) Strs(key string, strs ...string) *Event {
	if ev == nil {
		invariant.Sometimes(true, "Event.Strs event is nil")
		return nil
	}
	if len(strs) == 0 {
		invariant.Unreachable("Event.Strs has at least one variable")
		return ev
	}
	if key == "" {
		key = EmptyIndicatorString
	}
	invariant.XAlwaysNil(func() any { return ValidateKey(stringToBytesUnsafe(key)) }, "Log context key is valid")

	ev.Buffer = appendEscaped(ev.Buffer, stringToBytesUnsafe(key))
	ev.Buffer = append(ev.Buffer, KeyValDelimiter, '[', ' ')
	for _, str := range strs {
		ev.Buffer = append(ev.Buffer, Quote)
		ev.Buffer = appendEscaped(ev.Buffer, stringToBytesUnsafe(str))
		ev.Buffer = append(ev.Buffer, Quote, ' ')
	}
	ev.Buffer = append(ev.Buffer, ']', ComponentDelimiter)

	return ev
}

func (ev *Event) Err(err error) *Event {
	if ev == nil {
		invariant.Sometimes(ev == nil, "Event.Err Event is nil")
		return nil
	}
	if err == nil {
		invariant.Sometimes(true, "Event.Err got nil error")
		return ev
	}
	ev = ev.Str("error", err.Error())
	return ev
}

func (ev *Event) Errs(vals ...error) *Event {
	invariant.Always(len(vals) > 0, "Event.Errs takes at least one error")
	if ev == nil {
		invariant.Sometimes(ev == nil, "Event.Errs Event is nil")
		return nil
	}
	nilCount := 0
	for _, v := range vals {
		if v == nil {
			nilCount++
		} else {
			ev = ev.Err(v)
		}
	}
	invariant.Sometimes(nilCount == 0, "All arguments to Event.<level>.Errs are non-nil")
	invariant.Sometimes(nilCount > 0, "Some arguments to Event.<level>.Errs are non-nil")
	invariant.Sometimes(nilCount == len(vals), "All arguments to Event.<level>.Errs are nil")
	return ev
}

func (ev *Event) Int(key string, val int) *Event {
	if ev == nil {
		invariant.Sometimes(true, "Event.Int Event is nil")
		return nil
	}
	return ev.Int64(key, int64(val))
}

func (ev *Event) Int8(key string, val int8) *Event {
	if ev == nil {
		invariant.Sometimes(true, "Event.Int8 Event is nil")
		return nil
	}
	return ev.Int64(key, int64(val))
}

func (ev *Event) Int16(key string, val int16) *Event {
	if ev == nil {
		invariant.Sometimes(true, "Event.Int16 Event is nil")
		return nil
	}
	return ev.Int64(key, int64(val))
}

func (ev *Event) Int32(key string, val int32) *Event {
	if ev == nil {
		invariant.Sometimes(true, "Event.Int32 Event is nil")
		return nil
	}
	return ev.Int64(key, int64(val))
}

func (ev *Event) Int64(key string, val int64) *Event {
	invariant.Sometimes(key == "", "Event.Int64 key is not empty")
	if ev == nil {
		invariant.Sometimes(true, "Event.Int64 Event is nil")
		return nil
	}
	array := [64]byte{}
	buf := array[:0]
	buf = strconv.AppendInt(buf, val, 10)
	return ev.Data(stringToBytesUnsafe(key), buf)
}

func (ev *Event) Uint(key string, val uint) *Event {
	invariant.Sometimes(key == "", "Event.Uint key is not empty")
	if ev == nil {
		invariant.Sometimes(true, "Event.Uint Event is nil")
		return nil
	}
	return ev.Uint64(key, uint64(val))
}

func (ev *Event) Uint8(key string, val uint8) *Event {
	invariant.Sometimes(key == "", "Event.Uint8 key is not empty")
	if ev == nil {
		invariant.Sometimes(true, "Event.Uint8 Event is nil")
		return nil
	}
	return ev.Int64(key, int64(val))
}

func (ev *Event) Uint16(key string, val uint16) *Event {
	invariant.Sometimes(key == "", "Event.Uint16 key is not empty")
	if ev == nil {
		invariant.Sometimes(true, "Event.Uint16 Event is nil")
		return nil
	}
	return ev.Int64(key, int64(val))
}

func (ev *Event) Uint32(key string, val uint32) *Event {
	invariant.Sometimes(key == "", "Event.Uint32 key is not empty")
	if ev == nil {
		invariant.Sometimes(true, "Event.Uint32 Event is nil")
		return nil
	}
	return ev.Int64(key, int64(val))
}

func (ev *Event) Uint64(key string, val uint64) *Event {
	invariant.Sometimes(key == "", "Event.Uint64 key is not empty")
	if ev == nil {
		invariant.Sometimes(true, "Event.Uint64 Event is nil")
		return nil
	}
	array := [64]byte{}
	buf := array[:0]
	buf = strconv.AppendUint(buf, val, 10)
	return ev.Data(stringToBytesUnsafe(key), buf)
}

func (ev *Event) Float32(key string, val float32) *Event {
	invariant.Sometimes(key == "", "Event.Float32 key is empty")
	if ev == nil {
		invariant.Sometimes(true, "Event.Float32 Event is nil")
		return nil
	}
	// overcompensate
	array := [32]byte{}
	buf := array[:0]
	buf = strconv.AppendFloat(buf, float64(val), 'e', -1, 32)
	return ev.Data(stringToBytesUnsafe(key), buf)
}

func (ev *Event) Float64(key string, val float64) *Event {
	invariant.Sometimes(key == "", "Event.Float64 key is empty")
	if ev == nil {
		invariant.Sometimes(true, "Event.Float64 Event is nil")
		return nil
	}
	// overcompensate
	array := [64]byte{}
	buf := array[:0]
	buf = strconv.AppendFloat(buf, float64(val), 'e', -1, 64)
	return ev.Data(stringToBytesUnsafe(key), buf)
}

func (ev *Event) Bool(key string, cond bool) *Event {
	if ev == nil {
		invariant.Sometimes(true, "Event.Bool Event is nil")
		return nil
	}
	val := []byte{'f', 'a', 'l', 's', 'e'}
	if cond {
		invariant.Sometimes(cond, "Event.Bool cond is true")
		val = []byte{'t', 'r', 'u', 'e'}
	}
	invariant.Sometimes(!cond, "Event.Bool cond is false")

	return ev.Data(stringToBytesUnsafe(key), val)
}

func (ev *Event) Time(key string, t time.Time) *Event {
	invariant.Sometimes(key == "", "Event.Time key is empty")
	if ev == nil {
		invariant.Sometimes(true, "Event.Time Logger is nil")
		return nil
	}
	array := [TimestampCapacity]byte{}
	buf := array[:0]
	buf = appendTime(buf, t)
	return ev.Data(stringToBytesUnsafe(key), buf)
}

// Msg is a short summary of your log entry, similar to a git commit message.
// Msg asserts that msg does not contain a raw newline or raw null byte.
// If msg is longer than MessageCapacity, it gets truncated with no indicator.
func (ev *Event) Msg(msg string) {
	if ev == nil {
		invariant.Sometimes(true, "Event.Msg event is nil")
		return
	}
	defer ev.destroy()

	invariant.Sometimes(len(msg) < MessageCapacity, "Message didn't fill the sub buffer")
	invariant.Sometimes(len(msg) == MessageCapacity, "Message fills the sub buffer exactly")
	invariant.Sometimes(len(msg) > MessageCapacity, "Message overfills the sub buffer")
	if msg == "" {
		invariant.Sometimes(true, "Log message is empty")
	}

	// insert message
	{
		start := HeaderCapacity - MessageCapacity
		end := HeaderCapacity
		invariant.Always(len(ev.Buffer) >= end, "Length is unsafely set past message buffer during event init")

		buf := ev.Buffer[start:end]
		i := 0
		for ; i < min(len(buf), len(msg)); i++ {
			ch := msg[i]
			if ch == 0 || ch == '\n' {
				buf[i] = ' '
			} else {
				buf[i] = ch
			}
		}
		for ; i < len(buf); i++ {
			buf[i] = ' '
		}
	}

	// assert valid log
	{
		_, err := time.Parse(time.RFC3339, bytesToStringUnsafe(ev.Buffer[:TimestampCapacity]))
		invariant.Always(err == nil, "Timestamp is valid RFC3339")
		invariant.Always(ev.Buffer[TimestampCapacity] == ComponentDelimiter, "ComponentDelimiter found after timestamp")
		invariant.Always(func() bool {
			switch bytesToStringUnsafe(ev.Buffer[TimestampCapacity+1 : TimestampCapacity+1+LevelCapacity]) {
			case "DBG", "INF", "WRN", "ERR":
				return true
			default:
				return false
			}
		}(), "Timestamp level word is either DBG, INF, WRN, or ERR")
		invariant.Always(ev.Buffer[TimestampCapacity+1+LevelCapacity] == ComponentDelimiter, "ComponentDelimiter found after level word")

		{
			description := "Log message does not contain raw newlines or null bytes"
			invariant.XAlways(func() bool {
				for i := range ev.Buffer[TimestampCapacity+1+LevelCapacity+1 : HeaderCapacity] {
					i += TimestampCapacity + 1 + LevelCapacity + 1
					ch := ev.Buffer[i]
					if ch == '\n' {
						return false
					} else if ch == 0 {
						return false
					}
				}
				return true
			}, description)
		}

		{
			invariant.XAlways(func() bool {
				escaped := false
				for i := range ev.Buffer[HeaderCapacity+1:] {
					i += HeaderCapacity + 1
					ch := ev.Buffer[i]
					if ch == '\n' {
						return false
					} else if ch == '\\' {
						escaped = !escaped
						continue
					}
					escaped = false
				}
				return true
			}, "Log context contains raw newline")
		}
	}

	ev.Buffer = append(ev.Buffer, '\n')
	invariant.Always(ev.Writer != nil, "A logger with a nil writer never initializes an event")
	n, err := ev.Writer.Write(ev.Buffer)
	if err != nil {
		os.Stderr.Write(stringToBytesUnsafe("WRITE_ERROR|could not write log event\n"))
	}

	invariant.Sometimes(n > DefaultEventBufferCapacity, "Log exceeded default buffer size")
}

func (lgr *Logger) newEvent(level string) *Event {
	invariant.Always(lgr != nil, "Callers of Logger.newEvent don't propagate nil loggers")
	invariant.Always(len(level) == LevelMaxWordLength, "Level string is equal to LevelMaxWordLength")
	invariant.Sometimes(level == "DBG", "New event is level DBG")
	invariant.Sometimes(level == "INF", "New event is level INF")
	invariant.Sometimes(level == "WRN", "New event is level WRN")
	invariant.Sometimes(level == "ERR", "New event is level ERR")

	ev := EventPool.Get().(*Event)
	invariant.Sometimes(len(ev.Buffer) > 0, "sync.Pool reused Event with leftover data")
	ev.Buffer = ev.Buffer[:0]
	ev.Writer = lgr.Writer

	t := TickCallback().UTC()
	// Append YYYY-MM-DD
	invariant.Always(len(ev.Buffer) == 0, "Buffer was cleared before being written to")
	ev.Buffer = appendTime(ev.Buffer, t)
	ev.Buffer = append(ev.Buffer, ComponentDelimiter)
	invariant.Always(len(ev.Buffer) == TimestampCapacity+1, "Wrote exactly N bytes for Timestamp+ComponentDelimiter")

	ev.Buffer = append(ev.Buffer, stringToBytesUnsafe(level)...)
	invariant.Always(len(ev.Buffer) == TimestampCapacity+1+LevelCapacity, "Wrote exactly N bytes for Level")
	ev.Buffer = append(ev.Buffer, ComponentDelimiter)

	// Set the slice length past the Msg() portion, starting at the context.
	elementSize := int(unsafe.Sizeof(ev.Buffer[0]))
	previous := (*reflect.SliceHeader)(unsafe.Pointer(&ev.Buffer))
	current := reflect.SliceHeader{
		Data: previous.Data,
		Cap:  previous.Cap,
		Len:  (len(ev.Buffer) + MessageCapacity) * elementSize,
	}
	ev.Buffer = *(*[]byte)(unsafe.Pointer(&current))
	invariant.Always(len(ev.Buffer) == HeaderCapacity, "Skipped past the message sub buffer during initialization")
	invariant.Always(len(ev.Buffer)+1 < cap(ev.Buffer), "Default buffer size is greater than HeaderCapacity+ComponentDelimiter")
	ev.Buffer = append(ev.Buffer, ComponentDelimiter)
	invariant.Always(ev.Buffer[HeaderCapacity] == ComponentDelimiter, "Component separator after header was set during event initialization")
	ev.Buffer = append(ev.Buffer, lgr.Buffer...)
	return ev
}

func appendTime(dst []byte, t time.Time) []byte {
	append_zero_pad := func(buf []byte, v int) []byte {
		if v < 10 {
			buf = append(buf, '0')
		}
		return strconv.AppendInt(buf, int64(v), 10)
	}
	dst = strconv.AppendInt(dst, int64(t.Year()), 10)
	dst = append(dst, '-')
	dst = append_zero_pad(dst, int(t.Month()))
	dst = append(dst, '-')
	dst = append_zero_pad(dst, t.Day())
	dst = append(dst, 'T')
	dst = append_zero_pad(dst, t.Hour())
	dst = append(dst, ':')
	dst = append_zero_pad(dst, t.Minute())
	dst = append(dst, ':')
	dst = append_zero_pad(dst, t.Second())
	return append(dst, 'Z')
}

func (ev *Event) destroy() {
	if ev == nil {
		invariant.Unreachable("Event.Destroy caller does not propagate nil event")
		return
	}
	if cap(ev.Buffer) > DefaultEventBufferCapacity {
		invariant.Sometimes(true, "Event with oversized buffer isn't returned to the pool")
		noop()
	} else {
		EventPool.Put(ev)
	}
}

// Logger is a long-lived object that primarily holds context data to be
// inherited by all of its child Events. All of Logger's methods that append to
// the context buffer create a new copy of Logger.
type Logger struct {
	Writer io.Writer
	// To be inherited by a Event created by its methods.
	Buffer []byte
	Level  int
}

// Event is a transient object that should not be touched after writing to
// Writer. Minimize scope as much as possible, usually in one statement. If you
// find yourself passing this as a function parameter, embed that context in the
// Logger instead. Event methods modify the Event itself through a pointer
// receiver.
type Event struct {
	Writer io.Writer
	Buffer []byte
	// The log level is intentionally omitted from Event. Logger.<Level>()
	// methods return nil if the event should not be logged, allowing method
	// chains like Logger.Info().Str("key", "val").Msg("msg") to no-op
	// safely. This design eliminates the need to check the log level inside
	// Event itself.
}

var EventPool = &sync.Pool{
	New: func() any {
		return &Event{
			Buffer: make([]byte, 0, DefaultEventBufferCapacity),
		}
	},
}

func noop() {
}

func stringToBytesUnsafe(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func bytesToStringUnsafe(b []byte) string {
	return unsafe.String(&b[0], len(b))
}
