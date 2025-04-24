package port

import (
	"context"
	"errors"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
)

// SwiftRepository defines CRUD for SWIFT codes
type SwiftRepository interface {
	// SaveHeadquarters saves HQ list
	SaveHeadquarters(ctx context.Context, hqs []models.SwiftCode) (models.ImportSummary, error)

	// SaveBranches saves list of branches
	SaveBranches(ctx context.Context, branches []models.SwiftCode) (models.ImportSummary, error)

	// GetByCode saves SwiftCode (HQ or branch) by code
	GetByCode(ctx context.Context, code string) (models.SwiftCode, error)

	// GetByCountry downloads all codes by ISO2
	GetByCountry(ctx context.Context, iso2 string) ([]models.SwiftCode, error)

	// AddBranch adds branch for existing HQ
	AddBranch(ctx context.Context, hqCode string, branch models.SwiftBranch) error

	// Delete deletes by SWIFT
	Delete(ctx context.Context, code string) error

	Ping(ctx context.Context) error
}

var (
	ErrNotFound        = errors.New("swift code not found")
	ErrHQNotFound      = errors.New("headquarter not found")
	ErrBranchDuplicate = errors.New("branch already exists")
)
