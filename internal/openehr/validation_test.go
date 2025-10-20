package openehr

import "testing"

func TestValidateEHR(t *testing.T) {
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
