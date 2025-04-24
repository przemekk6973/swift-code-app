package util

import (
	"testing"
)

func TestValidateSwiftCode(t *testing.T) {
	valid := []string{
		"ABCDEFGH",    // 8 characters, HQ without XXX
		"ABCDEFGHXXX", // 11 characters, HQ
		"ABCDEFGH123", // 11 characters, branch
	}
	for _, code := range valid {
		if err := ValidateSwiftCode(code); err != nil {
			t.Errorf("expected valid, got error for %q: %v", code, err)
		}
	}

	invalid := []string{
		"ABC",           // too short
		"ABCDEFGHIJKLM", // too long
		"ABCDEFGH!@#",   // forbidden characters
		"ABCDEFGHXX",    // wrong amount of characters
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

func TestValidateSwiftSuffix(t *testing.T) {
	err := ValidateSwiftSuffix("ABCDEFGHXXX", false)
	if err == nil {
		t.Error("branch ending with XXX is invalid")
	}

	err = ValidateSwiftSuffix("ABCDEFGH001", true)
	if err == nil {
		t.Error("HQ without XXX is invalid")
	}

	err = ValidateSwiftSuffix("ABCDEFGHXXX", true)
	if err != nil {
		t.Errorf("valid HQ failed: %v", err)
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

func TestValidateCountryNameMatch_Errors(t *testing.T) {
	m := map[string]string{"PL": "Poland"}
	// Unknown ISO2
	if err := ValidateCountryNameMatch("XX", "X", m); err == nil {
		t.Error("expected error for unknown ISO2, got nil")
	}
	// Mismatch name
	if err := ValidateCountryNameMatch("PL", "Deutschland", m); err == nil {
		t.Error("expected error for name mismatch, got nil")
	}
}
