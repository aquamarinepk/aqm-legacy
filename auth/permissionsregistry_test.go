package auth

import (
	"testing"
)

func TestPermissionConstants(t *testing.T) {
	tests := []struct {
		perm Permission
		want string
	}{
		{PermSystemAdmin, "system:admin"},
		{PermSystemConfig, "system:config"},
		{PermUsersRead, "users:read"},
		{PermUsersWrite, "users:write"},
		{PermUsersDelete, "users:delete"},
		{PermRolesRead, "roles:read"},
		{PermRolesWrite, "roles:write"},
		{PermRolesDelete, "roles:delete"},
		{PermRolesManage, "roles:manage"},
		{PermGrantsRead, "grants:read"},
		{PermGrantsWrite, "grants:write"},
		{PermGrantsDelete, "grants:delete"},
		{PermGrantsManage, "grants:manage"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.perm) != tt.want {
				t.Errorf("Permission = %s, want %s", tt.perm, tt.want)
			}
		})
	}
}

func TestPermissionRegistryCategories(t *testing.T) {
	expectedCategories := []string{"System", "Authentication", "Authorization", "Access Control"}

	if len(PermissionRegistry) != len(expectedCategories) {
		t.Errorf("PermissionRegistry has %d categories, want %d", len(PermissionRegistry), len(expectedCategories))
	}

	for i, cat := range PermissionRegistry {
		if cat.Name != expectedCategories[i] {
			t.Errorf("Category %d name = %s, want %s", i, cat.Name, expectedCategories[i])
		}
	}
}

func TestPermissionRegistrySystemCategory(t *testing.T) {
	cat := PermissionRegistry[0]

	if cat.Name != "System" {
		t.Errorf("Name = %s, want System", cat.Name)
	}
	if len(cat.Permissions) != 2 {
		t.Errorf("System category has %d permissions, want 2", len(cat.Permissions))
	}
}

func TestPermissionInfoFields(t *testing.T) {
	info := PermissionRegistry[0].Permissions[0]

	if info.Code != PermSystemAdmin {
		t.Errorf("Code = %s, want %s", info.Code, PermSystemAdmin)
	}
	if info.Name != "System Administrator" {
		t.Errorf("Name = %s, want System Administrator", info.Name)
	}
	if info.Description != "Full system access" {
		t.Errorf("Description = %s, want Full system access", info.Description)
	}
}

func TestAllPermissions(t *testing.T) {
	perms := AllPermissions()

	if len(perms) == 0 {
		t.Error("AllPermissions should return non-empty slice")
	}

	// Count expected permissions
	expected := 0
	for _, cat := range PermissionRegistry {
		expected += len(cat.Permissions)
	}

	if len(perms) != expected {
		t.Errorf("AllPermissions returned %d permissions, want %d", len(perms), expected)
	}

	// Verify some known permissions are present
	found := make(map[Permission]bool)
	for _, p := range perms {
		found[p] = true
	}

	expectedPerms := []Permission{
		PermSystemAdmin,
		PermUsersRead,
		PermRolesWrite,
		PermGrantsManage,
	}

	for _, p := range expectedPerms {
		if !found[p] {
			t.Errorf("Permission %s not found in AllPermissions()", p)
		}
	}
}

func TestAllPermissionStrings(t *testing.T) {
	strings := AllPermissionStrings()

	if len(strings) == 0 {
		t.Error("AllPermissionStrings should return non-empty slice")
	}

	perms := AllPermissions()
	if len(strings) != len(perms) {
		t.Errorf("AllPermissionStrings returned %d strings, want %d", len(strings), len(perms))
	}

	// Verify strings match permissions
	for i, s := range strings {
		if s != string(perms[i]) {
			t.Errorf("strings[%d] = %s, want %s", i, s, perms[i])
		}
	}
}

func TestGetPermissionInfo(t *testing.T) {
	tests := []struct {
		code     Permission
		wantName string
		wantNil  bool
	}{
		{PermSystemAdmin, "System Administrator", false},
		{PermUsersRead, "Read Users", false},
		{PermRolesManage, "Manage Roles", false},
		{Permission("nonexistent"), "", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			info := GetPermissionInfo(tt.code)

			if tt.wantNil {
				if info != nil {
					t.Error("GetPermissionInfo should return nil for unknown permission")
				}
				return
			}

			if info == nil {
				t.Fatal("GetPermissionInfo returned nil")
			}
			if info.Name != tt.wantName {
				t.Errorf("Name = %s, want %s", info.Name, tt.wantName)
			}
			if info.Code != tt.code {
				t.Errorf("Code = %s, want %s", info.Code, tt.code)
			}
		})
	}
}

func TestGetPermissionInfoDescription(t *testing.T) {
	info := GetPermissionInfo(PermUsersWrite)

	if info == nil {
		t.Fatal("GetPermissionInfo returned nil")
	}
	if info.Description != "Create and update users" {
		t.Errorf("Description = %s, want Create and update users", info.Description)
	}
}

func TestPermissionCategoryStruct(t *testing.T) {
	cat := PermissionCategory{
		Name: "Custom",
		Permissions: []PermissionInfo{
			{Code: "custom:read", Name: "Read", Description: "Read access"},
		},
	}

	if cat.Name != "Custom" {
		t.Errorf("Name = %s, want Custom", cat.Name)
	}
	if len(cat.Permissions) != 1 {
		t.Errorf("Permissions length = %d, want 1", len(cat.Permissions))
	}
}

func TestPermissionType(t *testing.T) {
	var p Permission = "test:permission"

	if string(p) != "test:permission" {
		t.Errorf("Permission string conversion failed")
	}
}
