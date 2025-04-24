package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/przemekk6973/swift-code-app/app/internal/adapter/api"
	"github.com/przemekk6973/swift-code-app/app/internal/adapter/persistence"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/usecases"
	"github.com/przemekk6973/swift-code-app/app/internal/initializer"
)

// Integration test covering full API flow with live MongoDB.
func TestAPI_Integration(t *testing.T) {

	// Load local MongoDB
	uri := "mongodb://localhost:27017"

	// Create repo and load test data from CSV
	repo, err := persistence.NewMongoRepository(uri, "testdb", "swiftCodes")
	if err != nil {
		t.Fatalf("repo init: %v", err)
	}

	// Example csv
	csv := `COUNTRY ISO2 CODE,SWIFT CODE,NAME,ADDRESS,COUNTRY NAME
PL,TESTPLP1XXX,TestHQ,AddrHQ,POLAND
PL,TESTPLP1BR1,TestBR,AddrBR,POLAND
`
	tmp, err := os.CreateTemp("", "int_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.WriteString(csv)
	tmp.Close()

	if _, err := initializer.ImportCSV(repo, tmp.Name(), map[string]string{"PL": "POLAND"}); err != nil {
		t.Fatalf("import CSV: %v", err)
	}

	// Start service and router
	gin.SetMode(gin.TestMode)
	svc := usecases.NewSwiftService(repo)
	router := api.SetupRouter(svc)

	// Helper function
	do := func(method, path string, body interface{}) *httptest.ResponseRecorder {
		var req *http.Request
		if body != nil {
			b, _ := json.Marshal(body)
			req = httptest.NewRequest(method, path, bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(method, path, nil)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	// GET HQ
	w := do("GET", "/v1/swift-codes/TESTPLP1XXX", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("GET HQ expected 200, got %d: %s", w.Code, w.Body)
	}
	var hq models.SwiftCode
	if err := json.Unmarshal(w.Body.Bytes(), &hq); err != nil {
		t.Fatal(err)
	}
	if !hq.IsHeadquarter || len(hq.Branches) != 1 {
		t.Errorf("unexpected HQ payload: %+v", hq)
	}

	// GET Branch
	w = do("GET", "/v1/swift-codes/TESTPLP1BR1", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("GET branch expected 200, got %d: %s", w.Code, w.Body)
	}
	var br models.SwiftCode
	if err := json.Unmarshal(w.Body.Bytes(), &br); err != nil {
		t.Fatal(err)
	}
	if br.IsHeadquarter {
		t.Errorf("branch should have IsHeadquarter=false: %+v", br)
	}

	// GET by Country
	w = do("GET", "/v1/swift-codes/country/PL", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("GET country expected 200, got %d: %s", w.Code, w.Body)
	}
	var resp models.CountrySwiftCodesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.SwiftCodes) != 2 {
		t.Errorf("expected 2 codes, got %d", len(resp.SwiftCodes))
	}

	// POST new HQ
	newHQ := models.SwiftCode{
		SwiftCode:     "NEWPLPPLXXX",
		BankName:      "New Bank",
		Address:       "New Addr",
		CountryISO2:   "PL",
		CountryName:   "POLAND",
		IsHeadquarter: true,
	}
	w = do("POST", "/v1/swift-codes", newHQ)
	if w.Code != http.StatusOK {
		t.Fatalf("POST HQ expected 200, got %d: %s", w.Code, w.Body)
	}

	// DELETE new HQ
	w = do("DELETE", "/v1/swift-codes/NEWPLPPLXXX", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("DELETE HQ expected 200, got %d: %s", w.Code, w.Body)
	}
}
