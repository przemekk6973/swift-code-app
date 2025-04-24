package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
)

const (
	testURI        = "mongodb://localhost:27017"
	testDB         = "test_swift_repo"
	testCollection = "test_codes"
)

// getTestRepo tries to connect: skips the test if Mongo isn't running.
func getTestRepo(t *testing.T) *MongoRepository {
	repoIface, err := NewMongoRepository(testURI, testDB, testCollection)
	if err != nil {
		t.Skipf("skipping Mongo tests; cannot connect: %v", err)
	}
	repo := repoIface.(*MongoRepository)
	// clean slate
	repo.collection.Drop(context.Background())
	return repo
}

func TestSaveHeadquartersAndGetByCode(t *testing.T) {
	repo := getTestRepo(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hqs := []models.SwiftCode{
		{SwiftCode: "AAAAUS33XXX", BankName: "Bank A", Address: "Addr A", CountryISO2: "US", CountryName: "USA", IsHeadquarter: true},
		{SwiftCode: "BBBBDEFFXXX", BankName: "Bank B", Address: "Addr B", CountryISO2: "DE", CountryName: "GERMANY", IsHeadquarter: true},
	}
	summary, err := repo.SaveHeadquarters(ctx, hqs)
	if err != nil {
		t.Fatalf("SaveHeadquarters failed: %v", err)
	}
	if summary.HQAdded != 2 || summary.HQSkipped != 0 {
		t.Errorf("first import summary = %+v; want HQAdded=2,HQSkipped=0", summary)
	}

	// repeated import should skip both
	summary, err = repo.SaveHeadquarters(ctx, hqs)
	if err != nil {
		t.Fatal(err)
	}
	if summary.HQAdded != 0 || summary.HQSkipped != 2 {
		t.Errorf("second import summary = %+v; want HQAdded=0,HQSkipped=2", summary)
	}

	// fetch one HQ
	got, err := repo.GetByCode(ctx, "AAAAUS33XXX")
	if err != nil {
		t.Fatalf("GetByCode HQ failed: %v", err)
	}
	if got.SwiftCode != "AAAAUS33XXX" || !got.IsHeadquarter {
		t.Errorf("GetByCode HQ = %+v; want code=AAAAUS33XXX,isHQ=true", got)
	}
}

func TestSaveBranchesAndGetByCountryAndDelete(t *testing.T) {
	repo := getTestRepo(t)
	ctx := context.Background()

	// First insert a HQ to attach branches to
	_, _ = repo.SaveHeadquarters(ctx, []models.SwiftCode{
		{SwiftCode: "CCCCGB2LXXX", BankName: "Bank C", Address: "Addr C", CountryISO2: "GB", CountryName: "UK", IsHeadquarter: true},
	})

	// Branch with no HQ
	branches := []models.SwiftCode{
		{SwiftCode: "ZZZZGB2LAB1", BankName: "Branch Z", Address: "Addr Z", CountryISO2: "GB", CountryName: "UK", IsHeadquarter: false},
	}
	summary, err := repo.SaveBranches(ctx, branches)
	if err != nil {
		t.Fatal(err)
	}
	if summary.BranchesMissingHQ != 1 {
		t.Errorf("expected BranchesMissingHQ=1; got %+v", summary)
	}

	// Valid branch
	branches = []models.SwiftCode{
		{SwiftCode: "CCCCGB2LAB1", BankName: "Branch C1", Address: "Addr C1", CountryISO2: "GB", CountryName: "UK", IsHeadquarter: false},
	}
	summary, err = repo.SaveBranches(ctx, branches)
	if err != nil {
		t.Fatal(err)
	}
	if summary.BranchesAdded != 1 {
		t.Errorf("expected BranchesAdded=1; got %+v", summary)
	}

	// Duplicate branch
	summary, err = repo.SaveBranches(ctx, branches)
	if err != nil {
		t.Fatal(err)
	}
	if summary.BranchesDuplicate != 1 {
		t.Errorf("expected BranchesDuplicate=1; got %+v", summary)
	}

	// GetByCountry should return 1 HQ + its branch = 2 entries
	all, err := repo.GetByCountry(ctx, "GB")
	if err != nil {
		t.Fatalf("GetByCountry failed: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("GetByCountry returned %d items; want 2", len(all))
	}

	// Delete branch
	if err := repo.Delete(ctx, "CCCCGB2LAB1"); err != nil {
		t.Fatal(err)
	}
	// now only HQ remains
	all, _ = repo.GetByCountry(ctx, "GB")
	if len(all) != 1 {
		t.Errorf("after branch delete, got %d items; want 1", len(all))
	}

	// Delete HQ (and its branches, none now)
	if err := repo.Delete(ctx, "CCCCGB2LXXX"); err != nil {
		t.Fatal(err)
	}
	// country now empty → error
	if _, err := repo.GetByCountry(ctx, "GB"); err == nil {
		t.Errorf("expected not found after HQ delete")
	}
}
