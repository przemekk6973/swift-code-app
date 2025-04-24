package models

// CountrySwiftCodesResponse response structure for GET /v1/swift-codes/country/{iso2}
type CountrySwiftCodesResponse struct {
	CountryISO2 string        `json:"countryISO2"`
	CountryName string        `json:"countryName"`
	SwiftCodes  []SwiftBranch `json:"swiftCodes"`
}
