package models

import "gorm.io/gorm"

type CategoryBasic struct {
	gorm.Model
	Identity string `gorm:"column:identity;type:varchar(36);" json:"identity"` // 分类的唯一标识
	Name     string `gorm:"column:name;type:varchar(100);" json:"name"`        // 分类名
	ParentId string `gorm:"column:parent_id;type:int(11);" json:"parent_id"`   //
}

func (table *CategoryBasic) TableName() string {
	return "category_basic"
}
