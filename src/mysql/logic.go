package mysql

import (
	"fmt"
	"sync"
)

var last_trade_price float64
var initPirce sync.Once

func Get_last_price() float64 {
	initPirce.Do(func() {
		row := MysqlDb.QueryRow("select price from trade order by id desc limit 1")
		if err := row.Scan(&last_trade_price); err != nil {
			fmt.Printf("scan failed, err:%v", err)
			//return
		}
	})
	return last_trade_price

}

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
	price := Get_last_price()
	//var results sql.Result
	var out Kline_info
	row := MysqlDb.QueryRow("select id,time from `min` order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}
	min_start := time1 % 60
	if out.Time != min_start {
		results, _ := MysqlDb.Exec("insert INTO `min`(time,open,high,low) values(?,?,?,?)", time1-min_start, price, price, price)
		id, _ := results.LastInsertId()
		id_min <- id
	} else {

		id_min <- out.Id
	}

}

func Hours_check(time1 int64, id_hour chan int64) {
	price := Get_last_price()
	var out Kline_info
	row := MysqlDb.QueryRow("select id,time from ? order by id desc limit 1", "1hour")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}
	hours_start := time1 % 3600
	if out.Time != hours_start {
		results, _ := MysqlDb.Exec("insert INTO ?(time,open,high,low) values(?,?,?,?)", "1hour", time1-hours_start, price, price, price)
		id, _ := results.LastInsertId()
		id_hour <- id
	} else {

		id_hour <- out.Id
	}
}
func Five_min_check(time1 int64, id_five_min chan int64) {
	price := Get_last_price()
	var out Kline_info
	row := MysqlDb.QueryRow("select id,time from ? order by id desc limit 1", "5min")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}
	five_min_start := time1 % 300
	if out.Time != five_min_start {
		results, _ := MysqlDb.Exec("insert INTO ?(time,open,high,low) values(?,?,?,?)", "5min", time1-five_min_start, price, price, price)
		id, _ := results.LastInsertId()
		id_five_min <- id
	} else {

		id_five_min <- out.Id
	}
}

func Threty_min_check(time1 int64, id_threty_min chan int64) {
	price := Get_last_price()
	var out Kline_info
	row := MysqlDb.QueryRow("select id,time from ? order by id desc limit 1", "30min")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}
	threty_min_start := time1 % 1800
	if out.Time != threty_min_start {
		results, _ := MysqlDb.Exec("insert INTO ?(time,open,high,low) values(?,?,?,?)", "30min", time1-threty_min_start, price, price, price)
		id, _ := results.LastInsertId()
		id_threty_min <- id
	} else {

		id_threty_min <- out.Id
	}
}

func Twilve_hour_check(time1 int64, id_twilve_hour chan int64) {
	price := Get_last_price()
	var out Kline_info
	row := MysqlDb.QueryRow("select id,time from ? order by id desc limit 1", "12hour")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}
	twilve_hour := time1 % (3600 * 12)
	if out.Time != twilve_hour {
		results, _ := MysqlDb.Exec("insert INTO ?(time,open,high,low) values(?,?,?,?)", "12hour", time1-twilve_hour, price, price, price)
		id, _ := results.LastInsertId()
		id_twilve_hour <- id
	} else {

		id_twilve_hour <- out.Id
	}
}

func One_day_check(time1 int64, id_day chan int64) {
	price := Get_last_price()
	var out Kline_info
	row := MysqlDb.QueryRow("select id,time from ? order by id desc limit 1", "1day")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}
	one_day_start := time1 % (3600 * 24)
	if out.Time != one_day_start {
		results, _ := MysqlDb.Exec("insert INTO ?(time,open,high,low) values(?,?,?,?)", "1day", time1-one_day_start, price, price, price)
		id, _ := results.LastInsertId()
		id_day <- id
	} else {

		id_day <- out.Id
	}
}
func Sec_check(time1 int64, id_sec chan int64) {

	var out Kline_info
	row := MysqlDb.QueryRow("select id,time from secend order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}
	one_day_start := time1 % (3600 * 24)
	if out.Time != one_day_start {
		results, _ := MysqlDb.Exec("insert INTO secend(time) values(?)", time1-one_day_start)
		id, _ := results.LastInsertId()
		id_sec <- id
	} else {

		id_sec <- out.Id
	}
}
