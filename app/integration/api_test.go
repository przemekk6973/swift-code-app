// app/integration/api_test.go
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	"github.com/przemekk6973/swift-code-app/app/internal/adapter/api"
	"github.com/przemekk6973/swift-code-app/app/internal/adapter/persistence"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/usecases"
	"github.com/przemekk6973/swift-code-app/app/internal/initializer"
)

func TestAPI_EndToEnd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping integration test on Windows (Testcontainers rootless not supported)")
	}
	ctx := context.Background()

	// 1) Uruchom kontener MongoDB
	req := tc.ContainerRequest{
		Image:        "mongo:6.0",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections"),
	}
	mongoC, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer mongoC.Terminate(ctx)

	host, err := mongoC.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	mp, err := mongoC.MappedPort(ctx, "27017")
	if err != nil {
		t.Fatal(err)
	}
	uri := fmt.Sprintf("mongodb://%s:%s", host, mp.Port())

	// 2) Utwórz repo i zaimportuj mini‑CSV
	repo, err := persistence.NewMongoRepository(uri, "testdb", "swiftCodes")
	if err != nil {
		t.Fatalf("repo init: %v", err)
	}
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

	// 3) Przygotuj Gin + router
	gin.SetMode(gin.TestMode)
	svc := usecases.NewSwiftService(repo)
	router := api.SetupRouter(svc)

	// pomocnicza funkcja do zapytań
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

	// 4) GET HQ
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

	// 5) GET branch
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

	// 6) GET by country
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

	// 7) POST new HQ
	newHQ := models.SwiftCode{
		SwiftCode:     "NEWPLP1XXX",
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

	// 8) DELETE HQ
	w = do("DELETE", "/v1/swift-codes/NEWPLP1XXX", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("DELETE HQ expected 200, got %d: %s", w.Code, w.Body)
	}
}
