package test

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"testing"
)

type UserClaims struct {
	Name     string `json:"username"`
	Identity string `json:"identity"`
	jwt.StandardClaims
}

var myKey = []byte("gin-gorm-OG-key")

// 生成token
func TestGenerateTokenTest(t *testing.T) {
	UserClaim := &UserClaims{
		Name:           "Get",
		Identity:       "user",
		StandardClaims: jwt.StandardClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaim)
	ts, err := token.SignedString(myKey)
	//eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6IkdldCIsImlkZW50aXR5IjoidXNlciJ9.YrmICO6PzNDI7ttAeMWoEcprJnn8b8dD9hIBU1kV2E4
	fmt.Printf("%v %v", ts, err)
}

// 解析token
func TestParseTokenTest(t *testing.T) {
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6IkdldCIsImlkZW50aXR5IjoidXNlciJ9.YrmICO6PzNDI7ttAeMWoEcprJnn8b8dD9hIBU1kV2E4"
	userClaim := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, userClaim, func(token *jwt.Token) (interface{}, error) {
		return myKey, nil
	})
	if err != nil {
		fmt.Println("err:", err)
	}
	if token.Valid {
		fmt.Println(userClaim)
	}
}
