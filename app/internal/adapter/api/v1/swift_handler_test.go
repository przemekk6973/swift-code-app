// app/internal/adapter/api/v1/swift_handler_test.go
package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/usecases"
	"github.com/przemekk6973/swift-code-app/app/internal/port"
)

// stubRepo implements port.SwiftRepository with controllable behavior.
type stubRepo struct {
	getCode    func(ctx context.Context, code string) (models.SwiftCode, error)
	getCountry func(ctx context.Context, iso2 string) ([]models.SwiftCode, error)
	addCode    func(ctx context.Context, sc models.SwiftCode) error
	deleteCode func(ctx context.Context, code string) error
}

func (s *stubRepo) SaveHeadquarters(context.Context, []models.SwiftCode) (models.ImportSummary, error) {
	return models.ImportSummary{}, nil
}
func (s *stubRepo) SaveBranches(context.Context, []models.SwiftCode) (models.ImportSummary, error) {
	return models.ImportSummary{}, nil
}
func (s *stubRepo) GetByCode(ctx context.Context, code string) (models.SwiftCode, error) {
	return s.getCode(ctx, code)
}
func (s *stubRepo) GetByCountry(ctx context.Context, iso2 string) ([]models.SwiftCode, error) {
	return s.getCountry(ctx, iso2)
}
func (s *stubRepo) AddBranch(ctx context.Context, hqCode string, br models.SwiftBranch) error {
	// not used in handler tests
	return nil
}
func (s *stubRepo) Delete(ctx context.Context, code string) error {
	return s.deleteCode(ctx, code)
}
func (s *stubRepo) Ping(ctx context.Context) error {
	return nil
}

func setupRouterWithStub(repo port.SwiftRepository) *gin.Engine {
	svc := usecases.NewSwiftService(repo)
	handler := NewSwiftHandler(svc)
	r := gin.New()
	// only mount the handlers you want to test
	r.GET("/v1/swift-codes/:swift-code", handler.GetSwiftCode)
	r.GET("/v1/swift-codes/country/:countryISO2code", handler.GetSwiftCodesByCountry)
	r.POST("/v1/swift-codes", handler.AddSwiftCode)
	r.DELETE("/v1/swift-codes/:swift-code", handler.DeleteSwiftCode)
	return r
}

func TestGetSwiftCode_Success(t *testing.T) {
	expected := models.SwiftCode{
		SwiftCode:     "ABCDEFGHXXX",
		BankName:      "Test Bank",
		Address:       "Test Addr",
		CountryISO2:   "US",
		CountryName:   "UNITED STATES",
		IsHeadquarter: true,
	}
	repo := &stubRepo{
		getCode: func(ctx context.Context, code string) (models.SwiftCode, error) {
			return expected, nil
		},
	}
	router := setupRouterWithStub(repo)

	req := httptest.NewRequest("GET", "/v1/swift-codes/ABCDEFGHXXX", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp models.SwiftCode
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.SwiftCode != expected.SwiftCode {
		t.Errorf("unexpected body: %+v", resp)
	}
}

func TestGetSwiftCode_NotFound(t *testing.T) {
	repo := &stubRepo{
		getCode: func(ctx context.Context, code string) (models.SwiftCode, error) {
			return models.SwiftCode{}, port.ErrNotFound
		},
	}
	router := setupRouterWithStub(repo)

	req := httptest.NewRequest("GET", "/v1/swift-codes/ABCDEFGHXXX", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetSwiftCode_InternalError(t *testing.T) {
	repo := &stubRepo{
		getCode: func(ctx context.Context, code string) (models.SwiftCode, error) {
			return models.SwiftCode{}, errors.New("boom")
		},
	}
	router := setupRouterWithStub(repo)

	req := httptest.NewRequest("GET", "/v1/swift-codes/ABCDEFGHXXX", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
