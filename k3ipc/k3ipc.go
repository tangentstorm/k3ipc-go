package k3ipc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	K3LST = iota
	K3INT
	K3FLT
	K3CHR
	K3SYM
	K3DCT
	K3NUL
	K3FUN
)

// Arbitrary wrapper to mark incoming string as a symbol
type KSym struct{ s string }

func sym(s string) KSym       { return KSym{s} }
func (s KSym) String() string { return "`" + s.s }

// convert a string of space-separated numbers into a
// byte array, with one byte per number
func NumStrToBytes(s string) []byte {
	nums := strings.Split(s, " ")
	bytes := make([]byte, len(nums))
	for i, v := range nums {
		b, _ := strconv.Atoi(v)
		bytes[i] = byte(b)
	}
	return bytes
}

func BytesToNumStr(bytes []byte) string {
	nums := make([]string, len(bytes))
	for i, v := range bytes {
		nums[i] = strconv.Itoa(int(v))
	}
	return strings.Join(nums, " ")
}

type MsgHeader struct {
	byteOrder binary.ByteOrder
	msgType   byte
	msgLen    int32
	dataType  int32
	count     int32
}

func parseMessageHeader(r *bytes.Reader) (h MsgHeader) {
	endianFlag, _ := r.ReadByte()
	h.byteOrder = binary.LittleEndian
	if endianFlag != 1 {
		h.byteOrder = binary.BigEndian
	}
	r.ReadByte() // ignore
	r.ReadByte() // ignore
	h.msgType, _ = r.ReadByte()
	binary.Read(r, h.byteOrder, &h.msgLen)
	binary.Read(r, h.byteOrder, &h.dataType)
	if h.dataType < 0 {
		binary.Read(r, h.byteOrder, &h.count)
	}
	return
}

func readSym(r *bytes.Reader) KSym {
	var tok []byte
	for {
		b, _ := r.ReadByte()
		if b == 0 {
			break
		}
		tok = append(tok, b)
	}
	return sym(string(tok))
}

// K3 data from bytes
func Db(buf []byte) any {
	r := bytes.NewReader(buf)
	h := parseMessageHeader(r)
	const dataStart = 16
	switch h.dataType {
	case K3INT:
		var res int32
		binary.Read(r, h.byteOrder, &res)
		return res
	case K3FLT:
		var tmp uint64
		binary.Read(r, h.byteOrder, &tmp)
		return float64(math.Float64frombits(tmp))
	case K3CHR:
		var res byte
		res, _ = r.ReadByte()
		return res
	case -K3CHR:
		var res string
		res = string(buf[dataStart : dataStart+h.count])
		return res
	case K3SYM:
		// symbol doesn't include length, so data starts at 12
		// and the msgLen is 4 bytes longer, not 8
		// then there's a null terminator so subtract 1 for that too
		return sym(string(buf[12 : 12+h.msgLen-5]))
	case -K3SYM:
		var res []KSym
		for i := 0; i < int(h.count); i++ {
			res = append(res, readSym(r))
		}
		return res
	case K3NUL:
		return nil
	default:
		panic("TODO: parse type #" + fmt.Sprintf("%v", h.dataType))
	}
}

func writeI32(w *bytes.Buffer, ord binary.ByteOrder, n int32) {
	binary.Write(w, ord, n)
}

// bytes from K3 data
func Bd(val any) (res []byte) {
	ord := binary.LittleEndian
	buf := bytes.NewBuffer([]byte{
		1, 0, 0, 0, // message type
		0, 0, 0, 0, // message length
		0, 0, 0, 0, // data type
	})
	const tLen = 4
	dLen, dTyp := 0, K3INT
	switch val := val.(type) {
	case int32:
		dLen = 4
		writeI32(buf, ord, val)
	case byte:
		dTyp, dLen = K3CHR, 4
		buf.Write([]byte{val, 0, 0, 0})
	case KSym:
		dTyp, dLen = K3SYM, len(val.s)+1
		// no string length for symbols
		buf.Write([]byte(val.s))
		buf.WriteByte(0)
	case []KSym:
		strlens := 0
		for _, s := range val {
			strlens += len(s.s) + 1
		}
		dTyp, dLen = -K3SYM, 4+strlens
		writeI32(buf, ord, int32(len(val)))
		for _, s := range val {
			buf.Write([]byte(s.s))
			buf.WriteByte(0)
		}
	case string:
		dTyp = -K3CHR
		dLen = 4 + len(val) + 1 // strlen + \0
		// store the string length:
		writeI32(buf, ord, int32(dLen-5))
		buf.Write([]byte(val))
		buf.WriteByte(0)
	default:
		if val == nil {
			dTyp, dLen = K3NUL, 4
			buf.Write([]byte{0, 0, 0, 0})
			break
		}
		panic(fmt.Sprintf("don't know how to generate bytes for %v", val))
	}
	ord.PutUint32(buf.Bytes()[4:], uint32(dLen+tLen))
	ord.PutUint32(buf.Bytes()[8:], uint32(dTyp))
	return buf.Bytes()
}
