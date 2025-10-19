# laytan Github Gist: LLDB script to visualise Odin slices, maps, and strings 
# https://gist.github.com/laytan/a94c323a84cef7bcfbdf6d21987fd5a9
import lldb


def is_slice_type(t, internal_dict):
    return t.name.startswith("[]") or t.name.startswith("[dynamic]")


class SliceChildProvider:
    def __init__(self, val, dict):
        self.val = val


    def num_children(self):
        return self.val.GetChildMemberWithName("len").GetValueAsSigned()


    def get_child_at_index(self, index):
        assert index >= 0 and index < self.num_children()


        data = self.val.GetChildMemberWithName("data")
        assert data.type.is_pointer
        first = data.deref


        offset = index * first.size
        return first.CreateChildAtOffset(f"[{index}]", offset, first.type)


def is_string_type(t, internal_dict):
    return t.name == "string"


def string_summary(value, internal_dict):
    pointer = value.GetChildMemberWithName("data").GetValueAsUnsigned(0)
    length = value.GetChildMemberWithName("len").GetValueAsSigned(0)
    if pointer == 0:
        return False
    if length == 0:
        return '""'
    error = lldb.SBError()
    string_data = value.process.ReadMemory(pointer, length, error)
    return '"{}"'.format(string_data.decode("utf-8"))


def is_map_type(t, internal_dict):
    return t.name.startswith("map[")


class MapChildProvider:


    def __init__(self, val, dict):
        self.val = val


    def num_children(self):
        return (self.val.GetChildMemberWithName("len").GetValueAsSigned() * 2) + 1


    def get_child_at_index(self, index):
        data = self.val.GetChildMemberWithName("data")
        tkey = data.GetChildMemberWithName("key").type
        tval = data.GetChildMemberWithName("value").type
        hash_field = data.GetChildMemberWithName("hash")
        key_cell   = data.GetChildMemberWithName("key_cell")
        value_cell = data.GetChildMemberWithName("value_cell")


        raw_data = data.GetValueAsUnsigned()
        key_ptr = raw_data & ~63
        cap_log2 = raw_data & 63
        cap = 0 if cap_log2 <= 0 else 1 << cap_log2


        key_cell_info = self.cell_info(tkey, key_cell)
        value_cell_info = self.cell_info(tval, value_cell)


        size_of_hash = hash_field.size
        assert size_of_hash == 8


        value_ptr = self.cell_index(key_ptr, key_cell_info, cap)
        hash_ptr =  self.cell_index(value_ptr, value_cell_info, cap)


        error = lldb.SBError()


        # Last one, the capacity.
        if index == self.num_children()-1:
            cap_data = lldb.SBData.CreateDataFromInt(cap)
            return self.val.CreateValueFromData("cap", cap_data, self.val.GetChildMemberWithName("len").type)


        wants_key = index % 2 == 0
        index = int(index / 2)


        key_index = 0
        for i in range(cap):
            TOMBSTONE_MASK = 1 << (size_of_hash*8 - 1)


            offset_hash = hash_ptr + i * size_of_hash


            hash_val = self.val.process.ReadUnsignedFromMemory(offset_hash, size_of_hash, error)
            if not error.success:
                print(error)
                continue
            elif hash_val == 0 or (hash_val & TOMBSTONE_MASK) != 0:
                continue


            offset_key   = self.cell_index(key_ptr, key_cell_info, i)
            offset_value = self.cell_index(value_ptr, value_cell_info, i)


            if index == key_index:
                if wants_key:
                    return self.val.CreateValueFromAddress(f"[{i}]", offset_key, tkey)
                else:
                    return self.val.CreateValueFromAddress(f"[{i}]", offset_value, tval)


            key_index += 1


        print("not found")


    def cell_info(self, typev, cell_type):
        elements_per_cell = 0


        if typev.size != cell_type.size:
            array_type = cell_type.children[0].type
            if array_type.size > 0 and typev.size > 0:
                elements_per_cell = array_type.size / typev.size


        if elements_per_cell == 0:
            elements_per_cell = 1


        return CellInfo(typev.size, cell_type.size, elements_per_cell)


    def cell_index(self, base, info, index):
        cell_index = 0
        data_index = 0
        if info.elements_per_cell == 1:
            return base + (index * info.size_of_cell)
        elif info.elements_per_cell == 2:
            cell_index = index >> 1;
            data_index = index & 1;
        elif info.elements_per_cell == 4:
            cell_index = index >> 2;
            data_index = index & 3;
        elif info.elements_per_cell == 8:
            cell_index = index >> 3;
            data_index = index & 7;
        elif info.elements_per_cell == 16:
            cell_index = index >> 4;
            data_index = index & 15;
        elif info.elements_per_cell == 32:
            cell_index = index >> 5;
            data_index = index & 31;
        else:
            cell_index = index / elements_per_cell;
            data_index = index % elements_per_cell;


        return base + (cell_index * info.size_of_cell) + (data_index * info.size_of_type);


class CellInfo:
    def __init__(self, size_of_type, size_of_cell, elements_per_cell):
        self.size_of_type = size_of_type
        self.size_of_cell = size_of_cell
        self.elements_per_cell = elements_per_cell


def __lldb_init_module(debugger, unused):
    debugger.HandleCommand(
            "type summary add --recognizer-function --python-function odin.string_summary odin.is_string_type"
            )
    debugger.HandleCommand(
            "type synth add --recognizer-function --python-class odin.SliceChildProvider odin.is_slice_type"
            )
    debugger.HandleCommand(
            "type synth add --recognizer-function --python-class odin.MapChildProvider odin.is_map_type"
            )
