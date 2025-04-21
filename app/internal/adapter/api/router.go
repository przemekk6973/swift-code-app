package api

import (
	"github.com/gin-gonic/gin"
	"github.com/przemekk6973/swift-code-app/app/internal/adapter/api/v1"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/usecases"
)

// SetupRouter ustawia wszystkie endpointy
func SetupRouter(svc *usecases.SwiftService) *gin.Engine {
	r := gin.Default()
	// tu możesz dodać middleware: CORS, logging, recovery itd.

	handler := v1.NewSwiftHandler(svc)
	group := r.Group("/v1/swift-codes")
	{
		group.GET("/:swift-code", handler.GetSwiftCode)
		group.GET("/country/:countryISO2code", handler.GetSwiftCodesByCountry)
		group.POST("", handler.AddSwiftCode)
		group.DELETE("/:swift-code", handler.DeleteSwiftCode)
	}

	return r
}
