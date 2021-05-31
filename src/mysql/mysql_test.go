package mysql

import (
	"testing"
)

func Test_mysql(t *testing.T) {
	Connect()
	//a := Checkfile("handsome").Lock_money
	//fmt.Println(a)
	Deel_order(1, "sell")
}
