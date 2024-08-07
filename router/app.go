package router

import (
	_ "gin-gorm-OJ/docs"
	"gin-gorm-OJ/middleware"
	"gin-gorm-OJ/service"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()

	//swag配置
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	//配置路由规则

	//公用方法
	//问题
	r.GET("/problem-list", service.GetProblemList)
	r.GET("/problem-detail", service.GetProblemDetail)
	//用户
	r.GET("/user-detail", service.GetUserDetail)
	r.POST("/login", service.Login)
	r.POST("/send-code", service.SendCode)
	r.POST("/register", service.Register)
	//用户排行榜
	r.GET("/rank-list", service.GetRankList)
	//提交
	r.GET("/submit-list", service.GetSubmitList)

	//私有方法
	//管理员私有方法
	//admin := r.Group("/admin", middleware.AuthAdmin())
	admin := r.Group("/admin")
	//问题创建
	admin.POST("/problem-create", service.ProblemCreate)
	//问题修改
	admin.PUT("/problem-modify", service.ProblemModify)
	//分类列表
	admin.GET("/category-list", service.GetCategoryList)
	//分类创建
	admin.POST("/category-create", service.CategoryCreate)
	//分类修改
	admin.PUT("/category-modify", service.CategoryModify)
	//分类删除
	admin.DELETE("/category-delete", service.CategoryDelete)

	//用户的私有方法
	user := r.Group("/user", middleware.AuthUserCheck())
	//代码提交
	user.POST("/submit", service.Submit)
	return r
}
