package router

import (
	"awesomeProject/src/mysql"
	"awesomeProject/src/rlog"
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
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
}
type Trade_out struct {
	sell Out
	buy  Out
}

func InitRouter() {
	mysql.Connect()
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
		commonR.GET("/kline", Get_kline)
	}
	{
		order := apiR.Group("/order")
		order.Use(JWTAuth())
		order.POST("/create_order", create_order)
		order.POST("/get_open_order", Get_oppen_order)
		order.POST("/cancel_order", Cancel_order)
	}
	go mysql.Write_kline()
	router.Run()
}

func get_account(c *gin.Context) {

	token := c.Request.Header.Get("token")
	j := NewJWT()
	claims, err := j.ParseToken(token)
	if err != nil {
		return
	}
	Account := claims.Name
	//fmt.Println(Account)
	mysql.Connect()
	a := mysql.Check_same_account(Account)
	if a == true {
		var b wallet
		d := mysql.Checkfile(Account)
		b.Locked_money = d.Lock_money
		b.Nomal_coin = d.Normal_Coin
		b.Locked_coin = d.Lock_coin
		b.Nomal_money = d.Normal_Money
		//fmt.Println(b.Locked_coin, b.Locked_money, b.Nomal_money, b.Nomal_coin)
		//fmt.Println(b)

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
	//fmt.Println(coin, money)

	a := mysql.Check_same_account(Account)
	//fmt.Println(a)
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
	if err != nil {
		rlog.Error(err)
		c.JSON(400, gin.H{"message": "Post Data Err"})
		return
	}
	token := c.Request.Header.Get("token")
	j := NewJWT()
	claims, err := j.ParseToken(token)
	if err != nil {
		return
	}
	side := orderdata.Side
	price := orderdata.Price
	volume := orderdata.Volume
	account := claims.Name
	info := mysql.Checkfile(account)
	buy, sell := mysql.Get_side_info()
	use := price * volume
	fmt.Println("这次下单结果", side, price, volume, account)
	//fmt.Println("price=?", price)
	//fmt.Println("volume=?", volume)
	//fmt.Println("use=?", use)
	//fmt.Println("normal=?", info.Normal_Money)
	//fmt.Println(info.Lock_money)
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
							"code":     200,
							"message":  "success",
							"order_id": id,
						})
						return
					}
					buyindex := 0
					id := mysql.Create_order(account, price, volume, side, "online", t, info.Id, volume)
					mysql.Write_info(account, info.Normal_Money, info.Normal_Coin-volume, info.Lock_money, info.Lock_coin+volume)
					for {
						info = mysql.Checkfile(account)
						//fmt.Println(len(buy))
						//fmt.Println(buyindex)
						//fmt.Println("无敌",price,buy[buyindex].Price)
						info2 := mysql.Checkfile(buy[buyindex].Account)
						if price <= buy[buyindex].Price {

							if volume < buy[buyindex].Left {

								c.JSON(200, gin.H{
									"code":     200,
									"message":  "success",
									"order_id": id,
								})
								mysql.Write_trade_info(buy[buyindex].Price, volume, side, t, buy[buyindex].Account, account)
								mysql.Write_info(buy[buyindex].Account, info2.Normal_Money, info2.Normal_Coin+volume, info2.Lock_money-buy[buyindex].Left*buy[buyindex].Price, info2.Lock_coin)
								info = mysql.Checkfile(account) //买家(挂单)的余额修改
								fmt.Println("1")
								mysql.Write_info(account, info.Normal_Money+buy[buyindex].Left*buy[buyindex].Price, info.Normal_Coin-volume, info.Lock_money, info.Lock_coin) //卖家(下单)余额修改
								mysql.Write_order_info(buy[buyindex].Id, "online", buy[buyindex].Left-volume)
								mysql.Write_order_info(id, "dealed", volume-volume)
								break
							}
							if volume == buy[buyindex].Left {
								//id := mysql.Create_order(account, price, volume, side, "dealed", t, info.Id)
								c.JSON(200, gin.H{
									"code":     200,
									"message":  "success",
									"order_id": id,
								})
								mysql.Write_trade_info(buy[buyindex].Price, volume, side, t, buy[buyindex].Account, account)
								//fmt.Println(info2.Lock_money,buy[buyindex].Left)
								//fmt.Println("买家",buy[buyindex].Account, info2.Normal_Money, info2.Normal_Coin+volume, info2.Lock_money-buy[buyindex].Left*buy[buyindex].Price, info2.Lock_coin)

								mysql.Write_info(buy[buyindex].Account, info2.Normal_Money, info2.Normal_Coin+volume, info2.Lock_money-buy[buyindex].Left*buy[buyindex].Price, info2.Lock_coin) //买家的余额修改
								info = mysql.Checkfile(account)
								//fmt.Println("卖家",account, info.Normal_Money+buy[buyindex].Left*buy[buyindex].Price, info.Normal_Coin-volume, info.Lock_money, info.Lock_coin)
								mysql.Write_info(account, info.Normal_Money+buy[buyindex].Left*buy[buyindex].Price, info.Normal_Coin-volume, info.Lock_money, info.Lock_coin) //卖家余额修改
								mysql.Write_order_info(buy[buyindex].Id, "dealed", buy[buyindex].Left-volume)
								mysql.Write_order_info(id, "dealed", volume-volume)
								break

							}
							if volume > buy[buyindex].Left {

								mysql.Write_order_info(id, "online", volume-buy[buyindex].Left)
								volume = volume - buy[buyindex].Left
								//info2 = mysql.Checkfile(sell[sellindex].Account)
								mysql.Write_info(buy[buyindex].Account,
									info2.Normal_Money,
									info2.Normal_Coin+buy[buyindex].Left, info2.Lock_money-buy[buyindex].Price*buy[buyindex].Left,
									info2.Lock_coin)
								info = mysql.Checkfile(account)
								mysql.Write_info(account,
									info.Normal_Money+buy[buyindex].Price*buy[buyindex].Left,
									info.Normal_Coin-buy[buyindex].Left, info.Lock_money,
									info.Lock_coin)
								t2 := time.Now().Unix()
								mysql.Write_trade_info(buy[buyindex].Price, buy[buyindex].Left, side, t2, buy[buyindex].Account, account)
								mysql.Write_order_info(buy[buyindex].Id, "dealed", 0)
								mysql.Write_order_info(id, "online", volume)

								buyindex++

							}

						} else {
							//mysql.Write_info(account, info.Normal_Money, info.Normal_Coin-volume, info.Lock_money, info.Lock_coin+volume)
							//id := mysql.Create_order(account, price, volume, side, "online", t, info.Id)
							c.JSON(200, gin.H{
								"code":     200,
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
						c.JSON(200, gin.H{
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
							if volume < sell[sellindex].Left {
								//id := mysql.Create_order(account, price, volume, side, "dealed", t, info.Id, volume)
								c.JSON(200, gin.H{
									"code":     200,
									"message":  "success",
									"order_id": id,
								})
								mysql.Write_trade_info(sell[sellindex].Price, volume, side, t, account, sell[sellindex].Account)
								mysql.Write_info(account, info.Normal_Money-sell[sellindex].Left*sell[sellindex].Price, info.Normal_Coin+volume, info.Lock_money, info.Lock_coin)
								info = mysql.Checkfile(account) //买家的余额修改
								info2 = mysql.Checkfile(sell[sellindex].Account)
								mysql.Write_info(sell[sellindex].Account, info2.Normal_Money+sell[sellindex].Left*sell[sellindex].Price, info2.Normal_Coin, info2.Lock_money, info2.Lock_coin-volume) //卖家余额修改
								mysql.Write_order_info(sell[sellindex].Id, "online", sell[sellindex].Left-volume)
								mysql.Write_order_info(id, "dealed", volume-volume)
								//if info2.Lock_coin-volume<0{
								//	//fmt.Println("这里出错了1")
								//}
								break
							}
							if volume == sell[sellindex].Left {
								//id := mysql.Create_order(account, price, volume, side, "dealed", t, info.Id, volume)
								c.JSON(200, gin.H{
									"code":     200,
									"message":  "success",
									"order_id": id,
								})
								info = mysql.Checkfile(account) //买家的余额修改
								mysql.Write_trade_info(sell[sellindex].Price, volume, side, t, account, sell[sellindex].Account)
								mysql.Write_info(account, info.Normal_Money-sell[sellindex].Left*sell[sellindex].Price, info.Normal_Coin+volume, info.Lock_money, info.Lock_coin)

								info2 = mysql.Checkfile(sell[sellindex].Account)
								mysql.Write_info(sell[sellindex].Account, info2.Normal_Money+sell[sellindex].Price*volume, info2.Normal_Coin, info2.Lock_money, info2.Lock_coin-volume) //卖家余额修改
								fmt.Println("出错点的两个值", info2.Lock_money, volume)

								mysql.Write_order_info(sell[sellindex].Id, "dealed", sell[sellindex].Left-volume)

								mysql.Write_order_info(id, "dealed", volume-volume)
								break

							}
							if volume > sell[sellindex].Left {

								volume -= sell[sellindex].Left
								mysql.Write_info(sell[sellindex].Account,
									info2.Normal_Money+sell[sellindex].Price*sell[sellindex].Left,
									info2.Normal_Coin, info.Lock_money,
									info2.Lock_coin-sell[sellindex].Left)
								//if info2.Lock_coin-sell[sellindex].Left<0{
								//	fmt.Println("这里出错了4")
								//}
								info = mysql.Checkfile(account)
								info2 = mysql.Checkfile(sell[sellindex].Account)
								mysql.Write_info(account,
									info.Normal_Money-sell[sellindex].Price*sell[sellindex].Left,
									info.Normal_Coin+sell[sellindex].Left, info.Lock_money,
									info.Lock_coin)
								//if info.Lock_coin<0{
								//	fmt.Println("这里出错了5")
								//}
								info = mysql.Checkfile(account)
								info2 = mysql.Checkfile(sell[sellindex].Account)
								//fmt.Println("大雨符号", account,
								//	info.Normal_Money-sell[sellindex].Price*sell[sellindex].Left,
								//	info.Normal_Coin+sell[sellindex].Left, info.Lock_money,
								//	info.Lock_coin)
								t2 := time.Now().Unix()
								mysql.Write_trade_info(sell[sellindex].Price, sell[sellindex].Left, side, t2, account, sell[sellindex].Account)
								mysql.Write_order_info(sell[sellindex].Id, "dealed", 0)
								mysql.Write_order_info(id, "online", volume)
								sellindex++

							}

						} else {
							mysql.Write_info(account, info.Normal_Money-use, info.Normal_Coin, info.Lock_money+use, info.Lock_coin)
							//id := mysql.Create_order(account, price, volume, side, "online", t, info.Id, volume)
							c.JSON(200, gin.H{
								"code":     200,
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
	token := c.Request.Header.Get("token")
	j := NewJWT()
	claims, err := j.ParseToken(token)
	if err != nil {
		return
	}
	account := claims.Name
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
	var Iddata cancel
	err := c.BindJSON(&Iddata)
	if err != nil {
		rlog.Error(err)
		c.JSON(400, gin.H{"message": "Post Data Err"})
		return
	}
	token := c.Request.Header.Get("token")
	j := NewJWT()
	claims, err := j.ParseToken(token)
	account := claims.Name
	//account := c.Query("account")
	id := Iddata.Id
	s := mysql.Check_order_info(account, id)
	d := mysql.Checkfile(account)
	a := mysql.Check_same_account(account)
	//fmt.Println(s.Status)
	use := s.Price * s.Left
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

			mysql.Write_info(account, d.Normal_Money, d.Normal_Coin+s.Left, d.Lock_money, d.Lock_coin-s.Left)
		} else {
			mysql.Write_info(account, d.Normal_Money+use, d.Normal_Coin, d.Lock_money-use, d.Lock_coin)
		}

	}
}

func Get_depth(c *gin.Context) {
	var buy1 []Out
	var sell1 []Out
	var dep depth
	mysql.Connect()
	buy, sell := mysql.Get_side_info()
	println(len(buy))

	for i := 0; i < len(buy); i++ {
		buy1 = append(buy1, Out{buy[i].Price, buy[i].Volume})
		//fmt.Printf("%+v", buy1)

	}
	for i := 0; i < len(sell); i++ {
		sell1 = append(sell1, Out{sell[i].Price, sell[i].Volume})
	}
	//fmt.Println(buy1, sell1)
	dep.Bids = buy1
	dep.Asks = sell1
	//fmt.Println(dep)
	//out = append(out, buy1, sell1)
	//fmt.Printf("%+v", out)
	//ss:=
	c.JSON(200, gin.H{
		"code": 200,
		"data": dep,
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
	cc, _ := mysql.Get_ticker()
	//println(price,volume)

	c.JSON(200, gin.H{
		"code": 200,
		"data": cc,
	})

}

func Get_kline(c *gin.Context) {
	mysql.Connect()
	size, _ := strconv.ParseInt(c.DefaultQuery("size", "1000"), 0, 64)
	form := c.Query("form")
	back := mysql.Get_kline(form, size)
	c.JSON(200, gin.H{
		"code": 200,
		"data": back,
	})
}

type create struct {
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
	Side   string  `json:"side"`
}
type cancel struct {
	Id int64 `json:"id"`
}
type depth struct {
	Bids []Out `json:"bids"`
	Asks []Out `json:"asks"`
}

//func Deal_orders(sleep
//float64) {
//var leftOrder mysql.Order
//buy, sell := mysql.Get_side_info()
//
//}
