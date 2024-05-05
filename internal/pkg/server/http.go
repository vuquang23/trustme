package server

import (
	"context"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

	"github.com/vuquang23/trustme/internal/pkg/server/middleware"
	requestidpkg "github.com/vuquang23/trustme/internal/pkg/util/requestid"
	"github.com/vuquang23/trustme/pkg/logger"
)

func Run(ctx context.Context, address string, engine *gin.Engine) error {
	srv := &http.Server{
		Addr:    address,
		Handler: engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, err.Error())
		}
	}()
	<-ctx.Done()
	logger.Info(ctx, "HTTP server shutdown")
	return srv.Shutdown(ctx)
}

func GinEngine(config Config, logCfg logger.Config, logBackend logger.LoggerBackend) *gin.Engine {
	gin.SetMode(config.Mode)

	middlewares := []gin.HandlerFunc{
		requestid.New(
			requestid.WithCustomHeaderStrKey(requestidpkg.HeaderKeyRequestID),
			requestid.WithHandler(func(c *gin.Context, requestID string) {
				c.Request = c.Request.WithContext(requestidpkg.SetRequestIDToContext(c.Request.Context(), requestID))
			}),
		),
		middleware.NewLoggerMiddleware(logCfg, logBackend),
		gin.Recovery(),
	}

	engine := gin.New()
	engine.Use(middlewares...)

	setCORS(engine)

	return engine
}

func setCORS(engine *gin.Engine) {
	corsConfig := cors.DefaultConfig()
	corsConfig.AddAllowMethods(http.MethodOptions)
	corsConfig.AllowAllOrigins = true
	engine.Use(cors.New(corsConfig))
}
