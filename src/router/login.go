package router

import (
	"awesomeProject/src/rlog"
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
		//if reqData.Username != "" || reqData.Password != "" {
		//	c.JSON(200, gin.H{"message": "Pwd Wrong Err"})
		//	return
		//}
		OTAOK, err := NewGoogleAuth().VerifyCode("X5PNFM56OQTU4NPXV2R3LIKYJTXST4JR", reqData.OTA)
		if err != nil || !OTAOK {
			rlog.Error(err)
			c.JSON(200, gin.H{"message": "AUTH Data Err"})
			return
		}
		token := generateToken(c, reqData.Username)

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
}

// 生成令牌
func generateToken(c *gin.Context, user string) (token string) {
	j := NewJWT()
	claims := CustomClaims{
		user,
		user,
		user,
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
