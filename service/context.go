package service

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/scutrobotlab/bbs-search/database/query"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Context struct {
	Db      *gorm.DB
	Query   *query.Query
	Elastic *elasticsearch.Client
}

func NewContext(c Config) *Context {
	db, err := gorm.Open(mysql.Open(c.DataSource), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	elastic, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: c.ElasticConfig.Addresses,
		APIKey:    c.ElasticConfig.APIKey,
	})
	if err != nil {
		panic(err)
	}

	return &Context{
		Db:      db,
		Query:   query.Use(db),
		Elastic: elastic,
	}
}
