package mysql

func Check_same_account(Account string) bool {
	a := Checkfile(Account)
	if a.Id == 0 {
		return false
	} else {
		return true
	}

}
