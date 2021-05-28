package mysql

import "testing"

func Test_mysql(t *testing.T) {
	Connect()
	Checkfile("SSS")
}
