package route

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/service"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func MSearch(c *gin.Context) {
	reqBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logrus.Debugf("reqBody: %s", string(reqBody))

	mSearch, err := service.Ctx().Elastic.Msearch(bytes.NewReader(reqBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	mSearchBody, err := io.ReadAll(mSearch.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logrus.Debugf("mSearchBody: %s", string(mSearchBody))

	c.Data(200, "application/json; charset=utf-8", mSearchBody)
}
