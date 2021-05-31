package main

import (
	"awesomeProject/src/mysql"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

type wallet struct {
	Nomal_money  float64 `db:"nomal_money"`
	Locked_money float64 `db:"locked_money"`
	Nomal_coin   float64 `db:"nomal_coin"`
	Locked_coin  float64 `db:"locked_coin"`
}

func main() {

	router := gin.Default()
	router.GET("/get_account", get_account)
	router.POST("/sign_account", sign_account)
	router.POST("/deposit", doposit)
	order := router.Group("/order")
	order.POST("/create_order", create_order)
	order.POST("/get_open_order", Get_oppen_order)
	order.POST("/cancel_order", Cancel_order)
	router.Run()
}

func get_account(c *gin.Context) {
	Account := c.Query("account")

	mysql.Connect()
	a := mysql.Check_same_account(Account)
	if a == true {
		var b wallet
		d := mysql.Checkfile(Account)
		b.Locked_money = d.Lock_money
		b.Nomal_coin = d.Normal_Coin
		b.Locked_coin = d.Lock_coin
		b.Nomal_money = d.Normal_Money
		fmt.Println(b.Locked_coin, b.Locked_money, b.Nomal_money, b.Nomal_coin)
		fmt.Println(b)

		//Money := mysql.Checkfile(Account).Normal_Money
		//Coin := mysql.Checkfile(Account).normal_Coin
		//lock_money := mysql.Checkfile(Account).lock_money
		//lock_coin := mysql.Checkfile(Account).lock_coin
		c.JSON(200, gin.H{
			"code": "200",
			"data": b,
		})

	} else {
		c.JSON(200, gin.H{
			"code":  "10002",
			"error": "account not found",
		})
	}

}

func sign_account(c *gin.Context) {
	Account := c.Query("account")
	Password := c.Query("password")
	mysql.Connect()
	a := mysql.Check_same_account(Account)
	if a != true {
		mysql.Write_account(Account, Password)
		c.JSON(200, gin.H{
			"code":    200,
			"message": "success",
		})
	} else {
		mysql.Write_info(Account, 0, 0, 0, 0)
		c.JSON(200, gin.H{
			"code":    1003,
			"message": "already have this account",
		})

	}

}

func doposit(c *gin.Context) {
	Account := c.Query("account")
	coin, _ := strconv.ParseFloat(c.DefaultQuery("coin", "0"), 64)
	money, _ := strconv.ParseFloat(c.DefaultQuery("money", "0"), 64)
	fmt.Println(coin, money)

	a := mysql.Check_same_account(Account)
	fmt.Println(a)
	if a != true {
		c.JSON(200, gin.H{
			"code":    1001,
			"message": "account not found",
		})
		return

	} else {
		info := mysql.Checkfile(Account)
		mysql.Write_info(Account, info.Normal_Money+money, coin+info.Normal_Coin, info.Lock_money, info.Lock_coin)
		c.JSON(200, gin.H{
			"code":    200,
			"message": "success",
		})
	}

}

func create_order(c *gin.Context) {
	mysql.Connect()
	side := c.Query("side")
	price, _ := strconv.ParseFloat(c.Query("price"), 64)
	volume, _ := strconv.ParseFloat(c.Query("volume"), 64)
	account := c.Query("account")
	info := mysql.Checkfile(account)
	use := price * volume
	fmt.Println("price=?", price)
	fmt.Println("volume=?", volume)
	fmt.Println("use=?", use)
	fmt.Println("normal=?", info.Normal_Money)
	fmt.Println(info.Lock_money)
	if price == 0 || volume == 0 {
		c.JSON(200, gin.H{
			"code":    1006,
			"message": "wrong price or volume",
		})

	}

	if side != "sell" && side != "buy" {
		c.JSON(200, gin.H{
			"code":  1002,
			"error": "wrong side",
		})
	} else {

		a := mysql.Check_same_account(account)
		t := time.Now().Unix()
		if a == true {
			if side == "sell" {
				if volume > info.Normal_Coin {
					c.JSON(200, gin.H{
						"code":  1002,
						"error": "not enough coin",
					})

				} else {
					mysql.Write_info(account, info.Normal_Money, info.Normal_Coin-volume, info.Lock_money, info.Lock_coin+volume)
					id := mysql.Create_order(account, price, volume, side, "online", t, info.Id)
					c.JSON(200, gin.H{
						"code":     200,
						"message":  "success",
						"order_id": id,
					})
				}

			}
			if side == "buy" {
				if use > info.Normal_Money {
					c.JSON(200, gin.H{
						"code":  1001,
						"error": "not enough money",
					})
				} else {
					mysql.Write_info(account, info.Normal_Money-use, info.Normal_Coin, info.Lock_money+use, info.Lock_coin)
					id := mysql.Create_order(account, price, volume, side, "online", t, info.Id)
					c.JSON(200, gin.H{
						"code":     200,
						"message":  "success",
						"order_id": id,
					})
				}

			}

		} else {
			c.JSON(200, gin.H{
				"code":    1010,
				"message": "account not found",
			})
		}

	}

}

func Get_oppen_order(c *gin.Context) {
	mysql.Connect()
	account := c.Query("account")
	a := mysql.Check_same_account(account)
	if a != true {
		c.JSON(200, gin.H{
			"code":    1001,
			"message": "account not found",
		})
		return

	}
	b := mysql.Get_open(account)
	c.JSON(200, gin.H{
		"code": 200,
		"data": b,
	})

}

func Cancel_order(c *gin.Context) {
	mysql.Connect()
	account := c.Query("account")
	id, err := strconv.ParseInt(c.Query("id"), 0, 64)
	s := mysql.Check_order_info(account, id)
	d := mysql.Checkfile(account)
	a := mysql.Check_same_account(account)
	use := s.Price * s.Volume
	if a != true {
		c.JSON(200, gin.H{
			"code":    1001,
			"message": "account not found",
		})
		return

	}
	if err != nil {
		c.JSON(200, gin.H{
			"code":    1001,
			"message": "wrong id",
		})
		return
	}
	if s.Status != "online" {
		c.JSON(200, gin.H{
			"code":    1001,
			"message": "wrong status",
		})
		return
	}
	bb := mysql.Cancel_order(account, id)
	if bb != nil {
		c.JSON(200, gin.H{
			"code":    1001,
			"message": "wrong id",
		})
		return
	} else {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "success",
		})
		if s.Side == "sell" {
			mysql.Write_info(account, d.Normal_Money, d.Normal_Coin+s.Volume, d.Lock_money, d.Lock_coin-s.Volume)
		} else {
			mysql.Write_info(account, d.Normal_Money+use, d.Normal_Coin, d.Lock_money-use, d.Lock_coin)
		}

	}
}

func Deal_orders(sleep float64) {
	for {

	}

}
