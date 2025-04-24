package util

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
)

// LoadSwiftCodes reads CSV from csvPath and returns two lists:
//   - hqList: records with HQ (ending with "XXX")
//   - branchList: records with branch codes
func LoadSwiftCodes(csvPath string, countries map[string]string) ([]models.SwiftCode, []models.SwiftCode, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// load header
	header, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("error reading CSV header: %v", err)
	}

	// map column names to indexes
	indexes := make(map[string]int)
	for i, col := range header {
		key := strings.ToUpper(strings.TrimSpace(col))
		indexes[key] = i
	}

	var hqList []models.SwiftCode
	var branchList []models.SwiftCode

	// row processing
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("error reading CSV record: %v", err)
		}

		//Field extraction and normalization
		swiftCode := strings.ToUpper(strings.TrimSpace(record[indexes["SWIFT CODE"]]))
		countryISO2 := strings.ToUpper(strings.TrimSpace(record[indexes["COUNTRY ISO2 CODE"]]))
		bankName := strings.TrimSpace(record[indexes["NAME"]])
		address := strings.TrimSpace(record[indexes["ADDRESS"]])
		countryName := strings.TrimSpace(record[indexes["COUNTRY NAME"]])

		// validate code and country
		if err := ValidateSwiftCode(swiftCode); err != nil {
			continue
		}
		if err := ValidateCountryISO2(countryISO2); err != nil {
			continue
		}
		if err := ValidateCountryNameMatch(countryISO2, countryName, countries); err != nil {
			continue
		}

		// HQ and branch separation
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
