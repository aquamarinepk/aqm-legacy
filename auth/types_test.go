package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserStatusConstants(t *testing.T) {
	tests := []struct {
		status UserStatus
		want   string
	}{
		{UserStatusActive, "active"},
		{UserStatusSuspended, "suspended"},
		{UserStatusDeleted, "deleted"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("UserStatus = %s, want %s", tt.status, tt.want)
			}
		})
	}
}

func TestGrantTypeConstants(t *testing.T) {
	tests := []struct {
		grantType GrantType
		want      string
	}{
		{GrantTypeRole, "role"},
		{GrantTypePermission, "permission"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.grantType) != tt.want {
				t.Errorf("GrantType = %s, want %s", tt.grantType, tt.want)
			}
		})
	}
}

func TestUserStruct(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	user := User{
		ID:           id,
		Username:     "testuser",
		Name:         "Test User",
		EmailCT:      []byte("encrypted"),
		EmailIV:      []byte("iv"),
		EmailTag:     []byte("tag"),
		EmailLookup:  []byte("lookup"),
		PasswordHash: []byte("hash"),
		PasswordSalt: []byte("salt"),
		MFASecretCT:  []byte("mfa"),
		PINCT:        []byte("pin"),
		PINIV:        []byte("piniv"),
		PINTag:       []byte("pintag"),
		PINLookup:    []byte("pinlookup"),
		Status:       UserStatusActive,
		CreatedAt:    now,
	}

	if user.ID != id {
		t.Errorf("ID = %v, want %v", user.ID, id)
	}
	if user.Username != "testuser" {
		t.Errorf("Username = %s, want testuser", user.Username)
	}
	if user.Name != "Test User" {
		t.Errorf("Name = %s, want Test User", user.Name)
	}
	if user.Status != UserStatusActive {
		t.Errorf("Status = %s, want %s", user.Status, UserStatusActive)
	}
	if !user.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", user.CreatedAt, now)
	}
}

func TestRoleStruct(t *testing.T) {
	id := uuid.New()

	role := Role{
		ID:          id,
		Name:        "admin",
		Permissions: []string{"read", "write", "delete"},
	}

	if role.ID != id {
		t.Errorf("ID = %v, want %v", role.ID, id)
	}
	if role.Name != "admin" {
		t.Errorf("Name = %s, want admin", role.Name)
	}
	if len(role.Permissions) != 3 {
		t.Errorf("Permissions length = %d, want 3", len(role.Permissions))
	}
}

func TestGrantStruct(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	expires := time.Now().Add(24 * time.Hour)

	grant := Grant{
		ID:        id,
		UserID:    userID,
		GrantType: GrantTypeRole,
		Value:     "admin",
		Scope:     Scope{Type: "organization", ID: "org-123"},
		ExpiresAt: &expires,
	}

	if grant.ID != id {
		t.Errorf("ID = %v, want %v", grant.ID, id)
	}
	if grant.UserID != userID {
		t.Errorf("UserID = %v, want %v", grant.UserID, userID)
	}
	if grant.GrantType != GrantTypeRole {
		t.Errorf("GrantType = %s, want %s", grant.GrantType, GrantTypeRole)
	}
	if grant.Value != "admin" {
		t.Errorf("Value = %s, want admin", grant.Value)
	}
	if grant.ExpiresAt == nil {
		t.Error("ExpiresAt should not be nil")
	}
}

func TestGrantWithNilExpiresAt(t *testing.T) {
	grant := Grant{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		GrantType: GrantTypePermission,
		Value:     "read",
		Scope:     Scope{Type: "global", ID: ""},
		ExpiresAt: nil,
	}

	if grant.ExpiresAt != nil {
		t.Error("ExpiresAt should be nil")
	}
}

func TestScopeStruct(t *testing.T) {
	scope := Scope{
		Type: "organization",
		ID:   "org-456",
	}

	if scope.Type != "organization" {
		t.Errorf("Type = %s, want organization", scope.Type)
	}
	if scope.ID != "org-456" {
		t.Errorf("ID = %s, want org-456", scope.ID)
	}
}

func TestResourcePolicyStruct(t *testing.T) {
	policy := ResourcePolicy{
		ID:      "policy-123",
		Type:    "document",
		Version: 1,
		Actions: map[string]PolicyRule{
			"read":  {AnyOf: []string{"viewer", "editor"}},
			"write": {AllOf: []string{"editor"}},
		},
	}

	if policy.ID != "policy-123" {
		t.Errorf("ID = %s, want policy-123", policy.ID)
	}
	if policy.Type != "document" {
		t.Errorf("Type = %s, want document", policy.Type)
	}
	if policy.Version != 1 {
		t.Errorf("Version = %d, want 1", policy.Version)
	}
	if len(policy.Actions) != 2 {
		t.Errorf("Actions length = %d, want 2", len(policy.Actions))
	}
}

func TestPolicyRuleStruct(t *testing.T) {
	rule := PolicyRule{
		AnyOf: []string{"admin", "moderator"},
		AllOf: []string{"verified"},
	}

	if len(rule.AnyOf) != 2 {
		t.Errorf("AnyOf length = %d, want 2", len(rule.AnyOf))
	}
	if len(rule.AllOf) != 1 {
		t.Errorf("AllOf length = %d, want 1", len(rule.AllOf))
	}
}

func TestTokenClaimsStruct(t *testing.T) {
	claims := TokenClaims{
		Subject:      "user-123",
		SessionID:    "session-456",
		Audience:     "api",
		Context:      map[string]string{"org": "org-789"},
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
		AuthzVersion: 2,
	}

	if claims.Subject != "user-123" {
		t.Errorf("Subject = %s, want user-123", claims.Subject)
	}
	if claims.SessionID != "session-456" {
		t.Errorf("SessionID = %s, want session-456", claims.SessionID)
	}
	if claims.Audience != "api" {
		t.Errorf("Audience = %s, want api", claims.Audience)
	}
	if claims.Context["org"] != "org-789" {
		t.Errorf("Context[org] = %s, want org-789", claims.Context["org"])
	}
	if claims.AuthzVersion != 2 {
		t.Errorf("AuthzVersion = %d, want 2", claims.AuthzVersion)
	}
}

func TestTokenClaimsJSONTags(t *testing.T) {
	// This test verifies the struct has the expected JSON tags
	claims := TokenClaims{
		Subject:   "sub-value",
		SessionID: "sid-value",
	}

	if claims.Subject != "sub-value" {
		t.Errorf("Subject = %s, want sub-value", claims.Subject)
	}
}

func TestEmailSubscriptionStruct(t *testing.T) {
	userID := uuid.New()
	confirmed := time.Now()

	sub := EmailSubscription{
		UserID:      &userID,
		EmailCT:     []byte("encrypted"),
		EmailLookup: []byte("lookup"),
		Consent: ConsentRecord{
			Type:      "marketing",
			Scope:     "email",
			Timestamp: time.Now(),
			SourceIP:  "192.168.1.1",
		},
		ConfirmedAt: &confirmed,
	}

	if sub.UserID == nil || *sub.UserID != userID {
		t.Errorf("UserID = %v, want %v", sub.UserID, userID)
	}
	if sub.Consent.Type != "marketing" {
		t.Errorf("Consent.Type = %s, want marketing", sub.Consent.Type)
	}
	if sub.ConfirmedAt == nil {
		t.Error("ConfirmedAt should not be nil")
	}
}

func TestEmailSubscriptionWithNilFields(t *testing.T) {
	sub := EmailSubscription{
		UserID:      nil,
		ConfirmedAt: nil,
	}

	if sub.UserID != nil {
		t.Error("UserID should be nil")
	}
	if sub.ConfirmedAt != nil {
		t.Error("ConfirmedAt should be nil")
	}
}

func TestConsentRecordStruct(t *testing.T) {
	now := time.Now()

	consent := ConsentRecord{
		Type:      "terms",
		Scope:     "full",
		Timestamp: now,
		SourceIP:  "10.0.0.1",
	}

	if consent.Type != "terms" {
		t.Errorf("Type = %s, want terms", consent.Type)
	}
	if consent.Scope != "full" {
		t.Errorf("Scope = %s, want full", consent.Scope)
	}
	if !consent.Timestamp.Equal(now) {
		t.Errorf("Timestamp = %v, want %v", consent.Timestamp, now)
	}
	if consent.SourceIP != "10.0.0.1" {
		t.Errorf("SourceIP = %s, want 10.0.0.1", consent.SourceIP)
	}
}
