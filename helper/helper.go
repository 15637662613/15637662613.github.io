package helper

import (
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/jordan-wright/email"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"net/smtp"
	"os"
	"strconv"
	"time"
)

var myKey = []byte("gin-gorm-OG-key")

type UserClaims struct {
	Name     string `json:"username"`
	Identity string `json:"identity"`
	IsAdmin  int    `json:"is_admin"`
	jwt.StandardClaims
}

// GetMd5
// 生成md5
func GetMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s))) // 字符串 s 转换为字节切片,转为16进制
}

// GenerateToken
// 生成token
func GenerateToken(identity, name string, isAdmin int) (string, error) {
	UserClaim := &UserClaims{
		Name:           name,
		Identity:       identity,
		IsAdmin:        isAdmin,
		StandardClaims: jwt.StandardClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaim)
	ts, err := token.SignedString(myKey) // 将token转换为string
	if err != nil {
		return "", err
	}
	return ts, nil
}

// ParseToken
// 解析token
func ParseToken(tokenString string) (*UserClaims, error) {
	userClaim := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, userClaim, func(token *jwt.Token) (interface{}, error) {
		return myKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("token Is Invalid: %v", err)
	}
	return userClaim, nil
}

// SendCode
// 发送验证码
func SendCode(toUserEmail string, code string) error { // code是随机生成的6为数字
	e := email.NewEmail()
	e.From = "Cz <15637662613@163.com>"
	e.To = []string{toUserEmail}
	e.Subject = "验证码已发送，请查收"
	e.HTML = []byte("您的验证码是：<b>" + code + "</b>")
	//返回EOF时，关闭SSL重试
	return e.SendWithTLS("smtp.163.com:465", smtp.PlainAuth("", "15637662613@163.com", "OOYTFVQUSZNBVFVM", "smtp.163.com"),
		&tls.Config{InsecureSkipVerify: true, ServerName: "smtp.163.com"})
}

// GetUuid
// 生成唯一码
func GetUuid() string {
	return uuid.NewV4().String()
}

// GetRandom
// 生成验证码，随机生成6位验证码
func GetRandom() string {
	rand.Seed(time.Now().UnixNano())
	s := ""
	for i := 0; i < 6; i++ {
		s += strconv.Itoa(rand.Intn(10)) // 将int转为string
	}
	return s
}

// CodeSave
// 代码保存
func CodeSave(code []byte) (string, error) {
	dirName := "code/" + GetUuid() // 代码存放在code下，目录name是随机生成的uuid
	path := dirName + "/main.go"   // 代码的path，main里面放的是用户写的代码
	err := os.Mkdir(dirName, 0777) // 创建目录
	if err != nil {
		return "", err
	}
	file, err := os.Create(path) // 创建文件
	if err != nil {
		return "", err
	}
	file.Write(code) //写入内容
	defer file.Close()
	return path, nil
}
