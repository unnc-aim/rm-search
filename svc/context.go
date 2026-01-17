package svc

import (
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/go-tika/tika"
	"github.com/meilisearch/meilisearch-go"
	"github.com/patrickmn/go-cache"
	"github.com/scutrobotlab/rm-search/database/query"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var global *Context

type Context struct {
	Config  Config
	Db      *gorm.DB
	Query   *query.Query
	Elastic *elasticsearch.Client
	Meili   meilisearch.ServiceManager
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

	client := meilisearch.New(c.MeiliSearch.Address, meilisearch.WithAPIKey(c.MeiliSearch.APIKey))
	v, err := client.Version()
	if err != nil {
		panic(err)
	}
	logrus.Infof("MeiliSearch version: %s", v)

	global = &Context{
		Config: c,
		Db:     db,
		Query:  query.Use(db),
		Meili:  client,
		Tika:   tika.NewClient(http.DefaultClient, c.TikaHost),
		Cache:  cache.New(cache.DefaultExpiration, time.Minute),
	}
}
