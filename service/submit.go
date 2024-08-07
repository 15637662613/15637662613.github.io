package service

import (
	"bytes"
	"errors"
	"fmt"
	"gin-gorm-OJ/define"
	"gin-gorm-OJ/helper"
	"gin-gorm-OJ/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// GetSubmitList
// @Tags 公共方法
// @Summary 提交列表
// @Param page query int false "page"
// @Param size query int false "size"
// @Param problem_identity query string false "problem_identity"
// @Param user_identity query string false "user_identity"
// @Param status query int false "status"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /submit-list [get]
func GetSubmitList(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("Strconv Error : ", err)
		return
	}
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize)) // 一页显示几条数据
	page = (page - 1) * size                                            //计算机默认是从 0 开始
	var count int64
	list := make([]models.SubmitBasic, 0)
	problemIdentity := c.Query("problem_identity")
	userIdentity := c.Query("user_identity")
	status, _ := strconv.Atoi(c.Query("status"))
	tx := models.GetSubmitList(problemIdentity, userIdentity, status)
	err = tx.Count(&count).Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		log.Println("GetSubmitList Error : ", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "GetSubmitList Errors" + err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  list,
			"count": count,
		},
	})
}

// Submit
// @Tags 用户私有方法
// @Summary 代码提交
// @Param Authorization header string true "Authorization"
// @Param problem_identity query string true "problem_identity"
// @Param code body string true "code"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /user/submit [post]
func Submit(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	code, err := ioutil.ReadAll(c.Request.Body) // 读取body的内容
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "ReadAll Errors",
		})
		return
	}
	//代码保存
	path, err := helper.CodeSave(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "CodeSave Errors",
		})
		return
	}
	//提交
	u, _ := c.Get("user")
	userClaim := u.(*helper.UserClaims) //获取存储的值，并将其转换为 *helper.UserClaims类型
	sb := &models.SubmitBasic{
		Identity:        helper.GetUuid(),
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClaim.Identity, //userClaim是通过token解析得到的，包含用户的部分信息
		Path:            path,
	}
	//代码的判断
	pb := new(models.ProblemBasic)
	err = models.DB.Where("identity = ?", problemIdentity).Preload("TestCase").First(pb).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Problem Get Errors",
		})
		return
	}

	WA := make(chan int)  // 答案错误
	OOM := make(chan int) // 超内存
	CE := make(chan int)  // 编译错误
	passCount := 0        //通过测试案例的个数
	var lock sync.Mutex
	var msg string

	for _, testCase := range pb.TestCase {
		testCase := testCase
		go func() { // 每个测试案例的验证，同时进行
			//执行测试
			cmd := exec.Command("go", "run", path) //指向上诉路径的代码(用户的代码)
			var out, stdErr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stdErr
			//根据测试的输入案例，拿到输出结果和标准的输出结果作比对
			stdinPipe, err := cmd.StdinPipe()
			if err != nil {
				fmt.Println(err)
				return
			}
			io.WriteString(stdinPipe, testCase.Input)

			var bm runtime.MemStats
			runtime.ReadMemStats(&bm) // 运行前内存使用情况
			if err := cmd.Run(); err != nil {
				fmt.Println(err, stdErr.String())
				if err.Error() == "exit status 2" { // 编译错误
					CE <- 1
					msg = stdErr.String()
					return
				}
			}
			var em runtime.MemStats
			runtime.ReadMemStats(&em) // 运行后内存使用情况
			//答案错误
			if testCase.Output != out.String() {
				msg = "答案错误"
				WA <- 1
				return
			}
			//运行超内存
			if em.Alloc/1024-(bm.Alloc/1024) > uint64(pb.MaxMem) { // 换算为kb单位做计算
				msg = "运行超内存"
				OOM <- 1
				return
			}
			lock.Lock()
			passCount++
			lock.Unlock()
		}()
	}
	select {
	//-1-待判断，1-答案正确，2-答案错误，3-运行超时，4-运行超内存
	case <-WA:
		sb.Status = 2
	case <-OOM:
		sb.Status = 4
	case <-CE:
		sb.Status = 5
		//使用 time.After 函数创建一个定时器，在 pb.MaxRuntime 毫秒之后触发
	case <-time.After(time.Millisecond * time.Duration(pb.MaxRuntime)):
		if passCount == len(pb.TestCase) {
			msg = "成功"
			sb.Status = 1
		} else {
			msg = "错误"
			sb.Status = 3
		}
	}
	if err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 保存提交记录
		err = tx.Create(sb).Error
		if err != nil {
			return errors.New("Submit Record Save Errors" + err.Error())
		}
		// 更新项
		m := make(map[string]interface{})
		m["submit_num"] = gorm.Expr("submit_num + ?", 1) // 将 submit_num 字段的值增加 1
		if sb.Status == 1 {
			m["pass_num"] = gorm.Expr("pass_num + ?", 1)
		}
		//更新用户提交记录
		err = tx.Model(new(models.UserBasic)).Where("identity = ?", userClaim.Identity).Updates(m).Error
		if err != nil {
			return errors.New("UserBasic Modify Errors" + err.Error())
		}
		//更新问题提交记录
		err = tx.Model(new(models.ProblemBasic)).Where("identity = ?", problemIdentity).Updates(m).Error
		if err != nil {
			return errors.New("ProblemBasic Modify Errors" + err.Error())
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Submit Errors",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"status": sb.Status,
			"msg":    msg,
		},
	})
}
