package mysql

import (
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
}
type Find struct {
	price   float64 `db:"price"`
	volume  float64 `db:"volume"`
	account float64 `db:"account"`
}
type Order struct {
	Id      int64  `db:"id"`
	Price   float64  `db:"price"`
	Volume  float64  `db:"volume"`
	Side    string `db:"side"`
	Status  string `db:"status"`
	Time    int64  `db:"time"`
	Account string `db:"account"`
	User_id int64  `db:"user_id"`
}
type Open_order struct {
	Id   int64
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
	MysqlDb.SetMaxOpenConns(100)
	// 闲置连接数
	MysqlDb.SetMaxIdleConns(20)
	// 最大连接周期
	MysqlDb.SetConnMaxLifetime(100 * time.Second)

	if MysqlDbErr = MysqlDb.Ping(); nil != MysqlDbErr {
		panic("数据库链接失败: " + MysqlDbErr.Error())
	}
}
func Get_open(Account string) []Open_order  {
	var a Open_order
	orders := new(Order)
	row, err := MysqlDb.Query("select * from orders where account=? and status=?", Account,"online")
	if err != nil {
		log.Println(err)
	}
	var ss []Open_order
	for row.Next() {
		err = row.Scan(&orders.Id, &orders.Price, &orders.Volume, &orders.Side, &orders.Status, &orders.Time, &orders.Account, &orders.User_id)
		if err != nil {
			log.Println(err)
		}
		a.Id = orders.Id
		a.Side = orders.Side
		a.Time = orders.Time
		a.Volume = orders.Volume
		a.Price = orders.Price
		ss = append(ss, a)

	}

	return ss
}


func Checkfile(Account string) User {
	var a User
	user := new(User)
	row := MysqlDb.QueryRow("select * from Account where Account=?", Account)
	if err := row.Scan(&user.Id, &user.Account, &user.Password, &user.Normal_Money, &user.Normal_Coin, &user.Lock_money, &user.Lock_coin); err != nil {
		fmt.Printf("scan failed, err:%v", err)
		//return
	}
	fmt.Println(user.Id, user.Account, user.Password, user.Normal_Money, user.Normal_Coin, user.Lock_money, user.Lock_coin)
	a.Id = user.Id
	a.Account = user.Account
	a.Password = user.Password
	a.Normal_Money = user.Normal_Money
	a.Normal_Coin = user.Normal_Coin
	a.Lock_money = user.Lock_money
	a.Lock_coin = user.Lock_coin
	fmt.Println(a)

	return a

}
func Write_account(Account string, Password string) bool {
	a := Check_same_account(Account)
	if a == false {
		_, err := MysqlDb.Exec("insert INTO Account(Account,Password) values(?,?)", Account, Password)
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

func Create_order(Account string, price float64, volume float64, side string, status string, time int64, user_id int64) int64 {
	results, err := MysqlDb.Exec("insert INTO orders(account,price,volume,side,status,time,user_id) values(?,?,?,?,?,?,?)", Account, price,
		volume, side, status, time, user_id)
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

func deel_order(account string,price float64, volume float64, side string, time int64, buyer string, seller string) {
	if side =="sell"{
		row, err := MysqlDb.Query("select price, volume, account from orders where `status`=? and price>=? and side=?","online",
		price,"buy")
		if err!= nil{
			log.Println(err)
		}
	}


}

//func Find_order(side string, price float64) {
//	//order := []Find
//	var order []Find
//	if side == "sell" {
//		row := MysqlDb.QueryRow("select price, volume, account from orders where `status`=? and price>=? and side=?","online",
//			price,"buy")
//		if err := row.Scan(&order.price, &order.volume, &order.account); err != nil {
//			fmt.Printf("scan failed, err:%v", err)
//		}
//	}
//
//}

func Cancel_order(account string,id int64) error{
	_, err := MysqlDb.Exec("UPDATE orders set status=? where id=? and account=?", "canceled", id,account)
	if err != nil {
		panic(err)
	}
			return err
}

\