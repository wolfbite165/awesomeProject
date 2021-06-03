package router

import (
	"awesomeProject/src/mysql"
	"awesomeProject/src/rlog"
	"crypto/md5"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func authRouter(apiR *gin.RouterGroup) {

	apiR.POST("/login", func(c *gin.Context) {
		var reqData LoginReq
		err := c.BindJSON(&reqData)
		if err != nil {
			rlog.Error(err)
			c.JSON(400, gin.H{"message": "Post Data Err"})
			return
		}
		var pwdMd5 string
		{
			data := []byte(reqData.Password)
			has := md5.Sum(data)
			pwdMd5 = fmt.Sprintf("%x", has)
		}

		name, pwd, gsk, id, err := mysql.Get_account_info(reqData.Username)
		if err != nil {
			rlog.Error(err)
			c.JSON(200, gin.H{"message": "username wrong"})
			return
		}
		if reqData.Username != name || pwdMd5 != pwd {
			c.JSON(200, gin.H{"message": "Pwd Wrong Err"})
			return
		}
		OTAOK, err := NewGoogleAuth().VerifyCode(gsk, reqData.OTA)
		if err != nil || !OTAOK {
			rlog.Error(err)
			c.JSON(200, gin.H{"message": "AUTH Data Err"})
			return
		}
		token := generateToken(c, id, reqData.Username)

		data := LoginResult{
			User:  reqData.Username,
			Token: token,
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "登录成功！",
			"data": data,
		})
	})

	apiR.POST("/register", func(c *gin.Context) {

		Account := c.Query("account")
		Password := c.Query("password")
		mysql.Connect()
		a := mysql.Check_same_account(Account)
		if !a {
			googleSk := NewGoogleAuth().GetSecret()
			mysql.Write_account(Account, Password, googleSk)
			c.JSON(200, gin.H{
				"code":    200,
				"message": "success",
				"data":    googleSk,
			})
		} else {
			c.JSON(200, gin.H{
				"code":    1003,
				"message": "already have this account",
			})

		}
	})

}

// 生成令牌
func generateToken(c *gin.Context, userId int, userName string) (token string) {
	j := NewJWT()
	claims := CustomClaims{
		userId,
		userName,
		jwt.StandardClaims{
			NotBefore: int64(time.Now().Unix() - 1000), // 签名生效时间
			ExpiresAt: int64(time.Now().Unix() + 3600), // 过期时间 一小时
			Issuer:    "OhMyUniswaper",                 //签名的发行者
		},
	}

	token, err := j.CreateToken(claims)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 10000,
			"msg":  err.Error(),
		})
		return
	}

	return
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	OTA      string `json:"ota"`
}
type LoginResult struct {
	User  string `json:"user"`
	Token string `json:"token"`
}
