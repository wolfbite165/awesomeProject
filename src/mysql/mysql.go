package mysql

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

var MysqlDb *sql.DB
var MysqlDbErr error

const (
	USER_NAME = "root"
	PASS_WORD = "yyf262518"
	HOST      = "127.0.0.1"
	PORT      = "3306"
	DATABASE  = "account"
	//CHARSET   = "utf8"
)

type User struct {
	Id           int64   `db:"id"`
	Account      string  `db:"Account"`
	Password     string  `db:"Password"`
	Normal_Money float64 `db:"normal_Money"`
	Normal_Coin  float64 `db:"normal_Coin"`
	Lock_money   float64 `db:"lock_money"`
	Lock_coin    float64 `db:"lock_coin"`
	googleSK     string  `db:"GoogleSK"`
}
type Find struct {
	price   float64 `db:"price"`
	volume  float64 `db:"volume"`
	account float64 `db:"account"`
}
type Order struct {
	Id      int64   `db:"id"`
	Price   float64 `db:"price"`
	Volume  float64 `db:"volume"`
	Side    string  `db:"side"`
	Status  string  `db:"status"`
	Time    int64   `db:"time"`
	Account string  `db:"account"`
	User_id int64   `db:"user_id"`
	Left    float64 `db:"left"`
}
type Open_order struct {
	Id     int64
	Price  float64
	Volume float64
	Side   string
	Time   int64
}
type Deal_order struct {
	Price   float64
	Volume  float64
	account string
}
type Orders struct {
	Id      int64
	Price   float64
	Volume  float64
	Account string
}
type trades struct {
	Price  float64
	Volume float64
	Side   string
	Time   int64
}

func Connect() {
	dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", USER_NAME, PASS_WORD, HOST, PORT, DATABASE)
	fmt.Println(dbDSN)
	MysqlDb, MysqlDbErr = sql.Open("mysql", dbDSN)
	if MysqlDbErr != nil {
		log.Println("dbDSN: " + dbDSN)
		panic("数据源配置不正确: " + MysqlDbErr.Error())
	}

	// 最大连接数
	MysqlDb.SetMaxOpenConns(1000)
	// 闲置连接数
	MysqlDb.SetMaxIdleConns(20)
	// 最大连接周期
	MysqlDb.SetConnMaxLifetime(100 * time.Second)

	if MysqlDbErr = MysqlDb.Ping(); nil != MysqlDbErr {
		panic("数据库链接失败: " + MysqlDbErr.Error())
	}
}
func Get_open(Account string) []Open_order {
	var a Open_order
	orders := new(Order)
	row, err := MysqlDb.Query("select * from orders where account=? and status=?", Account, "online")
	if err != nil {
		log.Println(err)
	}
	var ss []Open_order
	for row.Next() {
		err = row.Scan(&orders.Id, &orders.Price, &orders.Volume, &orders.Side, &orders.Status, &orders.Time, &orders.Account, &orders.User_id, &orders.Left)
		if err != nil {
			log.Println(err)
		}
		a.Id = orders.Id
		a.Side = orders.Side
		a.Time = orders.Time
		a.Volume = orders.Left
		a.Price = orders.Price
		ss = append(ss, a)

	}

	return ss
}
func Check_order_info(account string, id int64) Order {
	var a Order
	orders := new(Order)
	row := MysqlDb.QueryRow("select * from orders where Account=? and id=?", account, id)
	if err := row.Scan(&orders.Id, &orders.Price, &orders.Volume, &orders.Side, &orders.Status, &orders.Time, &orders.Account, &orders.User_id, &orders.Left); err != nil {
		fmt.Printf("scan failed, err:%v", err)
		//return
	}
	//fmt.Println(order.Id, order.Account, order.Price, order.Volume, order.Normal_Coin, order.Lock_money, order.Lock_coin)
	a.Id = orders.Id
	a.Account = orders.Account
	a.Price = orders.Price
	a.Side = orders.Side
	a.Time = orders.Time
	a.Status = orders.Status
	a.User_id = orders.User_id
	a.Volume = orders.Volume
	a.Left = orders.Left
	fmt.Println(a)

	return a
}
func Checkfile(Account string) User {
	var a User
	user := new(User)
	row := MysqlDb.QueryRow("select * from Account where Account=?", Account)
	if err := row.Scan(&user.Id, &user.Account, &user.Password, &user.Normal_Money, &user.Normal_Coin, &user.Lock_money, &user.Lock_coin, &user.googleSK); err != nil {
		fmt.Printf("scan failed, err:%v", err)
		//return
	}
	//fmt.Println(user.Id, user.Account, user.Password, user.Normal_Money, user.Normal_Coin, user.Lock_money, user.Lock_coin)
	a.Id = user.Id
	a.Account = user.Account
	a.Password = user.Password
	a.Normal_Money = user.Normal_Money
	a.Normal_Coin = user.Normal_Coin
	a.Lock_money = user.Lock_money
	a.Lock_coin = user.Lock_coin
	a.googleSK = user.googleSK
	//fmt.Println(a)

	return a

}
func Write_account(Account string, Password string, googleSk string) bool {
	a := Check_same_account(Account)
	if a == false {
		{
			data := []byte(Password)
			has := md5.Sum(data)
			Password = fmt.Sprintf("%x", has)
		}
		_, err := MysqlDb.Exec("insert INTO Account(Account,Password,GoogleSK) values(?,?,?)", Account, Password, googleSk)
		if err != nil {
			fmt.Println(err)

		}
		return true

	} else {
		panic("已有重复的账户名")
		return false

	}

}
func Write_info(Account string, Money float64, Coin float64, lock_money float64, lock_coin float64) {
	_, err := MysqlDb.Exec("UPDATE Account set normal_Money=? where Account=?", Money, Account)
	if err != nil {
		fmt.Println(err)
	}
	_, err = MysqlDb.Exec("UPDATE Account set normal_Coin=? where Account=?", Coin, Account)

	if err != nil {
		fmt.Println(err)
	}
	_, err = MysqlDb.Exec("UPDATE Account set lock_coin=? where Account=?", lock_coin, Account)

	if err != nil {
		fmt.Println(err)
	}
	_, err = MysqlDb.Exec("UPDATE Account set lock_money=? where Account=?", lock_money, Account)

	if err != nil {
		fmt.Println(err)
	}
}
func Write_order_info(Id int64, status string, left float64) {
	_, err := MysqlDb.Exec("UPDATE orders set `status`=? where id=?", status, Id)
	if err != nil {
		fmt.Println(err)
	}
	_, err = MysqlDb.Exec("UPDATE orders set `left`=? where id=?", left, Id)
	if err != nil {
		fmt.Println(err)
	}

}
func Write_trade_info(price float64, volume float64, side string, time int64, buyer string, seller string) error {
	_, err := MysqlDb.Exec("insert INTO trade(price,volume,side,time,buyer,seller) values(?,?,?,?,?,?)", price,
		volume, side, time, buyer, seller)
	if err != nil {
		panic(err)
		return err
	}
	return err
	//results, err := MysqlDb.Exec("insert INTO secend(price,volume) value (?,?)", price, volume)
	//_, err = results.LastInsertId()
	//if err != nil {
	//	//panic(err)
	//	return err
	//}

}
func Create_order(Account string, price float64, volume float64, side string, status string, time int64, user_id int64, left float64) int64 {
	results, err := MysqlDb.Exec("insert INTO orders(account,price,volume,side,status,time,user_id,`left`) values(?,?,?,?,?,?,?,?)", Account, price,
		volume, side, status, time, user_id, left)
	if err != nil {
		panic(err)
	}
	id, err := results.LastInsertId()
	if err != nil {
		panic(err)
		return 0
	}

	return id
}
func Get_account_info(account string) (out acc_info, err error) {
	var rows *sql.Row
	rows = MysqlDb.QueryRow("select Account,Password,GoogleSK,id from account where Account=?",
		account)
	if err := rows.Scan(&out.Account, &out.Pwd, &out.Google, &out.Id); err != nil {
		fmt.Printf("scan failed, err:%v", err)
		//return
	}
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//var name string
	//var pwd string
	//var gsk string
	//var id int

	return

}
func Cancel_order(account string, id int64) error {
	_, err := MysqlDb.Exec("UPDATE orders set status=? where id=? and account=?", "canceled", id, account)
	if err != nil {
		panic(err)
	}
	return err
}
func Get_side_info() ([]Order, []Order) {
	var buy []Order
	var sell []Order
	rows, err := MysqlDb.Query("SELECT * FROM orders where `status`=? and side=? ORDER BY price DESC, `time` ", "online", "buy")
	if err != nil {
		log.Println(err)
	}
	for rows.Next() {
		var Id int64
		var Price float64
		var Volume float64
		var side string
		var status string
		var Time int64
		var account string
		var user_id int64
		var left float64

		err := rows.Scan(&Id, &Price, &Volume, &side, &status, &Time, &account, &user_id, &left)
		if err != nil {
			log.Println(err)
		}
		buy = append(buy, Order{Id, Price, Volume, side, status, Time, account, user_id, left})
		//fmt.Println(buy)
	}
	row, err := MysqlDb.Query("SELECT * FROM orders where `status`=? and side=? ORDER BY price, `time` ", "online", "sell")
	if err != nil {
		log.Println(err)
	}
	for row.Next() {
		var Id int64
		var Price float64
		var Volume float64
		var side string
		var status string
		var Time int64
		var account string
		var user_id int64
		var left float64

		err := row.Scan(&Id, &Price, &Volume, &side, &status, &Time, &account, &user_id, &left)
		if err != nil {
			log.Println(err)
		}
		sell = append(sell, Order{Id, Price, Volume, side, status, Time, account, user_id, left})
		//fmt.Println(sell)

	}
	//fmt.Println(buy, sell)

	return buy, sell
}
func Get_trade_list(num int64) []trades {
	var out []trades
	rows, err := MysqlDb.Query("select * from trade order by id desc limit 0,?", num)
	if err != nil {
		log.Println(err)
	}
	for rows.Next() {
		var Id int64
		var price float64
		var volume float64
		var side string
		var time int64
		var buyer string
		var seller string
		err := rows.Scan(&Id, &price, &volume, &side, &time, &buyer, &seller)
		if err != nil {
			log.Println(err)
		}
		out = append(out, trades{Price: price, Volume: volume, Side: side, Time: time})
	}
	return out
}
func Write_kline() {
	start := time.Now().Unix()
	last_min_start := start - (start % 60)
	last_hours_start := start - (start % 3600)
	last_threty_min_start := start - (start % 1800)
	last_five_min_start := start - (start % 300)
	last_twilve_hour := start - (start % (3600 * 12))
	last_one_day_start := start - (start % (3600 * 24))
	id_min := make(chan int64)
	id_hour := make(chan int64)
	id_five_min := make(chan int64)
	id_threty_min := make(chan int64)
	id_twilve_hour := make(chan int64)
	id_day := make(chan int64)
	go Min_check(start, id_min)
	go Hours_check(start, id_hour)
	go Five_min_check(start, id_five_min)
	go Threty_min_check(start, id_threty_min)
	go Twilve_hour_check(start, id_twilve_hour)
	go One_day_check(start, id_day)
	idmin, idhour, idfive, idthirty, idtwelve, idday := <-id_min,
		<-id_hour,
		<-id_five_min,
		<-id_threty_min,
		<-id_twilve_hour,
		<-id_day
	last_time := time.Now().Unix()

	for {
		now_time := time.Now().Unix()
		next_sec := last_time + 1
		next_min_start := last_min_start + 60
		next_hours_start := last_hours_start + 3600
		next_threty_min_start := last_threty_min_start + 1800
		next_five_min_start := last_five_min_start + 300
		next_twilve_hour := last_twilve_hour + (3600 * 12)
		next_one_day_start := last_one_day_start + (3600 * 24)
		if now_time >= next_sec {
			MysqlDb.Exec("insert INTO `1day`(time) values(?)", time1-one_day_start)

		}
	}
}

func Get_ticker() (out ticker, err error) {
	//tick :=new(ticker)
	row := MysqlDb.QueryRow("select price, volume, side from trade order by id desc limit 1")
	if err := row.Scan(&out.Price, &out.Volume, &out.Side); err != nil {
		fmt.Printf("scan failed, err:%v", err)
	}

	return

}

type ticker struct {
	Price  float64 `db:"price"`
	Volume float64 `db:"volume"`
	Side   string  `db:"side"`
}
type acc_info struct {
	Account string `db:"Account"`
	Pwd     string `db:"Password"`
	Google  string `db:"GoogleSK"`
	Id      int    `db:"id"`
}
type secend struct {
	Id     int     `db:"id"`
	Price  float64 `db:"price"`
	Volume float64 `db:"volume"`
}

//
//var leftOrder Order
//sellIndex := 0
//buyIndex := 1
//var okIds []int
//
//leftOrder = buy[buyIndex]
//for {
//	if len(sell) == sellIndex + 1 &&  len(buy) == buyIndex + 1 && sell[sellIndex].Price > buy[buyIndex].Price {
//		break
//	}
//
//	if leftOrder.Side == "buy" {
//		if sell[sellIndex].Volume > leftOrder.Volume{
//
//			sell[sellIndex].Volume -= leftOrder.Volume
//			leftOrder = sell[sellIndex]
//		} else {
//
//
//
//
//			sellIndex++
//		}
//
//	} else if leftOrder.Side == "sell"{
//
//	} else {
//		return err
//	}
//}
