package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/scutrobotlab/rm-search/common"
	"github.com/scutrobotlab/rm-search/index"
	"github.com/scutrobotlab/rm-search/job"
	"github.com/scutrobotlab/rm-search/server"
	"github.com/scutrobotlab/rm-search/svc"
	"github.com/sirupsen/logrus"
)

// Single-binary layout: the former cmd/ tools are subcommands so one
// static binary (~50MB) replaces five and image updates stay small.
//
//	rm-search [server] [-config etc/config.yaml] [-addr :8080]
//	rm-search setup-index [-config ...]
//	rm-search recreate-index [-config ...]
//	rm-search incremental-index [-config ...]
//	rm-search crawl [-config ...] [--posts-start 0 --posts-end 2000000
//	                 --posts-goroutines 50 --announce-start 0 --announce-end 0
//	                 --announce-goroutines 10]
//
// Tiny /usr/local/bin/<tool> wrapper scripts (written by the Dockerfile)
// keep the old entrypoints working.
func main() {
	cmd := "server"
	args := os.Args[1:]
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd = args[0]
		args = args[1:]
	}

	if common.IsProd() {
		logrus.SetLevel(logrus.InfoLevel)
	}

	switch cmd {
	case "server":
		runServer(args)
	case "setup-index":
		runSetupIndex(args)
	case "recreate-index":
		runRecreateIndex(args)
	case "incremental-index":
		runIncrementalIndex(args)
	case "crawl":
		runCrawl(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q; want server|setup-index|recreate-index|incremental-index|crawl\n", cmd)
		os.Exit(2)
	}
}

func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	configFile := fs.String("config", "etc/config.yaml", "config file")
	addr := fs.String("addr", ":8080", "server address")
	_ = fs.Parse(args)

	svc.InitContext(svc.ReadConfig(*configFile))

	// Apply index settings on every boot so a fresh Meilisearch instance
	// is usable immediately (filterable/sortable attributes); without
	// this, filtered searches silently degrade until setup-index runs.
	if err := index.NewIndexer(svc.Ctx()).UpdateIndexSettings(context.Background()); err != nil {
		logrus.Errorf("update index settings failed (search filters may not work): %v", err)
	}

	job.NewBase(svc.Ctx()).Start()
	server.Run(*addr)
}

func runSetupIndex(args []string) {
	fs := flag.NewFlagSet("setup-index", flag.ExitOnError)
	configFile := fs.String("config", "etc/config.yaml", "config file")
	_ = fs.Parse(args)

	svc.InitContext(svc.ReadConfig(*configFile))
	if err := index.NewIndexer(svc.Ctx()).UpdateIndexSettings(context.Background()); err != nil {
		panic(err)
	}
}

func runRecreateIndex(args []string) {
	fs := flag.NewFlagSet("recreate-index", flag.ExitOnError)
	configFile := fs.String("config", "etc/config.yaml", "config file")
	_ = fs.Parse(args)

	svc.InitContext(svc.ReadConfig(*configFile))
	if err := index.NewIndexer(svc.Ctx()).RecreateIndex(context.Background()); err != nil {
		panic(err)
	}
}

func runIncrementalIndex(args []string) {
	fs := flag.NewFlagSet("incremental-index", flag.ExitOnError)
	configFile := fs.String("config", "etc/config.yaml", "config file")
	_ = fs.Parse(args)

	svc.InitContext(svc.ReadConfig(*configFile))
	idx := index.NewIndexer(svc.Ctx())
	j := job.IncrementalJob{Indexer: idx, Base: *job.NewBase(svc.Ctx())}
	j.Run()
}

func runCrawl(args []string) {
	fs := flag.NewFlagSet("crawl", flag.ExitOnError)
	configFile := fs.String("config", "etc/config.yaml", "config file")
	postsStart := fs.Int64("posts-start", 0, "inclusive start of the bbs post id range")
	postsEnd := fs.Int64("posts-end", 2000000, "exclusive end of the bbs post id range, 0 disables")
	postsGoroutines := fs.Int("posts-goroutines", 50, "concurrent workers for bbs posts")
	announceStart := fs.Int64("announce-start", 0, "inclusive start of the announce id range")
	announceEnd := fs.Int64("announce-end", 0, "exclusive end of the announce id range, 0 disables")
	announceGoroutines := fs.Int("announce-goroutines", 10, "concurrent workers for announces")
	_ = fs.Parse(args)

	svc.InitContext(svc.ReadConfig(*configFile))
	idx := index.NewIndexer(svc.Ctx())
	ctx := context.Background()

	if *postsEnd > *postsStart {
		logrus.Infof("crawling bbs posts [%d, %d) with %d goroutines", *postsStart, *postsEnd, *postsGoroutines)
		if err := idx.BatchPersistenceRangeIfNotExist(ctx, *postsStart, *postsEnd, *postsGoroutines); err != nil {
			logrus.Fatalf("crawl bbs posts failed: %v", err)
		}
		logrus.Infof("bbs posts [%d, %d) crawled", *postsStart, *postsEnd)
	}

	if *announceEnd > *announceStart {
		logrus.Infof("crawling announces [%d, %d) with %d goroutines", *announceStart, *announceEnd, *announceGoroutines)
		if err := idx.BatchPersistenceAnnounceRange(ctx, *announceStart, *announceEnd, *announceGoroutines); err != nil {
			logrus.Fatalf("crawl announces failed: %v", err)
		}
		logrus.Infof("announces [%d, %d) crawled", *announceStart, *announceEnd)
	}

	logrus.Info("crawl finished")
}
