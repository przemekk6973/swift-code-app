package usecases

import (
	"context"

	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
	"github.com/przemekk6973/swift-code-app/app/internal/port"
	"github.com/przemekk6973/swift-code-app/app/internal/util"
)

// SwiftService realizuje operacje na SWIFTach
type SwiftService struct {
	repo port.SwiftRepository
}

// NewSwiftService tworzy instancję serwisu
func NewSwiftService(r port.SwiftRepository) *SwiftService {
	return &SwiftService{repo: r}
}

// GetSwiftCodeDetails zwraca dane HQ lub branch po kodzie
func (s *SwiftService) GetSwiftCodeDetails(ctx context.Context, code string) (models.SwiftCode, error) {
	// walidacja formatu SWIFT
	if err := util.ValidateSwiftCode(code); err != nil {
		return models.SwiftCode{}, util.BadRequest("invalid SWIFT code: %v", err)
	}
	// pobierz z repozytorium
	swift, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		if err == port.ErrNotFound {
			return models.SwiftCode{}, util.NotFound("SWIFT code %s not found", code)
		}
		return models.SwiftCode{}, util.Internal("error fetching SWIFT code: %v", err)
	}
	return swift, nil
}

// GetSwiftCodesByCountry zwraca wszystkie HQ i oddziały dla danego ISO2
func (s *SwiftService) GetSwiftCodesByCountry(ctx context.Context, iso2 string) (models.CountrySwiftCodesResponse, error) {
	// walidacja ISO2
	if err := util.ValidateCountryISO2(iso2); err != nil {
		return models.CountrySwiftCodesResponse{}, util.BadRequest("invalid country ISO2: %v", err)
	}
	list, err := s.repo.GetByCountry(ctx, iso2)
	if err != nil {
		if err == port.ErrNotFound {
			// no data for this country → 404
			return models.CountrySwiftCodesResponse{}, util.NotFound("no SWIFT codes for country %s", iso2)
		}
		// any other repo error → 500
		return models.CountrySwiftCodesResponse{}, util.Internal("error fetching by country: %v", err)
	}
	// (optional) if repo might return an empty slice without error:
	if len(list) == 0 {
		return models.CountrySwiftCodesResponse{}, util.NotFound("no SWIFT codes for country %s", iso2)
	}

	// nazwa kraju z pierwszego elementu
	countryName := list[0].CountryName

	// konsolidacja HQ i oddziałów w jedną listę SwiftBranch
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

// AddSwiftCode dodaje pojedyncze HQ lub oddział
func (s *SwiftService) AddSwiftCode(ctx context.Context, sc models.SwiftCode) error {
	// walidacje
	if err := util.ValidateSwiftCode(sc.SwiftCode); err != nil {
		return util.BadRequest("invalid SWIFT code: %v", err)
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

	// dodanie oddziału – szukamy HQ po pierwszych 8 znakach + "XXX"
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

// DeleteSwiftCode usuwa HQ (i wszystkie oddziały) lub pojedynczy oddział
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
