// app/internal/initializer/initializer.go
package initializer

import (
	"context"
	"fmt"
	"time"

	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
	"github.com/przemekk6973/swift-code-app/app/internal/port"
	"github.com/przemekk6973/swift-code-app/app/internal/util"
)

// ImportCSV parses CSV with csvPath and saves the code through the repository
func ImportCSV(repo port.SwiftRepository, csvPath string, countries map[string]string) (*models.ImportSummary, error) {
	start := time.Now()

	hqList, branchList, err := util.LoadSwiftCodes(csvPath, countries)
	if err != nil {
		return nil, fmt.Errorf("csv parse error: %w", err)
	}

	hqSum, err := repo.SaveHeadquarters(context.Background(), hqList)
	if err != nil {
		return nil, fmt.Errorf("save HQ error: %w", err)
	}
	brSum, err := repo.SaveBranches(context.Background(), branchList)
	if err != nil {
		return nil, fmt.Errorf("save branches error: %w", err)
	}

	summary := &models.ImportSummary{
		HQAdded:           hqSum.HQAdded,
		HQSkipped:         hqSum.HQSkipped,
		BranchesAdded:     brSum.BranchesAdded,
		BranchesDuplicate: brSum.BranchesDuplicate,
		BranchesMissingHQ: brSum.BranchesMissingHQ,
		BranchesSkipped:   brSum.BranchesSkipped,
	}

	elapsed := time.Since(start)
	fmt.Printf("CSV import done in %v: %+v\n", elapsed, summary)
	return summary, nil
}
