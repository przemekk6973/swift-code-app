package models

type SwiftCode struct {
	Address       string        `bson:"address"      json:"address"`
	BankName      string        `bson:"bankName"     json:"bankName"`
	CountryISO2   string        `bson:"countryISO2"  json:"countryISO2"`
	CountryName   string        `bson:"countryName"  json:"countryName"`
	IsHeadquarter bool          `bson:"isHeadquarter" json:"isHeadquarter"`
	SwiftCode     string        `bson:"swiftCode"    json:"swiftCode"`
	Branches      []SwiftBranch `bson:"branches,omitempty" json:"branches,omitempty"`
}
