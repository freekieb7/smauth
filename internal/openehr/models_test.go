package openehr

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestMarshalUnmarshalCycle(t *testing.T) {
	data, err := os.ReadFile("../../tests/fixture/composition.json")
	if err != nil {
		t.Fatalf("Failed to read fixture file: %v", err)
	}

	// Unmarshal original data
	var comp COMPOSITION
	if err := json.Unmarshal(data, &comp); err != nil {
		t.Fatalf("Failed to unmarshal Composition: %v", err)
	}

	// Marshal and unmarshal again
	marshaledData, err := json.Marshal(comp)
	if err != nil {
		t.Fatalf("Failed to marshal Composition: %v", err)
	}

	var comp2 COMPOSITION
	if err := json.Unmarshal(marshaledData, &comp2); err != nil {
		t.Fatalf("Failed to unmarshal marshaled Composition: %v", err)
	}

	// Compare the structs directly
	if !compositionsEqual(comp, comp2) {
		t.Errorf("Compositions don't match after marshal/unmarshal cycle")

		// For debugging, show the JSON differences
		original, _ := json.MarshalIndent(comp, "", "  ")
		roundtrip, _ := json.MarshalIndent(comp2, "", "  ")
		t.Logf("Original:\n%s", string(original))
		t.Logf("Roundtrip:\n%s", string(roundtrip))
	}
}

func TestMarshalRequiredFields(t *testing.T) {
	ehr := EHR{
		EHRID: HIER_OBJECT_ID{Value: uuid.New().String()},
		EHRStatus: OBJECT_REF{
			Namespace: "local",
			Type:      "EHR_STATUS",
			ID:        ANY_OBJECT_ID{Value: HIER_OBJECT_ID{Value: uuid.New().String()}},
		},
		EHRAccess: OBJECT_REF{
			Namespace: "local",
			Type:      "EHR_ACCESS",
			ID:        ANY_OBJECT_ID{Value: HIER_OBJECT_ID{Value: uuid.New().String()}},
		},
		TimeCreated: DV_DATE_TIME{Value: "2023-01-01T00:00:00Z"},
	}

	data, err := json.Marshal(ehr)
	if err != nil {
		t.Fatalf("Failed to marshal EHR: %v", err)
	}

	// Parse as generic JSON to check structure
	var jsonData map[string]any
	if err := json.Unmarshal(data, &jsonData); err != nil {
		t.Fatalf("Failed to unmarshal as JSON: %v", err)
	}

	// Check required fields
	requiredFields := []string{"_type", "ehr_id", "ehr_status", "ehr_access", "time_created"}
	for _, field := range requiredFields {
		if _, exists := jsonData[field]; !exists {
			t.Errorf("Required field missing: %s", field)
		}
	}
}

func TestOptionalFieldMarshaling(t *testing.T) {
	// Test with minimal EHR (only required fields)
	ehr := EHR{
		EHRID: HIER_OBJECT_ID{Value: uuid.New().String()},
		EHRStatus: OBJECT_REF{
			Namespace: "local",
			Type:      "EHR_STATUS",
			ID:        ANY_OBJECT_ID{Value: HIER_OBJECT_ID{Value: uuid.New().String()}},
		},
		EHRAccess: OBJECT_REF{
			Namespace: "local",
			Type:      "EHR_ACCESS",
			ID:        ANY_OBJECT_ID{Value: HIER_OBJECT_ID{Value: uuid.New().String()}},
		},
		TimeCreated: DV_DATE_TIME{Value: "2023-01-01T00:00:00Z"},
	}

	data, err := json.Marshal(ehr)
	if err != nil {
		t.Fatalf("Failed to marshal EHR: %v", err)
	}

	var jsonData map[string]interface{}
	json.Unmarshal(data, &jsonData)

	// Optional fields should not be present or be null/empty
	optionalFields := []string{"system_id", "contributions", "compositions", "directory", "folders"}
	for _, field := range optionalFields {
		if value, exists := jsonData[field]; exists && value != nil {
			t.Errorf("Optional field %s should not be present when not set, got: %v", field, value)
		}
	}
}

func compositionsEqual(a, b COMPOSITION) bool {
	// Marshal both to JSON and compare
	aJson, err1 := json.Marshal(a)
	bJson, err2 := json.Marshal(b)

	if err1 != nil || err2 != nil {
		return false
	}

	return string(aJson) == string(bJson)
}
