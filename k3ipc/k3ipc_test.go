package k3ipc

import (
	"testing"
)

func check(t *testing.T, expectValue any, expectedBytes string) {
	actualValue := Db(NumStrToBytes(expectedBytes))
	if actualValue != expectValue {
		t.Errorf("Db() failed. Expected %v but got %v", expectValue, actualValue)
	}
	actualBytes := BytesToNumStr(Bd(expectValue))
	if actualBytes != expectedBytes {
		t.Errorf("Bd() failed.\nexpect: %v\nactual: %v", expectedBytes, actualBytes)
	}
}
func TestK3INT(t *testing.T) {
	check(t, int32(0), "1 0 0 0 8 0 0 0 1 0 0 0 0 0 0 0")
	check(t, int32(1), "1 0 0 0 8 0 0 0 1 0 0 0 1 0 0 0")
	check(t, int32(0x7ffffffe), "1 0 0 0 8 0 0 0 1 0 0 0 254 255 255 127")
	check(t, int32(1234), "1 0 0 0 8 0 0 0 1 0 0 0 210 4 0 0")
}
func TestK3FLT(t *testing.T) {
	check(t, float64(1.1), "1 0 0 0 16 0 0 0 2 0 0 0 1 0 0 0 154 153 153 153 153 153 241 63")
}
func TestK3CHR(t *testing.T) {
	check(t, byte('x'), "1 0 0 0 8 0 0 0 3 0 0 0 120 0 0 0")
}
func TestK3CHRs(t *testing.T) {
	check(t, "hello",
		"1 0 0 0 14 0 0 0 253 255 255 255 5 0 0 0 104 101 108 108 111 0")
	check(t, "hi",
		"1 0 0 0 11 0 0 0 253 255 255 255 2 0 0 0 104 105 0")
}
func TestK3SYM(t *testing.T) {
	check(t, sym("abc"), "1 0 0 0 8 0 0 0 4 0 0 0 97 98 99 0")
}
func TestK3SYMs(t *testing.T) {
	check(t, []KSym{sym("abc"), sym("xyz")},
		"1 0 0 0 16 0 0 0 252 255 255 255 2 0 0 0 97 98 99 0 120 121 122 0")
}
func TestK3NUL(t *testing.T) {
	check(t, nil, "1 0 0 0 8 0 0 0 6 0 0 0 0 0 0 0")
}
func TestK3LST(t *testing.T) {
	check(t, []any{}, "1 0 0 0 8 0 0 0 0 0 0 0 0 0 0 0")
	check(t, []any{"abc", 1},
		"1 0 0 0 32 0 0 0 0 0 0 0 2 0 0 0 253 255 255 255 3 0 0 0 97 98 99 0 0 0 0 0 1 0 0 0 1 0 0 0")
	check(t, []any{'a', 'b', 'c', 1},
		"1 0 0 0 40 0 0 0 0 0 0 0 4 0 0 0 3 0 0 0 97 0 0 0 3 0 0 0 98 0 0 0 3 0 0 0 99 0 0 0 1 0 0 0 1 0 0 0")
	check(t, []any{[]any{[]any{}, []any{}}},
		"1 0 0 0 32 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 2 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0")
}
func TestK3DCT(t *testing.T) {
	check(t, map[string]any{}, "1 0 0 0  8 0 0 0 5 0 0 0 0 0 0 0")
	check(t, map[string]any{"k": 123}, "1 0 0 0 40 0 0 0 5 0 0 0 1 0 0 0 0 0 0 0 3 0 0 0 4 0 0 0 107 0 0 0 1 0 0 0 123 0 0 0 6 0 0 0 0 0 0 0")
}
