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
	Connect()
	//var results sql.Result
	//var out Kline_info
	out := new(Kline_info)
	min_start := time1 % 60
	row := MysqlDb.QueryRow("select id,time from `min` order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		results, err2 := MysqlDb.Exec("insert INTO `min`(`time`,`open`,high,low,`close`,volume) values(?,0,0,0,0,0)", time1-min_start)
		if err2 != nil {
			fmt.Println(err2)
		}
		id, _ := results.LastInsertId()
		id_min <- id
		return
	}
	if out.Time != min_start {
		results, _ := MysqlDb.Exec("insert INTO `min`(`time`,`open`,high,low,`close`,volume) values(?,0,0,0,0,0)", time1-min_start)
		id, _ := results.LastInsertId()
		id_min <- id
	} else {

		id_min <- out.Id
	}

}

func Hours_check(time1 int64, id_hour chan int64) {
	Connect()
	//var out Kline_info
	out := new(Kline_info)
	hours_start := time1 % 3600
	row := MysqlDb.QueryRow("select id,time from 1hour order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		results, err2 := MysqlDb.Exec("insert INTO 1hour(`time`,`open`,high,low,`close`,volume) values(?,0,0,0,0,0)", time1-hours_start)
		if err2 != nil {
			fmt.Println(err2)
		}
		id, _ := results.LastInsertId()
		id_hour <- id
		return
	}

	if out.Time != hours_start {
		results, _ := MysqlDb.Exec("insert INTO `1hour`(`time`,`open`,high,low,`close`,volume) values(?,0,0,0,0,0)", time1-hours_start)
		id, _ := results.LastInsertId()
		id_hour <- id
	} else {

		id_hour <- out.Id
	}
}
func Five_min_check(time1 int64, id_five_min chan int64) {
	Connect()
	//var out Kline_info
	out := new(Kline_info)
	five_min_start := time1 % 300
	row := MysqlDb.QueryRow("select id,time from 5min order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		results, err2 := MysqlDb.Exec("insert INTO 5min(`time`,`open`,high,low,`close`,volume) values(?,0,0,0,0,0)", time1-five_min_start)
		if err2 != nil {
			fmt.Println(err2)
		}
		id, _ := results.LastInsertId()
		id_five_min <- id
		return
	}

	if out.Time != five_min_start {
		results, _ := MysqlDb.Exec("insert INTO 5min(`time`,`open`,high,low,`close`,volume) values(?,0,0,0,0,0)", time1-five_min_start)
		id, _ := results.LastInsertId()
		id_five_min <- id
	} else {

		id_five_min <- out.Id
	}
}

func Threty_min_check(time1 int64, id_threty_min chan int64) {
	Connect()
	out := new(Kline_info)
	threty_min_start := time1 % 1800
	row := MysqlDb.QueryRow("select id,time from 30min order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		results, err2 := MysqlDb.Exec("insert INTO 30min(`time`,`open`,high,low,`close`,volume) values(?,0,0,0,0,0)", time1-threty_min_start)
		if err2 != nil {
			fmt.Println(err2)
		}
		id, _ := results.LastInsertId()
		id_threty_min <- id
		return
	}

	if out.Time != threty_min_start {
		results, _ := MysqlDb.Exec("insert INTO 30min(`time`,`open`,high,low,`close`,volume) values(?,0,0,0,0,0)", time1-threty_min_start)
		id, _ := results.LastInsertId()
		id_threty_min <- id
	} else {

		id_threty_min <- out.Id
	}
}

func Twilve_hour_check(time1 int64, id_twilve_hour chan int64) {
	Connect()
	//var out Kline_info
	out := new(Kline_info)
	twilve_hour := time1 % (3600 * 12)
	row := MysqlDb.QueryRow("select id,time from 12hour order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		results, err2 := MysqlDb.Exec("insert INTO 12hour(`time`,`open`,high,low,`close`,volume) values(?,0,0,0,0,0)", time1-twilve_hour)
		if err2 != nil {
			fmt.Println(err2)
		}
		id, _ := results.LastInsertId()
		id_twilve_hour <- id
		return
	}

	if out.Time != twilve_hour {
		results, _ := MysqlDb.Exec("insert INTO 12hour(`time`,`open`,high,low,`close`,volume) values(?,0,0,0,0,0)", time1-twilve_hour)
		id, _ := results.LastInsertId()
		id_twilve_hour <- id
	} else {

		id_twilve_hour <- out.Id
	}
}

func One_day_check(time1 int64, id_day chan int64) {
	Connect()
	//var out Kline_info
	out := new(Kline_info)
	one_day_start := time1 % (3600 * 24)
	row := MysqlDb.QueryRow("select id,time from 1day order by id desc limit 1")
	if err := row.Scan(&out.Id, &out.Time); err != nil {
		results, err2 := MysqlDb.Exec("insert INTO 1day(`time`,`open`,high,low,`close`,volume) values(?,0,0,0,0,0)", time1-one_day_start)
		if err2 != nil {
			fmt.Println(err2)
		}
		id, _ := results.LastInsertId()
		id_day <- id
		return
	}

	if out.Time != one_day_start {
		results, _ := MysqlDb.Exec("insert INTO 1day(`time`,`open`,high,low,`close`,volume) values(?,0,0,0,0,0)", time1-one_day_start)
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

func Get_info(id int64, form string) kline {
	var a kline
	info := new(kline)
	row := MysqlDb.QueryRow("select * from ? where id=?", id, form)
	err := row.Scan(&info.Id, &info.High, &info.Open, &info.Low, &info.Close, &info.Volume, &info.Time)
	if err != nil {
		panic(err)
	}
	a = kline{info.Id, info.High, info.Open, info.Low, info.Close, info.Volume, info.Time}
	return a
}
