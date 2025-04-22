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

// GetSwiftCode
// @Summary      Retrieve details for a single SWIFT code
// @Description  Returns the headquarter with its branches if the code is HQ, or a single branch object if branch code.
// @Tags         swift-codes
// @Accept       json
// @Produce      json
// @Param        swift-code   path      string            true  "SWIFT code (8 or 11 characters)"
// @Success      200          {object}  models.SwiftCode
// @Failure      400          {object}  map[string]string "invalid SWIFT code format"
// @Failure      404          {object}  map[string]string "SWIFT code not found"
// @Failure      500          {object}  map[string]string "internal server error"
// @Router       /v1/swift-codes/{swift-code} [get]
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

// GetSwiftCodesByCountry
// @Summary      Retrieve all SWIFT codes for a country
// @Description  Returns all headquarters and branches for a given country ISO2.
// @Tags         swift-codes
// @Accept       json
// @Produce      json
// @Param        countryISO2code  path      string                                  true  "Country ISO2 code"
// @Success      200              {object}  models.CountrySwiftCodesResponse
// @Failure      400              {object}  map[string]string                     "invalid ISO2 format"
// @Failure      404              {object}  map[string]string                     "no SWIFT codes for country"
// @Failure      500              {object}  map[string]string                     "internal server error"
// @Router       /v1/swift-codes/country/{countryISO2code} [get]
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

// AddSwiftCode
// @Summary      Create a new SWIFT code entry
// @Description  Adds either a headquarter (isHeadquarter=true) or a branch (isHeadquarter=false).
// @Tags         swift-codes
// @Accept       json
// @Produce      json
// @Param        payload  body      models.SwiftCode  true  "SWIFT code payload"
// @Success      200      {object}  map[string]string  "swift code added"
// @Failure      400      {object}  map[string]string  "invalid input or missing HQ for branch"
// @Failure      409      {object}  map[string]string  "duplicate code"
// @Failure      500      {object}  map[string]string  "internal server error"
// @Router       /v1/swift-codes [post]
func (h *SwiftHandler) AddSwiftCode(c *gin.Context) {
	var req models.SwiftCode
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid JSON payload"})
		return
	}
	// Ensure uppercase codes
	req.SwiftCode = strings.ToUpper(req.SwiftCode)
	req.CountryISO2 = strings.ToUpper(req.CountryISO2)

	if err := h.svc.AddSwiftCode(c.Request.Context(), req); err != nil {
		c.JSON(util.StatusCodeFromError(err), gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "swift code added"})
}

// DELETE /v1/swift-codes/:swift-code

// DeleteSwiftCode
// @Summary      Delete a SWIFT code entry
// @Description  Deletes a headquarter (and all its branches) if code ends with XXX, or a single branch otherwise.
// @Tags         swift-codes
// @Accept       json
// @Produce      json
// @Param        swift-code  path      string            true  "SWIFT code to delete"
// @Success      200         {object}  map[string]string  "swift code deleted"
// @Failure      400         {object}  map[string]string  "invalid SWIFT code format"
// @Failure      404         {object}  map[string]string  "SWIFT code not found"
// @Failure      500         {object}  map[string]string  "internal server error"
// @Router       /v1/swift-codes/{swift-code} [delete]
func (h *SwiftHandler) DeleteSwiftCode(c *gin.Context) {
	code := strings.ToUpper(c.Param(util.ParamSwiftCode))
	if err := h.svc.DeleteSwiftCode(c.Request.Context(), code); err != nil {
		c.JSON(util.StatusCodeFromError(err), gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "swift code deleted"})
}
