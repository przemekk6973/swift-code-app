package util

import (
	"testing"
)

func TestValidateSwiftCode(t *testing.T) {
	valid := []string{
		"ABCDEFGH",    // 8 znaków HQ
		"ABCDEFGHXXX", // 11 znaków HQ
		"ABCDEFGH123", // 11 znaków branch
	}
	for _, code := range valid {
		if err := ValidateSwiftCode(code); err != nil {
			t.Errorf("expected valid, got error for %q: %v", code, err)
		}
	}

	invalid := []string{
		"ABC",           // za krótkie
		"ABCDEFGHIJKLM", // za długie (13)
		"ABCDEFGH!",     // niedozwolony znak
	}
	for _, code := range invalid {
		if err := ValidateSwiftCode(code); err == nil {
			t.Errorf("expected error for invalid %q, got nil", code)
		}
	}
}

func TestValidateCountryISO2(t *testing.T) {
	good := []string{"PL", "US", "DE"}
	for _, iso := range good {
		if err := ValidateCountryISO2(iso); err != nil {
			t.Errorf("expected valid ISO2 %q, got %v", iso, err)
		}
	}
	bad := []string{"pl", "USA", "1A", ""}
	for _, iso := range bad {
		if err := ValidateCountryISO2(iso); err == nil {
			t.Errorf("expected error for invalid ISO2 %q, got nil", iso)
		}
	}
}

// Jeśli masz ValidateCountryNameMatch(country, name, map), dodaj tak:
func TestValidateCountryNameMatch(t *testing.T) {
	countries := map[string]string{"PL": "POLAND", "DE": "GERMANY"}
	if err := ValidateCountryNameMatch("PL", "POLAND", countries); err != nil {
		t.Fatal(err)
	}
	if err := ValidateCountryNameMatch("PL", "POLska", countries); err == nil {
		t.Errorf("expected mismatch error for country name different case/typo")
	}
}
