package svc

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/go-tika/tika"
	"github.com/patrickmn/go-cache"
	"github.com/scutrobotlab/rm-search/database/query"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"time"
)

var global *Context

type Context struct {
	Config  Config
	Db      *gorm.DB
	Query   *query.Query
	Elastic *elasticsearch.Client
	Tika    *tika.Client
	Cache   *cache.Cache
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
		Username:  c.ElasticConfig.Username,
		Password:  c.ElasticConfig.Password,
	})
	if err != nil {
		panic(err)
	}

	global = &Context{
		Config:  c,
		Db:      db,
		Query:   query.Use(db),
		Elastic: elastic,
		Tika:    tika.NewClient(http.DefaultClient, c.TikaHost),
		Cache:   cache.New(cache.DefaultExpiration, time.Minute),
	}
}
