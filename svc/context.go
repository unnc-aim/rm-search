package svc

import (
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/go-tika/tika"
	"github.com/patrickmn/go-cache"
	"github.com/scutrobotlab/rm-search/database/query"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	var db *gorm.DB
	var err error
	switch c.Driver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(c.DataSource), &gorm.Config{})
	case "postgres":
		db, err = gorm.Open(postgres.Open(c.DataSource), &gorm.Config{})
	default:
		panic(fmt.Sprintf("unsupported driver: %s", c.Driver))
	}
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
