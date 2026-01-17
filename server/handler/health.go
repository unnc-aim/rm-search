package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/svc"
)

type HealthResponse struct {
	Status     string                 `json:"status"`
	Components map[string]interface{} `json:"components"`
}

func Health(c *gin.Context) {
	ctx := svc.Ctx()
	resp := HealthResponse{
		Status:     "UP",
		Components: make(map[string]interface{}),
	}

	msHealth, err := ctx.Meili.HealthWithContext(c)
	if err != nil {
		resp.Status = "DOWN"
		resp.Components["meilisearch"] = err.Error()
	} else {
		resp.Components["meilisearch"] = msHealth
	}

	db, err := ctx.Db.DB()
	if err != nil {
		resp.Status = "DOWN"
		resp.Components["database"] = err.Error()
	} else {
		err = db.PingContext(c)
		if err != nil {
			resp.Status = "DOWN"
			resp.Components["database"] = err.Error()
		}
	}

	if resp.Status != "UP" {
		c.JSON(http.StatusServiceUnavailable, resp)
	} else {
		c.JSON(http.StatusOK, resp)
	}
}
