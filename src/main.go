package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()
	router.GET("/get_account", get_account)
	router.POST("/")

}
