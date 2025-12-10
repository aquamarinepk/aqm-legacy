package auth

import (
	"errors"
	"testing"
)

func TestErrorVariables(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrInvalidEmail", ErrInvalidEmail, "invalid email address"},
		{"ErrWeakPassword", ErrWeakPassword, "password does not meet security requirements"},
		{"ErrInvalidCredentials", ErrInvalidCredentials, "invalid credentials"},
		{"ErrUserNotFound", ErrUserNotFound, "user not found"},
		{"ErrUserSuspended", ErrUserSuspended, "user account suspended"},
		{"ErrUserDeleted", ErrUserDeleted, "user account deleted"},
		{"ErrSessionExpired", ErrSessionExpired, "session expired"},
		{"ErrInvalidToken", ErrInvalidToken, "invalid token"},
		{"ErrTokenExpired", ErrTokenExpired, "token expired"},
		{"ErrInvalidAudience", ErrInvalidAudience, "invalid token audience"},
		{"ErrInvalidScope", ErrInvalidScope, "invalid scope"},
		{"ErrPermissionDenied", ErrPermissionDenied, "permission denied"},
		{"ErrPolicyNotFound", ErrPolicyNotFound, "policy not found"},
		{"ErrInvalidPolicy", ErrInvalidPolicy, "invalid policy format"},
		{"ErrGrantExpired", ErrGrantExpired, "grant expired"},
		{"ErrRoleNotFound", ErrRoleNotFound, "role not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("Error() = %s, want %s", tt.err.Error(), tt.msg)
			}
		})
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	errs := []error{
		ErrInvalidEmail,
		ErrWeakPassword,
		ErrInvalidCredentials,
		ErrUserNotFound,
		ErrUserSuspended,
		ErrUserDeleted,
		ErrSessionExpired,
		ErrInvalidToken,
		ErrTokenExpired,
		ErrInvalidAudience,
		ErrInvalidScope,
		ErrPermissionDenied,
		ErrPolicyNotFound,
		ErrInvalidPolicy,
		ErrGrantExpired,
		ErrRoleNotFound,
	}

	for i, err1 := range errs {
		for j, err2 := range errs {
			if i != j && errors.Is(err1, err2) {
				t.Errorf("errors should be distinct: %v and %v", err1, err2)
			}
		}
	}
}

func TestValidationErrorError(t *testing.T) {
	err := ValidationError{
		Field:   "email",
		Code:    "invalid_format",
		Message: "invalid email format",
	}

	if err.Error() != "invalid email format" {
		t.Errorf("Error() = %s, want invalid email format", err.Error())
	}
}

func TestValidationErrorFields(t *testing.T) {
	err := ValidationError{
		Field:   "password",
		Code:    "too_short",
		Message: "password must be at least 8 characters",
	}

	if err.Field != "password" {
		t.Errorf("Field = %s, want password", err.Field)
	}
	if err.Code != "too_short" {
		t.Errorf("Code = %s, want too_short", err.Code)
	}
	if err.Message != "password must be at least 8 characters" {
		t.Errorf("Message = %s, want password must be at least 8 characters", err.Message)
	}
}

func TestValidationErrorsError(t *testing.T) {
	tests := []struct {
		name   string
		errors ValidationErrors
		want   string
	}{
		{
			name:   "empty",
			errors: ValidationErrors{},
			want:   "validation failed",
		},
		{
			name: "single error",
			errors: ValidationErrors{
				{Field: "email", Code: "invalid", Message: "invalid email"},
			},
			want: "invalid email",
		},
		{
			name: "multiple errors",
			errors: ValidationErrors{
				{Field: "email", Code: "invalid", Message: "first error"},
				{Field: "password", Code: "weak", Message: "second error"},
			},
			want: "first error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errors.Error(); got != tt.want {
				t.Errorf("Error() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestValidationErrorsHasErrors(t *testing.T) {
	tests := []struct {
		name   string
		errors ValidationErrors
		want   bool
	}{
		{
			name:   "empty",
			errors: ValidationErrors{},
			want:   false,
		},
		{
			name: "has errors",
			errors: ValidationErrors{
				{Field: "email", Code: "invalid", Message: "invalid email"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errors.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationErrorImplementsError(t *testing.T) {
	var _ error = ValidationError{}
	var _ error = ValidationErrors{}
}
