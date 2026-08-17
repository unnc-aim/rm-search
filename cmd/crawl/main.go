// cmd/crawl backfills the database from the RoboMaster forum and
// announcement pages. It is intended for fresh deployments; the running
// server keeps data fresh afterwards through its incremental jobs.
package main

import (
	"context"
	"flag"

	"github.com/sirupsen/logrus"

	"github.com/scutrobotlab/rm-search/index"
	"github.com/scutrobotlab/rm-search/svc"
)

func main() {
	configFile := flag.String("config", "etc/config.yaml", "config file")
	postsStart := flag.Int64("posts-start", 0, "inclusive start of the bbs post id range")
	postsEnd := flag.Int64("posts-end", 2000000, "exclusive end of the bbs post id range, 0 disables")
	postsGoroutines := flag.Int("posts-goroutines", 50, "concurrent workers for bbs posts")
	announceStart := flag.Int64("announce-start", 0, "inclusive start of the announce id range")
	announceEnd := flag.Int64("announce-end", 0, "exclusive end of the announce id range, 0 disables")
	announceGoroutines := flag.Int("announce-goroutines", 10, "concurrent workers for announces")
	flag.Parse()

	c := svc.ReadConfig(*configFile)
	svc.InitContext(c)

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
