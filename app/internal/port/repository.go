package port

import (
	"context"
	"errors"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
)

// SwiftRepository definiuje metody CRUD dla kodów SWIFT
type SwiftRepository interface {
	// SaveHeadquarters zapisuje listę kodów głównych (HQ)
	SaveHeadquarters(ctx context.Context, hqs []models.SwiftCode) (models.ImportSummary, error)

	// SaveBranches zapisuje listę kodów oddziałów
	SaveBranches(ctx context.Context, branches []models.SwiftCode) (models.ImportSummary, error)

	// GetByCode pobiera SwiftCode (HQ lub oddział) po kodzie
	GetByCode(ctx context.Context, code string) (models.SwiftCode, error)

	// GetByCountry pobiera wszystkie kody dla danego kraju (ISO2)
	GetByCountry(ctx context.Context, iso2 string) ([]models.SwiftCode, error)

	// AddBranch dodaje oddział do istniejącego HQ
	AddBranch(ctx context.Context, hqCode string, branch models.SwiftBranch) error

	// Delete usuwa wpis po podanym kodzie SWIFT
	Delete(ctx context.Context, code string) error
}

var (
	ErrNotFound        = errors.New("swift code not found")
	ErrHQNotFound      = errors.New("headquarter not found")
	ErrBranchDuplicate = errors.New("branch already exists")
)
