package util

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
)

func TestLoadSwiftCodes(t *testing.T) {
	sample := `COUNTRY ISO2 CODE,SWIFT CODE,NAME,ADDRESS,COUNTRY NAME
PL,AABBPLP1XXX,TestHQ1,Address1,POLAND
PL,AABBPLP1BR1,TestBR1,Address2,POLAND
DE,CCCCDEFFXXX,TestHQ2,Address3,GERMANY
`
	tmp, err := ioutil.TempFile("", "swift_test_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(sample); err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	// podstawowa mapa krajów
	countries := map[string]string{"PL": "POLAND", "DE": "GERMANY"}

	hqList, brList, err := LoadSwiftCodes(tmp.Name(), countries)
	if err != nil {
		t.Fatalf("LoadSwiftCodes error: %v", err)
	}

	if len(hqList) != 2 {
		t.Errorf("expected 2 HQ, got %d", len(hqList))
	}
	if len(brList) != 1 {
		t.Errorf("expected 1 branch, got %d", len(brList))
	}

	// checkk first HQ
	wantHQ := models.SwiftCode{
		SwiftCode:     "AABBPLP1XXX",
		BankName:      "TestHQ1",
		Address:       "Address1",
		CountryISO2:   "PL",
		CountryName:   "POLAND",
		IsHeadquarter: true,
		Branches:      []models.SwiftBranch{},
	}

	if !reflect.DeepEqual(hqList[0], wantHQ) {
		t.Errorf("HQ mismatch:\n got %+v\nwant %+v", hqList[0], wantHQ)
	}

	// check branch
	wantBR := models.SwiftCode{
		SwiftCode:     "AABBPLP1BR1",
		BankName:      "TestBR1",
		Address:       "Address2",
		CountryISO2:   "PL",
		CountryName:   "POLAND",
		IsHeadquarter: false,
	}
	if !reflect.DeepEqual(brList[0], wantBR) {
		t.Errorf("Branch mismatch:\n got %+v\nwant %+v", brList[0], wantBR)
	}
}
