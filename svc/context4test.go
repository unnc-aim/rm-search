package svc

import (
	"net/http"

	"github.com/google/go-tika/tika"
	"github.com/meilisearch/meilisearch-go"
	"github.com/scutrobotlab/rm-search/database/query"
	"github.com/scutrobotlab/rm-search/database/schema"
	"gorm.io/driver/postgres"
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
		const DataSource = "host=localhost port=5432 user=postgres password=123456 dbname=rm_search sslmode=disable"
		db, err := gorm.Open(postgres.Open(DataSource), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		if err := schema.Ensure(db); err != nil {
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
