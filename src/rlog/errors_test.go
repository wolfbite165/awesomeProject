package rlog

import (
	"log"
	"testing"
)

func TestCreateErr(t *testing.T) {
	Help_SetLogOutput(t)
	e1 := NewErrorf("an error!")
	log.Println(e1)
	e2 := TraceErr(e1)
	log.Println(e2)
	log.Println(GetStack(e2))
	log.Println(GetCause(e2))

	if NotEqual(e1, e2) {
		t.Errorf("应该一样")
	}
	e3 := NewErrorf("an error!")
	if NotEqual(e1, e3) {
		t.Errorf("应该一样")
	}
}
