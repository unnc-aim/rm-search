package service

import (
	"github.com/scutrobotlab/bbs-search/database/query"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Context struct {
	Db    *gorm.DB
	Query *query.Query
}

func NewContext(c Config) *Context {
	db, err := gorm.Open(mysql.Open(c.DataSource), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return &Context{
		Db:    db,
		Query: query.Use(db),
	}
}
