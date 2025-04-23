// app/internal/initializer/initializer_test.go
package initializer

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
)

// minimalRepo implements only the methods ImportCSV calls + Ping
type minimalRepo struct {
	hqSum, brSum models.ImportSummary
}

func (r *minimalRepo) SaveHeadquarters(_ context.Context, hqs []models.SwiftCode) (models.ImportSummary, error) {
	// expect exactly 2 HQ codes
	if len(hqs) != 2 {
		return models.ImportSummary{}, nil
	}
	return r.hqSum, nil
}
func (r *minimalRepo) SaveBranches(_ context.Context, brs []models.SwiftCode) (models.ImportSummary, error) {
	// expect exactly 1 branch
	if len(brs) != 1 {
		return models.ImportSummary{}, nil
	}
	return r.brSum, nil
}
func (r *minimalRepo) Ping(_ context.Context) error { return nil }
func (r *minimalRepo) GetByCode(_ context.Context, _ string) (models.SwiftCode, error) {
	panic("unused")
}
func (r *minimalRepo) GetByCountry(_ context.Context, _ string) ([]models.SwiftCode, error) {
	panic("unused")
}
func (r *minimalRepo) AddBranch(_ context.Context, _ string, _ models.SwiftBranch) error {
	panic("unused")
}
func (r *minimalRepo) Delete(_ context.Context, _ string) error { panic("unused") }

func TestImportCSV_Success(t *testing.T) {
	// valid 11‑char codes: two HQ and one branch
	csv := `COUNTRY ISO2 CODE,SWIFT CODE,NAME,ADDRESS,COUNTRY NAME
PL,AAAAPL1AXXX,HeadQ1,Addr1,POLAND
PL,AAAAPL1A001,Branch1,Addr2,POLAND
DE,BBBBDE2AXXX,HeadQ2,Addr3,GERMANY
`
	tmp, err := ioutil.TempFile("", "test_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.WriteString(csv)
	tmp.Close()

	repo := &minimalRepo{
		hqSum: models.ImportSummary{HQAdded: 2},
		brSum: models.ImportSummary{BranchesAdded: 1},
	}

	countries := map[string]string{"PL": "POLAND", "DE": "GERMANY"}

	sum, err := ImportCSV(repo, tmp.Name(), countries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// we only care that our stub sums are returned
	if sum.HQAdded != 2 || sum.BranchesAdded != 1 {
		t.Errorf("got summary %+v; want HQAdded=2, BranchesAdded=1", sum)
	}
}

func TestImportCSV_FileNotFound(t *testing.T) {
	repo := &minimalRepo{}
	if _, err := ImportCSV(repo, "no-such-file.csv", nil); err == nil {
		t.Fatal("expected error when CSV is missing, got nil")
	}
}
