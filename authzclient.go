package aqm

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aquamarinepk/aqm/auth"
)

// AuthzClient implements the auth.AuthzClient interface using ServiceClient.
type AuthzClient struct {
	client *ServiceClient
}

// NewAuthzClient creates a new authorization client.
func NewAuthzClient(baseURL string) *AuthzClient {
	return &AuthzClient{
		client: NewServiceClient(baseURL),
	}
}

// CheckPermission checks if a user has a specific permission on a resource.
func (c *AuthzClient) CheckPermission(ctx context.Context, userID, permission, resource string) (bool, error) {
	var scope map[string]interface{}
	if resource == "*" || resource == "" {
		scope = map[string]interface{}{
			"type": "global",
			"id":   "",
		}
	} else {
		scope = map[string]interface{}{
			"type": "resource",
			"id":   resource,
		}
	}

	requestBody := map[string]interface{}{
		"user_id":    userID,
		"permission": permission,
		"scope":      scope,
	}

	resp, err := c.client.Request(ctx, http.MethodPost, "/authz/policy/evaluate", requestBody)
	if err != nil {
		return false, fmt.Errorf("authz check failed: %w", err)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("invalid response format from authz service")
	}

	allowed, ok := data["allowed"].(bool)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'allowed' field in authz response")
	}

	return allowed, nil
}

// Ensure AuthzClient implements auth.AuthzClient interface
var _ auth.AuthzClient = (*AuthzClient)(nil)

// NewAuthzHelper creates a new authorization helper with caching.
// This wraps the AuthzClient with a cache layer to reduce load on the authz service.
// TODO: Eviction policies will be improved to provide more flexible cache invalidation strategies.
func NewAuthzHelper(client auth.AuthzClient, cacheTTL time.Duration) *auth.AuthzHelper {
	return auth.NewAuthzHelper(client, cacheTTL)
}
