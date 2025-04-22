package util

import (
	"testing"
)

func TestValidateSwiftCode(t *testing.T) {
	valid := []string{
		"ABCDEFGH",    // 8 znaków, HQ bez XXX
		"ABCDEFGHXXX", // 11 znaków, HQ
		"ABCDEFGH123", // 11 znaków, branch
	}
	for _, code := range valid {
		if err := ValidateSwiftCode(code); err != nil {
			t.Errorf("expected valid, got error for %q: %v", code, err)
		}
	}

	invalid := []string{
		"ABC",           // za krótkie
		"ABCDEFGHIJKLM", // za długie
		"ABCDEFGH!@#",   // zabronione znaki
		"ABCDEFGHXX",    // 10 znaków
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

func TestValidateCountryNameMatch(t *testing.T) {
	countries := map[string]string{
		"PL": "POLAND",
		"DE": "Germany", // mixed case
	}
	// exact match
	if err := ValidateCountryNameMatch("PL", "POLAND", countries); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// case‐insensitive match
	if err := ValidateCountryNameMatch("DE", "germany", countries); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// mismatch
	if err := ValidateCountryNameMatch("PL", "POLska", countries); err == nil {
		t.Errorf("expected mismatch error for %q vs %q", "PL", "POLska")
	}
	// missing key
	if err := ValidateCountryNameMatch("XX", "Xland", countries); err == nil {
		t.Errorf("expected error for missing country ISO2 key")
	}
}
