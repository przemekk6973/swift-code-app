package util

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
)

// LoadSwiftCodes czyta plik CSV pod ścieżką csvPath i zwraca dwie listy:
//   - hqList: rekordy z kodami HQ (kończące się na "XXX")
//   - branchList: rekordy z kodami oddziałów
func LoadSwiftCodes(csvPath string, countries map[string]string) ([]models.SwiftCode, []models.SwiftCode, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// Wczytaj nagłówek
	header, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("error reading CSV header: %v", err)
	}

	// Mapowanie nazw kolumn na indeksy
	indexes := make(map[string]int)
	for i, col := range header {
		key := strings.ToUpper(strings.TrimSpace(col))
		indexes[key] = i
	}

	var hqList []models.SwiftCode
	var branchList []models.SwiftCode

	// Przetwarzanie wierszy
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("error reading CSV record: %v", err)
		}

		// Ekstrakcja i normalizacja pól
		swiftCode := strings.ToUpper(strings.TrimSpace(record[indexes["SWIFT CODE"]]))
		countryISO2 := strings.ToUpper(strings.TrimSpace(record[indexes["COUNTRY ISO2 CODE"]]))
		bankName := strings.TrimSpace(record[indexes["NAME"]])
		address := strings.TrimSpace(record[indexes["ADDRESS"]])
		countryName := strings.TrimSpace(record[indexes["COUNTRY NAME"]])

		// Walidacja kodu i kraju
		if err := ValidateSwiftCode(swiftCode); err != nil {
			continue
		}
		if err := ValidateCountryISO2(countryISO2); err != nil {
			continue
		}
		if err := ValidateCountryNameMatch(countryISO2, countryName, countries); err != nil {
			continue
		}

		// Rozdzielenie HQ vs oddział
		isHQ := strings.HasSuffix(swiftCode, "XXX")
		if isHQ {
			hqList = append(hqList, models.SwiftCode{
				SwiftCode:     swiftCode,
				BankName:      bankName,
				Address:       address,
				CountryISO2:   countryISO2,
				CountryName:   countryName,
				IsHeadquarter: true,
				Branches:      []models.SwiftBranch{},
			})
		} else {
			branchList = append(branchList, models.SwiftCode{
				SwiftCode:     swiftCode,
				BankName:      bankName,
				Address:       address,
				CountryISO2:   countryISO2,
				CountryName:   countryName,
				IsHeadquarter: false,
			})
		}
	}

	return hqList, branchList, nil
}
