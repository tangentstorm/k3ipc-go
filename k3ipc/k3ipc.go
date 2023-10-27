package k3ipc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	K3LST = iota
	K3INT
	K3NUM
	K3CHR
	K3SYM
	K3DCT
	K3NUL
	K3FUN
)

// numstrToBytes
// convert a string of space-separated numbers into a byte array
// with one byte per number
func NumStrToBytes(s string) []byte {
	nums := strings.Split(s, " ")
	bytes := make([]byte, len(nums))
	for i, v := range nums {
		b, _ := strconv.Atoi(v)
		bytes[i] = byte(b)
	}
	return bytes
}

// inverse of NumStrToBytes
func BytesToNumStr(bytes []byte) string {
	nums := make([]string, len(bytes))
	for i, v := range bytes {
		nums[i] = strconv.Itoa(int(v))
	}
	return strings.Join(nums, " ")
}

// K3 data from bytes
func Db(buf []byte) any {
	// k3 buffer format (used by serialization and IPC):
	var endianFlag byte
	// then 2 ignored bytes, then 1 byte for k-ipc message type (which we don't need here)
	var len int32
	var typ int32

	r := bytes.NewReader(buf)

	endianFlag, _ = r.ReadByte()
	if endianFlag != 1 {
		panic("endianFlag != 1")
	}

	// fast forward past the bytes we don't care about
	_, err := r.Seek(4, io.SeekStart)
	if err != nil {
		fmt.Println("Seek failed")
		return nil
	}

	eness := binary.LittleEndian
	binary.Read(r, eness, &len)
	binary.Read(r, eness, &typ)

	switch typ {
	case K3INT:
		var res int32
		binary.Read(r, eness, &res)
		return res
	default:
		panic("TODO: parse type #" + fmt.Sprintf("%v", typ))
	}
}

// K3 data to bytes
func Bd(val any) (buf []byte) {
	switch val := val.(type) {
	case int32:
		buf = make([]byte, 16)
		buf[0] = 1
		len := 8 // length of the data (4 byte type + 4 byte value)
		endy := binary.LittleEndian
		endy.PutUint32(buf[4:], uint32(len))
		endy.PutUint32(buf[8:], uint32(K3INT))
		endy.PutUint32(buf[12:], uint32(val))
		return buf
	case byte:
		return []byte{
			1, 0, 0, 0,
			8, 0, 0, 0,
			byte(K3CHR), 0, 0, 0,
			val, 0, 0, 0}
	default:
		panic("don't know how to convert value to binary!")
	}
}
