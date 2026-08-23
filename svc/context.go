package svc

import (
	"net/http"
	"time"

	"github.com/google/go-tika/tika"
	"github.com/meilisearch/meilisearch-go"
	"github.com/patrickmn/go-cache"
	"github.com/scutrobotlab/rm-search/common"
	"github.com/scutrobotlab/rm-search/database/query"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var global *Context

type Context struct {
	Config Config
	Db     *gorm.DB
	Query  *query.Query
	Meili  meilisearch.ServiceManager
	Index  meilisearch.IndexManager
	Tika   *tika.Client
	Cache  *cache.Cache
}

func Ctx() *Context {
	return global
}

func InitContext(c Config) {
	db, err := gorm.Open(postgres.Open(c.DataSource), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	meili := meilisearch.New(c.MeiliSearch.Address, meilisearch.WithAPIKey(c.MeiliSearch.APIKey))

	global = &Context{
		Config: c,
		Db:     db,
		Query:  query.Use(db),
		Tika:   tika.NewClient(http.DefaultClient, c.TikaHost),
		Cache:  cache.New(cache.DefaultExpiration, time.Minute),
		Meili:  meili,
		Index:  meili.Index(common.IndexEntityName),
	}
}
