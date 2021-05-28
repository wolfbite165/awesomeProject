package rlog

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	r := Caller(0)
	fmt.Println(r.String())
}
