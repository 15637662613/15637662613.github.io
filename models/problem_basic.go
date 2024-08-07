package models

import (
	"gorm.io/gorm"
)

type ProblemBasic struct {
	gorm.Model
	Identity          string             `gorm:"column:identity;type:varchar(36);" json:"identity"`   // 问题的唯一标识
	ProblemCategories []*ProblemCategory `gorm:"foreignKey:problem_id;references:id;"`                // 关联问题分类表
	Title             string             `gorm:"column:title;type:varchar(255);" json:"title"`        // 文章标题
	Content           string             `gorm:"column:content;type:text;" json:"content"`            // 问题描述
	MaxRuntime        int                `gorm:"column:max_runtime;type:int(11);" json:"max_runtime"` // 最大运行时间
	MaxMem            int                `gorm:"column:max_mem;type:int(11);" json:"max_mem"`         // 最大运行内存
	TestCase          []*TestCase        `gorm:"foreignKey:problem_identity;references:identity;"`    // 关联测试用例表
	PassNum           int64              `gorm:"column:pass_num;type:int(11);" json:"pass_num"`       // 通过问题的个数
	SubmitNum         int64              `gorm:"column:submit_num;type:int(11);" json:"submit_num"`   // 提交次数
}

// TableName
// 每当 GORM 操作 ProblemBasic 结构体时，它将会使用 problem_basic 作为表名，而不是默认的 problem_basics
func (table *ProblemBasic) TableName() string {
	return "problem_basic"
}

func GetProblemList(keyword, categoryIdentity string) *gorm.DB {
	tx := DB.Model(new(ProblemBasic)).
		Preload("ProblemCategories").
		Preload("ProblemCategories.CategoryBasic").
		Where("title like ? OR content like ?", "%"+keyword+"%", "%"+keyword+"%")
	if categoryIdentity != "" {
		tx.Joins("right join problem_category pc on pc.problem_id = problem_basic.id").
			Where("pc.category_id =(select cb.id from category_basic cb where cb.identity =?)", categoryIdentity)
	}
	return tx
}
