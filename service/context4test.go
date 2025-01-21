package service

import (
	"github.com/scutrobotlab/bbs-search/database/query"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Option func(c *Context)

func NewContextForTest(options ...Option) *Context {
	c := &Context{}
	for _, option := range options {
		option(c)
	}
	return c
}

func WithDb() Option {
	return func(c *Context) {
		const DataSource = "root:123456@(localhost:3306)/bbs_search?charset=utf8mb4&parseTime=True&loc=Local"
		db, err := gorm.Open(mysql.Open(DataSource), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		c.Db = db
		c.Query = query.Use(db)
	}
}
