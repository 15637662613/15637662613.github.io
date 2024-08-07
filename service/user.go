package service

import (
	"errors"
	"gin-gorm-OJ/define"
	"gin-gorm-OJ/helper"
	"gin-gorm-OJ/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"time"
)

// GetUserDetail
// @Tags 公共方法
// @Summary 用户详情
// @Param identity query string false "user_identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /user-detail [get]
func GetUserDetail(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "用户唯一标识不能为空",
		})
		return
	}
	data := new(models.UserBasic)
	err := models.DB.Omit("password").Where("identity=?", identity).Find(&data).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Get User Detail By Identity" + identity + "Error" + err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code": -1,
		"data": data,
	})
}

// Login
// @Tags 公共方法
// @Summary 用户登录
// @Param username formData string false "username"
// @Param password formData string false "password"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /login [post]
func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "必填信息为空",
		})
		return
	}
	password = helper.GetMd5(password) // 将用户注册的密码加密再放入数据库
	print(password)

	data := new(models.UserBasic)
	err := models.DB.Where("name=? AND password=?", username, password).First(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "用户名或密码错误",
			})
			return
		}
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Get user Basic error:" + err.Error(),
		})
		return
	}
	token, err := helper.GenerateToken(data.Identity, data.Name, data.IsAdmin)
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "GenerateToken error:" + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"data": gin.H{
			"token": token,
		},
	})
}

// SendCode
// @Tags 公共方法
// @Summary 发送验证码
// @Param email formData string true "email"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /send-code [post]
func SendCode(c *gin.Context) {
	email := c.PostForm("email")
	if len(email) == 0 {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
	}
	code := helper.GetRandom()
	models.RDB.Set(c, email, code, time.Second*300) // 缓存到redis

	err := helper.SendCode(email, code)
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "验证码发送error：",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "验证码发送成功!",
	})
}

// Register
// @Tags 公共方法
// @Summary 用户注册
// @Param name formData string true "name"
// @Param code formData string true "code"
// @Param password formData string true "password"
// @Param mail formData string true "mail"
// @Param phone formData string false "phone"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /register [post]
func Register(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")
	mail := c.PostForm("mail")
	phone := c.PostForm("phone")
	userCode := c.PostForm("code") // 用户传递过来的验证码
	if len(name) == 0 || len(password) == 0 || len(mail) == 0 || len(userCode) == 0 {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "必填项不能为空",
		})
		return
	}
	//验证code是否正确
	sysCode, err := models.RDB.Get(c, mail).Result() // 系统生成的验证码
	if err != nil {
		log.Println("Get Code Error " + err.Error())
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "验证码错误,请从新获取",
		})
		return
	}
	if sysCode != userCode {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "验证码不正确	",
		})
		return
	}
	//判断邮箱是否已存在
	var cnt int64
	err = models.DB.Where("mail=?", mail).Model(new(models.UserBasic)).Count(&cnt).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "获取用户信息错误" + err.Error(),
		})
		return
	}
	if cnt > 0 {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "该邮箱已被注册",
		})
		return
	}
	//插入数据。需要将密码转换成md5再存入库中
	userIdentity := helper.GetUuid()
	data := &models.UserBasic{
		Identity: userIdentity,
		Name:     name,
		Password: helper.GetMd5(password),
		Mail:     mail,
		Phone:    phone,
	}
	err = models.DB.Create(data).Error
	if err != nil {
		log.Println("Create User Error:" + err.Error())
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Create User Error:",
		})
		return
	}

	//生成token
	token, err := helper.GenerateToken(userIdentity, name, data.IsAdmin)
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Generate Token Error:" + err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"data": gin.H{
			"token": token,
		},
	})
}

// GetRankList
// @Tags 公共方法
// @Summary 用户排行榜
// @Param page query int false "page"
// @Param size query int false "size"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /rank-list [get]
func GetRankList(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("Strconv Error : ", err)
		return
	}
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	page = (page - 1) * size
	var count int64
	list := make([]models.UserBasic, 0)
	err = models.DB.Model(new(models.UserBasic)).Count(&count).Order("pass_num DESC,submit_num ASC").
		Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Get Rank List Error:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  list,
			"count": count,
		},
	})
}
