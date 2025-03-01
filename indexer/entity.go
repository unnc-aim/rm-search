package indexer

import (
	"fmt"
)

type EntityType string

const (
	EntityTypeBbsPost = "bbs-post"
)

type BaseEntity struct {
	Id           string   `json:"id"`            // 主键
	Type         string   `json:"type"`          // 类型
	Title        string   `json:"title"`         // 标题
	Content      string   `json:"content"`       // 内容
	Image        string   `json:"image"`         // 图片
	Url          string   `json:"url"`           // 链接
	Season       string   `json:"season"`        // 赛季
	CategoryLvl0 []string `json:"category_lvl0"` // 一级分类
	CategoryLvl1 []string `json:"category_lvl1"` // 二级分类
	CollegeName  []string `json:"college_name"`  // 学校名称
}

type IndexEntity struct {
	BaseEntity
	BbsPost *BbsPost `json:"bbs_post,omitempty"`
}

func GetEntityId(entityType EntityType, id any) string {
	return fmt.Sprintf("%s:%v", entityType, id)
}
