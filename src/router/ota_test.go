package router

import (
	"awesomeProject/src/rlog"
	"testing"
)

func TestOTA(t *testing.T) {
	//str := "XUZ3VU4IL6GFAKPSUXW2UHDYNSCJPC7X"

	str := NewGoogleAuth().GetSecret()

	rlog.Info(str)

}
