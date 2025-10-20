package openehr

import "testing"

func TestValidateValidEHR(t *testing.T) {
	ehr := EHR{
		EHRID: HIER_OBJECT_ID{Value: "550e8400-e29b-41d4-a716-446655440000"},
		EHRStatus: OBJECT_REF{
			Namespace: "local",
			Type:      "VERSIONED_EHR_STATUS",
			ID:        ANY_OBJECT_ID{Value: HIER_OBJECT_ID{Value: "123e4567-e89b-12d3-a456-426614174000"}},
		},
		EHRAccess: OBJECT_REF{
			Namespace: "local",
			Type:      "VERSIONED_EHR_ACCESS",
			ID:        ANY_OBJECT_ID{Value: HIER_OBJECT_ID{Value: "987e6543-e21b-32d3-b456-426614174999"}},
		},
		TimeCreated: DV_DATE_TIME{Value: "2023-01-01T00:00:00Z"},
	}

	validator := Validator{}
	errors := validator.ValidateModel(ehr)
	if len(errors) != 0 {
		t.Errorf("Expected no validation errors, got %d", len(errors))
		for _, err := range errors {
			t.Logf("Validation error: %s", err.Message)
		}
	}
}

func TestValidateInvalidEHR(t *testing.T) {
	ehr := EHR{
		EHRID: HIER_OBJECT_ID{Value: "invalid-uuid."},
		EHRStatus: OBJECT_REF{
			Namespace: "local",
			Type:      "EHR_STATUS",
			ID:        ANY_OBJECT_ID{Value: HIER_OBJECT_ID{Value: "also-invalid-uuid"}},
		},
		EHRAccess: OBJECT_REF{
			Namespace: "local",
			Type:      "EHR_ACCESS",
			ID:        ANY_OBJECT_ID{Value: HIER_OBJECT_ID{Value: "yet-another-invalid-uuid"}},
		},
		TimeCreated: DV_DATE_TIME{Value: "2023-01-01T00:00:00Z"},
	}

	validator := Validator{}
	errors := validator.ValidateModel(ehr)
	if len(errors) != 3 {
		t.Errorf("Expected 3 validation errors, got %d", len(errors))
		for _, err := range errors {
			t.Logf("Validation error: %s", err.Message)
		}
	}
}
