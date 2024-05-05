package api

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/vuquang23/trustme/pkg/logger"
)

func SetupRoute(engine *gin.Engine, parser IParser) {
	rg := engine.Group("/api")

	rg.GET("/current-block", GetCurrentBlock(parser))
	rg.POST("/subscribe", SubscribeAddress(parser))
	rg.GET("/txs", GetTransactions(parser))
}

func GetCurrentBlock(parser IParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		RespondSuccess(c, parser.GetCurrentBlock())
	}
}

type SubscribeAddressParams struct {
	Address string `json:"address"`
}

func SubscribeAddress(parser IParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params SubscribeAddressParams
		if err := c.ShouldBindJSON(&params); err != nil {
			logger.Error(c, err.Error())
			RespondFailure(c, err)
			return
		}

		RespondSuccess(c, parser.Subscribe(strings.ToLower(params.Address)))
	}
}

type GetTransactionsParams struct {
	Address string `form:"address"`
}

func GetTransactions(parser IParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params GetTransactionsParams
		if err := c.ShouldBindQuery(&params); err != nil {
			logger.Error(c, err.Error())
			RespondFailure(c, err)
			return
		}

		RespondSuccess(c, parser.GetTransactions(strings.ToLower(params.Address)))
	}
}
