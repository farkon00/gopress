package commands

import (
	"encoding/binary"
	"fmt"
	"gopress/shared"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func parse_map_elem(input []byte, cursor *int, m map[shared.BitSeq]byte) (end bool) {
	key := shared.BitSeq{Length: int(binary.LittleEndian.Uint32(input[*cursor : *cursor+4]))}
	*cursor += 4
	if key.Length == 0 {
		end = true
		return
	}
	copy(key.Data[:], input[*cursor:*cursor+shared.Ceil_div(key.Length, 8)])

	*cursor += shared.Ceil_div(key.Length, 8)
	m[key] = input[*cursor]
	*cursor += 1
	return
}

func decode_bits(encoding_map map[shared.BitSeq]byte, length int, data []byte) (res []byte) {
	var acc shared.BitSeq
	for index, i := range data {
		end_bit := 8
		if index == len(data)-1 {
			end_bit = ((length - 1) % 8) + 1
		}
		for offset := 0; offset < end_bit; offset++ {
			acc = acc.Add_bit(i&(1<<offset) != 0)
			char, ok := encoding_map[acc]
			if ok {
				acc = shared.BitSeq{}
				res = append(res, char)
			}
		}
	}
	return
}

func Decode(config shared.Config, input []byte) {
	cursor := 0
	encoding_map := make(map[shared.BitSeq]byte)
	for !parse_map_elem(input, &cursor, encoding_map) {
	}

	length := int(binary.LittleEndian.Uint32(input[cursor:]))
	cursor += 4
	data := input[cursor : cursor+shared.Ceil_div(length, 8)]

	content := decode_bits(encoding_map, length, data)
	var filename string
	if config.Output == "" {
		filename = strings.TrimSuffix(config.Filename, filepath.Ext(config.Filename))
	} else {
		filename = config.Output
	}
	if len(filename) == len(config.Filename) {
		filename += "1"
	}

	err := os.WriteFile(filename, content, fs.FileMode(os.ModePerm))
	if err != nil {
		fmt.Println("Failed to write to file\n Traceback:")
		panic(err)
	}
}
