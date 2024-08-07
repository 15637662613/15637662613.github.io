package service

import (
	"encoding/json"
	"errors"
	"gin-gorm-OJ/define"
	"gin-gorm-OJ/helper"
	"gin-gorm-OJ/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
)

// GetProblemList
// @Tags 公共方法
// @Summary 问题列表
// @Param page query int false "page"
// @Param size query int false "size"
// @Param keyword query int false "keyword"
// @Param category_identity query string false "category_identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /problem-list [get]
func GetProblemList(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("Strconv Error : ", err)
		return
	}
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize)) // 一页显示几条数据
	page = (page - 1) * size                                            //计算机默认是从 0 开始
	var count int64

	keyword := c.Query("keyword")
	categoryIdentity := c.Query("category_identity")
	list := make([]*models.ProblemBasic, 0)
	tx := models.GetProblemList(keyword, categoryIdentity)
	err = tx.Count(&count).Omit("content").Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		log.Println("Get Problem List err:", err)
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  list,
			"count": count,
		},
	})
}

// GetProblemDetail
// @Tags 公共方法
// @Summary 问题详情
// @Param identity query string false "problem_identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /problem-detail [get]
func GetProblemDetail(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}

	data := new(models.ProblemBasic)
	err := models.DB.Where("identity=?", identity).
		Preload("ProblemCategories").
		Preload("ProblemCategories.CategoryBasic").
		First(&data).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "当前问题不存在",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get Problem Detail err:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": data,
	})
}

// ProblemCreate
// @Tags 管理员私有方法
// @Summary 问题创建
// @Param Authorization header string true "Authorization"
// @Param title formData string true "title"
// @Param content formData string true "content"
// @Param max_runtime formData int false "max_runtime"
// @Param max_mem formData int false "max_mem"
// @Param category_ids formData []string false "category_ids" collectionFormat(multi)
// @Param test_cases formData []string true "test_cases" collectionFormat(multi)
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/problem-create [post]
func ProblemCreate(c *gin.Context) {
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMem, _ := strconv.Atoi(c.PostForm("max_mem"))
	categoryIds := c.PostFormArray("category_ids")
	testCases := c.PostFormArray("test_cases")
	if len(title) == 0 || len(content) == 0 || len(testCases) == 0 || len(categoryIds) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "必填项不能为空",
		})
		return
	}
	identity := helper.GetUuid()
	data := &models.ProblemBasic{
		Identity:   identity,
		Title:      title,
		Content:    content,
		MaxRuntime: maxRuntime,
		MaxMem:     maxMem,
	}
	//处理分类
	problemCategories := make([]*models.ProblemCategory, 0)
	for _, id := range categoryIds {
		categoryId, _ := strconv.Atoi(id)
		problemCategories = append(problemCategories, &models.ProblemCategory{
			ProblemId:  data.ID,
			CategoryId: uint(categoryId),
		})
	}
	data.ProblemCategories = problemCategories
	//处理测试案例
	testCaseBasics := make([]*models.TestCase, 0)
	for _, testCase := range testCases {
		//案例：{"input":"1 2\n","output":"3\n"}
		//caseMap接收参数
		caseMap := make(map[string]string)
		err := json.Unmarshal([]byte(testCase), &caseMap)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "测试用例格式Error",
			})
			return
		}
		if _, ok := caseMap["input"]; !ok {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "测试用例格式Error",
			})
			return
		}
		if _, ok := caseMap["input"]; !ok {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "测试用例格式Error",
			})
			return
		}
		testCaseBasics = append(testCaseBasics, &models.TestCase{
			Identity:        helper.GetUuid(),
			ProblemIdentity: identity,
			Input:           caseMap["input"],
			Output:          caseMap["output"],
		})

	}
	data.TestCase = testCaseBasics
	err := models.DB.Create(&data).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "创建失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "问题创建成功",
	})
}

// ProblemModify
// @Tags 管理员私有方法
// @Summary 问题修改
// @Param Authorization header string true "Authorization"
// @Param identity formData string true "identity"
// @Param title formData string true "title"
// @Param content formData string true "content"
// @Param max_runtime formData int false "max_runtime"
// @Param max_mem formData int false "max_mem"
// @Param category_ids formData []string false "category_ids" collectionFormat(multi)
// @Param test_cases formData []string true "test_cases" collectionFormat(multi)
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/problem-modify [put]
func ProblemModify(c *gin.Context) {
	identity := c.PostForm("identity")
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMem, _ := strconv.Atoi(c.PostForm("max_mem"))
	categoryIds := c.PostFormArray("category_ids")
	testCases := c.PostFormArray("test_cases")
	if identity == "" || len(title) == 0 || len(content) == 0 || len(testCases) == 0 || len(categoryIds) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "必填项不能为空",
		})
		return
	}
	//使用事务确保修改部分满足ACID
	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		//问题基础信息更新
		problemBasic := &models.ProblemBasic{
			Identity:   identity,
			Title:      title,
			Content:    content,
			MaxRuntime: maxRuntime,
			MaxMem:     maxMem,
		}
		err := tx.Where("identity=?", identity).Updates(&problemBasic).Error
		if err != nil {
			return err
		}
		// 查询问题详情，方便关联表的更新
		err = tx.Where("identity=?", identity).Find(problemBasic).Error
		if err != nil {
			return err
		}
		//关联问题分类的更新（更新problem_category表）
		// 1删除存在的关联关系
		err = tx.Where("problem_id=?", problemBasic.ID).Delete(&models.ProblemCategory{}).Error
		if err != nil {
			return err
		}
		// 2新增新的关联关系
		pcs := make([]*models.ProblemCategory, 0)
		for _, id := range categoryIds {
			intID, _ := strconv.Atoi(id)
			pcs = append(pcs, &models.ProblemCategory{
				ProblemId:  problemBasic.ID,
				CategoryId: uint(intID),
			})
		}
		err = tx.Create(&pcs).Error
		if err != nil {
			return err
		}
		//测试案例的更新
		//1删除已存在的测试案例
		err = tx.Where("problem_identity=?", identity).Delete(&models.TestCase{}).Error
		if err != nil {
			return err
		}
		//2添加新的测试案例
		tcs := make([]*models.TestCase, 0)
		for i := range testCases {
			//{"input":"1 2\n","output":"3\n"}
			caseMap := make(map[string]string)
			err = json.Unmarshal([]byte(testCases[i]), &caseMap)
			if err != nil {
				return err
			}
			if _, ok := caseMap["input"]; !ok {
				c.JSON(http.StatusOK, gin.H{
					"code": -1,
					"msg":  "测试案例input格式错误",
				})
				return err
			}
			if _, ok := caseMap["output"]; !ok {
				c.JSON(http.StatusOK, gin.H{
					"code": -1,
					"msg":  "测试案例output格式错误",
				})
				return err
			}
			tcs = append(tcs, &models.TestCase{
				Identity:        helper.GetUuid(),
				ProblemIdentity: identity,
				Input:           caseMap["input"],
				Output:          caseMap["output"],
			})
		}
		err = tx.Create(&tcs).Error
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Problem Modify err:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "修改成功",
	})
}
