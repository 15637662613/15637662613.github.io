package middleware

import (
	"gin-gorm-OJ/helper"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AuthUserCheck
// 验证用户身份
func AuthUserCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		userClaim, err := helper.ParseToken(auth) // 解析token
		if err != nil {
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Unauthorized Authorization",
			})
			return
		}
		if userClaim == nil { //确定用户身份是user
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Unauthorized Admin",
			})
			return
		}
		c.Set("user", userClaim) // 这里的userClaim是UserClaim类型，Get后依然需要转换数据类型
		c.Next()
	}
}
