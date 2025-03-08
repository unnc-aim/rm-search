package handler

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/database/model"
	"github.com/scutrobotlab/rm-search/svc"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

func MSearch(c *gin.Context) {
	startTime := time.Now()
	remoteIP := c.Request.Header.Get("X-Real-Ip")
	if remoteIP == "" {
		remoteIP = c.RemoteIP()
	}

	var mSearchBody []byte
	log := model.SearchLog{
		RemoteIP:       remoteIP,
		UserAgent:      c.Request.Header.Get("User-Agent"),
		RequestBody:    "",
		RequestLength:  0,
		Query:          "",
		Status:         0,
		ResponseBody:   "",
		ResponseLength: 0,
		Latency:        0,
	}
	defer func() {
		log.Latency = int32(time.Since(startTime).Milliseconds())
		log.ResponseBody = string(mSearchBody)
		log.ResponseLength = int32(len(mSearchBody))
		go writeSearchLog(&log)
	}()

	reqBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Status = http.StatusInternalServerError
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logrus.Debugf("reqBody: %s", string(reqBody))
	log.RequestBody = string(reqBody)
	log.RequestLength = int32(len(reqBody))
	log.Query = extractQuery(reqBody)

	mSearch, err := svc.Ctx().Elastic.Msearch(bytes.NewReader(reqBody))
	if err != nil {
		log.Status = http.StatusInternalServerError
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mSearchBody, err = io.ReadAll(mSearch.Body)
	if err != nil {
		log.Status = http.StatusInternalServerError
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logrus.Debugf("mSearchBody: %s", string(mSearchBody))

	log.Status = http.StatusOK
	c.Data(200, "application/json; charset=utf-8", mSearchBody)
}

func writeSearchLog(log *model.SearchLog) {
	l := svc.Ctx().Query.SearchLog
	err := l.Create(log)
	if err != nil {
		logrus.Errorf("create search log failed: %v", err)
	}
}

func extractQuery(reqBody []byte) string {
	objects := strings.Split(string(reqBody), "\n")
	for _, obj := range objects {
		if obj == "" {
			continue
		}
		var search map[string]interface{}
		err := json.Unmarshal([]byte(obj), &search)
		if err != nil {
			logrus.Errorf("unmarshal search log failed: %v", err)
			continue
		}
		if query, ok := search["query"]; ok {
			if functionScore, ok := query.(map[string]interface{})["function_score"]; ok {
				if query, ok := functionScore.(map[string]interface{})["query"]; ok {
					if boolQuery, ok := query.(map[string]interface{})["bool"]; ok {
						if must, ok := boolQuery.(map[string]interface{})["must"]; ok {
							if boolQuery, ok := must.(map[string]interface{})["bool"]; ok {
								if should, ok := boolQuery.(map[string]interface{})["should"]; ok {
									for _, shouldQuery := range should.([]interface{}) {
										if multiMatch, ok := shouldQuery.(map[string]interface{})["multi_match"]; ok {
											if query, ok := multiMatch.(map[string]interface{})["query"]; ok {
												return query.(string)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return ""
}
