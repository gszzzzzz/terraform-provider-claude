package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gszzzzzz/terraform-provider-claude/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestParseAllowedInferenceGeos(t *testing.T) {
	tests := []struct {
		name         string
		input        json.RawMessage
		expectNull   bool
		expectValues []string
	}{
		{
			name:       "nil input",
			input:      nil,
			expectNull: true,
		},
		{
			name:         "string unrestricted",
			input:        json.RawMessage(`"unrestricted"`),
			expectNull:   false,
			expectValues: []string{"unrestricted"},
		},
		{
			name:         "array of geos",
			input:        json.RawMessage(`["us","eu"]`),
			expectNull:   false,
			expectValues: []string{"us", "eu"},
		},
		{
			name:         "single element array",
			input:        json.RawMessage(`["us"]`),
			expectNull:   false,
			expectValues: []string{"us"},
		},
		{
			name:         "empty array",
			input:        json.RawMessage(`[]`),
			expectNull:   false,
			expectValues: []string{},
		},
		{
			name:       "invalid JSON",
			input:      json.RawMessage(`not-json`),
			expectNull: true,
		},
		{
			name:       "JSON number",
			input:      json.RawMessage(`42`),
			expectNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := parseAllowedInferenceGeos(context.Background(), tt.input)
			if diags.HasError() {
				t.Fatalf("parseAllowedInferenceGeos returned errors: %v", diags)
			}

			if tt.expectNull {
				if !result.IsNull() {
					t.Errorf("expected null list, got %v", result)
				}
				return
			}

			if result.IsNull() {
				t.Fatal("expected non-null list, got null")
			}

			var elems []string
			diags = result.ElementsAs(context.Background(), &elems, false)
			if diags.HasError() {
				t.Fatalf("ElementsAs failed: %v", diags)
			}

			if len(elems) != len(tt.expectValues) {
				t.Fatalf("length = %d, want %d", len(elems), len(tt.expectValues))
			}
			for i, v := range tt.expectValues {
				if elems[i] != v {
					t.Errorf("elems[%d] = %q, want %q", i, elems[i], v)
				}
			}
		})
	}
}

func TestNormalizeAllowedGeosForAPI(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		expectStr   string
		expectSlice []string
	}{
		{
			name:      "unrestricted becomes string",
			input:     []string{"unrestricted"},
			expectStr: "unrestricted",
		},
		{
			name:        "multiple geos stays array",
			input:       []string{"us", "eu"},
			expectSlice: []string{"us", "eu"},
		},
		{
			name:        "single non-unrestricted stays array",
			input:       []string{"us"},
			expectSlice: []string{"us"},
		},
		{
			name:        "empty slice stays array",
			input:       []string{},
			expectSlice: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeAllowedGeosForAPI(tt.input)

			if tt.expectStr != "" {
				str, ok := result.(string)
				if !ok {
					t.Fatalf("expected string, got %T", result)
				}
				if str != tt.expectStr {
					t.Errorf("result = %q, want %q", str, tt.expectStr)
				}
				return
			}

			slice, ok := result.([]string)
			if !ok {
				t.Fatalf("expected []string, got %T", result)
			}
			if len(slice) != len(tt.expectSlice) {
				t.Fatalf("length = %d, want %d", len(slice), len(tt.expectSlice))
			}
			for i, v := range tt.expectSlice {
				if slice[i] != v {
					t.Errorf("slice[%d] = %q, want %q", i, slice[i], v)
				}
			}
		})
	}
}

func ptrString(s string) *string {
	return &s
}

func TestFlattenWorkspace(t *testing.T) {
	tests := []struct {
		name       string
		workspace  *client.Workspace
		checkModel func(t *testing.T, m workspaceResourceModel)
	}{
		{
			name: "basic workspace",
			workspace: &client.Workspace{
				ID:           "ws-1",
				Name:         "test",
				DisplayColor: "blue",
				CreatedAt:    "2024-01-01T00:00:00Z",
				ArchivedAt:   nil,
			},
			checkModel: func(t *testing.T, m workspaceResourceModel) {
				if m.ID.ValueString() != "ws-1" {
					t.Errorf("ID = %q, want %q", m.ID.ValueString(), "ws-1")
				}
				if m.Name.ValueString() != "test" {
					t.Errorf("Name = %q, want %q", m.Name.ValueString(), "test")
				}
				if m.DisplayColor.ValueString() != "blue" {
					t.Errorf("DisplayColor = %q, want %q", m.DisplayColor.ValueString(), "blue")
				}
				if !m.ArchivedAt.IsNull() {
					t.Errorf("ArchivedAt should be null, got %q", m.ArchivedAt.ValueString())
				}
				if m.DataResidency != nil {
					t.Errorf("DataResidency should be nil")
				}
			},
		},
		{
			name: "archived workspace",
			workspace: &client.Workspace{
				ID:           "ws-2",
				Name:         "archived",
				DisplayColor: "gray",
				CreatedAt:    "2024-01-01T00:00:00Z",
				ArchivedAt:   ptrString("2024-06-01T00:00:00Z"),
			},
			checkModel: func(t *testing.T, m workspaceResourceModel) {
				if m.ArchivedAt.IsNull() {
					t.Fatal("ArchivedAt should not be null")
				}
				if m.ArchivedAt.ValueString() != "2024-06-01T00:00:00Z" {
					t.Errorf("ArchivedAt = %q, want %q", m.ArchivedAt.ValueString(), "2024-06-01T00:00:00Z")
				}
			},
		},
		{
			name: "data residency with string geos",
			workspace: &client.Workspace{
				ID:           "ws-3",
				Name:         "dr-string",
				DisplayColor: "green",
				CreatedAt:    "2024-01-01T00:00:00Z",
				DataResidency: &client.DataResidency{
					WorkspaceGeo:         "us",
					DefaultInferenceGeo:  "us",
					AllowedInferenceGeos: json.RawMessage(`"unrestricted"`),
				},
			},
			checkModel: func(t *testing.T, m workspaceResourceModel) {
				if m.DataResidency == nil {
					t.Fatal("DataResidency should not be nil")
				}
				if m.DataResidency.WorkspaceGeo.ValueString() != "us" {
					t.Errorf("WorkspaceGeo = %q, want %q", m.DataResidency.WorkspaceGeo.ValueString(), "us")
				}
				var elems []string
				m.DataResidency.AllowedInferenceGeos.ElementsAs(context.Background(), &elems, false)
				if len(elems) != 1 || elems[0] != "unrestricted" {
					t.Errorf("AllowedInferenceGeos = %v, want [unrestricted]", elems)
				}
			},
		},
		{
			name: "data residency with array geos",
			workspace: &client.Workspace{
				ID:           "ws-4",
				Name:         "dr-array",
				DisplayColor: "red",
				CreatedAt:    "2024-01-01T00:00:00Z",
				DataResidency: &client.DataResidency{
					WorkspaceGeo:         "eu",
					DefaultInferenceGeo:  "eu",
					AllowedInferenceGeos: json.RawMessage(`["us","eu"]`),
				},
			},
			checkModel: func(t *testing.T, m workspaceResourceModel) {
				if m.DataResidency == nil {
					t.Fatal("DataResidency should not be nil")
				}
				var elems []string
				m.DataResidency.AllowedInferenceGeos.ElementsAs(context.Background(), &elems, false)
				if len(elems) != 2 || elems[0] != "us" || elems[1] != "eu" {
					t.Errorf("AllowedInferenceGeos = %v, want [us eu]", elems)
				}
			},
		},
		{
			name: "data residency with nil geos",
			workspace: &client.Workspace{
				ID:           "ws-5",
				Name:         "dr-nil",
				DisplayColor: "yellow",
				CreatedAt:    "2024-01-01T00:00:00Z",
				DataResidency: &client.DataResidency{
					WorkspaceGeo:         "us",
					DefaultInferenceGeo:  "us",
					AllowedInferenceGeos: nil,
				},
			},
			checkModel: func(t *testing.T, m workspaceResourceModel) {
				if m.DataResidency == nil {
					t.Fatal("DataResidency should not be nil")
				}
				if !m.DataResidency.AllowedInferenceGeos.IsNull() {
					t.Errorf("AllowedInferenceGeos should be null")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, diags := flattenWorkspace(context.Background(), tt.workspace)
			if diags.HasError() {
				t.Fatalf("flattenWorkspace returned errors: %v", diags)
			}
			tt.checkModel(t, model)
		})
	}
}

func TestNormalizeAllowedGeosForAPI_TypeAssertions(t *testing.T) {
	// Verify the return type is correct for each case
	strResult := normalizeAllowedGeosForAPI([]string{"unrestricted"})
	if _, ok := strResult.(string); !ok {
		t.Errorf("expected string type for unrestricted, got %T", strResult)
	}

	sliceResult := normalizeAllowedGeosForAPI([]string{"us", "eu"})
	if _, ok := sliceResult.([]string); !ok {
		t.Errorf("expected []string type for multiple geos, got %T", sliceResult)
	}
}

// Verify types.List behavior for edge cases
func TestParseAllowedInferenceGeos_ListProperties(t *testing.T) {
	// Non-null result should have correct element type
	result, diags := parseAllowedInferenceGeos(context.Background(), json.RawMessage(`["us"]`))
	if diags.HasError() {
		t.Fatalf("parseAllowedInferenceGeos returned errors: %v", diags)
	}
	if result.IsNull() {
		t.Fatal("expected non-null list")
	}
	if result.IsUnknown() {
		t.Fatal("expected known list")
	}
	if result.ElementType(context.Background()) != types.StringType {
		t.Errorf("element type = %v, want StringType", result.ElementType(context.Background()))
	}
}
