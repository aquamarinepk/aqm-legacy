package auth

import (
	"context"
	"fmt"
	"time"
)

// AuthzClient interface for authorization service calls.
// Applications should implement this interface for their specific authz service.
type AuthzClient interface {
	CheckPermission(ctx context.Context, userID, permission, resource string) (bool, error)
}

// AuthzHelper provides transparent caching for authorization checks.
// This is a convenience wrapper around StringTTLCache for permission checks.
type AuthzHelper struct {
	client AuthzClient
	cache  *StringTTLCache[bool]
}

// NewAuthzHelper creates a new authorization helper with caching.
func NewAuthzHelper(client AuthzClient, cacheTTL time.Duration) *AuthzHelper {
	return &AuthzHelper{
		client: client,
		cache:  NewStringTTLCache[bool](cacheTTL),
	}
}

// CheckPermission checks if user has permission with transparent caching.
// This is the main function used throughout the application.
func (h *AuthzHelper) CheckPermission(ctx context.Context, userID, permission, resource string) (bool, error) {
	key := h.cacheKey(userID, permission, resource)

	// Try cache first
	if allowed, found := h.cache.Get(key); found {
		return allowed, nil
	}

	// Cache miss - call AuthZ service
	allowed, err := h.client.CheckPermission(ctx, userID, permission, resource)
	if err != nil {
		return false, err
	}

	// Cache the result
	h.cache.Set(key, allowed)

	return allowed, nil
}

// CheckMultiplePermissions checks multiple permissions efficiently.
// Returns map of permission results - useful for UI rendering.
func (h *AuthzHelper) CheckMultiplePermissions(ctx context.Context, userID string, checks []PermissionCheck) (map[string]bool, error) {
	results := make(map[string]bool)

	for _, check := range checks {
		allowed, err := h.CheckPermission(ctx, userID, check.Permission, check.Resource)
		if err != nil {
			return nil, fmt.Errorf("error check %s:%s: %w", check.Permission, check.Resource, err)
		}

		key := fmt.Sprintf("%s:%s", check.Permission, check.Resource)
		results[key] = allowed
	}

	return results, nil
}

// PermissionCheck represents a permission to check.
type PermissionCheck struct {
	Permission string
	Resource   string
}

// cacheKey generates a unique cache key for a permission check.
func (h *AuthzHelper) cacheKey(userID, permission, resource string) string {
	return fmt.Sprintf("%s:%s:%s", userID, permission, resource)
}

// ClearUserCache removes all cached permissions for a specific user.
// Useful when user permissions change.
func (h *AuthzHelper) ClearUserCache(userID string) {
	h.cache.DeleteByPrefix(userID + ":")
}

// ClearExpiredCache removes expired entries from cache.
// Should be called periodically to prevent memory leaks.
func (h *AuthzHelper) ClearExpiredCache() {
	h.cache.ClearExpired()
}

// Pure helper functions for common permission patterns

// HasAnyPermission checks if user has any of the specified permissions (OR logic)
func HasAnyPermission(ctx context.Context, helper *AuthzHelper, userID string, permissions []string, resource string) (bool, error) {
	for _, permission := range permissions {
		allowed, err := helper.CheckPermission(ctx, userID, permission, resource)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions checks if user has all specified permissions (AND logic)
func HasAllPermissions(ctx context.Context, helper *AuthzHelper, userID string, permissions []string, resource string) (bool, error) {
	for _, permission := range permissions {
		allowed, err := helper.CheckPermission(ctx, userID, permission, resource)
		if err != nil {
			return false, err
		}
		if !allowed {
			return false, nil
		}
	}
	return true, nil
}

// IsResourceOwner checks if user owns/created the resource
func IsResourceOwner(ctx context.Context, helper *AuthzHelper, userID, resourceID string) (bool, error) {
	return helper.CheckPermission(ctx, userID, "own", resourceID)
}
