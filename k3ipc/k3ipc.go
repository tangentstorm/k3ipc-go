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
}

type chunkHeader struct {
	dataType int32
	count    int32
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
	return
}

func parseChunkHeader(ord binary.ByteOrder, r *bytes.Reader) (h chunkHeader) {
	binary.Read(r, ord, &h.dataType)
	if h.dataType <= 0 {
		binary.Read(r, ord, &h.count)
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
	mh := parseMessageHeader(r)
	return readDb(mh.byteOrder, r, false)
}

func readDb(ord binary.ByteOrder, r *bytes.Reader, align bool) any {
	ch := parseChunkHeader(ord, r)
	switch ch.dataType {
	case K3INT:
		var res int32
		binary.Read(r, ord, &res)
		return res
	case K3FLT:
		var tmp uint64
		var pad uint32
		binary.Read(r, ord, &pad)
		binary.Read(r, ord, &tmp)
		return float64(math.Float64frombits(tmp))
	case K3CHR:
		var res byte
		res, _ = r.ReadByte()
		if align {
			r.Read(make([]byte, 3)) // ignore padding
		}
		return res
	case -K3CHR:
		buf := make([]byte, ch.count)
		r.Read(buf)
		r.ReadByte() // ignore \0
		if align && ((ch.count+1)%8 != 0) {
			for i := 0; i < 8-(int(ch.count)+1)%8; i++ {
				r.ReadByte() // ignore padding
			}
		}
		return string(buf)
	case K3SYM:
		// symbol doesn't include length. We just read until \0
		str := ""
		for {
			b, _ := r.ReadByte()
			if b == 0 {
				break
			} else {
				str += string(b)
			}
		}
		return sym(str)
	case -K3SYM:
		var res []KSym
		for i := 0; i < int(ch.count); i++ {
			res = append(res, readSym(r))
		}
		return res
	case K3NUL:
		return nil
	case K3LST:
		res := []any{}
		for i := 0; i < int(ch.count); i++ {
			res = append(res, readDb(ord, r, true))
		}
		return res
	case K3DCT:
		res := map[string]any{}
		return res
	default:
		panic("TODO: parse type #" + fmt.Sprintf("%v", ch.dataType))
	}
}

// bytes from K3 data
func Bd(val any) (res []byte) {
	ord := binary.LittleEndian
	buf := bytes.NewBuffer([]byte{
		1, 0, 0, 0, // message type
		0, 0, 0, 0, // message length
	})
	dLen := emitBd(buf, ord, val)
	ord.PutUint32(buf.Bytes()[4:], uint32(dLen))
	return buf.Bytes()
}

// recursively write data into the buffer
func emitBd(buf *bytes.Buffer, ord binary.ByteOrder, val any) (dLen int) {
	switch v := val.(type) {
	case int:
		// TODO: handle 64-bit ints ?
		return emitBd(buf, ord, int32(v))
	case int32:
		binary.Write(buf, ord, int32(K3INT))
		binary.Write(buf, ord, int32(v))
		return 8
	case []int32:
		panic("todo: []int32")
	case float64:
		dLen = 16
		binary.Write(buf, ord, int32(K3FLT))
		// k sticks an extra int here to keep it 64-bit aligned
		binary.Write(buf, ord, int32(1))
		binary.Write(buf, ord, v)
	case []float64:
		panic("todo: []float64")
	case byte: // note 'x' is an int32. this is byte('x')
		dLen = 8 // KCHR is always padded
		binary.Write(buf, ord, int32(K3CHR))
		buf.Write([]byte{v, 0, 0, 0})
	case string: // TODO: handle utf-8?
		dLen = 8 + len(v) + 1 // strlen + \0
		binary.Write(buf, ord, int32(-K3CHR))
		binary.Write(buf, ord, int32(len(v)))
		buf.Write([]byte(v))
		buf.WriteByte(0)
	case KSym: // sym("abc")
		dLen = 4 + len(v.s) + 1
		// no string length for symbols
		binary.Write(buf, ord, int32(K3SYM))
		buf.Write([]byte(v.s))
		buf.WriteByte(0)
	case []KSym: // []KSym{sym("abc")}
		binary.Write(buf, ord, int32(-K3SYM))
		binary.Write(buf, ord, int32(len(v)))
		strlens := 0
		for _, s := range v {
			strlens += len(s.s) + 1
			buf.Write([]byte(s.s))
			buf.WriteByte(0)
		}
		dLen = 8 + strlens
	case []any:
		binary.Write(buf, ord, int32(K3LST))
		binary.Write(buf, ord, int32(len(v)))
		dLen = 8 // those two ints, plus...
		for _, v := range v {
			dLen += emitBd(buf, ord, v)
			// pad to 8-byte boundary
			for buf.Len()%8 != 0 {
				buf.WriteByte(0)
				dLen++
			}
		}
	case map[string]any:
		panic("todo: map[string]any")
	default:
		if v == nil {
			dLen = 8
			binary.Write(buf, ord, int32(K3NUL))
			binary.Write(buf, ord, int32(0))
			break
		}
		panic(fmt.Sprintf("Db: don't know how to generate bytes for %v (type:%T)", v, v))
	}
	return
}
