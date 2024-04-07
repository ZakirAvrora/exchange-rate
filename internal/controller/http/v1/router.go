package v1

import (
	"net/http"

	"github.com/ZakirAvrora/exchange-rate/internal/exchangerates"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Swagger docs.
	_ "github.com/ZakirAvrora/exchange-rate/docs"
)

// NewRouter -.
// Swagger spec:
// @title       Currency Exchange Rate API
// @description Description of a currency exchange service endpoints
// @version     1.0
// @host        localhost:8080
// @BasePath    /v1
func NewRouter(handler *gin.Engine, t exchangerates.RecodsService) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	// Swagger
	swaggerHandler := ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER_HTTP_HANDLER")
	handler.GET("/swagger/*any", swaggerHandler)

	// Health check
	handler.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Routers
	h := handler.Group("/v1")
	{
		newExchangeRatesRoutes(h, t)
	}
}
