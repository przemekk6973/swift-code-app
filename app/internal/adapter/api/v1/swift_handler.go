package v1

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/usecases"
	"github.com/przemekk6973/swift-code-app/app/internal/util"
)

type SwiftHandler struct {
	svc *usecases.SwiftService
}

func NewSwiftHandler(svc *usecases.SwiftService) *SwiftHandler {
	return &SwiftHandler{svc: svc}
}

// GET /v1/swift-codes/:swift-code
func (h *SwiftHandler) GetSwiftCode(c *gin.Context) {
	code := strings.ToUpper(c.Param(util.ParamSwiftCode))
	swift, err := h.svc.GetSwiftCodeDetails(c.Request.Context(), code)
	if err != nil {
		c.JSON(util.StatusCodeFromError(err), gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, swift)
}

// GET /v1/swift-codes/country/:countryISO2code
func (h *SwiftHandler) GetSwiftCodesByCountry(c *gin.Context) {
	iso2 := strings.ToUpper(c.Param(util.ParamCountryISO2))
	resp, err := h.svc.GetSwiftCodesByCountry(c.Request.Context(), iso2)
	if err != nil {
		c.JSON(util.StatusCodeFromError(err), gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, resp)
}

// POST /v1/swift-codes
func (h *SwiftHandler) AddSwiftCode(c *gin.Context) {
	var req models.SwiftCode
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid JSON payload"})
		return
	}
	// Upewnij się, że kody są uppercase
	req.SwiftCode = strings.ToUpper(req.SwiftCode)
	req.CountryISO2 = strings.ToUpper(req.CountryISO2)

	if err := h.svc.AddSwiftCode(c.Request.Context(), req); err != nil {
		c.JSON(util.StatusCodeFromError(err), gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "swift code added"})
}

// DELETE /v1/swift-codes/:swift-code
func (h *SwiftHandler) DeleteSwiftCode(c *gin.Context) {
	code := strings.ToUpper(c.Param(util.ParamSwiftCode))
	if err := h.svc.DeleteSwiftCode(c.Request.Context(), code); err != nil {
		c.JSON(util.StatusCodeFromError(err), gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "swift code deleted"})
}
