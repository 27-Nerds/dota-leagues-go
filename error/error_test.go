package e

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestWrappedErrorClassification(t *testing.T) {
	inner := &Error{Code: ENOTFOUND, Message: "missing match", Err: context.Canceled}
	err := fmt.Errorf("request: %w", &Error{Op: "Get", Err: inner})
	if !errors.Is(err, context.Canceled) || !IsNotFound(err) || ErrorMessage(err) != "missing match" {
		t.Fatalf("lost wrapped error information: %v", err)
	}
	var appError *Error
	if !errors.As(err, &appError) || appError.Op != "Get" {
		t.Fatal("errors.As did not find the application error")
	}
}
