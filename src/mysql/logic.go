package mysql

import "fmt"

func Check_same_account(Account string) bool {
	a := Checkfile(Account)
	if a.Id == 0 {
		return false
	} else {
		return true
	}

}

type Kline_info struct {
	Time int64 `db:"time"`
	Id   int64 `db:"id"`
}

func Min_check(time1 int64, id_min chan int64) {
	//var results sql.Result
	var out Kline_info
	row := MysqlDb.QueryRow("select id,time from `min` order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}
	min_start := time1 % 60
	if out.Time != min_start {
		results, _ := MysqlDb.Exec("insert INTO `min`(time) values(?)", time1-min_start)
		id, _ := results.LastInsertId()
		id_min <- id
	} else {

		id_min <- out.Id
	}

}

func Hours_check(time1 int64, id_hour chan int64) {

	var out Kline_info
	row := MysqlDb.QueryRow("select id,time from `1hour` order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}
	hours_start := time1 % 3600
	if out.Time != hours_start {
		results, _ := MysqlDb.Exec("insert INTO `1hour`(time) values(?)", time1-hours_start)
		id, _ := results.LastInsertId()
		id_hour <- id
	} else {

		id_hour <- out.Id
	}
}
func Five_min_check(time1 int64, id_five_min chan int64) {

	var out Kline_info
	row := MysqlDb.QueryRow("select id,time from `5min` order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}
	five_min_start := time1 % 300
	if out.Time != five_min_start {
		results, _ := MysqlDb.Exec("insert INTO `5min`(time) values(?)", time1-five_min_start)
		id, _ := results.LastInsertId()
		id_five_min <- id
	} else {

		id_five_min <- out.Id
	}
}

func Threty_min_check(time1 int64, id_threty_min chan int64) {

	var out Kline_info
	row := MysqlDb.QueryRow("select id,time from `30min` order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}
	threty_min_start := time1 % 1800
	if out.Time != threty_min_start {
		results, _ := MysqlDb.Exec("insert INTO `30min`(time) values(?)", time1-threty_min_start)
		id, _ := results.LastInsertId()
		id_threty_min <- id
	} else {

		id_threty_min <- out.Id
	}
}

func Twilve_hour_check(time1 int64, id_twilve_hour chan int64) {

	var out Kline_info
	row := MysqlDb.QueryRow("select id,time from `12hour` order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}
	twilve_hour := time1 % (3600 * 12)
	if out.Time != twilve_hour {
		results, _ := MysqlDb.Exec("insert INTO `12hour`(time) values(?)", time1-twilve_hour)
		id, _ := results.LastInsertId()
		id_twilve_hour <- id
	} else {

		id_twilve_hour <- out.Id
	}
}

func One_day_check(time1 int64, id_day chan int64) {

	var out Kline_info
	row := MysqlDb.QueryRow("select id,time from `1day` order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}
	one_day_start := time1 % (3600 * 24)
	if out.Time != one_day_start {
		results, _ := MysqlDb.Exec("insert INTO `1day`(time) values(?)", time1-one_day_start)
		id, _ := results.LastInsertId()
		id_day <- id
	} else {

		id_day <- out.Id
	}
}
