// Copyright 2025 Danzig James Orcales
// 
// 
// Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:
// 
// 
// 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
// 
// 
// 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or
// other materials provided with the distribution.
// 
// 
// 3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without
// specific prior written permission.
// 
// 
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS “AS IS” AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY
// DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS
// OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
// NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
package main


import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)


var (
	HOME              = filepath.Clean(os.Getenv("HOME"))
	PATH              = os.Getenv("PATH")
	BIG_BANG_ROOT     = filepath.Clean(os.Getenv("BIG_BANG_ROOT"))
	//                A mirror of the home directory but only hosts dotfiles.
	BIG_BANG_DOTFILES = filepath.Clean(os.Getenv("BIG_BANG_DOTFILES"))
	BIG_BANG_TMP      = filepath.Clean(os.Getenv("BIG_BANG_TMP"))
	BIG_BANG_SHARE    = filepath.Clean(os.Getenv("BIG_BANG_SHARE"))
	BIG_BANG_BIN      = filepath.Clean(os.Getenv("BIG_BANG_BIN"))
	CARGO_HOME        = filepath.Clen(os.Getenv("CARGO_HOME"))
	RUSTUP_HOME       = filepath.Clean(os.Getenv("RUSTUP_HOME"))


	big_bang_dotfiles_common      = filepath.Join(BIG_BANG_DOTFILES, "common")
	big_bang_dotfiles_os_specific = func() string {
		switch runtime.GOOS {
		case "darwin":
			return filepath.Join(BIG_BANG_DOTFILES, "macos")
		case "linux":
			return filepath.Join(BIG_BANG_DOTFILES, "debian")
		default:
			crash("unsupported os")
		}
	}()
)


// TODO: Have checksums for artifacts list and homebrew list where you're forced to update these manually just like with nix. This would need type
// 	 Artifact to implement Stringer
// TODO: CLI commands
func main() {
	artifacts := []Artifact{
		{
			Name: "brew",
			Install: func(logger *Logger) {
				if path := which("brew"); path != "" {
					logger.Info().Msg("homebrew is already installed")
					return 
				}
				logger.Info().Begin("installing")
				defer logger.Info().Done("installing")
				if err := shell_command(nil, "", "sudo", "--validate"); err != nil {
					logger.Error(err).Msg("user must be an administrator to install homebrew")
					return
				}
				if err := shell_command(
					[]string{"NONINTERACTIVE=1"},
					"",
					"/bin/bash", "-c", 
					`$(curl --fail --silent --show-error --location 'https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh')`,
				); err != nil {
					logger.Error(err).Msg("user must be an administrator to install homebrew")
					return
				}
			},
		},
		{
			Name: "cargo",
			Install: func(logger *Logger) {
				has_missing_dependency := false
				for _, dependency := range []string {
					"rustup",
					"cargo",
					"rustc",
				} {
					if path := which(dependency); path != "" {
						assert(filepath.IsAbs(path))
						assert(filepath.IsAbs(BIG_BANG_ROOT))
						if !strings.HasPrefix(path, BIG_BANG_ROOT) {
							has_missing_dependency = true
							break
						}
					} else {
						has_missing_dependency = true
						break
					}
				}
				if !has_missing_dependency {
					return
				}
				logger.Info().Begin("installing")
				defer logger.Info().Done("installing")


				assert(filepath.IsAbs(BIG_BANG_SHARE))
				assert(filepath.IsAbs(CARGO_HOME))
				assert(filepath.IsAbs(RUSTUP_HOME))
				assert(strings.HasPrefix(filepath.Clean(CARGO_HOME),  BIG_BANG_SHARE))
				assert(strings.HasPrefix(RUSTUP_HOME, BIG_BANG_SHARE))
				// Ideally, we want to assert that $CARGO_HOME/bin is in PATH, but this check can break with consecutive path separators. As a
				// workaround, we assert that $CARGO_HOME is in PATH, which still gives us reasonable confidence it's correctly set.
				// Also, don't clean the path here. We want CARGO_HOME to match exactly with PATH
				assert(strings.Contains(PATH, CARGO_HOME))


				if path := which("curl"); path == "" {
					assert(filepath.IsAbs(path))
					assert(filepath.IsAbs(BIG_BANG_ROOT))
					logger.Error().Msg("curl is needed to install rustup")
					return
				}
				if err := shell_command(
					nil, 
					"",
                                        "sh", "-c", 
					`curl --proto '=https' --tlsv1.2 --silent --show-error --fail https://sh.rustup.rs | 
					 	sh -s -- -y --no-modify-path --default-toolchain=stable`,
				); err != nil {
					logger.Error(err).Msg("installing rustup")
					return
				}
			},
		},
		{
			Name:    "fish",
			Install: func(logger *Logger) {
				if path := which("fish"); path != "" {
					assert(filepath.IsAbs(path))
					assert(filepath.IsAbs(BIG_BANG_ROOT))
					if strings.HasPrefix(path, BIG_BANG_ROOT) {
						version_check := exec.Command("fish", "--version")
						expect        := "fish, version 4.0.2-gf1456f970"
						actual, _     := version_check.Output()
						if slices.Equal(actual, string_to_bytes(expect)) {
							return
						} else {
                                                        // TODO: Info().Bytes()
							logger.Info().Str("current_version", string(actual)).Str("target_version", expect).Msg("outdated installation")
						}
					}
				}
				logger.Info().Begin("installing")
				defer logger.Info().Done("installing")
				if path := which("cargo"); path != "" {
					assert(filepath.IsAbs(path))
					assert(filepath.IsAbs(BIG_BANG_SHARE))
					if !strings.HasPrefix(path, BIG_BANG_SHARE) {
						logger.Error().Msg("cargo installation is not within BIG_BANG_SHARE")
						return
					}
				}
				if path := which("git"); path != "" {
					assert(filepath.IsAbs(path))
				}
				assert(strings.HasPrefix(filepath.Clean(os.Getenv("CARGO_HOME")), BIG_BANG_SHARE))


				tmp_dir := filepath.Join(BIG_BANG_TMP, "fish-shell")
				assert(filepath.IsAbs(tmp_dir))
				if err := shell_command(
					nil,
					"",
					"git", "clone", "--quiet", "--depth=1", "--branch=4.0.2", "https://github.com/fish-shell/fish-shell/", tmp_dir,
				); err != nil {
					logger.Error(err).Msg("cloning git repo")
					return
				}
				if err := shell_command(
					nil,
					tmp_dir,
					"cargo", "--quiet", "vendor",
				); err != nil {
					logger.Error(err).Msg("cargo vendor")
					return
				}
				if err := shell_command(
					// Fabian Boehm: https://github.com/fish-shell/fish-shell/issues/10935#issuecomment-2558599433
					[]string{"RUSTFLAGS=-C target-feature=+crt-static"}, 
					tmp_dir,
					"cargo",  "install", "--quiet", "--offline", "--path=.",
					// https://users.rust-lang.org/t/the-source-requires-a-lock-file-to-be-present-first-before-it-can-be-used-against-vendored-source-code/122648
					"--locked",
					// auto generated by `cargo vendor`
					"--config", `source.crates-io.replace-with="vendored-sources"`,
					"--config", `source."git+https://github.com/fish-shell/rust-pcre2?tag=0.2.9-utf32".git="https://github.com/fish-shell/rust-pcre2"`,
					"--config", `source."git+https://github.com/fish-shell/rust-pcre2?tag=0.2.9-utf32".tag="0.2.9-utf32"`,
					"--config", `source."git+https://github.com/fish-shell/rust-pcre2?tag=0.2.9-utf32".replace-with="vendored-sources"`,
					"--config", `source.vendored-sources.directory="vendor"`,
				); err != nil {
					logger.Error(err).Msg("cargo install")
					return
				}
				return
			},
		},
		{
			Name: "tokei",
			Install: func(logger *Logger) {
				if path := which("tokei"); path != "" {
					assert(filepath.IsAbs(path))
					assert(filepath.IsAbs(BIG_BANG_ROOT))
					if strings.HasPrefix(path, BIG_BANG_ROOT) {
                                                return
					}
				}
				logger.Info().Begin("installing")
				defer logger.Info().Done("installing")
				if err := shell_command(nil, "", "cargo", "install", "--quiet", "tokei", "--version=12.1.2"); err != nil {
					logger.Error(err).Msg("cargo install")
				}
			},
		},
		{
			Name:          "nvim",
			Download_Link: "https://github.com/neovim/neovim/releases/download/v0.11.3/nvim-macos-arm64.tar.gz",
			Checksum:      "17d22826f19fe28a11f9ab4bee13c43399fdcce485eabfa2bea6c5b3d660740f",
			Retain_Installation_Dir: true,
		},
		{
			Name:          "fzf",
			Download_Link: "https://github.com/junegunn/fzf/releases/download/v0.64.0/fzf-0.64.0-darwin_arm64.tar.gz",
			Checksum:      "c71d2528e090de5d4765017d745f8a4fed44b43703f93247a28f6dc2aa4c7c01",
		},
		{
			Name:          "fd",
			Download_Link: "https://github.com/sharkdp/fd/releases/download/v10.2.0/fd-v10.2.0-aarch64-apple-darwin.tar.gz",
			Checksum:      "ae6327ba8c9a487cd63edd8bddd97da0207887a66d61e067dfe80c1430c5ae36", //manual
		},
		{
			Name:          "rg",
			Download_Link: "https://github.com/BurntSushi/ripgrep/releases/download/14.1.1/ripgrep-14.1.1-aarch64-apple-darwin.tar.gz",
			Checksum:      "24ad76777745fbff131c8fbc466742b011f925bfa4fffa2ded6def23b5b937be",
		},
		{
			Name:          "lazygit",
			Download_Link: "https://github.com/jesseduffield/lazygit/releases/download/v0.54.1/lazygit_0.54.1_darwin_arm64.tar.gz",
			Checksum:      "25710495177762f9df2dccaf5e7deed8e5ec70871b7ad385cffa8f7de0646d1d",
		},
		{
			Name:          "hyperfine",
			Download_Link: "https://github.com/sharkdp/hyperfine/releases/download/v1.19.0/hyperfine-v1.19.0-aarch64-apple-darwin.tar.gz",
			Checksum:      "502e7c7f99e7e1919321eaa23a4a694c34b1b92d99cbd773a4a2497e100e088f", // manual
		},
		{
			Name:          "fastfetch",
			Download_Link: "https://github.com/fastfetch-cli/fastfetch/releases/download/2.48.1/fastfetch-macos-aarch64.zip",
			Checksum:      "a1279a5a12ab22f33bcede94108ae501c9c8b27a20629b23481f155f69b7f62d",
		},
	}
	brew_file := `
cask "ghostty"
cask "firefox"
cask "microsoft-edge"
cask "cryptomator"
cask "veracrypt"
cask "karabiner-elements"
cask "obs"
`


	logger := New_Logger(Log_Level_Debug)
	switch runtime.GOOS {
	case "windows": 
		crash("it's a cold day in hell eh?")
	case "darwin":
		if runtime.GOARCH != "arm64" {
			crash("let that rest in peace.")
		}
	case "linux": 
		crash("haven't tested this script here. cover x86_64 and arm64. check distro with /etc/os-release")
	default: 
		crash("os unsupported")
	}
	assert(runtime.Version() == "go1.23.12", "only one supported go version")


	prerequisites := map[string]string{
		"git":    "clones the big bang repo hosting big_bang.sh, big_bang.go, and the dotfiles",
		"sh":     "big_bang.sh: shell to execute",
		"curl":   "big_bang.sh: downloads golang",
		"sha256": "big_bang.sh: checksums golang",
		"tar":    "big_bang.sh: unpacks go<version>.tar.gz. also unpacks .xz files because go doesn't have it in the std lib. darn fish shell",
	}
	for dependency := range prerequisites {
		if path, _ := exec.LookPath(dependency); path != "" {
			delete(prerequisites, dependency)
		}
	}
	assert(len(prerequisites) == 0, "%#v", prerequisites)


	var err_setup = func() error {
		for _, dir := range []string{BIG_BANG_ROOT, BIG_BANG_DOTFILES, BIG_BANG_TMP, BIG_BANG_SHARE, BIG_BANG_BIN} {
			assert(filepath.IsAbs(dir), "exported in $ZDOTDIR/.zprofile and sourced by big_bang.sh before calling this script")
		}


		assert(dir_exists(BIG_BANG_ROOT),                 "the repo is cloned into the hardcoded BIG_BANG_ROOT")
		assert(dir_exists(BIG_BANG_DOTFILES),             "included in the big bang repo")
		assert(dir_exists(BIG_BANG_SHARE),                "created by big_bang.sh hosting GOROOT and GOPATH")
		assert(dir_exists(BIG_BANG_BIN),                  "created by big_bang.sh hosting go.exe")
		assert(dir_exists(big_bang_dotfiles_common),      "included in the big bang repo")
		assert(dir_exists(big_bang_dotfiles_os_specific), "included in the big bang repo")


		if err := os.RemoveAll(BIG_BANG_TMP); err != nil { 
			return err 
		}
		if err := os.MkdirAll(BIG_BANG_TMP, 0755); err != nil { 
			return err 
		}
		// Just a safety measure in case I mess up paths. I still use absolute paths for everything.
		if err := os.Chdir(BIG_BANG_ROOT); err != nil { 
			return err 
		}
		return nil
	}()
	defer os.RemoveAll(BIG_BANG_TMP)
	if err_setup != nil {
		logger.Error(err_setup).Msg("initiliazing environment")
		return
	}


	func() {
		type Info struct {
			description string
			action      func()
		}
		commands := map[string]Info{
			"dotfiles_check": Info{
				description: "checks if actual dotfiles match those in BIG_BANG_DOTFILES",
				action: func() {
					files_set := mismatched_dotfiles(logger)
					if len(files_set) == 0 {
						logger.Info().Msg("no mismatches found")
					} else {
						files := make([]string, 0, len(files_set))
						for file := range files_set {
							files = append(files, file)
						}
						logger.Info().Int("count", len(files)).List("file", files...).Msg("found mismatches")
					}
				},
			},
			"dotfiles_sync": Info{
				description: "Synchronizes dotfiles from $BIG_BANG_DOTFILES to $HOME by creating missing files or truncating existing files. It will never delete other files.",
				action: func() {
					files := mismatched_dotfiles(logger)
					if len(files) == 0 {
						logger.Info().Msg("dotfiles are already synced")
						return
					}
					logger.Info().Begin("syncing dotfiles")
					defer logger.Info().Done("syncing dotfiles")
					for relative_path := range files {
						assert(!filepath.IsAbs(relative_path))
						var err_sync = func() error {
							var src = func() string {
								if os_specific_file := filepath.Join(big_bang_dotfiles_os_specific, relative_path); file_exists(os_specific_file) {
									return os_specific_file
								} else {
									return filepath.Join(big_bang_dotfiles_common, relative_path)
								}
							}()
							assert(!is_dir(src))
							contents, err := os.ReadFile(src)
							if err != nil {
								return err
							}
							dst := filepath.Join(HOME, relative_path)
							if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
								return err
							}
							if err := os.WriteFile(filepath.Join(HOME, relative_path), contents, 0644); err != nil {
								return err
							}
							logger.Info().Str("file", strings.TrimPrefix(src, BIG_BANG_DOTFILES)).Msg("updated dotfile")
							return nil
						}()
						if err_sync != nil {
							logger.Error(err_sync).Msg("syncing dotfiles")
							return
						}
					}
				},
			},
			"dependencies_check": Info{
				description: "WIP",
				action: func() {
				},
			},
			"dependencies_sync":  Info{
				description: "WIP",
				action: func() {
				},
			},
		}
		var print_help = func() {
			fmt.Println("big bang replicates NixOS reproducibility on my work machines (MacOS, Debian, NixOS).")
			fmt.Println("Command Overview:")
			fmt.Println("")
			var longest_command_length int
			for command := range commands {
				if len(command) > longest_command_length {
					longest_command_length = len(command)
				}
			}
			assert(longest_command_length > 0)
			for command, info := range commands {
				fmt.Printf("%*s    %s\n", -longest_command_length, command, info.description)
			}
		}
		if len(os.Args) <= 1 {
			print_help()
			os.Exit(0)
		} else {
			command := os.Args[1]
			info, is_valid_command := commands[command]
			if !is_valid_command {
				fmt.Printf("invalid command: %s\n", command)
				print_help()
				os.Exit(1)
			}
			info.action()
		}
	}()
	return


	total_download_time := time.Minute * 10
	ctx, cancel := context.WithTimeout(context.Background(), total_download_time)
	defer cancel()
	var wg sync.WaitGroup
	artifacts_installed := 0
	for _, artifact := range artifacts {
		if path, _ := exec.LookPath(artifact.Name); strings.HasPrefix(path, BIG_BANG_ROOT) {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger_with_artifact_name := logger.With_Str("artifact", artifact.Name)
			if artifact.Install != nil {
				artifact.Install(logger_with_artifact_name)
			} else {
				artifact_handle := download_artifact(ctx, artifact, logger_with_artifact_name)
				if artifact_handle == nil {
					return
				}
				if ok := install_artifact(artifact, artifact_handle, logger_with_artifact_name); !ok {
					return
				}
			}
		}()
	}
	wg.Wait()
	logger.Info().Int("artifacts.installed", artifacts_installed).Int("artifacts.total", len(artifacts)).Done("installing all artifacts!")


	// Don't do this asynchronously.
	func() {
		logger.Info().Begin("installing homebrew bundle")
		defer logger.Info().Done("installing homebrew bundle")
		brew_file_path := filepath.Join(BIG_BANG_ROOT, "Brewfile")
		if err := os.WriteFile(brew_file_path, string_to_bytes(brew_file), 0644); err != nil {
			logger.Error(err).Msg("creating brewfile")
			return
		}
		bundle_install := exec.Command("brew", "bundle", "install", "--file", brew_file_path)
		bundle_install.Stdout = os.Stdout
		bundle_install.Stderr = os.Stderr
		if err := bundle_install.Run(); err != nil {
			logger.Error(err).Msg("something went wrong")
		}
	}()
}


// The returned file paths are relative to BIG_BANG_DOTFILES. Note that BIG_BANG_DOTFILES is a mirror of the home directory.
func mismatched_dotfiles(logger *Logger) (mismatched_files map[string]struct{}) {
	defer func() {
		if len(mismatched_files) > 0 {
			for file := range mismatched_files {
				assert(!filepath.IsAbs(file))
			}
		}
	}()
	assert(filepath.IsAbs(BIG_BANG_DOTFILES))
	assert(dir_exists(BIG_BANG_DOTFILES))
	logger.Info().Begin("finding mismatches")
	defer logger.Info().Done("finding mismatches")


	var same_file_contents = func(a, b string) bool {
		assert(filepath.IsAbs(a))
		assert(filepath.IsAbs(b))
		a_info, err := os.Lstat(a)
		if err != nil {
			return false
		}
		b_info, err := os.Lstat(b)
		if err != nil {
			return false
		}
		assert_location(!a_info.IsDir() && !b_info.IsDir(), "", 1)
		if a_info.Size() != b_info.Size() {
			return false
		} else {
			a_contents, err := os.ReadFile(a)
			if err != nil {
				return false
			}
			b_contents, err := os.ReadFile(b)
			if err != nil {
				return false
			}
			return slices.Equal(a_contents, b_contents)
		}
	}


	mismatched_files = make(map[string]struct{})
	working_directory := big_bang_dotfiles_os_specific
	if error_find_mismatches := filepath.WalkDir(working_directory, func(src_path string, src fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		src_path_relative, _ := filepath.Rel(working_directory, src_path)
		dst_path             := filepath.Join(HOME, src_path_relative)
		if dir_exists(src_path) || dir_exists(dst_path) {
			return nil
		} 
		if !same_file_contents(src_path, dst_path) {
			mismatched_files[src_path_relative] = struct{}{}
		}
		return nil
	}); error_find_mismatches != nil {
		logger.Error(error_find_mismatches).Msg("finding mismatches between big bang dotfiles and actual dotfiles (os_specific)")
		return nil
	}
	working_directory = big_bang_dotfiles_common
	if error_find_mismatches := filepath.WalkDir(working_directory, func(src_path string, src fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		src_path_relative, _ := filepath.Rel(working_directory, src_path)
		dst_path             := filepath.Join(HOME, src_path_relative)
		oss_path             := filepath.Join(big_bang_dotfiles_os_specific, src_path_relative)
		if _, err := os.Lstat(oss_path); err == nil || dir_exists(src_path) || dir_exists(dst_path) {
			noop("skip the os-specific file which is higher priority or skip directories")
			return nil
		} 
		if !same_file_contents(src_path, dst_path) {
			mismatched_files[src_path_relative] = struct{}{}
		}
		return nil
	}); error_find_mismatches != nil {
		logger.Error(error_find_mismatches).Msg("finding mismatches between big bang dotfiles and actual dotfiles (common)")
		return nil
	}
	return mismatched_files
}


// Caller must provide a context.WithTimeout() because this will retry indefinitely.
// If artifact downloads successfully, caller is responsible for closing the file handle. 
// There's only one case where handle == nil and path != "", and thats when resetting the handle cursor positiion. The caller can instead try to open the file
// again.
func download_artifact(ctx context.Context, artifact Artifact, logger *Logger) (artifact_handle *os.File) {
	waiting_time := time.Minute * 5
	retry_event  := logger.Warn()
	var ok bool
	defer func() {
		if ok {
			logger.Info().Done("downloading")
		} else {
			if artifact_handle != nil {
				artifact_handle.Close()
				artifact_handle = nil
			}
		}
	}()
	ok = func() bool {
		first_iteration := true
		for retry_delay_ns := time.Second * 2;; retry_delay_ns *= 2 { 
			if first_iteration {
				select {
				case <- ctx.Done(): return false
				default: first_iteration = false
				}
			} else {
				retry_event.Number("retry_delay(s)", int64(retry_delay_ns / time.Second)).Msg("Retry artifact download")
				retry_event = logger.Warn()
				select {
				case <- ctx.Done(): return false
				case <- time.After(retry_delay_ns): 
					if artifact_handle != nil {
						   err1 := artifact_handle.Truncate(0);           
						_, err2 := artifact_handle.Seek(0, io.SeekStart); 
						if err1 != nil || err2 != nil{ 
							logger.Error(err1, err2).Msg("truncating temporary file for retrying download")
							return false 
						}
					}
				}
			}
			attempt_ctx, attempt_cancel := context.WithTimeout(ctx, waiting_time)
			// TODO: should all of this be one closure? Right now, it'd just pile up goroutines whiie retrying indefinitely
			defer attempt_cancel()
			request,  err := http.NewRequestWithContext(attempt_ctx, http.MethodGet, artifact.Download_Link, nil)
			if err != nil { 
				logger.Error(err).Msg("initializing http request")
				return false
			}
			logger.Info().Begin("downloading")
			response, err := http.DefaultClient.Do(request)
			if err != nil { 
				retry_event.Err(err)
				continue
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusOK {
				retry_event.Int("status_code", response.StatusCode)
				continue
			}
			if artifact_handle == nil {
				filename := func(content_disposition string) string {
					content_disposition_parts := strings.Split(content_disposition, ";")
					if len(content_disposition_parts) < 2 || content_disposition_parts[0] != "attachment"  {
						return ""
					}
					// TODO: support `filename*=UTF-8`
					// https://datatracker.ietf.org/doc/html/rfc5987#section-3.2
					// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Disposition
					filename_key_val := strings.Split(content_disposition_parts[1], "=")
					if len(filename_key_val) != 2 || strings.TrimSpace(filename_key_val[0]) != "filename" {
						return ""
					}
					return strings.Trim(filename_key_val[1], `" `)
				}(response.Header.Get("Content-Disposition"))
				if filename == "" {
					retry_event.Err(errors.New("invalid Content-Disposition header"))
					continue
				}
				download_path := filepath.Join(BIG_BANG_TMP, filepath.Clean(filename))
				var err_artifact_file error
				artifact_handle, err_artifact_file = os.Create(download_path)
				if err_artifact_file != nil {
					logger.Error(err).Str("file", download_path).Msg("creating file to write downloaded artifact to")
					return false
				}
			}
			io.Copy(artifact_handle, response.Body)
			artifact_handle.Sync()
			actual_checksum := hex.EncodeToString(file_checksum(artifact_handle, logger))
			if artifact.Checksum != "" {
				if actual_checksum != artifact.Checksum {
					retry_event.
						Str("expected", artifact.Checksum).
						Str("actual", actual_checksum).
						Err(errors.New("checksum mismatch"))
					continue
				}
			} else {
				logger.Error().Str("checksum", actual_checksum).Msg("unset checksum. copy the calculated checksum and set it in the source code")
				return false
			}
			break
		}
		return true
	}()
	return artifact_handle
}


func install_artifact(artifact Artifact, artifact_handle *os.File, logger *Logger) (ok bool) {
	defer func() {
		if ok {
			logger.Info().Done("installing")
		}
		artifact_handle.Close()
	}()
	_    = artifact_handle.Sync()
	_, _ = artifact_handle.Seek(0, io.SeekStart)


	artifact_tmp_dir  := filepath.Join(BIG_BANG_TMP, artifact.Name)
	os.Mkdir(artifact_tmp_dir, 0755)
	artifact_fullpath := artifact_handle.Name()
	artifact_filename := filepath.Base(artifact_fullpath)
	switch {
	default: 
		logger.Error().Str("file", artifact_filename).Msg("unsupported extension")
		return false
	case strings.HasSuffix(artifact_filename, ".tar.gz"), strings.HasSuffix(artifact_filename, ".tar.xz"):
		var compression_flag string
		switch {
		case strings.HasSuffix(artifact_filename, ".gz"): compression_flag = "--gzip"
		case strings.HasSuffix(artifact_filename, ".xz"): compression_flag = "--xz"
		default: 
			logger.Error().Str("file", artifact_filename).Msg("unsupported tar compresison")
			return false
		}
		tar_xz := exec.Command(
			"tar", 
			"--extract",   compression_flag,
			"--file",      artifact_fullpath,
			"--directory", artifact_tmp_dir,
		)
		if out, err := tar_xz.CombinedOutput(); err != nil {
			logger.Error(errors.New(string(out))).Msg("unpacking .xz file with external tool")
			return false
		}
	case strings.HasSuffix(artifact_filename, ".zip"):
		var unpacking_error = func() error {
			info, err := artifact_handle.Stat(); 
			if err != nil { 
				return err 
			}
			zip_reader, err := zip.NewReader(artifact_handle, info.Size())
			if err != nil { 
				return err 
			}
			for _, entry := range zip_reader.File {
				if strings.Contains(entry.Name, "__MACOSX") {
					continue
				}
				extraction_path := filepath.Join(artifact_tmp_dir, filepath.Clean(entry.Name))
				if entry.FileInfo().IsDir() || filepath.Ext(entry.Name) == ".app" { 
					if err := os.MkdirAll(extraction_path, 0755); err != nil { return err }
					continue 
				} else {
					src, err := entry.Open() 
					if err != nil { 
						return err
					}
					dst, err := os.Create(extraction_path)
					if err != nil { 
						return err
					}
					if _, err := io.CopyN(dst, src, int64(entry.UncompressedSize64)); err != nil { 
						return err 
					}
					src.Close()
					dst.Close()
				}
			}
			return nil
		}()
		if unpacking_error != nil {
			logger.Error(unpacking_error).Msg("unpacking zip file")
		}
	}
	var find_binary func(string) string
	find_binary = func(directory string) string {
		entries, err := os.ReadDir(directory)
		if err != nil {
			logger.Error(err).Str("directory", directory).Msg("finding binary")
			return ""
		}
		var directories []string
		for _, entry := range entries {
			if entry.IsDir() {
				directories = append(directories, entry.Name())
				continue
			}
			binary_path := entry.Name()
			if filepath.Base(binary_path) == artifact.Name {
				return filepath.Join(directory, binary_path)
			}
		}
		for _, child_dir := range directories {
			binary_path := find_binary(filepath.Join(directory, child_dir))
			if binary_path != "" {
				return binary_path
			}
		}
		return ""
	}
	artifact_binary_destination := filepath.Join(BIG_BANG_BIN, artifact.Name)
	os.RemoveAll(artifact_binary_destination)
	if artifact.Retain_Installation_Dir {
		artifact_root_dir := filepath.Join(BIG_BANG_SHARE, artifact.Name)
		os.RemoveAll(artifact_root_dir)
		if err := os.Rename(artifact_tmp_dir, artifact_root_dir); err != nil {
			logger.Error(err).Msg("finalizing artifact installation")
			return false
		}
		artifact_binary_source := find_binary(artifact_root_dir)
		if artifact_binary_source == "" {
			logger.Error().Msg("binary was not found")
			return false
		}
		if err := os.Symlink(artifact_binary_source, artifact_binary_destination); err != nil {
			logger.Error(err).Msg("symlinking artifact binary")
			return false
		}
	} else {
		artifact_binary_source := find_binary(artifact_tmp_dir)
		if artifact_binary_source == "" {
			logger.Error().Msg("binary was not found")
			return false
		}
		if err := os.Rename(artifact_binary_source, artifact_binary_destination); err != nil {
			logger.Error(err).Str("artifact_binary_source", artifact_binary_source).Msg("moving binary to BIG_BANG_BIN")
			return false
		}
	}
	if err := os.Chmod(artifact_binary_destination, 0755); err != nil {
		logger.Error(err).Msg("making artifact binary executable")
		return false
	}
	return true
}


func file_checksum(source *os.File, logger *Logger) []byte {
	debug := logger.Debug().Str("file", source.Name())
	if err := source.Sync(); err != nil { 
		debug.Err(err).Msg("flushing file")
		return nil
	}
	original_position, err := source.Seek(0, io.SeekCurrent)
	if err != nil { 
		debug.Err(err).Msg("saving initial offset")
		return nil
	}
	defer func() {
		if _, err := source.Seek(original_position, io.SeekStart); err != nil { 
			debug.Err(err).Msg("restoring offset")
			return
		}
	}()
	hasher := sha256.New()
	if _, err := source.Seek(0, io.SeekStart); err != nil { debug.Err(err).Msg("resetting offset"); return nil }
	if _, err := io.Copy(hasher, source);      err != nil { debug.Err(err).Msg("hashing file");     return nil }
	return hasher.Sum(nil)
}


func os_remove_if_exists(file_path string) error { 
	if err := os.Remove(file_path); !errors.Is(err, fs.ErrNotExist) { 
		return err 
	} 
	return nil
}


func file_exists(path string) bool {
	_, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	if err != nil {
		crash("checking if file exists: %s", err)
	}
	return true
}


type Artifact struct {
	// the same as the executable name
	Name              string
	Download_Link     string
	Checksum          string
	// If false, deletes BIG_BANG_ROOT/<PROGRAM>/ after installation.
	// Useful for self-contained executables with no other files unlike Golang with its stdlib or nvim with its runtime directories.
	// Instead of symlinking the executable to BIG_BANG_BIN, it gets moved there instead.
	Retain_Installation_Dir bool
	Install	func(*Logger)
	Healthcheck     []string
}


func crash(format_string string, args... any) {		
	fmt.Printf("ERROR "+format_string+"\n", args...)
	os.Exit(1)
}
func echo(format_string string, args... any) {		
	fmt.Printf(format_string+"\n", args...)
}
func assert(cond bool, message_and_args ...any) {
	const (
		stackframe_of_this_func int = iota
		stackframe_of_the_caller_of_this_func
	)
	if cond {
		return
	}
	msg := ""
	if len(message_and_args) > 0 {
                var is_string bool
		msg, is_string = message_and_args[0].(string)
                if is_string {
			if len(message_and_args) > 1 {
				args := message_and_args[1:]
				msg = fmt.Sprintf(msg, args...)
			}
		} else {
			msg = "formatting string is invalid"
		}
	}
	assert_location(cond, msg, stackframe_of_the_caller_of_this_func)
}
func assert_location(cond bool, message string, skip int) {
	if !cond {
		_, file, line, ok := runtime.Caller(skip + 1)
		if ok {
			fmt.Printf("%s:%d ASSERTION FAILED: %s\n", file, line, message)
		} else {
			fmt.Printf("ASSERTION FAILED: %s", message)
		}
		os.Exit(1)
	}
}


/*
https://patorjk.com/software/taag/#p=display&v=0&f=Terrace&t=logger.go
░██                                                                                
░██                                                                                
░██  ░███████   ░████████  ░████████  ░███████  ░██░████      ░████████  ░███████  
░██ ░██    ░██ ░██    ░██ ░██    ░██ ░██    ░██ ░███         ░██    ░██ ░██    ░██ 
░██ ░██    ░██ ░██    ░██ ░██    ░██ ░█████████ ░██          ░██    ░██ ░██    ░██ 
░██ ░██    ░██ ░██   ░███ ░██   ░███ ░██        ░██          ░██   ░███ ░██    ░██ 
░██  ░███████   ░█████░██  ░█████░██  ░███████  ░██      ░██  ░█████░██  ░███████  
                      ░██        ░██                                ░██            
                ░███████   ░███████                           ░███████             
*/
var Log_Writer io.Writer = os.Stdout
const (
	// These defaults cover most cases. Note that these buffers can still grow when the need arises, in which case, they don't get returned to the
	// pool but are instead left for the garbage collector.
	Log_Default_Buffer_Capacity = 
		Log_Time_Capacity    + len(" ") +
		Log_Level_Capacity   + len(" ") +
		Log_Message_Capacity + len(" ") +
		Log_Context_Capacity + len("\n")
	Log_Time_Capacity    = len(time.RFC3339) - len("07:30") // it will always be in UTC
	Log_Level_Capacity   = len("ERROR")
	Log_Message_Capacity = 100
	Log_Context_Capacity = 400


	Log_Level_Debug    = -10
	Log_Level_Info     =   0
	Log_Level_Warn     =  10
	Log_Level_Error    =  20
	Log_Level_Disabled =  50


	// TODO: implement. This allows only needing to escape the log separator instead of currently space and double quotes. It also is stricter on the
	// separation of components. But this comes at the cost of readability.
	Log_Component_Separator = '|'
)
func New_Logger(level int) *Logger {
	if level >= Log_Level_Disabled || Log_Writer == nil {
		return nil
	}
	return &Logger{
		Buffer: make([]byte, 0, Log_Context_Capacity / 2),
		Level: level,
	}
}


// TODO: func (logger *Logger) Assert(cond bool) *Log Event { ... }
func (logger *Logger) Debug() *Log_Event { if logger == nil || logger.Level > Log_Level_Debug { return nil } else { return init_log_event(logger, "DEBUG ") } }
func (logger *Logger) Info()  *Log_Event { if logger == nil || logger.Level > Log_Level_Info  { return nil } else { return init_log_event(logger, "INFO  ") } }
func (logger *Logger) Warn()  *Log_Event { if logger == nil || logger.Level > Log_Level_Warn  { return nil } else { return init_log_event(logger, "WARN  ") } }
// The errs parameter is merely for convenience.
func (logger *Logger) Error(errs ...error) *Log_Event { 
	if logger == nil || logger.Level > Log_Level_Error { return nil } 
	event := init_log_event(logger, "ERROR ") 
	switch len(errs) {
	case 0:  noop()
	case 1:  event = event.Err(errs[0])
	default: event = event.Errs(errs...)
	}
	return event
}


func (logger *Logger) With_Str (key string, val string) *Logger { if logger == nil { return nil }; return logger.With_Data_Quoted(key, val         ) }
func (logger *Logger) With_Err (key string, val error ) *Logger { if logger == nil { return nil }; return logger.With_Data_Quoted(key, val.Error() ) }
func (logger *Logger) With_Int (key string, val int   ) *Logger { if logger == nil { return nil }; return logger.With_Data(key, strconv.Itoa(val)  ) }
func (logger *Logger) With_Bool(key string, cond bool ) *Logger { 
	if logger == nil { 
		return nil 
	}
	val := "false"
	if cond {
		val = "true"
	}
	return logger.With_Data(key, val)
}
func (logger *Logger) With_Data(key string, val string) *Logger { 
	if logger == nil { 
		return nil 
	} 
	dst := New_Logger(logger.Level)
	copy(dst.Buffer, logger.Buffer)
	var space, underscore byte = ' ', '_'
	append_and_replace(&logger.Buffer, string_to_bytes(key), space, underscore)
	dst.Buffer = append(dst.Buffer, '=')
	dst.Buffer = append(dst.Buffer, string_to_bytes(val)...)
	dst.Buffer = append(dst.Buffer, ' ')
	return dst
}
func (logger *Logger) With_Data_Quoted(key string, val string) *Logger { 
	if logger == nil { 
		return nil 
	} 
	dst := New_Logger(logger.Level)
	copy(dst.Buffer, logger.Buffer)
	var space, underscore byte = ' ', '_'
	append_and_replace(&dst.Buffer, string_to_bytes(key), space, underscore)
	dst.Buffer = append(dst.Buffer, '=', '"')
	append_and_replace(&dst.Buffer, string_to_bytes(val), '"', '\\', '"') 
	dst.Buffer = append(dst.Buffer, '"', ' ')
	return dst
}


// func (event *Log_Event) Bytes (key string, val string) *Log_Event { if event == nil { return nil }; return event.Data_Quoted(key, val) }
func (event *Log_Event) Str (key string, val string) *Log_Event { if event == nil { return nil }; return event.Data_Quoted(key, val) }
func (event *Log_Event) Err (err error ) *Log_Event { 
	if event == nil { 
		return nil 
	} 
	err_str := "nil" 
	if err != nil { 
		err_str = err.Error() 
	} 
	return event.Data_Quoted("error", err_str) 
}
func (event *Log_Event) Errs(vals ...error) *Log_Event { 
	if event == nil || len(vals) == 0  { return nil }
	first_error := true
	for _, v := range vals {
		if v == nil {
			continue
		}
		if first_error {
			event.Buffer = append(event.Buffer, string_to_bytes("errors")...)
			event.Buffer = append(event.Buffer, '=', '[')
		}
		first_error = false
		event.Buffer = append(event.Buffer, ' ', '"')
		append_and_replace(&event.Buffer, string_to_bytes(v.Error()), '"', '\\', '"')
		event.Buffer = append(event.Buffer, '"', ' ')
	}
	event.Buffer = append(event.Buffer, ' ', ']', ' ')
	return event
}


func (event *Log_Event) Int   (key string, val int    ) *Log_Event { if event == nil { return nil }; return event.Number(key,  int64(val)) }
func (event *Log_Event) Int8  (key string, val int8   ) *Log_Event { if event == nil { return nil }; return event.Number(key,  int64(val)) }
func (event *Log_Event) Int16 (key string, val int16  ) *Log_Event { if event == nil { return nil }; return event.Number(key,  int64(val)) }
func (event *Log_Event) Int32 (key string, val int32  ) *Log_Event { if event == nil { return nil }; return event.Number(key,  int64(val)) }
func (event *Log_Event) Int64 (key string, val int64  ) *Log_Event { if event == nil { return nil }; return event.Number(key,  int64(val)) }
func (event *Log_Event) Uint  (key string, val uint   ) *Log_Event { if event == nil { return nil }; return event.Uint64(key, uint64(val)) }
func (event *Log_Event) Uint8 (key string, val uint8  ) *Log_Event { if event == nil { return nil }; return event.Number(key,  int64(val)) }
func (event *Log_Event) Uint16(key string, val uint16 ) *Log_Event { if event == nil { return nil }; return event.Number(key,  int64(val)) }
func (event *Log_Event) Uint32(key string, val uint32 ) *Log_Event { if event == nil { return nil }; return event.Number(key,  int64(val)) }
func (event *Log_Event) Uint64(key string, val uint64 ) *Log_Event { 
	if event == nil { 
		return nil 
	}
	var space, underscore byte = ' ', '_'
	append_and_replace(&event.Buffer, string_to_bytes(key), space, underscore)
	event.Buffer = append(event.Buffer, '=')
	event.Buffer = strconv.AppendUint(event.Buffer, val, 10)
	event.Buffer = append(event.Buffer, ' ')
	return event
}
func (event *Log_Event) Number(key string, val int64 ) *Log_Event { 
	if event == nil { 
		return nil 
	}
	var space, underscore byte = ' ', '_'
	append_and_replace(&event.Buffer, string_to_bytes(key), space, underscore)
	event.Buffer = append(event.Buffer, '=')
	event.Buffer = strconv.AppendInt(event.Buffer, val, 10)
	event.Buffer = append(event.Buffer, ' ')
	return event
}
func (event *Log_Event) Bool(key string, cond bool) *Log_Event { 
	if event == nil { 
		return nil 
	}
	val := "false"
	if cond {
		val = "true"
	}
	return event.Data(key, val)
}
func (event *Log_Event) List(key string, vals ...string) *Log_Event {
	if event == nil {
		return nil
	}
	var space, underscore byte = ' ', '_'
	append_and_replace(&event.Buffer, string_to_bytes(key), space, underscore)
	event.Buffer = append(event.Buffer, '=', '[')
	for _, v := range vals {
		event.Buffer = append(event.Buffer, space)
		event.Buffer = append(event.Buffer, string_to_bytes(v)...)
	}
	event.Buffer = append(event.Buffer, space, ']', space)
	return event
}
func (event *Log_Event) Data(key, val string) *Log_Event {
	if event == nil {
		return nil
	}
	var space, underscore byte = ' ', '_'
	append_and_replace(&event.Buffer, string_to_bytes(key), space, underscore)
	event.Buffer = append(event.Buffer, '=')
	event.Buffer = append(event.Buffer, string_to_bytes(val)...)
	event.Buffer = append(event.Buffer, ' ')
	return event
}
func (event *Log_Event) Data_Quoted(key, val string) *Log_Event {
	if event == nil {
		return nil
	}
	var space, underscore byte = ' ', '_'
	append_and_replace(&event.Buffer, string_to_bytes(key), space, underscore)
	event.Buffer = append(event.Buffer, '=', '"')
	append_and_replace(&event.Buffer, string_to_bytes(val), '"', '\\', '"') 
	event.Buffer = append(event.Buffer, '"', ' ')
	return event
}
func (event *Log_Event) Begin(msg string) { if event == nil { return }; event.Msg("begin "+msg) }
func (event *Log_Event) Done (msg string) { if event == nil { return }; event.Msg("done  "+msg) }
func (event *Log_Event) Message(msg string) *Log_Event {
	if event == nil {
		return nil
	}
	start := Log_Time_Capacity + len(" ") + Log_Level_Capacity + len(" ")
	end   := start + Log_Message_Capacity
	message_buffer := event.Buffer[start:end]
	if len(msg) <= Log_Message_Capacity { 
		for offset := range msg {
			message_buffer[offset] = msg[offset]
		}
		for offset := range message_buffer[len(msg):] {
			message_buffer[len(msg)+offset] = ' '
		}
	} else {
		placeholder := "..."
		before := message_buffer[:len(message_buffer) - len(placeholder) ]
		after  := message_buffer[ len(message_buffer) - len(placeholder):]
		for offset := range before {
			before[offset] = msg[offset]
		}
		for offset := range after {
			after[offset] = '.'
		}
	}
	return event
}
func (event *Log_Event) Msg  (message string) { if event == nil { return }; event.Message(message).Send() }
func (event *Log_Event) Fatal(message string) { if event == nil { return }; event.Message("fatal: "+ message).Send(); os.Exit(1) }
func (event *Log_Event) Send() {
	if event == nil {
		return
	}
	defer deinit_log_event(event)
	event.Buffer = append(event.Buffer, '\n')
	assert(func() bool {
		for offset := range event.Buffer {
			if event.Buffer[offset] == 0 {
				return false
			}
		}
		return true
	}())
	if _, err := Log_Writer.Write(event.Buffer); err != nil {
		Log_Writer.Write(string_to_bytes("could not write log event"))
	}
}


func init_log_event(logger *Logger, log_level_str string) *Log_Event {
	if logger == nil {
		return nil
	}
	event := Log_Event_Pool.Get().(*Log_Event)
	event.Buffer = event.Buffer[:0]


	append_zero_padded_int := func(buf []byte, v int) []byte {
		if v < 10 {
			buf = append(buf, '0')
		}
		return strconv.AppendInt(buf, int64(v), 10)
	}
	t := time.Now().UTC()
	// Append YYYY-MM-DD
	event.Buffer = strconv.AppendInt(event.Buffer, int64(t.Year()), 10)
	event.Buffer = append(event.Buffer, '-')
	event.Buffer = append_zero_padded_int(event.Buffer, int(t.Month()))
	event.Buffer = append(event.Buffer, '-')
	event.Buffer = append_zero_padded_int(event.Buffer, t.Day())
	// Append THH:mm:ss
	event.Buffer = append(event.Buffer, 'T')
	event.Buffer = append_zero_padded_int(event.Buffer, t.Hour())
	event.Buffer = append(event.Buffer, ':')
	event.Buffer = append_zero_padded_int(event.Buffer, t.Minute())
	event.Buffer = append(event.Buffer, ':')
	event.Buffer = append_zero_padded_int(event.Buffer, t.Second())
	event.Buffer = append(event.Buffer, 'Z', ' ')


	// assert.Always(strings.HasSuffix(log_level_str, " "))
	// assert.Always(len(log_level_str) == 6)
	event.Buffer = append(event.Buffer, string_to_bytes(log_level_str)...)


	// Set the slice length past the Msg() portion, starting at the context.
	element_size := int(unsafe.Sizeof(event.Buffer[0]))
	space := 1
	previous := (*reflect.SliceHeader)(unsafe.Pointer(&event.Buffer))
	current  := reflect.SliceHeader{
		Data: previous.Data,
		Cap:  previous.Cap, // doesn't matter because we must stay within the capacity
		Len: 
			Log_Time_Capacity    * element_size + space * element_size +
			Log_Level_Capacity   * element_size + space * element_size +
			Log_Message_Capacity * element_size,
	}
	event.Buffer = *(*[]byte)(unsafe.Pointer(&current))
	event.Buffer = append(event.Buffer, logger.Buffer...)
	event.Buffer = append(event.Buffer, ' ')
	return event
}


func deinit_log_event(event *Log_Event) {
	if cap(event.Buffer) > Log_Default_Buffer_Capacity {
		return
	}
	Log_Event_Pool.Put(event)
}


type Logger struct {
	// To be inherited by a Log_Event created by its methods.
	Buffer     []byte
	Level      int
}
// A transient object that should not be touched after writing to stdout. Minimize scope as much as possible, usually in one statement. If you're passing this
// as a function parameter, embed that context in the Logger instead.
type  Log_Event struct {
	Buffer []byte
	// Doesn't need to hold log level. Logger.<Method>() *Log_Event {...} returns a nil event if a message shouldn't be logged. This nil check will be a
	// no-op that permeates through the rest of the method chain e.g. logger.Info().Str("key", "value").Msg("foo bar baz").
	// Level  int
}
var Log_Event_Pool = &sync.Pool{
	New: func() any { 
		return &Log_Event{
			Buffer: make([]byte, 0, Log_Default_Buffer_Capacity),
		} 
	},
}


func string_to_bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
func append_and_replace(dst *[]byte, src []byte, old byte, new ...byte) {
	for offset := 0; offset < len(src); offset += 1 {
		ch := src[offset]
		if ch == old {
			*dst = append(*dst, new...)
		} else {
			*dst = append(*dst, ch)
		}
	}
}


func is_dir(path string) bool {
	return dir_exists(path)
}
func dir_exists(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}


// no operation function used for explicitness
func noop(_ ...string) {}


func shell_command(environment []string, working_directory string, binary string, arguments ...string) error {
	cmd := exec.Command(binary, arguments...)
	if len(environment) > 0 {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, environment...)
	}
	if working_directory != "" {
		assert(filepath.IsAbs(working_directory))
		os.MkdirAll(working_directory, 0755)
		cmd.Dir = working_directory
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin  = os.Stdin
	return cmd.Run()
}


func which(name string) (path string) {
	path_raw, err := exec.Command("command", "-v", name).Output()
	if err != nil {
		return ""
	}
        path = filepath.Clean(string(path_raw))
	assert(filepath.IsAbs(path))
	return path
}
