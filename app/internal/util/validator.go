package util

import (
	"strings"
	"unicode"
)

// ValidateSwiftCode sprawdza, czy kod SWIFT ma długość 8 lub 11, tylko litery i cyfry
func ValidateSwiftCode(code string) error {
	length := len(code)
	if length != 8 && length != 11 {
		return WrapError(ErrBadRequest, "SWIFT code must be 8 or 11 characters, got %d", length)
	}
	for _, r := range code {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return WrapError(ErrBadRequest, "SWIFT code can only contain letters and digits")
		}
	}
	return nil
}

// ValidateSwiftSuffix sprawdza, czy kody HQ (XXX) i oddziałów są poprawne
func ValidateSwiftSuffix(code string, isHQ bool) error {
	if isHQ && !(len(code) == 11 && code[8:] == "XXX") {
		return WrapError(ErrBadRequest, "HQ SWIFT code must end with 'XXX'")
	}
	if !isHQ && len(code) == 11 && code[8:] == "XXX" {
		return WrapError(ErrBadRequest, "Branch SWIFT code cannot end with 'XXX'")
	}
	return nil
}

// ValidateCountryISO2 ensure code has exactly two uppercase letters
func ValidateCountryISO2(iso2 string) error {
	if len(iso2) != 2 {
		return WrapError(ErrBadRequest, "country ISO2 must be 2 characters")
	}
	for _, r := range iso2 {
		if r < 'A' || r > 'Z' {
			return WrapError(ErrBadRequest, "country ISO2 must be uppercase letters")
		}
	}
	return nil
}

// ValidateCountryNameMatch checks if nazwa kraju pasuje do ISO2 (dostępnego z mapy)
func ValidateCountryNameMatch(iso2, inputName string, nameMap map[string]string) error {
	expected, ok := nameMap[iso2]
	if !ok {
		return WrapError(ErrBadRequest, "unknown country ISO2: %s", iso2)
	}
	if !strings.EqualFold(inputName, expected) {
		return WrapError(ErrBadRequest, "country name '%s' does not match ISO2 '%s' (expected '%s')", inputName, iso2, expected)
	}
	return nil
}
