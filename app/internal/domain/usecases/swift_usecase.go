package usecases

import (
	"context"

	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
	"github.com/przemekk6973/swift-code-app/app/internal/port"
	"github.com/przemekk6973/swift-code-app/app/internal/util"
)

// SwiftService does operations on SWIFT coes
type SwiftService struct {
	repo port.SwiftRepository
}

// NewSwiftService creates new insance of service
func NewSwiftService(r port.SwiftRepository) *SwiftService {
	return &SwiftService{repo: r}
}

// GetSwiftCodeDetails returns data of HQ or branch by code
func (s *SwiftService) GetSwiftCodeDetails(ctx context.Context, code string) (models.SwiftCode, error) {
	// walidacja formatu SWIFT
	if err := util.ValidateSwiftCode(code); err != nil {
		return models.SwiftCode{}, util.BadRequest("invalid SWIFT code: %v", err)
	}
	// get from repository
	swift, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		if err == port.ErrNotFound {
			return models.SwiftCode{}, util.NotFound("SWIFT code %s not found", code)
		}
		return models.SwiftCode{}, util.Internal("error fetching SWIFT code: %v", err)
	}
	return swift, nil
}

// GetSwiftCodesByCountry returns every HQ and branch for every ISO2
func (s *SwiftService) GetSwiftCodesByCountry(ctx context.Context, iso2 string) (models.CountrySwiftCodesResponse, error) {
	// walidacja ISO2
	if err := util.ValidateCountryISO2(iso2); err != nil {
		return models.CountrySwiftCodesResponse{}, util.BadRequest("invalid country ISO2: %v", err)
	}
	list, err := s.repo.GetByCountry(ctx, iso2)
	if err != nil {
		if err == port.ErrNotFound {
			// no data for this country: 404
			return models.CountrySwiftCodesResponse{}, util.NotFound("no SWIFT codes for country %s", iso2)
		}
		// any other repo error: 500
		return models.CountrySwiftCodesResponse{}, util.Internal("error fetching by country: %v", err)
	}
	// if repo might return an empty slice without error:
	if len(list) == 0 {
		return models.CountrySwiftCodesResponse{}, util.NotFound("no SWIFT codes for country %s", iso2)
	}

	// country name from the first element
	countryName := list[0].CountryName

	// consolidate HQ and branches into one list SwiftBranch
	var branches []models.SwiftBranch
	for _, sc := range list {
		branches = append(branches, models.SwiftBranch{
			Address:       sc.Address,
			BankName:      sc.BankName,
			CountryISO2:   sc.CountryISO2,
			CountryName:   sc.CountryName,
			IsHeadquarter: sc.IsHeadquarter,
			SwiftCode:     sc.SwiftCode,
		})
	}
	return models.CountrySwiftCodesResponse{
		CountryISO2: iso2,
		CountryName: countryName,
		SwiftCodes:  branches,
	}, nil
}

// AddSwiftCode adds single HQ or branch
func (s *SwiftService) AddSwiftCode(ctx context.Context, sc models.SwiftCode) error {
	// validate
	if err := util.ValidateSwiftCode(sc.SwiftCode); err != nil {
		return util.BadRequest("invalid SWIFT code: %v", err)
	}
	if err := util.ValidateSwiftSuffix(sc.SwiftCode, sc.IsHeadquarter); err != nil {
		return err
	}
	if err := util.ValidateCountryISO2(sc.CountryISO2); err != nil {
		return util.BadRequest("invalid country ISO2: %v", err)
	}

	if sc.IsHeadquarter {
		summary, err := s.repo.SaveHeadquarters(ctx, []models.SwiftCode{sc})
		if err != nil {
			return util.Internal("error saving HQ: %v", err)
		}
		if summary.HQSkipped > 0 {
			return util.Conflict("headquarter %s already exists", sc.SwiftCode)
		}
		return nil
	}

	// add branch – search HQ with first 8 characters + "XXX"
	hqCode := sc.SwiftCode[:8] + "XXX"
	branch := models.SwiftBranch{
		Address:       sc.Address,
		BankName:      sc.BankName,
		CountryISO2:   sc.CountryISO2,
		CountryName:   sc.CountryName,
		IsHeadquarter: false,
		SwiftCode:     sc.SwiftCode,
	}
	if err := s.repo.AddBranch(ctx, hqCode, branch); err != nil {
		switch err {
		case port.ErrHQNotFound:
			return util.BadRequest("headquarter %s not found for branch %s", hqCode, sc.SwiftCode)
		case port.ErrBranchDuplicate:
			return util.Conflict("branch %s already exists", sc.SwiftCode)
		default:
			return util.Internal("error adding branch: %v", err)
		}
	}
	return nil
}

// DeleteSwiftCode removes HQ (and its branches) or single branch
func (s *SwiftService) DeleteSwiftCode(ctx context.Context, code string) error {
	if err := util.ValidateSwiftCode(code); err != nil {
		return util.BadRequest("invalid SWIFT code: %v", err)
	}
	if err := s.repo.Delete(ctx, code); err != nil {
		if err == port.ErrNotFound {
			return util.NotFound("SWIFT code %s not found", code)
		}
		return util.Internal("error deleting SWIFT code: %v", err)
	}
	return nil
}

// HealthCheck pings the database to check if it is available
func (s *SwiftService) HealthCheck(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
