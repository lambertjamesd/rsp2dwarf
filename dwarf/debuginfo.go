package dwarf

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/lambertjamesd/rsp2dwarf/elf"
)

type DW_TAG uint32

const (
	DW_TAG_array_type             DW_TAG = 0x01
	DW_TAG_class_type             DW_TAG = 0x02
	DW_TAG_entry_point            DW_TAG = 0x03
	DW_TAG_enumeration_type       DW_TAG = 0x04
	DW_TAG_formal_parameter       DW_TAG = 0x05
	DW_TAG_imported_declaration   DW_TAG = 0x08
	DW_TAG_label                  DW_TAG = 0x0a
	DW_TAG_lexical_block          DW_TAG = 0x0b
	DW_TAG_member                 DW_TAG = 0x0d
	DW_TAG_pointer_type           DW_TAG = 0x0f
	DW_TAG_reference_type         DW_TAG = 0x10
	DW_TAG_compile_unit           DW_TAG = 0x11
	DW_TAG_string_type            DW_TAG = 0x12
	DW_TAG_structure_type         DW_TAG = 0x13
	DW_TAG_subroutine_type        DW_TAG = 0x15
	DW_TAG_typedef                DW_TAG = 0x16
	DW_TAG_union_type             DW_TAG = 0x17
	DW_TAG_unspecified_parameters DW_TAG = 0x18
	DW_TAG_variant                DW_TAG = 0x19
	DW_TAG_common_block           DW_TAG = 0x1a
	DW_TAG_common_inclusion       DW_TAG = 0x1b
	DW_TAG_inheritance            DW_TAG = 0x1c
	DW_TAG_inlined_subroutine     DW_TAG = 0x1d
	DW_TAG_module                 DW_TAG = 0x1e
	DW_TAG_ptr_to_member_type     DW_TAG = 0x1f
	DW_TAG_set_type               DW_TAG = 0x20
	DW_TAG_subrange_type          DW_TAG = 0x21
	DW_TAG_with_stmt              DW_TAG = 0x22
	DW_TAG_access_declaration     DW_TAG = 0x23
	DW_TAG_base_type              DW_TAG = 0x24
	DW_TAG_catch_block            DW_TAG = 0x25
	DW_TAG_const_type             DW_TAG = 0x26
	DW_TAG_constant               DW_TAG = 0x27
	DW_TAG_enumerator             DW_TAG = 0x28
	DW_TAG_file_type              DW_TAG = 0x29
	DW_TAG_friend                 DW_TAG = 0x2a
	DW_TAG_namelist               DW_TAG = 0x2b
	DW_TAG_namelist_item          DW_TAG = 0x2c
	DW_TAG_packed_type            DW_TAG = 0x2d
	DW_TAG_subprogram             DW_TAG = 0x2e
	DW_TAG_template_type_param    DW_TAG = 0x2f
	DW_TAG_template_value_param   DW_TAG = 0x30
	DW_TAG_thrown_type            DW_TAG = 0x31
	DW_TAG_try_block              DW_TAG = 0x32
	DW_TAG_variant_part           DW_TAG = 0x33
	DW_TAG_variable               DW_TAG = 0x34
	DW_TAG_volatile_type          DW_TAG = 0x35
	DW_TAG_lo_user                DW_TAG = 0x4080
	DW_TAG_hi_user                DW_TAG = 0xffff
)

type DW_AT uint32

const (
	DW_AT_sibling              DW_AT = 0x01
	DW_AT_location             DW_AT = 0x02
	DW_AT_name                 DW_AT = 0x03
	DW_AT_ordering             DW_AT = 0x09
	DW_AT_byte_size            DW_AT = 0x0b
	DW_AT_bit_offset           DW_AT = 0x0c
	DW_AT_bit_size             DW_AT = 0x0d
	DW_AT_stmt_list            DW_AT = 0x10
	DW_AT_low_pc               DW_AT = 0x11
	DW_AT_high_pc              DW_AT = 0x12
	DW_AT_language             DW_AT = 0x13
	DW_AT_discr                DW_AT = 0x15
	DW_AT_discr_value          DW_AT = 0x16
	DW_AT_visibility           DW_AT = 0x17
	DW_AT_import               DW_AT = 0x18
	DW_AT_string_length        DW_AT = 0x19
	DW_AT_common_reference     DW_AT = 0x1a
	DW_AT_comp_dir             DW_AT = 0x1b
	DW_AT_const_value          DW_AT = 0x1c
	DW_AT_containing_type      DW_AT = 0x1d
	DW_AT_default_value        DW_AT = 0x1e
	DW_AT_inline               DW_AT = 0x20
	DW_AT_is_optional          DW_AT = 0x21
	DW_AT_lower_bound          DW_AT = 0x22
	DW_AT_producer             DW_AT = 0x25
	DW_AT_prototyped           DW_AT = 0x27
	DW_AT_return_addr          DW_AT = 0x2a
	DW_AT_start_scope          DW_AT = 0x2c
	DW_AT_stride_size          DW_AT = 0x2e
	DW_AT_upper_bound          DW_AT = 0x2f
	DW_AT_abstract_origin      DW_AT = 0x31
	DW_AT_accessibility        DW_AT = 0x32
	DW_AT_address_class        DW_AT = 0x33
	DW_AT_artificial           DW_AT = 0x34
	DW_AT_base_types           DW_AT = 0x35
	DW_AT_calling_convention   DW_AT = 0x36
	DW_AT_count                DW_AT = 0x37
	DW_AT_data_member_location DW_AT = 0x38
	DW_AT_decl_column          DW_AT = 0x39
	DW_AT_decl_file            DW_AT = 0x3a
	DW_AT_decl_line            DW_AT = 0x3b
	DW_AT_declaration          DW_AT = 0x3c
	DW_AT_discr_list           DW_AT = 0x3d
	DW_AT_encoding             DW_AT = 0x3e
	DW_AT_external             DW_AT = 0x3f
	DW_AT_frame_base           DW_AT = 0x40
	DW_AT_friend               DW_AT = 0x41
	DW_AT_identifier_case      DW_AT = 0x42
	DW_AT_macro_info           DW_AT = 0x43
	DW_AT_namelist_item        DW_AT = 0x44
	DW_AT_priority             DW_AT = 0x45
	DW_AT_segment              DW_AT = 0x46
	DW_AT_specification        DW_AT = 0x47
	DW_AT_static_link          DW_AT = 0x48
	DW_AT_type                 DW_AT = 0x49
	DW_AT_use_location         DW_AT = 0x4a
	DW_AT_variable_parameter   DW_AT = 0x4b
	DW_AT_virtuality           DW_AT = 0x4c
	DW_AT_vtable_elem_location DW_AT = 0x4d
	DW_AT_lo_user              DW_AT = 0x2000
	DW_AT_hi_user              DW_AT = 0x3fff
)

type DW_FORM uint32

const (
	DW_FORM_addr      DW_FORM = 0x01
	DW_FORM_block2    DW_FORM = 0x03
	DW_FORM_block4    DW_FORM = 0x04
	DW_FORM_data2     DW_FORM = 0x05
	DW_FORM_data4     DW_FORM = 0x06
	DW_FORM_data8     DW_FORM = 0x07
	DW_FORM_string    DW_FORM = 0x08
	DW_FORM_block     DW_FORM = 0x09
	DW_FORM_block1    DW_FORM = 0x0a
	DW_FORM_data1     DW_FORM = 0x0b
	DW_FORM_flag      DW_FORM = 0x0c
	DW_FORM_sdata     DW_FORM = 0x0d
	DW_FORM_strp      DW_FORM = 0x0e
	DW_FORM_udata     DW_FORM = 0x0f
	DW_FORM_ref_addr  DW_FORM = 0x10
	DW_FORM_ref1      DW_FORM = 0x11
	DW_FORM_ref2      DW_FORM = 0x12
	DW_FORM_ref4      DW_FORM = 0x13
	DW_FORM_ref8      DW_FORM = 0x14
	DW_FORM_ref_udata DW_FORM = 0x15
	DW_FORM_indirect  DW_FORM = 0x16
)

type AttributeValue interface {
	WriteOut(writer io.Writer, byteOrder binary.ByteOrder, debugStr []byte) []byte
}

type NumberValue struct {
	Value int64
	Size  uint32
}

func writeOutNumber(writer io.Writer, byteOrder binary.ByteOrder, value int64, size uint32) {
	switch size {
	case 0:
		writeULEB128(writer, uint64(value))
	case 1:
		var asByte = uint8(value)
		binary.Write(writer, byteOrder, &asByte)
	case 2:
		var asShort = uint16(value)
		binary.Write(writer, byteOrder, &asShort)
	case 4:
		var asWord = uint32(value)
		binary.Write(writer, byteOrder, &asWord)
	case 8:
		binary.Write(writer, byteOrder, &value)
	}
}

func (value NumberValue) WriteOut(writer io.Writer, byteOrder binary.ByteOrder, debugStr []byte) []byte {
	writeOutNumber(writer, byteOrder, value.Value, value.Size)
	return debugStr
}

type StringValue struct {
	Value  string
	Inline bool
}

func (value StringValue) WriteOut(writer io.Writer, byteOrder binary.ByteOrder, debugStr []byte) []byte {
	if value.Inline {
		writer.Write([]byte(value.Value))
		writer.Write(make([]byte, 0))

		return debugStr
	} else {
		result, valueToWrite := elf.AddStringToSection(debugStr, value.Value)

		var asWord = uint32(valueToWrite)
		binary.Write(writer, byteOrder, &asWord)

		return result
	}
}

type BlockValue struct {
	Value []byte
	Size  uint32
}

func (value BlockValue) WriteOut(writer io.Writer, byteOrder binary.ByteOrder, debugStr []byte) []byte {
	writeOutNumber(writer, byteOrder, int64(len(value.Value)), value.Size)
	writer.Write(value.Value)
	return debugStr
}

type Indirect struct {
	Form  DW_FORM
	Value AttributeValue
}

func (value Indirect) WriteOut(writer io.Writer, byteOrder binary.ByteOrder, debugStr []byte) []byte {
	writeULEB128(writer, uint64(value.Form))
	return value.Value.WriteOut(writer, byteOrder, debugStr)
}

type AbbrevAttr struct {
	Type  DW_AT
	Form  DW_FORM
	Value AttributeValue
}

func CreateAddrAttr(at DW_AT, data int64) AbbrevAttr {
	return AbbrevAttr{
		at,
		DW_FORM_addr,
		NumberValue{data, 4},
	}
}

func CreateConstantAttr(at DW_AT, data int64, size uint32) AbbrevAttr {
	var dwType DW_FORM

	switch size {
	case 0:
		dwType = DW_FORM_udata
	case 1:
		dwType = DW_FORM_data1
	case 2:
		dwType = DW_FORM_data2
	case 4:
		dwType = DW_FORM_data4
	case 8:
		dwType = DW_FORM_data8
	}

	return AbbrevAttr{
		at,
		dwType,
		NumberValue{data, size},
	}
}

func CreateStringAttr(at DW_AT, data string, inline bool) AbbrevAttr {
	var dwType DW_FORM

	if inline {
		dwType = DW_FORM_string
	} else {
		dwType = DW_FORM_strp
	}

	return AbbrevAttr{
		at,
		dwType,
		StringValue{data, inline},
	}
}

type AbbrevTreeNode struct {
	Tag        DW_TAG
	Attributes []AbbrevAttr
	Children   []*AbbrevTreeNode
}

type InfoData struct {
	Info     []byte
	RelInfo  *elf.RelocationBuilder
	Abbrev   []byte
	DebugStr []byte
}

func generateAbbrv(input []*AbbrevTreeNode, result *bytes.Buffer, currId int, idMapping map[*AbbrevTreeNode]int) int {
	for _, node := range input {
		idMapping[node] = currId

		writeULEB128(result, uint64(currId))
		writeULEB128(result, uint64(node.Tag))

		if len(node.Children) > 0 {
			result.WriteByte(1)
			currId = generateAbbrv(node.Children, result, currId, idMapping)
		} else {
			result.WriteByte(0)
		}

		for _, attr := range node.Attributes {
			writeULEB128(result, uint64(attr.Type))
			writeULEB128(result, uint64(attr.Form))
		}

		// double null terminate attributes
		result.WriteByte(0)
		result.WriteByte(0)

		currId++
	}

	// null termiante node list
	result.WriteByte(0)

	return currId
}

func generateInfo(input []*AbbrevTreeNode, result *bytes.Buffer, rel *elf.RelocationBuilder, strBytes []byte, byteOrder binary.ByteOrder, idMapping map[*AbbrevTreeNode]int) []byte {
	for _, node := range input {
		id, ok := idMapping[node]

		if ok {
			writeULEB128(result, uint64(id))

			for _, attr := range node.Attributes {
				if attr.Form == DW_FORM_addr {
					// kinda hacky but works for now
					rel.AddEntry(uint32(result.Len()), ".text", elf.R_MIPS_32)
				}

				strBytes = attr.Value.WriteOut(result, byteOrder, strBytes)
			}

			strBytes = generateInfo(node.Children, result, rel, strBytes, byteOrder, idMapping)
		}
	}

	return strBytes
}

func GenerateInfoAndAbbrev(input []*AbbrevTreeNode, byteOrder binary.ByteOrder) InfoData {
	var result InfoData
	var relBuilder = elf.NewRelocationBuilder()

	var idMapping = make(map[*AbbrevTreeNode]int)
	var abbrevBytes bytes.Buffer

	generateAbbrv(input, &abbrevBytes, 1, idMapping)

	result.Abbrev = abbrevBytes.Bytes()

	var infoBytes bytes.Buffer

	result.DebugStr = generateInfo(input, &infoBytes, relBuilder, make([]byte, 1), byteOrder, idMapping)

	var finalInfo bytes.Buffer
	var totalLength = uint32(infoBytes.Len()) + 7
	binary.Write(&finalInfo, byteOrder, &totalLength)

	var version = uint16(2)
	binary.Write(&finalInfo, byteOrder, &version)

	var offset = uint32(0)
	binary.Write(&finalInfo, byteOrder, &offset)
	finalInfo.WriteByte(4)

	relBuilder.AddOffset(uint32(finalInfo.Len()))
	finalInfo.Write(infoBytes.Bytes())

	result.Info = finalInfo.Bytes()
	result.RelInfo = relBuilder

	return result
}
