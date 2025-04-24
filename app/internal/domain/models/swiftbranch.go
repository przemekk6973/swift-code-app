package models

// SwiftBranch structure of SWIFT branch
type SwiftBranch struct {
	Address       string `bson:"address"       json:"address"`
	BankName      string `bson:"bankName"      json:"bankName"`
	CountryISO2   string `bson:"countryISO2"   json:"countryISO2"`
	CountryName   string `bson:"countryName,omitempty" json:"countryName,omitempty"`
	IsHeadquarter bool   `bson:"isHeadquarter" json:"isHeadquarter"`
	SwiftCode     string `bson:"swiftCode"     json:"swiftCode"`
}
