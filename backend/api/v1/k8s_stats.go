package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"devops-backend/model/response"
	"devops-backend/service"
)

type StatsController struct{}

var statsService = &service.StatsService{}

func (c *StatsController) GetClusterStats(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail("无效的集群ID"))
		return
	}

	stats, err := statsService.GetClusterStats(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(stats))
}
