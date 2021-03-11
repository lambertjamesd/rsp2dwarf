package dwarf

import "io"

const DW_LANG_Mips_Assembler = 0x8001

func writeLEB128(writer io.Writer, value int64) {
	var hasMore = true
	var tmp = make([]byte, 1)
	for hasMore {
		tmp[0] = byte(value & 0x7f)
		value = value >> 7
		hasMore = value != 0

		if hasMore {
			tmp[0] |= 0x80
		}

		writer.Write(tmp)
	}
}

func readLEB128(reader io.Reader) int64 {
	var result int64 = 0

	var hasMore = true
	var tmp = make([]byte, 1)

	for hasMore {
		reader.Read(tmp)

		result = result << 7
		result = result | int64(tmp[0]&0x7f)

		hasMore = tmp[0]&0x80 != 0
	}

	return result
}
