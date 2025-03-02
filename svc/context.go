package svc

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/go-tika/tika"
	"github.com/scutrobotlab/rm-search/database/query"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var global *Context

type Context struct {
	Db      *gorm.DB
	Query   *query.Query
	Elastic *elasticsearch.Client
	Tika    *tika.Client
}

func Ctx() *Context {
	return global
}

func InitContext(c Config) {
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

	global = &Context{
		Db:      db,
		Query:   query.Use(db),
		Elastic: elastic,
	}
}
