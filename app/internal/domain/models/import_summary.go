package models

// ImportSummary podsumowuje wynik importu CSV
type ImportSummary struct {
	HQAdded           int `json:"hqAdded"`
	HQSkipped         int `json:"hqSkipped"`
	BranchesAdded     int `json:"branchesAdded"`
	BranchesDuplicate int `json:"branchesDuplicate"`
	BranchesMissingHQ int `json:"branchesMissingHQ"`
	BranchesSkipped   int `json:"branchesSkipped"`
}
