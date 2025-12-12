// Package errors provides custom error types for Matrix MUD.
// These typed errors enable better error handling and allow callers
// to distinguish between different failure modes.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for common failure conditions
var (
	// ErrNotFound indicates a requested resource doesn't exist
	ErrNotFound = errors.New("not found")

	// ErrInvalidInput indicates malformed or invalid user input
	ErrInvalidInput = errors.New("invalid input")

	// ErrPermissionDenied indicates the action is not allowed
	ErrPermissionDenied = errors.New("permission denied")

	// ErrInventoryFull indicates the player's inventory is at capacity
	ErrInventoryFull = errors.New("inventory full")

	// ErrInsufficientFunds indicates not enough money for purchase
	ErrInsufficientFunds = errors.New("insufficient funds")

	// ErrInCombat indicates an action cannot be performed while in combat
	ErrInCombat = errors.New("cannot do that while in combat")

	// ErrNotInCombat indicates an action requires being in combat
	ErrNotInCombat = errors.New("not in combat")

	// ErrTargetNotFound indicates the specified target doesn't exist
	ErrTargetNotFound = errors.New("target not found")

	// ErrItemNotEquippable indicates an item cannot be equipped
	ErrItemNotEquippable = errors.New("item cannot be equipped")

	// ErrSlotEmpty indicates no item in the specified equipment slot
	ErrSlotEmpty = errors.New("nothing equipped in that slot")

	// ErrNotAtVendor indicates player is not at a vendor location
	ErrNotAtVendor = errors.New("no vendor here")

	// ErrNotAtBank indicates player is not at a bank/storage location
	ErrNotAtBank = errors.New("no storage facility here")

	// ErrRateLimited indicates too many requests
	ErrRateLimited = errors.New("rate limited")

	// ErrConnectionTimeout indicates the connection timed out
	ErrConnectionTimeout = errors.New("connection timeout")

	// ErrAuthFailed indicates authentication failed
	ErrAuthFailed = errors.New("authentication failed")
)

// GameError wraps an error with additional context
type GameError struct {
	Op      string // Operation that failed (e.g., "GetItem", "MovePlayer")
	Err     error  // Underlying error
	Context string // Additional context (e.g., item name, room ID)
}

// Error implements the error interface
func (e *GameError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Op, e.Err.Error(), e.Context)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Err.Error())
}

// Unwrap returns the underlying error for errors.Is/As
func (e *GameError) Unwrap() error {
	return e.Err
}

// NewGameError creates a new GameError
func NewGameError(op string, err error, context string) *GameError {
	return &GameError{
		Op:      op,
		Err:     err,
		Context: context,
	}
}

// Result represents the outcome of a game operation
type Result struct {
	Success bool
	Message string
	Error   error
	Data    interface{} // Optional data payload
}

// NewSuccess creates a successful result
func NewSuccess(message string) Result {
	return Result{
		Success: true,
		Message: message,
	}
}

// NewSuccessWithData creates a successful result with data
func NewSuccessWithData(message string, data interface{}) Result {
	return Result{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewError creates a failed result
func NewError(err error, message string) Result {
	return Result{
		Success: false,
		Message: message,
		Error:   err,
	}
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsRateLimited checks if the error is a rate limit error
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsPermissionDenied checks if the error is a permission error
func IsPermissionDenied(err error) bool {
	return errors.Is(err, ErrPermissionDenied)
}
