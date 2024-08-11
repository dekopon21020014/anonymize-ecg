package mfer

import (
	"encoding/binary"
	"errors"
)

func Binary2Uint16(byteOrder byte, bytes ...byte) (uint16, error) {
	if len(bytes) > 2 {
		return 0, errors.New("len(bytes) must be less than 2")
	}
	b := make([]byte, len(bytes))
	copy(b, bytes)
	padding := make([]byte, 2-len(b))
	if byteOrder == 0 { // big endian
		b = append(padding, b...)
		return binary.BigEndian.Uint16(b), nil
	} else if byteOrder == 1 { // little endian
		b = append(b, padding...)
		return binary.LittleEndian.Uint16(b), nil
	} else {
		return 0, errors.New("undefined byte order or something went wrong")
	}
}

// 4byte以内の文字列をuint32に変換する
func Binary2Uint32(byteOrder byte, bytes ...byte) (uint32, error) {
	if len(bytes) > 4 {
		return 0, errors.New("len(bytes) must be less than 4")
	}

	b := make([]byte, len(bytes))
	copy(b, bytes)
	padding := make([]byte, 4-len(b))
	if byteOrder == 0 { // big endian
		b = append(padding, b...)
		return binary.BigEndian.Uint32(b), nil
	} else if byteOrder == 1 { // little endian
		b = append(b, padding...)
		return binary.LittleEndian.Uint32(b), nil
	} else {
		return 0, errors.New("undefined byte order or something went wrong")
	}
}

func Binary2Uint64(byteOrder byte, bytes ...byte) (uint64, error) {
	if len(bytes) > 8 {
		return 0, errors.New("len(bytes) must be less than 8")
	}

	b := make([]byte, len(bytes))
	copy(b, bytes)
	padding := make([]byte, 8-len(b))
	if byteOrder == 0 { // big endian
		b = append(padding, b...)
		return binary.BigEndian.Uint64(b), nil
	} else if byteOrder == 1 { // little endian
		b = append(b, padding...)
		return binary.LittleEndian.Uint64(b), nil
	} else {
		return 0, errors.New("undefined byte order or something went wrong")
	}
}
