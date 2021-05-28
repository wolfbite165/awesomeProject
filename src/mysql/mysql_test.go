package mysql

import (
	"testing"
)

func Test_mysql(t *testing.T) {
	Connect()
	//a := Checkfile("handsome").Lock_money
	//fmt.Println(a)
	Cancel_order(5)
}
