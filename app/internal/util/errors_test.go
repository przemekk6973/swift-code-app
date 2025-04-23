package util

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestAppErrorAndStatusCode(t *testing.T) {
	// BadRequest
	e := BadRequest("oops %d", 42)
	if !strings.Contains(e.Error(), "oops 42") {
		t.Errorf("BadRequest message = %q; want contains %q", e.Error(), "oops 42")
	}
	if e.StatusCode != 400 {
		t.Errorf("BadRequest status = %d; want 400", e.StatusCode)
	}
	// Conflict
	e2 := Conflict("dup %s", "X")
	if e2.StatusCode != 409 {
		t.Errorf("Conflict status = %d; want 409", e2.StatusCode)
	}
	// Internal
	e3 := Internal("err %v", fmt.Errorf("inner"))
	if e3.StatusCode != 500 {
		t.Errorf("Internal status = %d; want 500", e3.StatusCode)
	}
	// StatusCodeFromError
	if StatusCodeFromError(e2) != 409 {
		t.Errorf("StatusCodeFromError = %d; want 409", StatusCodeFromError(e2))
	}
	if StatusCodeFromError(errors.New("x")) != 500 {
		t.Errorf("StatusCodeFromError(non-AppError) = %d; want 500", StatusCodeFromError(errors.New("x")))
	}
}
