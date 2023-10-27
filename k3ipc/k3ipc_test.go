package k3ipc

import (
	"fmt"
	"testing"
)

type Sym struct{ s string }

func sym(s string) Sym       { return Sym{s} }
func (s Sym) String() string { return "`" + s.s }

func ExampleK3List() {
	ls := []any{"cat", 123, sym("cat"), []int{1, 2, 3}}
	fmt.Printf("%v\n", ls)
	// Output: [cat 123 `cat [1 2 3]]
}

func check(t *testing.T, expectValue any, expectedBytes string) {
	actualValue := Db(NumStrToBytes(expectedBytes))
	if actualValue != expectValue {
		t.Errorf("Db() failed. Expected %v but got %v", expectValue, actualValue)
	}
	actualBytes := BytesToNumStr(Bd(expectValue))
	if actualBytes != expectedBytes {
		t.Errorf("Bd() failed. Expected %v but got %v", expectedBytes, actualBytes)
	}
}

func TestK3(t *testing.T) {
	check(t, int32(0), "1 0 0 0 8 0 0 0 1 0 0 0 0 0 0 0")
	check(t, int32(1), "1 0 0 0 8 0 0 0 1 0 0 0 1 0 0 0")
	check(t, int32(0x7ffffffe), "1 0 0 0 8 0 0 0 1 0 0 0 254 255 255 127")
	check(t, int32(1234), "1 0 0 0 8 0 0 0 1 0 0 0 210 4 0 0")
}
