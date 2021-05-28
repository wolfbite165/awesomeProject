package mysql

import (
	"fmt"
	"testing"
)

func Test_mysql(t *testing.T) {
	Connect()
	a := Checkfile("handsome").Lock_money
	fmt.Println(a)
}
