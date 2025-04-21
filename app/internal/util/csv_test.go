package util

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestLoadSwiftCodes(t *testing.T) {
	csv := `COUNTRY ISO2 CODE,SWIFT CODE,NAME,ADDRESS,COUNTRY NAME
PL,AABBPLP1XXX,BANK1,ADDR1,POLAND
PL,AABBPLP1BR1,BANK1,ADDR2,POLAND
DE,CCCCDEFFXXX,BANK2,ADDR3,GERMANY
`
	tmp, err := ioutil.TempFile("", "swift*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(csv); err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	countries := map[string]string{"PL": "POLAND", "DE": "GERMANY"}
	hqs, branches, err := LoadSwiftCodes(tmp.Name(), countries)
	if err != nil {
		t.Fatal(err)
	}
	if len(hqs) != 2 {
		t.Errorf("expected 2 HQ, got %d", len(hqs))
	}
	if len(branches) != 1 {
		t.Errorf("expected 1 branch, got %d", len(branches))
	}
	// dalej możesz sprawdzić, że hqs[0].SwiftCode=="AABBPLP1XXX", itd.
}
