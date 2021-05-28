package main

import (
	"awesomeProject/src/mysql"
	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()
	router.GET("/get_account", get_account)
	//router.POST("/sign_account", sign)
	//router.POST("/deposit", doposit)
	//order := router.Group("/order")
	//order.POST("/create_order", create_order)
	//order.POST("/cancel_order", cancel)
	router.Run()
}

func get_account(c *gin.Context) {
	Account := c.Query("account")

	mysql.Connect()
	a := mysql.Check_same_account(Account)
	if a == true {
		b := mysql.Checkfile(Account)
		//Money := mysql.Checkfile(Account).normal_Money
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
