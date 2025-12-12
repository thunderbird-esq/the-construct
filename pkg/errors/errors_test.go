package errors

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	// Verify sentinel errors are properly defined
	sentinels := []error{
		ErrNotFound,
		ErrInvalidInput,
		ErrPermissionDenied,
		ErrInventoryFull,
		ErrInsufficientFunds,
		ErrInCombat,
		ErrNotInCombat,
		ErrTargetNotFound,
		ErrItemNotEquippable,
		ErrSlotEmpty,
		ErrNotAtVendor,
		ErrNotAtBank,
		ErrRateLimited,
		ErrConnectionTimeout,
		ErrAuthFailed,
	}

	for _, err := range sentinels {
		if err == nil {
			t.Error("Sentinel error should not be nil")
		}
		if err.Error() == "" {
			t.Error("Sentinel error should have message")
		}
	}
}

func TestGameError(t *testing.T) {
	ge := NewGameError("GetItem", ErrNotFound, "katana")

	if ge.Op != "GetItem" {
		t.Errorf("Op = %q, want GetItem", ge.Op)
	}
	if ge.Err != ErrNotFound {
		t.Error("Err should be ErrNotFound")
	}
	if ge.Context != "katana" {
		t.Errorf("Context = %q, want katana", ge.Context)
	}

	// Test error message
	msg := ge.Error()
	if msg == "" {
		t.Error("Error message should not be empty")
	}
	if !errors.Is(ge, ErrNotFound) {
		t.Error("Should unwrap to ErrNotFound")
	}
}

func TestGameErrorWithoutContext(t *testing.T) {
	ge := NewGameError("MovePlayer", ErrPermissionDenied, "")

	msg := ge.Error()
	if msg == "" {
		t.Error("Error message should not be empty")
	}
	// Should not contain empty parentheses
	if ge.Context != "" {
		t.Error("Context should be empty")
	}
}

func TestResultSuccess(t *testing.T) {
	r := NewSuccess("Item picked up")

	if !r.Success {
		t.Error("Success should be true")
	}
	if r.Message != "Item picked up" {
		t.Errorf("Message = %q, want 'Item picked up'", r.Message)
	}
	if r.Error != nil {
		t.Error("Error should be nil for success")
	}
}

func TestResultSuccessWithData(t *testing.T) {
	data := map[string]int{"damage": 10}
	r := NewSuccessWithData("Attack landed", data)

	if !r.Success {
		t.Error("Success should be true")
	}
	if r.Data == nil {
		t.Error("Data should not be nil")
	}
}

func TestResultError(t *testing.T) {
	r := NewError(ErrNotFound, "Item not found")

	if r.Success {
		t.Error("Success should be false")
	}
	if r.Error != ErrNotFound {
		t.Error("Error should be ErrNotFound")
	}
	if r.Message != "Item not found" {
		t.Errorf("Message = %q, want 'Item not found'", r.Message)
	}
}

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(ErrNotFound) {
		t.Error("IsNotFound should return true for ErrNotFound")
	}

	ge := NewGameError("GetItem", ErrNotFound, "sword")
	if !IsNotFound(ge) {
		t.Error("IsNotFound should return true for wrapped ErrNotFound")
	}

	if IsNotFound(ErrPermissionDenied) {
		t.Error("IsNotFound should return false for other errors")
	}
}

func TestIsRateLimited(t *testing.T) {
	if !IsRateLimited(ErrRateLimited) {
		t.Error("IsRateLimited should return true for ErrRateLimited")
	}

	if IsRateLimited(ErrNotFound) {
		t.Error("IsRateLimited should return false for other errors")
	}
}

func TestIsPermissionDenied(t *testing.T) {
	if !IsPermissionDenied(ErrPermissionDenied) {
		t.Error("IsPermissionDenied should return true")
	}

	if IsPermissionDenied(ErrNotFound) {
		t.Error("IsPermissionDenied should return false for other errors")
	}
}

func TestUnwrap(t *testing.T) {
	ge := NewGameError("Test", ErrNotFound, "context")

	unwrapped := ge.Unwrap()
	if unwrapped != ErrNotFound {
		t.Error("Unwrap should return underlying error")
	}
}
