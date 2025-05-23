package usecases

import (
	"context"
	"net/http"
	"testing"

	"github.com/przemekk6973/swift-code-app/app/internal/domain/models"
	"github.com/przemekk6973/swift-code-app/app/internal/port"
	"github.com/przemekk6973/swift-code-app/app/internal/util"
)

type stubRepo struct {
	existing     map[string]bool
	byCode       map[string]models.SwiftCode
	byCountry    map[string][]models.SwiftCode
	addBranchErr error
	deleteErr    error
}

func (s *stubRepo) Ping(ctx context.Context) error {
	return nil
}

func (s *stubRepo) SaveHeadquarters(ctx context.Context, hqs []models.SwiftCode) (models.ImportSummary, error) {
	if s.existing == nil {
		s.existing = make(map[string]bool)
	}
	var summary models.ImportSummary
	for _, hq := range hqs {
		if s.existing[hq.SwiftCode] {
			summary.HQSkipped++
		} else {
			s.existing[hq.SwiftCode] = true
			summary.HQAdded++
		}
	}
	return summary, nil
}
func (s *stubRepo) SaveBranches(ctx context.Context, branches []models.SwiftCode) (models.ImportSummary, error) {
	return models.ImportSummary{}, nil
}
func (s *stubRepo) GetByCode(ctx context.Context, code string) (models.SwiftCode, error) {
	if v, ok := s.byCode[code]; ok {
		return v, nil
	}
	return models.SwiftCode{}, port.ErrNotFound
}
func (s *stubRepo) GetByCountry(ctx context.Context, iso2 string) ([]models.SwiftCode, error) {
	if v, ok := s.byCountry[iso2]; ok {
		return v, nil
	}
	return nil, port.ErrNotFound
}
func (s *stubRepo) AddBranch(ctx context.Context, hqCode string, branch models.SwiftBranch) error {
	return s.addBranchErr
}
func (s *stubRepo) Delete(ctx context.Context, code string) error {
	return s.deleteErr
}

func TestGetSwiftCodeDetails_NotFound(t *testing.T) {
	svc := NewSwiftService(&stubRepo{})
	_, err := svc.GetSwiftCodeDetails(context.Background(), "ABCDEFGH")
	if e, ok := err.(*util.AppError); !ok || e.StatusCode != 404 {
		t.Errorf("expected 404 AppError, got %v", err)
	}
}

func TestAddSwiftCode_BranchWithoutHQ(t *testing.T) {
	svc := NewSwiftService(&stubRepo{addBranchErr: port.ErrHQNotFound})
	err := svc.AddSwiftCode(context.Background(), models.SwiftCode{
		SwiftCode:     "ABCDEFGHBR1",
		BankName:      "B",
		Address:       "A",
		CountryISO2:   "PL",
		CountryName:   "POLAND",
		IsHeadquarter: false,
	})
	if e, ok := err.(*util.AppError); !ok || e.StatusCode != 400 {
		t.Errorf("expected 400 BadRequest, got %v", err)
	}
}

func TestDeleteSwiftCode_NotFound(t *testing.T) {
	svc := NewSwiftService(&stubRepo{deleteErr: port.ErrNotFound})
	err := svc.DeleteSwiftCode(context.Background(), "ABCDEFGHXXX")
	if e, ok := err.(*util.AppError); !ok || e.StatusCode != 404 {
		t.Errorf("expected 404 NotFound, got %v", err)
	}
}
func TestAddSwiftCode_HQ_Duplicate(t *testing.T) {
	code := models.SwiftCode{
		SwiftCode:     "DUPLPLP1XXX",
		BankName:      "Bank X",
		Address:       "HQ Street",
		CountryISO2:   "PL",
		CountryName:   "POLAND",
		IsHeadquarter: true,
	}

	repo := &stubRepo{existing: make(map[string]bool)}
	svc := NewSwiftService(repo)

	// Add HQ
	if err := svc.AddSwiftCode(context.Background(), code); err != nil {
		t.Fatalf("unexpected error on first insert: %v", err)
	}

	// Conflict
	err := svc.AddSwiftCode(context.Background(), code)
	if e, ok := err.(*util.AppError); !ok || e.StatusCode != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %v", err)
	}
}

func TestAddSwiftCode_InvalidSuffix(t *testing.T) {
	svc := NewSwiftService(&stubRepo{})
	code := models.SwiftCode{
		SwiftCode:     "INVALID1ABC", // not ending with XXX
		BankName:      "Bank X",
		Address:       "HQ Street",
		CountryISO2:   "PL",
		CountryName:   "POLAND",
		IsHeadquarter: true,
	}
	err := svc.AddSwiftCode(context.Background(), code)
	if e, ok := err.(*util.AppError); !ok || e.StatusCode != 400 {
		t.Errorf("expected 400 BadRequest for invalid HQ suffix, got %v", err)
	}
}
