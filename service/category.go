package service

import (
	"gin-gorm-OJ/define"
	"gin-gorm-OJ/helper"
	"gin-gorm-OJ/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

// GetCategoryList
// @Tags 管理员私有方法
// @Summary 分类列表
// @Param Authorization header string true "Authorization"
// @Param page query int false "page"
// @Param size query int false "size"
// @Param keyword query int false "keyword"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/category-list [get]
func GetCategoryList(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("Strconv Error : ", err)
		return
	}
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize)) // 一页显示几条数据
	page = (page - 1) * size                                            //计算机默认是从 0 开始
	var count int64

	keyword := c.Query("keyword")
	list := make([]*models.CategoryBasic, 0)
	err = models.DB.Model(new(models.CategoryBasic)).Where("name like ?", "%"+keyword+"%").
		Count(&count).Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		log.Println("CategoryList Error : ", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取分类列表失败",
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

// CategoryCreate
// @Tags 管理员私有方法
// @Summary 分类创建
// @Param Authorization header string true "Authorization"
// @Param name formData string true "name"
// @Param parentId formData string false "parentId"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/category-create [post]
func CategoryCreate(c *gin.Context) {
	name := c.PostForm("name")
	parentId, _ := strconv.Atoi(c.PostForm("parent_id"))
	category := &models.CategoryBasic{
		Identity: helper.GetUuid(),
		Name:     name,
		ParentId: strconv.Itoa(parentId),
	}
	err := models.DB.Create(category).Error
	if err != nil {
		log.Println("CategoryCreate Error : ", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "CategoryCreate Error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "CategoryCreate success",
	})
}

// CategoryModify
// @Tags 管理员私有方法
// @Summary 分类修改
// @Param Authorization header string true "Authorization"
// @Param identity formData string true "identity"
// @Param name formData string true "name"
// @Param parentId formData string false "parentId"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/category-modify [put]
func CategoryModify(c *gin.Context) {
	name := c.PostForm("name")
	identity := c.PostForm("identity")
	parentId := c.PostForm("parentId")
	if name == "" || identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	category := &models.CategoryBasic{
		Identity: identity,
		Name:     name,
		ParentId: parentId,
	}
	err := models.DB.Model(new(models.CategoryBasic)).Where("identity=?", identity).Updates(category).Error
	if err != nil {
		log.Println("CategoryModify Error : ", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "CategoryModify Error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "CategoryModify success",
	})
}

// CategoryDelete
// @Tags 管理员私有方法
// @Summary 分类删除
// @Param Authorization header string true "Authorization"
// @Param identity query string true "identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/category-delete [delete]
func CategoryDelete(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	var count int64
	err := models.DB.Model(new(models.ProblemCategory)).
		Where("category_id = (SELECT id FROM category_basic  WHERE identity = ? LIMIT 1)", identity).
		Count(&count).Error
	if err != nil {
		log.Println("CategoryDelete Error : ", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取分类关联的问题失败",
		})
		return
	}
	if count > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "该分类下面已存在问题，不可删除",
		})
		return
	}
	err = models.DB.Model(new(models.CategoryBasic)).Where("identity=?", identity).Delete(&models.CategoryBasic{}).Error
	if err != nil {
		log.Println("CategoryDelete Error : ", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "删除失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": -1,
		"msg":  "成功删除",
	})
	return
}
