package router

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

type Out struct {
	price  float64
	volume float64
	side   string
}
type Trade_out struct {
	sell Out
	buy  Out
}

func InitRouter() {

	router := gin.Default()
	apiR := router.Group("/api")
	authRouter(apiR)
	{
		userR := apiR.Group("/user")
		userR.Use(JWTAuth())
		userR.POST("/get_account", get_account)
	}
	{
		commonR := apiR.Group("/common")
		commonR.POST("/deposit", doposit)
		commonR.GET("/get_depth", Get_depth)
		commonR.GET("/get_trade_history", Get_trade_history)
		commonR.GET("/ticker", Get_ticker)
	}
	{
		order := apiR.Group("/order")
		order.Use(JWTAuth())
		order.POST("/create_order", create_order)
		order.POST("/get_open_order", Get_oppen_order)
		order.POST("/cancel_order", Cancel_order)
	}
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
			"code": 200,
			"data": b,
		})

	} else {
		c.JSON(200, gin.H{
			"code":  "10002",
			"error": "account not found",
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
	var orderdata create
	err := c.BindJSON(&orderdata)
	side := c.Query("side")
	price, _ := strconv.ParseFloat(c.Query("price"), 64)
	volume, _ := strconv.ParseFloat(c.Query("volume"), 64)
	account := c.Query("account")
	info := mysql.Checkfile(account)
	buy, sell := mysql.Get_side_info()
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
					if len(buy) == 0 {
						t := time.Now().Unix()
						mysql.Write_info(account, info.Normal_Money, info.Normal_Coin-volume, info.Lock_money, info.Lock_coin+volume)
						id := mysql.Create_order(account, price, volume, side, "online", t, info.Id, volume)
						c.JSON(200, gin.H{
							"code":     201,
							"message":  "success",
							"order_id": id,
						})
						return
					}
					buyindex := 0
					id := mysql.Create_order(account, price, volume, side, "online", t, info.Id, volume)
					//mysql.Write_info(account, info.Normal_Money, info.Normal_Coin-volume, info.Lock_money, info.Lock_coin+volume)
					for {
						info = mysql.Checkfile(account)
						fmt.Println(len(buy))
						fmt.Println(buyindex)
						//fmt.Println("无敌",price,buy[buyindex].Price)
						info2 := mysql.Checkfile(buy[buyindex].Account)
						if price <= buy[buyindex].Price {

							if volume < buy[buyindex].Volume {

								c.JSON(200, gin.H{
									"code":     202,
									"message":  "success",
									"order_id": id,
								})
								mysql.Write_trade_info(buy[buyindex].Price, volume, side, t, buy[buyindex].Account, account)
								mysql.Write_info(buy[buyindex].Account, info2.Normal_Money, info2.Normal_Coin+volume, info2.Lock_money-buy[buyindex].Volume*buy[buyindex].Price, info2.Lock_coin) //买家(挂单)的余额修改
								mysql.Write_info(account, info.Normal_Money+buy[buyindex].Volume*buy[buyindex].Price, info.Normal_Coin-volume, info.Lock_money, info.Lock_coin)                   //卖家(下单)余额修改
								mysql.Write_order_info(buy[buyindex].Id, "online", buy[buyindex].Volume-volume)
								mysql.Write_order_info(id, "dealed", volume-volume)
								break
							}
							if volume == buy[buyindex].Volume {
								//id := mysql.Create_order(account, price, volume, side, "dealed", t, info.Id)
								c.JSON(200, gin.H{
									"code":     203,
									"message":  "success",
									"order_id": id,
								})
								mysql.Write_trade_info(buy[buyindex].Price, volume, side, t, buy[buyindex].Account, account)
								mysql.Write_info(buy[buyindex].Account, info2.Normal_Money, info2.Normal_Coin+volume, info2.Lock_money-buy[buyindex].Volume*buy[buyindex].Price, info2.Lock_coin) //买家的余额修改
								mysql.Write_info(account, info.Normal_Money+buy[buyindex].Volume*buy[buyindex].Price, info.Normal_Coin-volume, info.Lock_money, info.Lock_coin)                   //卖家余额修改
								mysql.Write_order_info(buy[buyindex].Id, "dealed", buy[buyindex].Volume-volume)
								mysql.Write_order_info(id, "dealed", volume-volume)
								break

							}
							if volume > buy[buyindex].Volume {

								mysql.Write_order_info(id, "online", volume-buy[buyindex].Volume)
								volume = volume - buy[buyindex].Volume
								mysql.Write_info(buy[buyindex].Account,
									info2.Normal_Money,
									info2.Normal_Coin+buy[buyindex].Volume, info2.Lock_money-buy[buyindex].Price*buy[buyindex].Volume,
									info2.Lock_coin)
								mysql.Write_info(account,
									info.Normal_Money+buy[buyindex].Price*buy[buyindex].Volume,
									info.Normal_Coin-buy[buyindex].Volume, info.Lock_money,
									info.Lock_coin)
								t2 := time.Now().Unix()
								mysql.Write_trade_info(buy[buyindex].Price, buy[buyindex].Volume, side, t2, buy[buyindex].Account, account)
								mysql.Write_order_info(buy[buyindex].Id, "dealed", 0)
								mysql.Write_order_info(id, "online", volume)

								buyindex++

							}

						} else {
							//mysql.Write_info(account, info.Normal_Money, info.Normal_Coin-volume, info.Lock_money, info.Lock_coin+volume)
							//id := mysql.Create_order(account, price, volume, side, "online", t, info.Id)
							c.JSON(200, gin.H{
								"code":     204,
								"message":  "success",
								"order_id": id,
							})
							break
						}

					}

				}

			}
			if side == "buy" {
				if use > info.Normal_Money {
					c.JSON(200, gin.H{
						"code":  1001,
						"error": "not enough money",
					})
				} else {
					if len(sell) == 0 {
						t = time.Now().Unix()
						mysql.Write_info(account, info.Normal_Money-use, info.Normal_Coin, info.Lock_money+use, info.Lock_coin)
						id := mysql.Create_order(account, price, volume, side, "online", t, info.Id, volume)
						c.JSON(206, gin.H{
							"code":     200,
							"message":  "success",
							"order_id": id,
						})
						return
					}
					sellindex := 0
					//mysql.Write_info(account, info.Normal_Money-use, info.Normal_Coin, info.Lock_money+use, info.Lock_coin)
					id := mysql.Create_order(account, price, volume, side, "online", t, info.Id, volume)
					for {
						info = mysql.Checkfile(account)
						//fmt.Println(len(sell))
						//fmt.Println(sellindex)
						info2 := mysql.Checkfile(sell[sellindex].Account)
						if price >= sell[sellindex].Price {
							if volume < sell[sellindex].Volume {
								//id := mysql.Create_order(account, price, volume, side, "dealed", t, info.Id, volume)
								c.JSON(200, gin.H{
									"code":     200,
									"message":  "success",
									"order_id": id,
								})
								mysql.Write_trade_info(sell[sellindex].Price, volume, side, t, account, sell[sellindex].Account)
								mysql.Write_info(account, info.Normal_Money-sell[sellindex].Volume*sell[sellindex].Price, info.Normal_Coin+volume, info.Lock_money, info.Lock_coin)                     //买家的余额修改
								mysql.Write_info(sell[sellindex].Account, info2.Normal_Money+sell[sellindex].Volume*sell[sellindex].Price, info2.Normal_Coin, info2.Lock_money, info2.Lock_coin-volume) //卖家余额修改
								mysql.Write_order_info(sell[sellindex].Id, "online", sell[sellindex].Volume-volume)
								mysql.Write_order_info(id, "dealed", volume-volume)
								break
							}
							if volume == sell[sellindex].Volume {
								//id := mysql.Create_order(account, price, volume, side, "dealed", t, info.Id, volume)
								c.JSON(200, gin.H{
									"code":     207,
									"message":  "success",
									"order_id": id,
								})
								mysql.Write_trade_info(sell[sellindex].Price, volume, side, t, account, sell[sellindex].Account)
								mysql.Write_info(account, info.Normal_Money-sell[sellindex].Volume*sell[sellindex].Price, info.Normal_Coin+volume, info.Lock_money, info.Lock_coin)
								fmt.Println("等于富豪", account, info.Normal_Money-sell[sellindex].Volume*sell[sellindex].Price, info.Normal_Coin+volume, info.Lock_money, info.Lock_coin)  //买家的余额修改
								mysql.Write_info(sell[sellindex].Account, info2.Normal_Money+sell[sellindex].Price*volume, info2.Normal_Coin, info2.Lock_money, info2.Lock_coin-volume) //卖家余额修改
								//fmt.Println(sell[sellindex].Id)
								mysql.Write_order_info(sell[sellindex].Id, "dealed", sell[sellindex].Volume-volume)
								mysql.Write_order_info(id, "dealed", volume-volume)
								break

							}
							if volume > sell[sellindex].Volume {

								volume -= sell[sellindex].Volume
								mysql.Write_info(sell[sellindex].Account,
									info2.Normal_Money+sell[sellindex].Price*sell[sellindex].Volume,
									info2.Normal_Coin, info.Lock_money,
									info2.Lock_coin-sell[sellindex].Volume)
								mysql.Write_info(account,
									info.Normal_Money-sell[sellindex].Price*sell[sellindex].Volume,
									info.Normal_Coin+sell[sellindex].Volume, info.Lock_money,
									info.Lock_coin)
								fmt.Println("大雨符号", account,
									info.Normal_Money-sell[sellindex].Price*sell[sellindex].Volume,
									info.Normal_Coin+sell[sellindex].Volume, info.Lock_money,
									info.Lock_coin)
								t2 := time.Now().Unix()
								mysql.Write_trade_info(sell[sellindex].Price, sell[sellindex].Volume, side, t2, account, sell[sellindex].Account)
								mysql.Write_order_info(sell[sellindex].Id, "dealed", 0)
								mysql.Write_order_info(id, "online", volume)
								sellindex++

							}

						} else {
							mysql.Write_info(account, info.Normal_Money-use, info.Normal_Coin, info.Lock_money+use, info.Lock_coin)
							//id := mysql.Create_order(account, price, volume, side, "online", t, info.Id, volume)
							c.JSON(200, gin.H{
								"code":     208,
								"message":  "success",
								"order_id": id,
							})
							break
						}

					}
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
	token := c.Request.Header.Get("token")
	j := NewJWT()
	claims, err := j.ParseToken(token)
	account := claims.Name
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

func Get_depth(c *gin.Context) {
	var buy1 []Out
	var sell1 []Out
	var out [][]Out
	mysql.Connect()
	buy, sell := mysql.Get_side_info()
	for i := 0; i == len(buy); i++ {
		buy1 = append(buy1, Out{buy[i].Price, buy[i].Volume, buy[i].Side})
		sell1 = append(sell1, Out{sell[i].Price, sell[i].Volume, sell[i].Side})
	}
	out = append(out, buy1, sell1)
	c.JSON(200, gin.H{
		"code": 200,
		"data": out,
	})

}

func Get_trade_history(c *gin.Context) {
	mysql.Connect()
	times, _ := strconv.ParseInt(c.DefaultQuery("length", "100"), 0, 64)
	if times > 100 || times <= 0 {
		c.JSON(200, gin.H{
			"code":    10010,
			"message": "length out range",
		})
	} else {
		cc := mysql.Get_trade_list(times)
		c.JSON(200, gin.H{
			"code":    200,
			"message": cc,
		})

	}

}

func Get_ticker(c *gin.Context) {
	mysql.Connect()
	var tick []float64
	price, volume := mysql.Get_ticker()
	tick = append(tick, price, volume)
	c.JSON(200, gin.H{
		"code": 200,
		"data": tick,
	})

}

type create struct {
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
	Side   string  `json:"side"`
}

//func Deal_orders(sleep
//float64) {
//var leftOrder mysql.Order
//buy, sell := mysql.Get_side_info()
//
//}
