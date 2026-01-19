package svc

import (
	"net/http"

	"github.com/google/go-tika/tika"
	"github.com/meilisearch/meilisearch-go"
	"github.com/scutrobotlab/rm-search/database/query"
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
		const DataSource = "root:123456@(localhost:3306)/rm_search?charset=utf8mb4&parseTime=True&loc=Local"
		db, err := gorm.Open(mysql.Open(DataSource), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		c.Db = db
		c.Query = query.Use(db)
	}
}

func WithMeili() Option {
	return func(c *Context) {
		const MeiliSearchAddress = "http://localhost:7700"
		const MeiliSearchAPIKey = "123456"
		client := meilisearch.New(MeiliSearchAddress, meilisearch.WithAPIKey(MeiliSearchAPIKey))
		c.Meili = client
	}
}

func WithTika() Option {
	return func(c *Context) {
		const TikaHost = "http://localhost:9998"
		client := tika.NewClient(http.DefaultClient, TikaHost)
		c.Tika = client
	}
}

func WithConfig() Option {
	return func(c *Context) {
		c.Config = ReadConfig("../etc/config.yaml")
	}
}
