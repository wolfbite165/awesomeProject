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
	PASS_WORD = "BlockPulse"
	HOST      = "81.69.224.151"
	PORT      = "3306"
	DATABASE  = "Account"
	//CHARSET   = "utf8"
)

type User struct {
	Id       int64   `db:"id"`
	Account  string  `db:"Account"`
	Password string  `db:"Password"`
	Money    float64 `db:"Money"`
	Coin     float64 `db:"Coin"`
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

func Checkfile(Account string) User {
	var a User
	user := new(User)
	row := MysqlDb.QueryRow("select * from Account where Account=?", Account)
	if err := row.Scan(&user.Id, &user.Account, &user.Password, &user.Money, &user.Coin); err != nil {
		fmt.Printf("scan failed, err:%v", err)
		//return
	}
	fmt.Println(user.Id, user.Account, user.Password, user.Money, user.Coin)
	a.Id = user.Id
	a.Account = user.Account
	a.Password = user.Password
	a.Money = user.Money
	a.Coin = user.Coin
	fmt.Println(a.Id)

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

func Write_info(Account string, Money float64, Coin float64) {
	_, err := MysqlDb.Exec("UPDATE Account set Money=? where Account=?", Money, Account)
	if err != nil {
		fmt.Println(err)
	}
	_, err = MysqlDb.Exec("UPDATE Account set Coin=? where Account=?", Coin, Account)

	if err != nil {
		fmt.Println(err)
	}
}

func Create_order(Account string, price float64, volume float64, side string, status string, time time.Time) {
	_, err := MysqlDb.Exec("insert INTO orders(account,price,volume,side,status,time) values(?,?,?,?,?,?)", Account, price,
		volume, side, status, time)
	if err != nil {
		panic(err)
	}

}

func deel_order(price float64, volume float64, side string, time time.Time, buyer string, seller string) {
	_, err := MysqlDb.Exec("insert INTO trade(price,volume,side,time,buyer,seller) values(?,"+
		"?,?,?,?,?)", price, volume, side, time, buyer, seller)
	if err != nil {
		panic(err)
	}
}
