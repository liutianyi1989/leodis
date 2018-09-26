package ds

import (
	"fmt"
	"testing"
)

func TestMove(t *testing.T) {
	iniSet := NewIntSet()
	iniSet.encoding = INTSET_ENC_INT16
	iniSet.contents = []byte{0, 1, 0, 2, 0, 0}

	iniSet.Move(1, 2)

	fmt.Println(iniSet.contents)
}
