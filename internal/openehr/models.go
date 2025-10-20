package openehr

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/freekieb7/smauth/internal/openehr/terminology"
	"github.com/freekieb7/smauth/internal/util"
)

type OpenEHRType interface {
	SetMetaType()
	Validate() []ValidationError
}

// Helper struct for extracting _type field
type TypeExtractor struct {
	MetaType string `json:"_type,omitzero"`
}

// -----------------------------------
// EHR
// -----------------------------------

const EHR_META_TYPE string = "EHR"

type EHR struct {
	MetaType      util.Optional[string]         `json:"_type,omitzero"`
	SystemID      util.Optional[HIER_OBJECT_ID] `json:"system_id,omitzero"`
	EHRID         HIER_OBJECT_ID                `json:"ehr_id"`
	Contributions util.Optional[[]OBJECT_REF]   `json:"contributions,omitzero"`
	EHRStatus     OBJECT_REF                    `json:"ehr_status"`
	EHRAccess     OBJECT_REF                    `json:"ehr_access"`
	Compositions  util.Optional[[]OBJECT_REF]   `json:"compositions,omitzero"`
	// Directory     util.Optional[OBJECT_REF]     `json:"directory,omitzero"`
	TimeCreated DV_DATE_TIME                `json:"time_created"`
	Folders     util.Optional[[]OBJECT_REF] `json:"folders,omitzero"`
	Tags        util.Optional[[]OBJECT_REF] `json:"tags,omitzero"`
}

func (e EHR) Validate() []ValidationError {
	var errors []ValidationError

	if e.MetaType.IsSet() && e.MetaType.V != EHR_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          EHR_META_TYPE,
			Path:           "._type",
			Message:        fmt.Sprintf("invalid EHR _type field: %s", e.MetaType.V),
			Recommendation: "Ensure the _type field is set to 'EHR'",
		})
	}

	if e.EHRStatus.Type != VERSIONED_EHR_STATUS_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          EHR_META_TYPE,
			Path:           ".ehr_status",
			Message:        fmt.Sprintf("invalid EHR status type: %s", e.EHRStatus.Type),
			Recommendation: fmt.Sprintf("Ensure ehr_status _type field is set to '%s'", VERSIONED_EHR_STATUS_META_TYPE),
		})
	}

	if e.EHRAccess.Type != VERSIONED_EHR_ACCESS_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          EHR_META_TYPE,
			Path:           ".ehr_access",
			Message:        fmt.Sprintf("invalid EHR access type: %s", e.EHRAccess.Type),
			Recommendation: fmt.Sprintf("Ensure ehr_access _type field is set to '%s'", VERSIONED_EHR_ACCESS_META_TYPE),
		})
	}

	if e.Compositions.IsSet() {
		for i, compRef := range e.Compositions.V {
			if compRef.Type != VERSIONED_COMPOSITION_META_TYPE {
				errors = append(errors, ValidationError{
					Model:          EHR_META_TYPE,
					Path:           fmt.Sprintf(".compositions[%d]", i),
					Message:        fmt.Sprintf("invalid composition type: %s", compRef.Type),
					Recommendation: fmt.Sprintf("Ensure compositions[%d] _type field is set to '%s'", i, VERSIONED_COMPOSITION_META_TYPE),
				})
			}
		}
	}

	if e.Contributions.IsSet() {
		for i, contribRef := range e.Contributions.V {
			if contribRef.Type != CONTRIBUTION_META_TYPE {
				errors = append(errors, ValidationError{
					Model:          EHR_META_TYPE,
					Path:           fmt.Sprintf(".contributions[%d]", i),
					Message:        fmt.Sprintf("invalid contribution type: %s", contribRef.Type),
					Recommendation: fmt.Sprintf("Ensure contributions[%d] _type field is set to '%s'", i, CONTRIBUTION_META_TYPE),
				})
			}
		}
	}

	if e.Folders.IsSet() {
		for i, folderRef := range e.Folders.V {
			if folderRef.Type != VERSIONED_FOLDER_META_TYPE {
				errors = append(errors, ValidationError{
					Model:          EHR_META_TYPE,
					Path:           fmt.Sprintf(".folders[%d]", i),
					Message:        fmt.Sprintf("invalid folder type: %s", folderRef.Type),
					Recommendation: fmt.Sprintf("Ensure folders[%d] _type field is set to '%s'", i, VERSIONED_FOLDER_META_TYPE),
				})
			}
		}
	}

	return errors
}

func (e *EHR) SetMetaType() {
	e.MetaType = util.Some(EHR_META_TYPE)
}

const VERSIONED_EHR_ACCESS_META_TYPE string = "VERSIONED_EHR_ACCESS"

type VERSIONED_EHR_ACCESS struct {
	MetaType    util.Optional[string] `json:"_type,omitzero"`
	UID         HIER_OBJECT_ID        `json:"uid"`
	OwnerID     OBJECT_REF            `json:"owner_id"`
	TimeCreated DV_DATE_TIME          `json:"time_created"`
}

func (e VERSIONED_EHR_ACCESS) Validate() []ValidationError {
	var errors []ValidationError

	if e.MetaType.IsSet() && e.MetaType.V != VERSIONED_EHR_ACCESS_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          VERSIONED_EHR_ACCESS_META_TYPE,
			Path:           "._type",
			Message:        fmt.Sprintf("invalid VERSIONED_EHR_ACCESS _type field: %s", e.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", VERSIONED_EHR_ACCESS_META_TYPE),
		})
	}

	return errors
}

func (e *VERSIONED_EHR_ACCESS) SetMetaType() {
	e.MetaType = util.Some(VERSIONED_EHR_ACCESS_META_TYPE)
}

const EHR_ACCESS_META_TYPE string = "EHR_ACCESS"

type EHR_ACCESS struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	// Settings         util.Optional[ACCESS_CONTROL_SETTINGS] `json:"settings,omitzero"`
}

func (e EHR_ACCESS) Validate() []ValidationError {
	var errors []ValidationError

	if e.MetaType.IsSet() && e.MetaType.V != EHR_ACCESS_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          EHR_ACCESS_META_TYPE,
			Path:           "._type",
			Message:        fmt.Sprintf("invalid EHR_ACCESS _type field: %s", e.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", EHR_ACCESS_META_TYPE),
		})
	}

	return errors
}

func (e *EHR_ACCESS) SetMetaType() {
	e.MetaType = util.Some(EHR_ACCESS_META_TYPE)
}

const VERSIONED_EHR_STATUS_META_TYPE string = "VERSIONED_EHR_STATUS"

type VERSIONED_EHR_STATUS struct {
	MetaType    util.Optional[string] `json:"_type,omitzero"`
	UID         HIER_OBJECT_ID        `json:"uid"`
	OwnerID     OBJECT_REF            `json:"owner_id"`
	TimeCreated DV_DATE_TIME          `json:"time_created"`
}

func (e VERSIONED_EHR_STATUS) Validate() []ValidationError {
	var errors []ValidationError

	if e.MetaType.IsSet() && e.MetaType.V != VERSIONED_EHR_STATUS_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          VERSIONED_EHR_STATUS_META_TYPE,
			Path:           "._type",
			Message:        fmt.Sprintf("invalid VERSIONED_EHR_STATUS _type field: %s", e.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", VERSIONED_EHR_STATUS_META_TYPE),
		})
	}

	return errors
}

func (e *VERSIONED_EHR_STATUS) SetMetaType() {
	e.MetaType = util.Some(VERSIONED_EHR_STATUS_META_TYPE)
}

const EHR_STATUS_META_TYPE string = "EHR_STATUS"

type EHR_STATUS struct {
	MetaType         util.Optional[string]             `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                       `json:"name"`
	ArchetypeNodeID  string                            `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID]   `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]             `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]         `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]       `json:"feeder_audit,omitzero"`
	Subject          PARTY_SELF                        `json:"subject"`
	IsQueryable      bool                              `json:"is_queryable"`
	IsModifiable     bool                              `json:"is_modifiable"`
	OtherDetails     util.Optional[ANY_ITEM_STRUCTURE] `json:"other_details,omitzero"`
}

func (e EHR_STATUS) Validate() []ValidationError {
	var errors []ValidationError

	if e.MetaType.IsSet() && e.MetaType.V != EHR_STATUS_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          EHR_STATUS_META_TYPE,
			Path:           "._type",
			Message:        fmt.Sprintf("invalid EHR_STATUS _type field: %s", e.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", EHR_STATUS_META_TYPE),
		})
	}

	return errors
}

func (e *EHR_STATUS) SetMetaType() {
	e.MetaType = util.Some(EHR_STATUS_META_TYPE)
}

const VERSIONED_COMPOSITION_META_TYPE string = "VERSIONED_COMPOSITION"

type VERSIONED_COMPOSITION struct {
	MetaType    util.Optional[string] `json:"_type,omitzero"`
	UID         HIER_OBJECT_ID        `json:"uid"`
	OwnerID     OBJECT_REF            `json:"owner_id"`
	TimeCreated DV_DATE_TIME          `json:"time_created"`
}

func (e VERSIONED_COMPOSITION) Validate() []ValidationError {
	var errors []ValidationError

	if e.MetaType.IsSet() && e.MetaType.V != VERSIONED_COMPOSITION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          VERSIONED_COMPOSITION_META_TYPE,
			Path:           "._type",
			Message:        fmt.Sprintf("invalid VERSIONED_COMPOSITION _type field: %s", e.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", VERSIONED_COMPOSITION_META_TYPE),
		})
	}

	return errors
}

func (e *VERSIONED_COMPOSITION) SetMetaType() {
	e.MetaType = util.Some(VERSIONED_COMPOSITION_META_TYPE)
}

const COMPOSITION_META_TYPE string = "COMPOSITION"

type COMPOSITION struct {
	MetaType         util.Optional[string]             `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                       `json:"name"`
	ArchetypeNodeID  string                            `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID]   `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]             `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]         `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]       `json:"feeder_audit,omitzero"`
	Language         CODE_PHRASE                       `json:"language"`
	Territory        CODE_PHRASE                       `json:"territory"`
	Category         DV_CODED_TEXT                     `json:"category"`
	Context          util.Optional[EVENT_CONTEXT]      `json:"context,omitzero"`
	Composer         ANY_PARTY_PROXY                   `json:"composer"`
	Content          util.Optional[[]ANY_CONTENT_ITEM] `json:"content,omitzero"`
}

func (c COMPOSITION) Validate() []ValidationError {
	var errors []ValidationError

	// Validate _type field
	if c.MetaType.IsSet() && c.MetaType.V != COMPOSITION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          COMPOSITION_META_TYPE,
			Path:           "._type",
			Message:        fmt.Sprintf("invalid COMPOSITION _type field: %s", c.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", COMPOSITION_META_TYPE),
		})
	}

	// Validate archetype_details field
	if c.ArchetypeDetails.IsSet() {
		if c.ArchetypeDetails.V.ArchetypeID.Value != c.ArchetypeNodeID {
			errors = append(errors, ValidationError{
				Model:          COMPOSITION_META_TYPE,
				Path:           ".archetype_details.archetype_id.value",
				Message:        fmt.Sprintf("invalid COMPOSITION archetype field: %s", c.ArchetypeDetails.V.ArchetypeID.Value),
				Recommendation: fmt.Sprintf("Ensure archetype ID is set to '%s'", c.ArchetypeNodeID),
			})
		}
	}

	// Validate language field
	if !terminology.IsValidLanguageTerminologyID(c.Language.TerminologyId.Value) {
		errors = append(errors, ValidationError{
			Model:          COMPOSITION_META_TYPE,
			Path:           ".language.terminology_id.value",
			Message:        fmt.Sprintf("invalid COMPOSITION language field: %s", c.Language.TerminologyId.Value),
			Recommendation: "Ensure language terminology ID is valid according to openEHR standards",
		})
	}

	if !terminology.IsValidLanguageCode(c.Language.CodeString) {
		errors = append(errors, ValidationError{
			Model:          COMPOSITION_META_TYPE,
			Path:           ".language.code_string",
			Message:        fmt.Sprintf("invalid COMPOSITION language field: %s", c.Language.CodeString),
			Recommendation: "Ensure language code string is valid according to openEHR standards",
		})
	}

	// Validate territory field
	if !terminology.IsValidCountryTerminologyID(c.Territory.TerminologyId.Value) {
		errors = append(errors, ValidationError{
			Model:          COMPOSITION_META_TYPE,
			Path:           ".territory.terminology_id.value",
			Message:        fmt.Sprintf("invalid COMPOSITION territory field: %s", c.Territory.TerminologyId.Value),
			Recommendation: "Ensure territory terminology ID is valid according to openEHR standards",
		})
	}

	if !terminology.IsValidCountryCode(c.Territory.CodeString) {
		errors = append(errors, ValidationError{
			Model:          COMPOSITION_META_TYPE,
			Path:           ".territory",
			Message:        fmt.Sprintf("invalid COMPOSITION territory field: %s", c.Territory.CodeString),
			Recommendation: "Ensure territory code string is valid according to openEHR standards",
		})
	}

	// Validate category field
	if !terminology.IsValidCompositionCategoryTerminologyID(c.Category.DefiningCode.TerminologyId.Value) {
		errors = append(errors, ValidationError{
			Model:          COMPOSITION_META_TYPE,
			Path:           ".category.",
			Message:        fmt.Sprintf("invalid COMPOSITION category field: %s", c.Category.DefiningCode.TerminologyId.Value),
			Recommendation: "Ensure category terminology ID is 'openehr'",
		})
	}

	if !terminology.IsValidCompositionCategoryCode(c.Category.DefiningCode.CodeString) {
		errors = append(errors, ValidationError{
			Model:          COMPOSITION_META_TYPE,
			Path:           ".category",
			Message:        fmt.Sprintf("invalid COMPOSITION category field: %s", c.Category.DefiningCode.CodeString),
			Recommendation: "Ensure category code string is valid according to openEHR composition_category terminology",
		})
	}

	if c.Category.Value != terminology.GetCompositionCategoryNameCode(c.Category.DefiningCode.CodeString) {
		errors = append(errors, ValidationError{
			Model:          COMPOSITION_META_TYPE,
			Path:           ".category.value",
			Message:        fmt.Sprintf("invalid COMPOSITION category field: %s", c.Category.Value),
			Recommendation: "Ensure category value is valid according to openEHR composition_category terminology",
		})
	}

	return errors
}

func (e *COMPOSITION) SetMetaType() {
	e.MetaType = util.Some(COMPOSITION_META_TYPE)
}

const EVENT_CONTEXT_META_TYPE string = "EVENT_CONTEXT"

type EVENT_CONTEXT struct {
	MetaType           util.Optional[string]             `json:"_type,omitzero"`
	StartTime          DV_DATE_TIME                      `json:"start_time"`
	EndTime            util.Optional[DV_DATE_TIME]       `json:"end_time,omitzero"`
	Location           util.Optional[string]             `json:"location,omitzero"`
	Setting            DV_CODED_TEXT                     `json:"setting"`
	OtherContext       util.Optional[ANY_ITEM_STRUCTURE] `json:"other_context,omitzero"`
	HealthCareFacility util.Optional[PARTY_IDENTIFIED]   `json:"health_care_facility,omitzero"`
	Participations     util.Optional[[]PARTICIPATION]    `json:"participations,omitzero"`
}

func (e EVENT_CONTEXT) Validate() []ValidationError {
	var errors []ValidationError

	// Validate _type field
	if e.MetaType.IsSet() && e.MetaType.V != EVENT_CONTEXT_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "EVENT_CONTEXT",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid EVENT_CONTEXT _type field: %s", e.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", EVENT_CONTEXT_META_TYPE),
		})
	}

	// Validate setting field
	if !terminology.IsValidSettingTerminologyID(e.Setting.DefiningCode.TerminologyId.Value) {
		errors = append(errors, ValidationError{
			Model:          "EVENT_CONTEXT",
			Path:           ".setting.defining_code.terminology_id.value",
			Message:        fmt.Sprintf("invalid EVENT_CONTEXT setting field: %s", e.Setting.DefiningCode.TerminologyId.Value),
			Recommendation: "Ensure setting terminology ID is 'openehr'",
		})
	}

	if !terminology.IsValidSettingCode(e.Setting.DefiningCode.CodeString) {
		errors = append(errors, ValidationError{
			Model:          "EVENT_CONTEXT",
			Path:           ".setting.defining_code.code_string",
			Message:        fmt.Sprintf("invalid EVENT_CONTEXT setting field: %s", e.Setting.DefiningCode.CodeString),
			Recommendation: "Ensure setting code string is valid according to openEHR setting terminology",
		})
	}

	return errors
}

func (e *EVENT_CONTEXT) SetMetaType() {
	e.MetaType = util.Some(EVENT_CONTEXT_META_TYPE)
}

const CONTENT_ITEM_META_TYPE string = "CONTENT_ITEM"

// Abstract type
type CONTENT_ITEM struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
}

func (c CONTENT_ITEM) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract CONTENT_ITEM type")
}

func (c *CONTENT_ITEM) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract CONTENT_ITEM type")
}

type ANY_CONTENT_ITEM struct {
	Value any
}

// Implement UnionType interface
func (c ANY_CONTENT_ITEM) GetBaseType() reflect.Type {
	return reflect.TypeFor[CONTENT_ITEM]()
}

func (c ANY_CONTENT_ITEM) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Value)
}

func (c *ANY_CONTENT_ITEM) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case SECTION_META_TYPE:
		c.Value = new(SECTION)
	case ADMIN_ENTRY_META_TYPE:
		c.Value = new(ADMIN_ENTRY)
	case OBSERVATION_META_TYPE:
		c.Value = new(OBSERVATION)
	case EVALUATION_META_TYPE:
		c.Value = new(EVALUATION)
	case INSTRUCTION_META_TYPE:
		c.Value = new(INSTRUCTION)
	case ACTIVITY_META_TYPE:
		c.Value = new(ACTIVITY)
	case ACTION_META_TYPE:
		c.Value = new(ACTION)
	case GENERIC_ENTRY_META_TYPE:
		c.Value = new(GENERIC_ENTRY)
	case "":
		return fmt.Errorf("missing CONTENT_ITEM _type field")
	default:
		return fmt.Errorf("invalid CONTENT_ITEM type: %s", t)
	}

	return json.Unmarshal(data, c.Value)
}

const SECTION_META_TYPE string = "SECTION"

type SECTION struct {
	MetaType         util.Optional[string]             `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                       `json:"name"`
	ArchetypeNodeID  string                            `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID]   `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]             `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]         `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]       `json:"feeder_audit,omitzero"`
	Items            util.Optional[[]ANY_CONTENT_ITEM] `json:"items,omitzero"`
}

func (c SECTION) Validate() []ValidationError {
	var errors []ValidationError

	if c.MetaType.IsSet() && c.MetaType.V != SECTION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:   "SECTION",
			Path:    "._type",
			Message: fmt.Sprintf("invalid SECTION _type field: %s", c.MetaType.V),
		})
	}

	return errors
}

func (c *SECTION) SetMetaType() {
	c.MetaType = util.Some(SECTION_META_TYPE)
}

const GENERIC_ENTRY_META_TYPE string = "GENERIC_ENTRY"

type GENERIC_ENTRY struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	Data             ANY_ITEM                        `json:"data"`
}

func (c GENERIC_ENTRY) Validate() []ValidationError {
	var errors []ValidationError

	if c.MetaType.IsSet() && c.MetaType.V != GENERIC_ENTRY_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "GENERIC_ENTRY",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid GENERIC_ENTRY _type field: %s", c.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", GENERIC_ENTRY_META_TYPE),
		})
	}

	return errors
}

func (c *GENERIC_ENTRY) SetMetaType() {
	c.MetaType = util.Some(GENERIC_ENTRY_META_TYPE)
}

const ENTRY_META_TYPE string = "ENTRY"

// Abstract type
type ENTRY struct {
	MetaType            util.Optional[string]           `json:"_type,omitzero"`
	Name                ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID     string                          `json:"archetype_node_id"`
	UID                 util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links               util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails    util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit         util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	Language            CODE_PHRASE                     `json:"language"`
	Encoding            CODE_PHRASE                     `json:"encoding"`
	OtherParticipations util.Optional[[]PARTICIPATION]  `json:"other_participations,omitzero"`
	WorkflowID          util.Optional[OBJECT_REF]       `json:"workflow_id,omitzero"`
	Subject             ANY_PARTY_PROXY                 `json:"subject"`
	Provider            util.Optional[ANY_PARTY_PROXY]  `json:"provider,omitzero"`
}

func (c ENTRY) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract ENTRY type")
}

func (c *ENTRY) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract ENTRY type")
}

type ANY_ENTRY struct {
	Value any
}

// Implement UnionType interface
func (c ANY_ENTRY) GetBaseType() reflect.Type {
	return reflect.TypeFor[ENTRY]()
}

func (c ANY_ENTRY) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Value)
}

func (c *ANY_ENTRY) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case ADMIN_ENTRY_META_TYPE:
		c.Value = new(ADMIN_ENTRY)
	case OBSERVATION_META_TYPE:
		c.Value = new(OBSERVATION)
	case EVALUATION_META_TYPE:
		c.Value = new(EVALUATION)
	case INSTRUCTION_META_TYPE:
		c.Value = new(INSTRUCTION)
	case ACTIVITY_META_TYPE:
		c.Value = new(ACTIVITY)
	case ACTION_META_TYPE:
		c.Value = new(ACTION)
	case "":
		return fmt.Errorf("missing ENTRY _type field")
	default:
		return fmt.Errorf("invalid ENTRY type: %s", t)
	}

	return json.Unmarshal(data, c.Value)
}

const ADMIN_ENTRY_META_TYPE string = "ADMIN_ENTRY"

type ADMIN_ENTRY struct {
	MetaType            util.Optional[string]           `json:"_type,omitzero"`
	Name                ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID     string                          `json:"archetype_node_id"`
	UID                 util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links               util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails    util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit         util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	Language            CODE_PHRASE                     `json:"language"`
	Encoding            CODE_PHRASE                     `json:"encoding"`
	OtherParticipations util.Optional[[]PARTICIPATION]  `json:"other_participations,omitzero"`
	WorkflowID          util.Optional[OBJECT_REF]       `json:"workflow_id,omitzero"`
	Subject             ANY_PARTY_PROXY                 `json:"subject"`
	Provider            util.Optional[ANY_PARTY_PROXY]  `json:"provider,omitzero"`
	Data                ANY_ITEM_STRUCTURE              `json:"data"`
}

func (c ADMIN_ENTRY) Validate() []ValidationError {
	var errors []ValidationError

	if c.MetaType.IsSet() && c.MetaType.V != ADMIN_ENTRY_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ADMIN_ENTRY",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ADMIN_ENTRY _type field: %s", c.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ADMIN_ENTRY_META_TYPE),
		})
	}

	return errors
}

func (c *ADMIN_ENTRY) SetMetaType() {
	c.MetaType = util.Some(ADMIN_ENTRY_META_TYPE)
}

const CARE_ENTRY_META_TYPE string = "CARE_ENTRY"

// Abstract type
type CARE_ENTRY struct {
	MetaType            util.Optional[string]             `json:"_type,omitzero"`
	Name                ANY_DV_TEXT                       `json:"name"`
	ArchetypeNodeID     string                            `json:"archetype_node_id"`
	UID                 util.Optional[ANY_UID_BASED_ID]   `json:"uid,omitzero"`
	Links               util.Optional[[]LINK]             `json:"links,omitzero"`
	ArchetypeDetails    util.Optional[ARCHETYPED]         `json:"archetype_details,omitzero"`
	FeederAudit         util.Optional[FEEDER_AUDIT]       `json:"feeder_audit,omitzero"`
	Language            CODE_PHRASE                       `json:"language"`
	Encoding            CODE_PHRASE                       `json:"encoding"`
	OtherParticipations util.Optional[[]PARTICIPATION]    `json:"other_participations,omitzero"`
	WorkflowID          util.Optional[OBJECT_REF]         `json:"workflow_id,omitzero"`
	Subject             ANY_PARTY_PROXY                   `json:"subject"`
	Provider            util.Optional[ANY_PARTY_PROXY]    `json:"provider,omitzero"`
	Protocol            util.Optional[ANY_ITEM_STRUCTURE] `json:"protocol,omitzero"`
	GuidelineID         util.Optional[OBJECT_REF]         `json:"guideline_id,omitzero"`
}

func (c CARE_ENTRY) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract CARE_ENTRY type")
}

func (c *CARE_ENTRY) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract CARE_ENTRY type")
}

type ANY_CARE_ENTRY_TYPE struct {
	Value any
}

// Implement UnionType interface
func (c ANY_CARE_ENTRY_TYPE) GetBaseType() reflect.Type {
	return reflect.TypeFor[CARE_ENTRY]()
}

func (c ANY_CARE_ENTRY_TYPE) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Value)
}

func (c *ANY_CARE_ENTRY_TYPE) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case OBSERVATION_META_TYPE:
		c.Value = new(OBSERVATION)
	case EVALUATION_META_TYPE:
		c.Value = new(EVALUATION)
	case INSTRUCTION_META_TYPE:
		c.Value = new(INSTRUCTION)
	case ACTIVITY_META_TYPE:
		c.Value = new(ACTIVITY)
	case ACTION_META_TYPE:
		c.Value = new(ACTION)
	case "":
		return fmt.Errorf("missing CARE_ENTRY _type field")
	default:
		return fmt.Errorf("CARE_ENTRY unexpected _type %s", t)
	}

	return json.Unmarshal(data, c.Value)
}

const OBSERVATION_META_TYPE string = "OBSERVATION"

type OBSERVATION struct {
	MetaType            util.Optional[string]                      `json:"_type,omitzero"`
	Name                ANY_DV_TEXT                                `json:"name"`
	ArchetypeNodeID     string                                     `json:"archetype_node_id"`
	UID                 util.Optional[ANY_UID_BASED_ID]            `json:"uid,omitzero"`
	Links               util.Optional[[]LINK]                      `json:"links,omitzero"`
	ArchetypeDetails    util.Optional[ARCHETYPED]                  `json:"archetype_details,omitzero"`
	FeederAudit         util.Optional[FEEDER_AUDIT]                `json:"feeder_audit,omitzero"`
	Language            CODE_PHRASE                                `json:"language"`
	Encoding            CODE_PHRASE                                `json:"encoding"`
	OtherParticipations util.Optional[[]PARTICIPATION]             `json:"other_participations,omitzero"`
	WorkflowID          util.Optional[OBJECT_REF]                  `json:"workflow_id,omitzero"`
	Subject             ANY_PARTY_PROXY                            `json:"subject"`
	Provider            util.Optional[ANY_PARTY_PROXY]             `json:"provider,omitzero"`
	Protocol            util.Optional[ANY_ITEM_STRUCTURE]          `json:"protocol,omitzero"`
	GuidelineID         util.Optional[OBJECT_REF]                  `json:"guideline_id,omitzero"`
	Data                HISTORY[ANY_ITEM_STRUCTURE]                `json:"data"`
	State               util.Optional[HISTORY[ANY_ITEM_STRUCTURE]] `json:"state,omitzero"`
}

func (o OBSERVATION) Validate() []ValidationError {
	var errors []ValidationError

	if o.MetaType.IsSet() && o.MetaType.V != OBSERVATION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "OBSERVATION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid OBSERVATION _type field: %s", o.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", OBSERVATION_META_TYPE),
		})
	}

	return errors
}

func (o *OBSERVATION) SetMetaType() {
	o.MetaType = util.Some(OBSERVATION_META_TYPE)
}

const EVALUATION_META_TYPE string = "EVALUATION"

type EVALUATION struct {
	MetaType            util.Optional[string]             `json:"_type,omitzero"`
	Name                ANY_DV_TEXT                       `json:"name"`
	ArchetypeNodeID     string                            `json:"archetype_node_id"`
	UID                 util.Optional[ANY_UID_BASED_ID]   `json:"uid,omitzero"`
	Links               util.Optional[[]LINK]             `json:"links,omitzero"`
	ArchetypeDetails    util.Optional[ARCHETYPED]         `json:"archetype_details,omitzero"`
	FeederAudit         util.Optional[FEEDER_AUDIT]       `json:"feeder_audit,omitzero"`
	Language            CODE_PHRASE                       `json:"language"`
	Encoding            CODE_PHRASE                       `json:"encoding"`
	OtherParticipations util.Optional[[]PARTICIPATION]    `json:"other_participations,omitzero"`
	WorkflowID          util.Optional[OBJECT_REF]         `json:"workflow_id,omitzero"`
	Subject             ANY_PARTY_PROXY                   `json:"subject"`
	Provider            util.Optional[ANY_PARTY_PROXY]    `json:"provider,omitzero"`
	Protocol            util.Optional[ANY_ITEM_STRUCTURE] `json:"protocol,omitzero"`
	GuidelineID         util.Optional[OBJECT_REF]         `json:"guideline_id,omitzero"`
	Data                ANY_ITEM_STRUCTURE                `json:"data"`
}

func (e EVALUATION) Validate() []ValidationError {
	var errors []ValidationError

	if e.MetaType.IsSet() && e.MetaType.V != EVALUATION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "EVALUATION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid EVALUATION _type field: %s", e.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", EVALUATION_META_TYPE),
		})
	}

	return errors
}

func (e *EVALUATION) SetMetaType() {
	e.MetaType = util.Some(EVALUATION_META_TYPE)
}

const INSTRUCTION_META_TYPE string = "INSTRUCTION"

type INSTRUCTION struct {
	MetaType            util.Optional[string]             `json:"_type,omitzero"`
	Name                ANY_DV_TEXT                       `json:"name"`
	ArchetypeNodeID     string                            `json:"archetype_node_id"`
	UID                 util.Optional[ANY_UID_BASED_ID]   `json:"uid,omitzero"`
	Links               util.Optional[[]LINK]             `json:"links,omitzero"`
	ArchetypeDetails    util.Optional[ARCHETYPED]         `json:"archetype_details,omitzero"`
	FeederAudit         util.Optional[FEEDER_AUDIT]       `json:"feeder_audit,omitzero"`
	Language            CODE_PHRASE                       `json:"language"`
	Encoding            CODE_PHRASE                       `json:"encoding"`
	OtherParticipations util.Optional[[]PARTICIPATION]    `json:"other_participations,omitzero"`
	WorkflowID          util.Optional[OBJECT_REF]         `json:"workflow_id,omitzero"`
	Subject             ANY_PARTY_PROXY                   `json:"subject"`
	Provider            util.Optional[ANY_PARTY_PROXY]    `json:"provider,omitzero"`
	Protocol            util.Optional[ANY_ITEM_STRUCTURE] `json:"protocol,omitzero"`
	GuidelineID         util.Optional[OBJECT_REF]         `json:"guideline_id,omitzero"`
	Narrative           ANY_DV_TEXT                       `json:"narrative"`
	ExpiryTime          util.Optional[DV_DATE_TIME]       `json:"expiry_time,omitzero"`
	WFDefinition        util.Optional[DV_PARSABLE]        `json:"wf_definition,omitzero"`
	Activities          util.Optional[[]ACTIVITY]         `json:"activities,omitzero"`
}

func (i INSTRUCTION) Validate() []ValidationError {
	var errors []ValidationError

	if i.MetaType.IsSet() && i.MetaType.V != INSTRUCTION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "INSTRUCTION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid INSTRUCTION _type field: %s", i.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", INSTRUCTION_META_TYPE),
		})
	}

	return errors
}

func (i *INSTRUCTION) SetMetaType() {
	i.MetaType = util.Some(INSTRUCTION_META_TYPE)
}

const ACTIVITY_META_TYPE string = "ACTIVITY"

type ACTIVITY struct {
	MetaType          util.Optional[string]           `json:"_type,omitzero"`
	Name              ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID   string                          `json:"archetype_node_id"`
	UID               util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links             util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails  util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit       util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	Timing            util.Optional[DV_PARSABLE]      `json:"timing,omitzero"`
	ActionArchetypeID string                          `json:"action_archetype_id"`
	Description       ANY_ITEM_STRUCTURE              `json:"description"`
}

func (a ACTIVITY) Validate() []ValidationError {
	var errors []ValidationError

	if a.MetaType.IsSet() && a.MetaType.V != ACTIVITY_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ACTIVITY",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ACTIVITY _type field: %s", a.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ACTIVITY_META_TYPE),
		})
	}

	return errors
}

func (a *ACTIVITY) SetMetaType() {
	a.MetaType = util.Some(ACTIVITY_META_TYPE)
}

const ACTION_META_TYPE string = "ACTION"

type ACTION struct {
	MetaType            util.Optional[string]              `json:"_type,omitzero"`
	Name                ANY_DV_TEXT                        `json:"name"`
	ArchetypeNodeID     string                             `json:"archetype_node_id"`
	UID                 util.Optional[ANY_UID_BASED_ID]    `json:"uid,omitzero"`
	Links               util.Optional[[]LINK]              `json:"links,omitzero"`
	ArchetypeDetails    util.Optional[ARCHETYPED]          `json:"archetype_details,omitzero"`
	FeederAudit         util.Optional[FEEDER_AUDIT]        `json:"feeder_audit,omitzero"`
	Language            CODE_PHRASE                        `json:"language"`
	Encoding            CODE_PHRASE                        `json:"encoding"`
	OtherParticipations util.Optional[[]PARTICIPATION]     `json:"other_participations,omitzero"`
	WorkflowID          util.Optional[OBJECT_REF]          `json:"workflow_id,omitzero"`
	Subject             ANY_PARTY_PROXY                    `json:"subject"`
	Provider            util.Optional[ANY_PARTY_PROXY]     `json:"provider,omitzero"`
	Protocol            util.Optional[ANY_ITEM_STRUCTURE]  `json:"protocol,omitzero"`
	GuidelineID         util.Optional[OBJECT_REF]          `json:"guideline_id,omitzero"`
	Time                DV_DATE_TIME                       `json:"time"`
	IsmTransition       ISM_TRANSITION                     `json:"ism_transition"`
	InstructionDetails  util.Optional[INSTRUCTION_DETAILS] `json:"instruction_details,omitzero"`
	Description         ANY_ITEM_STRUCTURE                 `json:"description"`
}

func (a ACTION) Validate() []ValidationError {
	var errors []ValidationError

	if a.MetaType.IsSet() && a.MetaType.V != ACTION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ACTION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ACTION _type field: %s", a.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ACTION_META_TYPE),
		})
	}

	return errors
}

func (a *ACTION) SetMetaType() {
	a.MetaType = util.Some(ACTION_META_TYPE)
}

const INSTRUCTION_DETAILS_META_TYPE string = "INSTRUCTION_DETAILS"

type INSTRUCTION_DETAILS struct {
	MetaType      util.Optional[string]             `json:"_type,omitzero"`
	InstructionID LOCATABLE_REF                     `json:"instruction_id"`
	ActivityID    string                            `json:"activity"`
	WfDetails     util.Optional[ANY_ITEM_STRUCTURE] `json:"wf_details,omitzero"`
}

func (i INSTRUCTION_DETAILS) Validate() []ValidationError {
	var errors []ValidationError

	if i.MetaType.IsSet() && i.MetaType.V != INSTRUCTION_DETAILS_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "INSTRUCTION_DETAILS",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid INSTRUCTION_DETAILS _type field: %s", i.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", INSTRUCTION_DETAILS_META_TYPE),
		})
	}

	return errors
}

func (i *INSTRUCTION_DETAILS) SetMetaType() {
	i.MetaType = util.Some(INSTRUCTION_DETAILS_META_TYPE)
}

const ISM_TRANSITION_META_TYPE string = "ISM_TRANSITION"

type ISM_TRANSITION struct {
	MetaType     util.Optional[string]        `json:"_type,omitzero"`
	CurrentState DV_CODED_TEXT                `json:"current_state"`
	Transition   util.Optional[DV_CODED_TEXT] `json:"transition,omitzero"`
	CareflowStep util.Optional[DV_CODED_TEXT] `json:"cateflow_step,omitzero"`
	Reason       util.Optional[ANY_DV_TEXT]   `json:"reason,omitzero"`
}

func (i ISM_TRANSITION) Validate() []ValidationError {
	var errors []ValidationError

	if i.MetaType.IsSet() && i.MetaType.V != ISM_TRANSITION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ISM_TRANSITION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ISM_TRANSITION _type field: %s", i.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ISM_TRANSITION_META_TYPE),
		})
	}

	return errors
}

func (i *ISM_TRANSITION) SetMetaType() {
	i.MetaType = util.Some(ISM_TRANSITION_META_TYPE)
}

// -----------------------------------
// COMMON
// -----------------------------------

const PATHABLE_META_TYPE string = "PATHABLE"

// Abstract type
type PATHABLE struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
}

func (p PATHABLE) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract PATHABLE type")
}

func (p *PATHABLE) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract PATHABLE type")
}

type ANY_PATHABLE struct {
	Value any
}

// Implement UnionType interface
func (c ANY_PATHABLE) GetBaseType() reflect.Type {
	return reflect.TypeFor[PATHABLE]()
}

func (p ANY_PATHABLE) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Value)
}

func (p *ANY_PATHABLE) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case EHR_ACCESS_META_TYPE:
		p.Value = new(EHR_ACCESS)
	case EHR_STATUS_META_TYPE:
		p.Value = new(EHR_STATUS)
	case COMPOSITION_META_TYPE:
		p.Value = new(COMPOSITION)
	case SECTION_META_TYPE:
		p.Value = new(SECTION)
	case ADMIN_ENTRY_META_TYPE:
		p.Value = new(ADMIN_ENTRY)
	case OBSERVATION_META_TYPE:
		p.Value = new(OBSERVATION)
	case EVALUATION_META_TYPE:
		p.Value = new(EVALUATION)
	case INSTRUCTION_META_TYPE:
		p.Value = new(INSTRUCTION)
	case ACTIVITY_META_TYPE:
		p.Value = new(ACTIVITY)
	case ACTION_META_TYPE:
		p.Value = new(ACTION)
	case FOLDER_META_TYPE:
		p.Value = new(FOLDER)
	case ITEM_SINGLE_META_TYPE:
		p.Value = new(ITEM_SINGLE)
	case ITEM_LIST_META_TYPE:
		p.Value = new(ITEM_LIST)
	case ITEM_TABLE_META_TYPE:
		p.Value = new(ITEM_TABLE)
	case ITEM_TREE_META_TYPE:
		p.Value = new(ITEM_TREE)
	case CLUSTER_META_TYPE:
		p.Value = new(CLUSTER)
	case ELEMENT_META_TYPE:
		p.Value = new(ELEMENT)
	case HISTORY_META_TYPE:
		p.Value = new(HISTORY[any])
	case POINT_EVENT_META_TYPE:
		p.Value = new(POINT_EVENT[any])
	case "":
		return fmt.Errorf("missing PATHABLE _type field")
	default:
		return fmt.Errorf("invalid PATHABLE _type: %s", t)
	}

	return json.Unmarshal(data, p.Value)
}

const LOCATABLE_META_TYPE string = "LOCATABLE"

// Abstract type
type LOCATABLE struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
}

func (l LOCATABLE) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract LOCATABLE type")
}

func (l *LOCATABLE) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract LOCATABLE type")
}

type ANY_LOCATABLE struct {
	Value any
}

// Implement UnionType interface
func (l ANY_LOCATABLE) GetBaseType() reflect.Type {
	return reflect.TypeFor[LOCATABLE]()
}

func (l ANY_LOCATABLE) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Value)
}

func (l *ANY_LOCATABLE) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case EHR_ACCESS_META_TYPE:
		l.Value = new(EHR_ACCESS)
	case EHR_STATUS_META_TYPE:
		l.Value = new(EHR_STATUS)
	case COMPOSITION_META_TYPE:
		l.Value = new(COMPOSITION)
	case SECTION_META_TYPE:
		l.Value = new(SECTION)
	case GENERIC_ENTRY_META_TYPE:
		l.Value = new(GENERIC_ENTRY)
	case ADMIN_ENTRY_META_TYPE:
		l.Value = new(ADMIN_ENTRY)
	case OBSERVATION_META_TYPE:
		l.Value = new(OBSERVATION)
	case EVALUATION_META_TYPE:
		l.Value = new(EVALUATION)
	case INSTRUCTION_META_TYPE:
		l.Value = new(INSTRUCTION)
	case ACTIVITY_META_TYPE:
		l.Value = new(ACTIVITY)
	case ACTION_META_TYPE:
		l.Value = new(ACTION)
	case FOLDER_META_TYPE:
		l.Value = new(FOLDER)
	case ITEM_SINGLE_META_TYPE:
		l.Value = new(ITEM_SINGLE)
	case ITEM_LIST_META_TYPE:
		l.Value = new(ITEM_LIST)
	case ITEM_TABLE_META_TYPE:
		l.Value = new(ITEM_TABLE)
	case ITEM_TREE_META_TYPE:
		l.Value = new(ITEM_TREE)
	case CLUSTER_META_TYPE:
		l.Value = new(CLUSTER)
	case ELEMENT_META_TYPE:
		l.Value = new(ELEMENT)
	case HISTORY_META_TYPE:
		l.Value = new(HISTORY[any])
	case POINT_EVENT_META_TYPE:
		l.Value = new(POINT_EVENT[any])
	case "":
		return fmt.Errorf("missing LOCATABLE _type field")
	default:
		return fmt.Errorf("invalid LOCATABLE _type %s", t)
	}

	return json.Unmarshal(data, l.Value)
}

const ARCHETYPED_META_TYPE string = "ARCHETYPED"

type ARCHETYPED struct {
	MetaType    util.Optional[string]      `json:"_type,omitzero"`
	ArchetypeID ARCHETYPE_ID               `json:"archetype_id"`
	TemplateID  util.Optional[TEMPLATE_ID] `json:"template_id,omitzero"`
	RMVersion   string                     `json:"rm_version"`
}

func (a ARCHETYPED) Validate() []ValidationError {
	var errors []ValidationError

	if a.MetaType.IsSet() && a.MetaType.V != ARCHETYPED_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ARCHETYPED",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ARCHETYPED _type field: %s", a.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ARCHETYPED_META_TYPE),
		})
	}

	return errors
}

func (a *ARCHETYPED) SetMetaType() {
	a.MetaType = util.Some(ARCHETYPED_META_TYPE)
}

const LINK_META_TYPE string = "LINK"

type LINK struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Meaning  ANY_DV_TEXT           `json:"meaning"`
	Type     ANY_DV_TEXT           `json:"type"`
	Target   DV_EHR_URI            `json:"target"`
}

func (l LINK) Validate() []ValidationError {
	var errors []ValidationError

	if l.MetaType.IsSet() && l.MetaType.V != LINK_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "LINK",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid LINK _type field: %s", l.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", LINK_META_TYPE),
		})
	}

	return errors
}

func (l *LINK) SetMetaType() {
	l.MetaType = util.Some(LINK_META_TYPE)
}

const FEEDER_AUDIT_META_TYPE string = "FEEDER_AUDIT"

type FEEDER_AUDIT struct {
	MetaType                 util.Optional[string]               `json:"_type,omitzero"`
	OriginatingSystemItemIDs util.Optional[[]DV_IDENTIFIER]      `json:"originating_system_item_ids,omitzero"`
	FeederSystemItemIDs      util.Optional[[]DV_IDENTIFIER]      `json:"feeder_system_item_ids,omitzero"`
	OriginalContent          util.Optional[ANY_DV_ENCAPSULATED]  `json:"original_content,omitzero"`
	OriginatingSystemAudit   FEEDER_AUDIT_DETAILS                `json:"originating_system_audit"`
	FeederSystemAudit        util.Optional[FEEDER_AUDIT_DETAILS] `json:"feeder_system_audit,omitzero"`
}

func (f FEEDER_AUDIT) Validate() []ValidationError {
	var errors []ValidationError

	if f.MetaType.IsSet() && f.MetaType.V != FEEDER_AUDIT_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "FEEDER_AUDIT",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid FEEDER_AUDIT _type field: %s", f.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", FEEDER_AUDIT_META_TYPE),
		})
	}

	return errors
}

func (f *FEEDER_AUDIT) SetMetaType() {
	f.MetaType = util.Some(FEEDER_AUDIT_META_TYPE)
}

const FEEDER_AUDIT_DETAILS_META_TYPE string = "FEEDER_AUDIT_DETAILS"

type FEEDER_AUDIT_DETAILS struct {
	MetaType     util.Optional[string]             `json:"_type,omitzero"`
	SystemID     string                            `json:"system_id"`
	Location     util.Optional[PARTY_IDENTIFIED]   `json:"location,omitzero"`
	Subject      util.Optional[ANY_PARTY_PROXY]    `json:"subject,omitzero"`
	Provider     util.Optional[PARTY_IDENTIFIED]   `json:"provider,omitzero"`
	Time         util.Optional[DV_DATE_TIME]       `json:"time,omitzero"`
	VersionID    util.Optional[string]             `json:"version_id,omitzero"`
	OtherDetails util.Optional[ANY_ITEM_STRUCTURE] `json:"other_details,omitzero"`
}

func (f FEEDER_AUDIT_DETAILS) Validate() []ValidationError {
	var errors []ValidationError

	if f.MetaType.IsSet() && f.MetaType.V != FEEDER_AUDIT_DETAILS_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "FEEDER_AUDIT_DETAILS",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid FEEDER_AUDIT_DETAILS _type field: %s", f.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", FEEDER_AUDIT_DETAILS_META_TYPE),
		})
	}

	return errors
}

func (f *FEEDER_AUDIT_DETAILS) SetMetaType() {
	f.MetaType = util.Some(FEEDER_AUDIT_DETAILS_META_TYPE)
}

const PARTY_PROXY_META_TYPE string = "PARTY_PROXY"

// Abstract type
type PARTY_PROXY struct {
	MetaType    util.Optional[string]    `json:"_type,omitzero"`
	ExternalRef util.Optional[PARTY_REF] `json:"external_ref,omitzero"`
}

func (p PARTY_PROXY) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract PARTY_PROXY type")
}

func (p *PARTY_PROXY) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract PARTY_PROXY type")
}

type ANY_PARTY_PROXY struct {
	Value any
}

// Implement UnionType interface
func (p ANY_PARTY_PROXY) GetBaseType() reflect.Type {
	return reflect.TypeFor[PARTY_PROXY]()
}

func (p ANY_PARTY_PROXY) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Value)
}

func (p *ANY_PARTY_PROXY) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case PARTY_SELF_META_TYPE:
		p.Value = new(PARTY_SELF)
	case PARTY_IDENTIFIED_META_TYPE:
		p.Value = new(PARTY_IDENTIFIED)
	case PARTY_RELATED_META_TYPE:
		p.Value = new(PARTY_RELATED)
	case "":
		return fmt.Errorf("missing PARTY_PROXY _type field")
	default:
		return fmt.Errorf("PARTY_PROXY unexpected _type %s", t)
	}

	return json.Unmarshal(data, p.Value)
}

const PARTY_SELF_META_TYPE string = "PARTY_SELF"

type PARTY_SELF struct {
	MetaType    util.Optional[string]    `json:"_type,omitzero"`
	ExternalRef util.Optional[PARTY_REF] `json:"external_ref,omitzero"`
}

func (p PARTY_SELF) Validate() []ValidationError {
	var errors []ValidationError

	if p.MetaType.IsSet() && p.MetaType.V != PARTY_SELF_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "PARTY_SELF",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid PARTY_SELF _type field: %s", p.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", PARTY_SELF_META_TYPE),
		})
	}

	return errors
}

func (p *PARTY_SELF) SetMetaType() {
	p.MetaType = util.Some(PARTY_SELF_META_TYPE)
}

const PARTY_IDENTIFIED_META_TYPE string = "PARTY_IDENTIFIED"

type PARTY_IDENTIFIED struct {
	MetaType    util.Optional[string]          `json:"_type,omitzero"`
	ExternalRef util.Optional[PARTY_REF]       `json:"external_ref,omitzero"`
	Name        util.Optional[string]          `json:"name,omitzero"`
	Identifiers util.Optional[[]DV_IDENTIFIER] `json:"identifiers,omitzero"`
}

func (p PARTY_IDENTIFIED) Validate() []ValidationError {
	var errors []ValidationError

	if p.MetaType.IsSet() && p.MetaType.V != PARTY_IDENTIFIED_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "PARTY_IDENTIFIED",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid PARTY_IDENTIFIED _type field: %s", p.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", PARTY_IDENTIFIED_META_TYPE),
		})
	}

	return errors
}

func (p *PARTY_IDENTIFIED) SetMetaType() {
	p.MetaType = util.Some(PARTY_IDENTIFIED_META_TYPE)
}

const PARTY_RELATED_META_TYPE string = "PARTY_RELATED"

type PARTY_RELATED struct {
	MetaType     util.Optional[string]          `json:"_type,omitzero"`
	ExternalRef  util.Optional[PARTY_REF]       `json:"external_ref,omitzero"`
	Name         util.Optional[string]          `json:"name,omitzero"`
	Identifiers  util.Optional[[]DV_IDENTIFIER] `json:"identifiers,omitzero"`
	Relationship DV_CODED_TEXT                  `json:"relationship"`
}

func (p PARTY_RELATED) Validate() []ValidationError {
	var errors []ValidationError

	if p.MetaType.IsSet() && p.MetaType.V != PARTY_RELATED_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "PARTY_RELATED",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid PARTY_RELATED _type field: %s", p.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", PARTY_RELATED_META_TYPE),
		})
	}

	return errors
}

func (p *PARTY_RELATED) SetMetaType() {
	p.MetaType = util.Some(PARTY_RELATED_META_TYPE)
}

const PARTICIPATION_META_TYPE string = "PARTICIPATION"

type PARTICIPATION struct {
	MetaType  util.Optional[string]        `json:"_type,omitzero"`
	Function  ANY_DV_TEXT                  `json:"function"`
	Mode      util.Optional[DV_CODED_TEXT] `json:"mode,omitzero"`
	Performer ANY_PARTY_PROXY              `json:"performer"`
	Time      util.Optional[DV_INTERVAL]   `json:"time,omitzero"`
}

func (p PARTICIPATION) Validate() []ValidationError {
	var errors []ValidationError

	if p.MetaType.IsSet() && p.MetaType.V != PARTICIPATION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "PARTICIPATION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid PARTICIPATION _type field: %s", p.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", PARTICIPATION_META_TYPE),
		})
	}

	return errors
}

func (p *PARTICIPATION) SetMetaType() {
	p.MetaType = util.Some(PARTICIPATION_META_TYPE)
}

const AUDIT_DETAILS_META_TYPE string = "AUDIT_DETAILS"

type AUDIT_DETAILS struct {
	MetaType      util.Optional[string]      `json:"_type,omitzero"`
	SystemID      string                     `json:"system_id"`
	TimeCommitted DV_DATE_TIME               `json:"time_committed"`
	ChangeType    DV_CODED_TEXT              `json:"change_type"`
	Description   util.Optional[ANY_DV_TEXT] `json:"description,omitzero"`
	Committer     ANY_PARTY_PROXY            `json:"committer"`
}

func (a AUDIT_DETAILS) Validate() []ValidationError {
	var errors []ValidationError

	if a.MetaType.IsSet() && a.MetaType.V != AUDIT_DETAILS_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "AUDIT_DETAILS",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid AUDIT_DETAILS _type field: %s", a.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", AUDIT_DETAILS_META_TYPE),
		})
	}

	return errors
}

func (a *AUDIT_DETAILS) SetMetaType() {
	a.MetaType = util.Some(AUDIT_DETAILS_META_TYPE)
}

const ATTESTATION_META_TYPE string = "ATTESTATION"

type ATTESTATION struct {
	MetaType      util.Optional[string]        `json:"_type,omitzero"`
	SystemID      string                       `json:"system_id"`
	TimeCommitted DV_DATE_TIME                 `json:"time_committed"`
	ChangeType    DV_CODED_TEXT                `json:"change_type"`
	Description   util.Optional[ANY_DV_TEXT]   `json:"description,omitzero"`
	Committer     ANY_PARTY_PROXY              `json:"committer"`
	AttestedView  util.Optional[DV_MULTIMEDIA] `json:"attested_view,omitzero"`
	Proof         util.Optional[string]        `json:"proof,omitzero"`
	Items         util.Optional[[]DV_EHR_URI]  `json:"items,omitzero"`
	Reason        ANY_DV_TEXT                  `json:"reason"`
	IsPending     bool                         `json:"is_pending"`
}

func (a ATTESTATION) Validate() []ValidationError {
	var errors []ValidationError

	if a.MetaType.IsSet() && a.MetaType.V != ATTESTATION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ATTESTATION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ATTESTATION _type field: %s", a.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ATTESTATION_META_TYPE),
		})
	}

	return errors
}

func (a *ATTESTATION) SetMetaType() {
	a.MetaType = util.Some(ATTESTATION_META_TYPE)
}

const REVISION_HISTORY_META_TYPE string = "REVISION_HISTORY"

type REVISION_HISTORY struct {
	MetaType util.Optional[string]   `json:"_type,omitzero"`
	Items    []REVISION_HISTORY_ITEM `json:"items"`
}

func (r REVISION_HISTORY) Validate() []ValidationError {
	var errors []ValidationError

	if r.MetaType.IsSet() && r.MetaType.V != REVISION_HISTORY_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "REVISION_HISTORY",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid REVISION_HISTORY _type field: %s", r.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", REVISION_HISTORY_META_TYPE),
		})
	}

	return errors
}

func (r *REVISION_HISTORY) SetMetaType() {
	r.MetaType = util.Some(REVISION_HISTORY_META_TYPE)
}

const REVISION_HISTORY_ITEM_META_TYPE string = "REVISION_HISTORY_ITEM"

type REVISION_HISTORY_ITEM struct {
	MetaType  util.Optional[string] `json:"_type,omitzero"`
	VersionID OBJECT_VERSION_ID     `json:"version_id"`
	Audits    []AUDIT_DETAILS       `json:"audits"`
}

func (r REVISION_HISTORY_ITEM) Validate() []ValidationError {
	var errors []ValidationError

	if r.MetaType.IsSet() && r.MetaType.V != REVISION_HISTORY_ITEM_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "REVISION_HISTORY_ITEM",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid REVISION_HISTORY_ITEM _type field: %s", r.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", REVISION_HISTORY_ITEM_META_TYPE),
		})
	}

	return errors
}

func (r *REVISION_HISTORY_ITEM) SetMetaType() {
	r.MetaType = util.Some(REVISION_HISTORY_ITEM_META_TYPE)
}

const VERSIONED_FOLDER_META_TYPE string = "VERSIONED_FOLDER"

type VERSIONED_FOLDER struct {
	MetaType    util.Optional[string] `json:"_type,omitzero"`
	UID         HIER_OBJECT_ID        `json:"uid"`
	OwnerID     OBJECT_REF            `json:"owner_id"`
	TimeCreated DV_DATE_TIME          `json:"time_created"`
}

func (v VERSIONED_FOLDER) Validate() []ValidationError {
	var errors []ValidationError

	if v.MetaType.IsSet() && v.MetaType.V != VERSIONED_FOLDER_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "VERSIONED_FOLDER",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid VERSIONED_FOLDER _type field: %s", v.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", VERSIONED_FOLDER_META_TYPE),
		})
	}

	return errors
}

func (v *VERSIONED_FOLDER) SetMetaType() {
	v.MetaType = util.Some(VERSIONED_FOLDER_META_TYPE)
}

const FOLDER_META_TYPE string = "FOLDER"

type FOLDER struct {
	MetaType         util.Optional[string]             `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                       `json:"name"`
	ArchetypeNodeID  string                            `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID]   `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]             `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]         `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]       `json:"feeder_audit,omitzero"`
	Items            util.Optional[[]OBJECT_REF]       `json:"items,omitzero"`
	Folders          util.Optional[[]FOLDER]           `json:"folders,omitzero"`
	Details          util.Optional[ANY_ITEM_STRUCTURE] `json:"details,omitzero"`
}

func (f FOLDER) Validate() []ValidationError {
	var errors []ValidationError

	if f.MetaType.IsSet() && f.MetaType.V != FOLDER_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "FOLDER",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid FOLDER _type field: %s", f.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", FOLDER_META_TYPE),
		})
	}

	return errors
}

func (f *FOLDER) SetMetaType() {
	f.MetaType = util.Some(FOLDER_META_TYPE)
}

const VERSIONED_OBJECT_META_TYPE string = "VERSIONED_OBJECT"

// Abstract type
type VERSIONED_OBJECT struct {
	MetaType    util.Optional[string] `json:"_type,omitzero"`
	UID         HIER_OBJECT_ID        `json:"uid"`
	OwnerID     OBJECT_REF            `json:"owner_id"`
	TimeCreated DV_DATE_TIME          `json:"time_created"`
}

func (v VERSIONED_OBJECT) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract VERSIONED_OBJECT type")
}

func (v *VERSIONED_OBJECT) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract VERSIONED_OBJECT type")
}

type ANY_VERSIONED_OBJECT struct {
	Value any
}

// Implement UnionType interface
func (c ANY_VERSIONED_OBJECT) GetBaseType() reflect.Type {
	return reflect.TypeFor[VERSIONED_OBJECT]()
}

func (v ANY_VERSIONED_OBJECT) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Value)
}

func (v *ANY_VERSIONED_OBJECT) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case VERSIONED_COMPOSITION_META_TYPE:
		v.Value = new(VERSIONED_COMPOSITION)
	case VERSIONED_EHR_STATUS_META_TYPE:
		v.Value = new(VERSIONED_EHR_STATUS)
	case VERSIONED_EHR_ACCESS_META_TYPE:
		v.Value = new(VERSIONED_EHR_ACCESS)
	case VERSIONED_FOLDER_META_TYPE:
		v.Value = new(VERSIONED_FOLDER)
	case VERSIONED_PARTY_META_TYPE:
		v.Value = new(VERSIONED_PARTY)
	case "":
		return fmt.Errorf("missing VERSIONED_OBJECT _type field")
	default:
		return fmt.Errorf("VERSIONED_OBJECT unexpected _type %s", t)
	}

	return json.Unmarshal(data, v.Value)
}

const VERSION_META_TYPE string = "VERSION"

// Abstract type
type VERSION struct {
	MetaType     util.Optional[string] `json:"_type,omitzero"`
	Contribution OBJECT_REF            `json:"contribution"`
	Signature    util.Optional[string] `json:"signature,omitzero"`
	CommitAudit  AUDIT_DETAILS         `json:"commit_audit"`
}

func (v VERSION) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract VERSION type")
}

func (v *VERSION) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract VERSION type")
}

type ANY_VERSION struct {
	Value any
}

// Implement UnionType interface
func (v ANY_VERSION) GetBaseType() reflect.Type {
	return reflect.TypeFor[VERSION]()
}

func (v ANY_VERSION) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Value)
}

func (v *ANY_VERSION) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case ORIGINAL_VERSION_META_TYPE:
		v.Value = new(ORIGINAL_VERSION)
	case IMPORTED_VERSION_META_TYPE:
		v.Value = new(IMPORTED_VERSION)
	case "":
		return fmt.Errorf("missing VERSION _type field")
	default:
		return fmt.Errorf("VERSION unexpected _type %s", t)
	}

	return json.Unmarshal(data, v.Value)
}

const ORIGINAL_VERSION_META_TYPE string = "ORIGINAL_VERSION"

type ORIGINAL_VERSION struct {
	MetaType              util.Optional[string]              `json:"_type,omitzero"`
	Contribution          OBJECT_REF                         `json:"contribution"`
	Signature             util.Optional[string]              `json:"signature,omitzero"`
	CommitAudit           AUDIT_DETAILS                      `json:"commit_audit"`
	UID                   OBJECT_VERSION_ID                  `json:"uid"`
	PrecedingVersionUID   util.Optional[OBJECT_VERSION_ID]   `json:"preceding_version_uid,omitzero"`
	OtherInputVersionUIDs util.Optional[[]OBJECT_VERSION_ID] `json:"other_input_version_uids,omitzero"`
	LifecycleState        DV_CODED_TEXT                      `json:"lifecycle_state"`
	Attestations          util.Optional[[]ATTESTATION]       `json:"attestations,omitzero"`
	Data                  any                                `json:"data"`
}

func (o ORIGINAL_VERSION) Validate() []ValidationError {
	var errors []ValidationError

	if o.MetaType.IsSet() && o.MetaType.V != ORIGINAL_VERSION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ORIGINAL_VERSION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ORIGINAL_VERSION _type field: %s", o.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ORIGINAL_VERSION_META_TYPE),
		})
	}

	return errors
}

func (o *ORIGINAL_VERSION) SetMetaType() {
	o.MetaType = util.Some(ORIGINAL_VERSION_META_TYPE)
}

const IMPORTED_VERSION_META_TYPE string = "IMPORTED_VERSION"

type IMPORTED_VERSION struct {
	MetaType     util.Optional[string] `json:"_type,omitzero"`
	Contribution OBJECT_REF            `json:"contribution"`
	Signature    util.Optional[string] `json:"signature,omitzero"`
	CommitAudit  AUDIT_DETAILS         `json:"commit_audit"`
	Item         ORIGINAL_VERSION      `json:"item"`
}

func (i IMPORTED_VERSION) Validate() []ValidationError {
	var errors []ValidationError

	if i.MetaType.IsSet() && i.MetaType.V != IMPORTED_VERSION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "IMPORTED_VERSION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid IMPORTED_VERSION _type field: %s", i.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", IMPORTED_VERSION_META_TYPE),
		})
	}

	return errors
}

func (i *IMPORTED_VERSION) SetMetaType() {
	i.MetaType = util.Some(IMPORTED_VERSION_META_TYPE)
}

const CONTRIBUTION_META_TYPE string = "CONTRIBUTION"

type CONTRIBUTION struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	UID      HIER_OBJECT_ID        `json:"uid"`
	Versions []OBJECT_REF          `json:"versions"`
	Audit    AUDIT_DETAILS         `json:"audit"`
}

func (c CONTRIBUTION) Validate() []ValidationError {
	var errors []ValidationError

	if c.MetaType.IsSet() && c.MetaType.V != CONTRIBUTION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "CONTRIBUTION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid CONTRIBUTION _type field: %s", c.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", CONTRIBUTION_META_TYPE),
		})
	}

	return errors
}

func (c *CONTRIBUTION) SetMetaType() {
	c.MetaType = util.Some(CONTRIBUTION_META_TYPE)
}

// idk what these are for yet
// pub const AUTHORED_RESOURCE = struct {};
// pub const TRANSLATION_DETAILS = struct {};
// pub const RESOURCE_DESCRIPTION = struct {};
// pub const RESOURCE_DESCRIPTION_ITEM = struct {};

// -----------------------------------
// DATA_STRUCTURES
// -----------------------------------

const ITEM_STRUCTURE_META_TYPE string = "ITEM_STRUCTURE"

// Abstract type
type ITEM_STRUCTURE struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[*FEEDER_AUDIT]    `json:"feeder_audit,omitzero"`
}

func (i ITEM_STRUCTURE) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract ITEM_STRUCTURE type")
}

func (i *ITEM_STRUCTURE) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract ITEM_STRUCTURE type")
}

type ANY_ITEM_STRUCTURE struct {
	Value any
}

// Implement UnionType interface
func (i ANY_ITEM_STRUCTURE) GetBaseType() reflect.Type {
	return reflect.TypeFor[ITEM_STRUCTURE]()
}

func (i ANY_ITEM_STRUCTURE) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Value)
}

func (i *ANY_ITEM_STRUCTURE) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case ITEM_SINGLE_META_TYPE:
		i.Value = new(ITEM_SINGLE)
	case ITEM_LIST_META_TYPE:
		i.Value = new(ITEM_LIST)
	case ITEM_TABLE_META_TYPE:
		i.Value = new(ITEM_TABLE)
	case ITEM_TREE_META_TYPE:
		i.Value = new(ITEM_TREE)
	case "":
		return fmt.Errorf("missing ITEM_STRUCTURE _type field")
	default:
		return fmt.Errorf("ITEM_STRUCTURE unexpected _type %s", t)
	}

	return json.Unmarshal(data, i.Value)
}

const ITEM_SINGLE_META_TYPE string = "ITEM_SINGLE"

type ITEM_SINGLE struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[LINK]             `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	Item             ELEMENT                         `json:"item"`
}

func (i ITEM_SINGLE) Validate() []ValidationError {
	var errors []ValidationError

	if i.MetaType.IsSet() && i.MetaType.V != ITEM_SINGLE_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ITEM_SINGLE",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ITEM_SINGLE _type field: %s", i.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ITEM_SINGLE_META_TYPE),
		})
	}

	return errors
}

func (i *ITEM_SINGLE) SetMetaType() {
	i.MetaType = util.Some(ITEM_SINGLE_META_TYPE)
}

const ITEM_LIST_META_TYPE string = "ITEM_LIST"

type ITEM_LIST struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	Items            util.Optional[[]ELEMENT]        `json:"items,omitzero"`
}

func (i ITEM_LIST) Validate() []ValidationError {
	var errors []ValidationError

	if i.MetaType.IsSet() && i.MetaType.V != ITEM_LIST_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ITEM_LIST",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ITEM_LIST _type field: %s", i.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ITEM_LIST_META_TYPE),
		})
	}

	return errors
}

func (i *ITEM_LIST) SetMetaType() {
	i.MetaType = util.Some(ITEM_LIST_META_TYPE)
}

const ITEM_TABLE_META_TYPE string = "ITEM_TABLE"

type ITEM_TABLE struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	Rows             util.Optional[[]CLUSTER]        `json:"rows,omitzero"`
}

func (i ITEM_TABLE) Validate() []ValidationError {
	var errors []ValidationError

	if i.MetaType.IsSet() && i.MetaType.V != ITEM_TABLE_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ITEM_TABLE",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ITEM_TABLE _type field: %s", i.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ITEM_TABLE_META_TYPE),
		})
	}

	return errors
}

func (i *ITEM_TABLE) SetMetaType() {
	i.MetaType = util.Some(ITEM_TABLE_META_TYPE)
}

const ITEM_TREE_META_TYPE string = "ITEM_TREE"

type ITEM_TREE struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	Items            util.Optional[[]ANY_ITEM]       `json:"items,omitzero"`
}

func (i ITEM_TREE) Validate() []ValidationError {
	var errors []ValidationError

	if i.MetaType.IsSet() && i.MetaType.V != ITEM_TREE_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ITEM_TREE",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ITEM_TREE _type field: %s", i.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ITEM_TREE_META_TYPE),
		})
	}

	return errors
}

func (i *ITEM_TREE) SetMetaType() {
	i.MetaType = util.Some(ITEM_TREE_META_TYPE)
}

const ITEM_META_TYPE string = "ITEM"

// Abstract type
type ITEM struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
}

func (i ITEM) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract ITEM type")
}

func (i *ITEM) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract ITEM type")
}

type ANY_ITEM struct {
	Value any
}

// Implement UnionType interface
func (i ANY_ITEM) GetBaseType() reflect.Type {
	return reflect.TypeFor[ITEM]()
}

func (i ANY_ITEM) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Value)
}

func (i *ANY_ITEM) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case CLUSTER_META_TYPE:
		i.Value = new(CLUSTER)
	case ELEMENT_META_TYPE:
		i.Value = new(ELEMENT)
	case "":
		return fmt.Errorf("missing ITEM _type field")
	default:
		return fmt.Errorf("ITEM unexpected _type %s", t)
	}

	return json.Unmarshal(data, i.Value)
}

const CLUSTER_META_TYPE string = "CLUSTER"

type CLUSTER struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	Items            []ANY_ITEM                      `json:"items"`
}

func (c CLUSTER) Validate() []ValidationError {
	var errors []ValidationError

	if c.MetaType.IsSet() && c.MetaType.V != CLUSTER_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "CLUSTER",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid CLUSTER _type field: %s", c.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", CLUSTER_META_TYPE),
		})
	}

	return errors
}

func (c *CLUSTER) SetMetaType() {
	c.MetaType = util.Some(CLUSTER_META_TYPE)
}

const ELEMENT_META_TYPE string = "ELEMENT"

type ELEMENT struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	NullFlavour      util.Optional[DV_CODED_TEXT]    `json:"null_flavour,omitzero"`
	Value            util.Optional[ANY_DATA_VALUE]   `json:"value,omitzero"`
	NullReason       util.Optional[ANY_DV_TEXT]      `json:"null_reason,omitzero"`
}

func (e ELEMENT) Validate() []ValidationError {
	var errors []ValidationError

	if e.MetaType.IsSet() && e.MetaType.V != ELEMENT_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ELEMENT",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ELEMENT type: %s", e.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ELEMENT_META_TYPE),
		})
	}

	return errors
}

func (e *ELEMENT) SetMetaType() {
	e.MetaType = util.Some(ELEMENT_META_TYPE)
}

const HISTORY_META_TYPE string = "HISTORY"

type HISTORY[T any] struct {
	MetaType         util.Optional[string]             `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                       `json:"name"`
	ArchetypeNodeID  string                            `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID]   `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]             `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]         `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]       `json:"feeder_audit,omitzero"`
	Origin           DV_DATE_TIME                      `json:"origin"`
	Period           util.Optional[DV_DURATION]        `json:"period,omitzero"`
	Duration         util.Optional[DV_DURATION]        `json:"duration,omitzero"`
	Summary          util.Optional[ANY_ITEM_STRUCTURE] `json:"summary,omitzero"`
	Events           util.Optional[[]ANY_EVENT[T]]     `json:"events,omitzero"`
}

func (h HISTORY[T]) Validate() []ValidationError {
	var errors []ValidationError

	if h.MetaType.IsSet() && h.MetaType.V != HISTORY_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "HISTORY",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid HISTORY _type field: %s", h.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", HISTORY_META_TYPE),
		})
	}

	return errors
}

const EVENT_META_TYPE string = "EVENT"

// Abstract type
type EVENT[T any] struct {
	MetaType         util.Optional[string]             `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                       `json:"name"`
	ArchetypeNodeID  string                            `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID]   `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]             `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]         `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]       `json:"feeder_audit,omitzero"`
	Time             DV_DATE_TIME                      `json:"time"`
	State            util.Optional[ANY_ITEM_STRUCTURE] `json:"state,omitzero"`
	Data             T                                 `json:"data"`
}

func (e EVENT[T]) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract EVENT type")
}

func (e *EVENT[T]) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract EVENT type")
}

type ANY_EVENT[T any] struct {
	Value any
}

// Implement UnionType interface
func (e ANY_EVENT[T]) GetBaseType() reflect.Type {
	return reflect.TypeFor[EVENT[T]]()
}

func (e ANY_EVENT[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Value)
}

func (e *ANY_EVENT[T]) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case POINT_EVENT_META_TYPE:
		e.Value = new(POINT_EVENT[T])
	case INTERVAL_EVENT_META_TYPE:
		e.Value = new(INTERVAL_EVENT[T])
	case "":
		return fmt.Errorf("missing EVENT _type field")
	default:
		return fmt.Errorf("EVENT unexpected _type %s", t)
	}

	return json.Unmarshal(data, e.Value)
}

const POINT_EVENT_META_TYPE string = "POINT_EVENT"

type POINT_EVENT[T any] struct {
	MetaType         util.Optional[string]             `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                       `json:"name"`
	ArchetypeNodeID  string                            `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID]   `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]             `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]         `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]       `json:"feeder_audit,omitzero"`
	Time             DV_DATE_TIME                      `json:"time"`
	State            util.Optional[ANY_ITEM_STRUCTURE] `json:"state,omitzero"`
	Data             T                                 `json:"data"`
}

func (p POINT_EVENT[T]) Validate() []ValidationError {
	var errors []ValidationError

	if p.MetaType.IsSet() && p.MetaType.V != POINT_EVENT_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "POINT_EVENT",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid POINT_EVENT _type field: %s", p.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", POINT_EVENT_META_TYPE),
		})
	}

	return errors
}

const INTERVAL_EVENT_META_TYPE string = "INTERVAL_EVENT"

type INTERVAL_EVENT[T any] struct {
	MetaType         util.Optional[string]             `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                       `json:"name"`
	ArchetypeNodeID  string                            `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID]   `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]             `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]         `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]       `json:"feeder_audit,omitzero"`
	Time             DV_DATE_TIME                      `json:"time"`
	State            util.Optional[ANY_ITEM_STRUCTURE] `json:"state,omitzero"`
	Data             T                                 `json:"data"`
	Width            DV_DURATION                       `json:"width"`
	SampleCount      util.Optional[int64]              `json:"sample_count,omitzero"`
	MathFunction     DV_CODED_TEXT                     `json:"math_function"`
}

func (i INTERVAL_EVENT[T]) Validate() []ValidationError {
	var errors []ValidationError

	if i.MetaType.IsSet() && i.MetaType.V != INTERVAL_EVENT_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "INTERVAL_EVENT",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid INTERVAL_EVENT _type field: %s", i.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", INTERVAL_EVENT_META_TYPE),
		})
	}

	return errors
}

// -----------------------------------
// DATA_STRUCTURES
// -----------------------------------

const DATA_VALUE_META_TYPE string = "DATA_VALUE"

// Abstract type
type DATA_VALUE struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
}

func (d DATA_VALUE) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract DATA_VALUE type")
}

func (d *DATA_VALUE) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract DATA_VALUE type")
}

type ANY_DATA_VALUE struct {
	Value any
}

// Implement UnionType interface
func (d ANY_DATA_VALUE) GetBaseType() reflect.Type {
	return reflect.TypeFor[DATA_VALUE]()
}

func (d ANY_DATA_VALUE) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Value)
}

func (d *ANY_DATA_VALUE) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case DV_BOOLEAN_META_TYPE:
		d.Value = new(DV_BOOLEAN)
	case DV_STATE_META_TYPE:
		d.Value = new(DV_STATE)
	case DV_IDENTIFIER_META_TYPE:
		d.Value = new(DV_IDENTIFIER)
	case DV_TEXT_META_TYPE:
		d.Value = new(DV_TEXT)
	case DV_CODED_TEXT_META_TYPE:
		d.Value = new(DV_CODED_TEXT)
	case DV_PARAGRAPH_META_TYPE:
		d.Value = new(DV_PARAGRAPH)
	case DV_INTERVAL_META_TYPE:
		d.Value = new(DV_INTERVAL)
	case DV_ORDINAL_META_TYPE:
		d.Value = new(DV_ORDINAL)
	case DV_SCALE_META_TYPE:
		d.Value = new(DV_SCALE)
	case DV_QUANTITY_META_TYPE:
		d.Value = new(DV_QUANTITY)
	case DV_COUNT_META_TYPE:
		d.Value = new(DV_COUNT)
	case DV_PROPORTION_META_TYPE:
		d.Value = new(DV_PROPORTION)
	case DV_DATE_META_TYPE:
		d.Value = new(DV_DATE)
	case DV_TIME_META_TYPE:
		d.Value = new(DV_TIME)
	case DV_DATE_TIME_META_TYPE:
		d.Value = new(DV_DATE_TIME)
	case DV_DURATION_META_TYPE:
		d.Value = new(DV_DURATION)
	case DV_PERIODIC_TIME_SPECIFICATION_META_TYPE:
		d.Value = new(DV_PERIODIC_TIME_SPECIFICATION)
	case DV_GENERAL_TIME_SPECIFICATION_META_TYPE:
		d.Value = new(DV_GENERAL_TIME_SPECIFICATION)
	case DV_MULTIMEDIA_META_TYPE:
		d.Value = new(DV_MULTIMEDIA)
	case DV_PARSABLE_META_TYPE:
		d.Value = new(DV_PARSABLE)
	case DV_URI_META_TYPE:
		d.Value = new(DV_URI)
	case DV_EHR_URI_META_TYPE:
		d.Value = new(DV_EHR_URI)
	case "":
		return fmt.Errorf("missing DATA_VALUE _type field")
	default:
		return fmt.Errorf("DATA_VALUE unexpected _type %s", t)
	}

	return json.Unmarshal(data, d.Value)
}

const DV_BOOLEAN_META_TYPE string = "DV_BOOLEAN"

type DV_BOOLEAN struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    bool                  `json:"value"`
}

func (d DV_BOOLEAN) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_BOOLEAN_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_BOOLEAN",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_BOOLEAN _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_BOOLEAN_META_TYPE),
		})
	}

	return errors
}

func (d *DV_BOOLEAN) SetMetaType() {
	d.MetaType = util.Some(DV_BOOLEAN_META_TYPE)
}

const DV_STATE_META_TYPE string = "DV_STATE"

type DV_STATE struct {
	MetaType   util.Optional[string] `json:"_type,omitzero"`
	Value      DV_CODED_TEXT         `json:"value"`
	IsTerminal bool                  `json:"is_terminal"`
}

func (d DV_STATE) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_STATE_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_STATE",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_STATE _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_STATE_META_TYPE),
		})
	}

	return errors
}

func (d *DV_STATE) SetMetaType() {
	d.MetaType = util.Some(DV_STATE_META_TYPE)
}

const DV_IDENTIFIER_META_TYPE string = "DV_IDENTIFIER"

type DV_IDENTIFIER struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Issuer   util.Optional[string] `json:"issuer,omitzero"`
	Assigner util.Optional[string] `json:"assigner,omitzero"`
	ID       string                `json:"id"`
	Type     util.Optional[string] `json:"type,omitzero"`
}

func (d DV_IDENTIFIER) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_IDENTIFIER_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_IDENTIFIER",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_IDENTIFIER _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_IDENTIFIER_META_TYPE),
		})
	}

	return errors
}

func (d *DV_IDENTIFIER) SetMetaType() {
	d.MetaType = util.Some(DV_IDENTIFIER_META_TYPE)
}

const DV_TEXT_META_TYPE string = "DV_TEXT"

type DV_TEXT struct {
	MetaType   util.Optional[string]         `json:"_type,omitzero"`
	Value      string                        `json:"value"`
	Hyperlink  util.Optional[DV_URI]         `json:"hyperlink,omitzero"`
	Formatting util.Optional[string]         `json:"formatting,omitzero"`
	Mappings   util.Optional[[]TERM_MAPPING] `json:"mappings,omitzero"`
	Language   util.Optional[CODE_PHRASE]    `json:"language,omitzero"`
	Encoding   util.Optional[CODE_PHRASE]    `json:"encoding,omitzero"`
}

func (d DV_TEXT) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_TEXT_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_TEXT",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_TEXT _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_TEXT_META_TYPE),
		})
	}

	return errors
}

func (d *DV_TEXT) SetMetaType() {
	d.MetaType = util.Some(DV_TEXT_META_TYPE)
}

type ANY_DV_TEXT struct {
	Value any
}

func (a ANY_DV_TEXT) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Value)
}

func (a *ANY_DV_TEXT) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case DV_TEXT_META_TYPE:
		a.Value = new(DV_TEXT)
	case DV_CODED_TEXT_META_TYPE:
		a.Value = new(DV_CODED_TEXT)
	case "":
		return fmt.Errorf("missing DV_TEXT _type field")
	default:
		return fmt.Errorf("DV_TEXT unexpected _type %s", t)
	}

	return json.Unmarshal(data, a.Value)
}

const TERM_MAPPING_META_TYPE string = "TERM_MAPPING"

type TERM_MAPPING struct {
	MetaType util.Optional[string]        `json:"_type,omitzero"`
	Match    byte                         `json:"match"`
	Purpose  util.Optional[DV_CODED_TEXT] `json:"purpose,omitzero"`
	Target   CODE_PHRASE                  `json:"target"`
}

func (t TERM_MAPPING) Validate() []ValidationError {
	var errors []ValidationError

	if t.MetaType.IsSet() && t.MetaType.V != TERM_MAPPING_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "TERM_MAPPING",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid TERM_MAPPING _type field: %s", t.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", TERM_MAPPING_META_TYPE),
		})
	}

	return errors
}

func (t *TERM_MAPPING) SetMetaType() {
	t.MetaType = util.Some(TERM_MAPPING_META_TYPE)
}

const CODE_PHRASE_META_TYPE string = "CODE_PHRASE"

type CODE_PHRASE struct {
	MetaType      util.Optional[string] `json:"_type,omitzero"`
	TerminologyId TERMINOLOGY_ID        `json:"terminology_id"`
	CodeString    string                `json:"code_string"`
	PreferredTerm util.Optional[string] `json:"preferred_term,omitzero"`
}

func (c CODE_PHRASE) Validate() []ValidationError {
	var errors []ValidationError

	if c.MetaType.IsSet() && c.MetaType.V != CODE_PHRASE_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "CODE_PHRASE",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid CODE_PHRASE _type field: %s", c.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", CODE_PHRASE_META_TYPE),
		})
	}

	return errors
}

func (c *CODE_PHRASE) SetMetaType() {
	c.MetaType = util.Some(CODE_PHRASE_META_TYPE)
}

const DV_CODED_TEXT_META_TYPE string = "DV_CODED_TEXT"

type DV_CODED_TEXT struct {
	MetaType     util.Optional[string]         `json:"_type,omitzero"`
	Value        string                        `json:"value"`
	Hyperlink    util.Optional[DV_URI]         `json:"hyperlink,omitzero"`
	Formatting   util.Optional[string]         `json:"formatting,omitzero"`
	Mappings     util.Optional[[]TERM_MAPPING] `json:"mappings,omitzero"`
	Language     util.Optional[CODE_PHRASE]    `json:"language,omitzero"`
	Encoding     util.Optional[CODE_PHRASE]    `json:"encoding,omitzero"`
	DefiningCode CODE_PHRASE                   `json:"defining_code"`
}

func (d DV_CODED_TEXT) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_CODED_TEXT_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_CODED_TEXT",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_CODED_TEXT _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_CODED_TEXT_META_TYPE),
		})
	}

	return errors
}

func (d *DV_CODED_TEXT) SetMetaType() {
	d.MetaType = util.Some(DV_CODED_TEXT_META_TYPE)
}

const DV_PARAGRAPH_META_TYPE string = "DV_PARAGRAPH"

type DV_PARAGRAPH struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Items    []ANY_DV_TEXT         `json:"items"`
}

func (d DV_PARAGRAPH) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_PARAGRAPH_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_PARAGRAPH",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_PARAGRAPH _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_PARAGRAPH_META_TYPE),
		})
	}

	return errors
}

func (d *DV_PARAGRAPH) SetMetaType() {
	d.MetaType = util.Some(DV_PARAGRAPH_META_TYPE)
}

const DV_ORDERED_META_TYPE string = "DV_ORDERED"

// Abstract type
type DV_ORDERED struct {
	MetaType             util.Optional[string]            `json:"_type,omitzero"`
	NormalStatus         util.Optional[CODE_PHRASE]       `json:"normal_status,omitzero"`
	NormalRange          util.Optional[DV_INTERVAL]       `json:"normal_range,omitzero"`
	OtherReferenceRanges util.Optional[[]REFERENCE_RANGE] `json:"other_reference_ranges,omitzero"`
}

func (d DV_ORDERED) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract DV_ORDERED type")
}

func (d *DV_ORDERED) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract DV_ORDERED type")
}

type ANY_DV_ORDERED struct {
	Value any
}

// Implement UnionType interface
func (d ANY_DV_ORDERED) GetBaseType() reflect.Type {
	return reflect.TypeFor[DV_ORDERED]()
}

func (d ANY_DV_ORDERED) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Value)
}

func (d *ANY_DV_ORDERED) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case DV_ORDINAL_META_TYPE:
		d.Value = new(DV_ORDINAL)
	case DV_SCALE_META_TYPE:
		d.Value = new(DV_SCALE)
	case DV_QUANTITY_META_TYPE:
		d.Value = new(DV_QUANTITY)
	case DV_COUNT_META_TYPE:
		d.Value = new(DV_COUNT)
	case DV_PROPORTION_META_TYPE:
		d.Value = new(DV_PROPORTION)
	case DV_DATE_META_TYPE:
		d.Value = new(DV_DATE)
	case DV_TIME_META_TYPE:
		d.Value = new(DV_TIME)
	case DV_DATE_TIME_META_TYPE:
		d.Value = new(DV_DATE_TIME)
	case DV_DURATION_META_TYPE:
		d.Value = new(DV_DURATION)
	case "":
		return fmt.Errorf("missing DV_ORDERED _type field")
	default:
		{
			return fmt.Errorf("DV_ORDERED unexpected _type %s", t)
		}
	}

	return json.Unmarshal(data, d.Value)
}

const DV_INTERVAL_META_TYPE string = "DV_INTERVAL"

type DV_INTERVAL struct {
	MetaType       util.Optional[string] `json:"_type,omitzero"`
	Lower          any                   `json:"lower"`
	Upper          any                   `json:"upper"`
	LowerUnbounded bool                  `json:"lower_unbounded"`
	UpperUnbounded bool                  `json:"upper_unbounded"`
	LowerIncluded  bool                  `json:"lower_included"`
	UpperIncluded  bool                  `json:"upper_included"`
}

func (d DV_INTERVAL) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_INTERVAL_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_INTERVAL",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_INTERVAL _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_INTERVAL_META_TYPE),
		})
	}

	return errors
}

func (d *DV_INTERVAL) SetMetaType() {
	d.MetaType = util.Some(DV_INTERVAL_META_TYPE)
}

const REFERENCE_RANGE_META_TYPE string = "REFERENCE_RANGE"

type REFERENCE_RANGE struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Meaning  ANY_DV_TEXT           `json:"meaning"`
	Range    DV_INTERVAL           `json:"range"`
}

func (r REFERENCE_RANGE) Validate() []ValidationError {
	var errors []ValidationError

	if r.MetaType.IsSet() && r.MetaType.V != REFERENCE_RANGE_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "REFERENCE_RANGE",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid REFERENCE_RANGE _type field: %s", r.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", REFERENCE_RANGE_META_TYPE),
		})
	}

	return errors
}

func (r *REFERENCE_RANGE) SetMetaType() {
	r.MetaType = util.Some(REFERENCE_RANGE_META_TYPE)
}

const DV_ORDINAL_META_TYPE string = "DV_ORDINAL"

type DV_ORDINAL struct {
	MetaType             util.Optional[string]          `json:"_type,omitzero"`
	NormalStatus         util.Optional[CODE_PHRASE]     `json:"normal_status,omitzero"`
	NormalRange          util.Optional[DV_INTERVAL]     `json:"normal_range,omitzero"`
	OtherReferenceRanges util.Optional[REFERENCE_RANGE] `json:"other_reference_ranges,omitzero"`
	Symbol               DV_CODED_TEXT                  `json:"symbol"`
	Value                int64                          `json:"value"`
}

func (d DV_ORDINAL) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_ORDINAL_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_ORDINAL",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_ORDINAL _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_ORDINAL_META_TYPE),
		})
	}

	return errors
}

func (d *DV_ORDINAL) SetMetaType() {
	d.MetaType = util.Some(DV_ORDINAL_META_TYPE)
}

const DV_SCALE_META_TYPE string = "DV_SCALE"

type DV_SCALE struct {
	MetaType             util.Optional[string]            `json:"_type,omitzero"`
	NormalStatus         util.Optional[CODE_PHRASE]       `json:"normal_status,omitzero"`
	NormalRange          util.Optional[DV_INTERVAL]       `json:"normal_range,omitzero"`
	OtherReferenceRanges util.Optional[[]REFERENCE_RANGE] `json:"other_reference_ranges,omitzero"`
	Symbol               DV_CODED_TEXT                    `json:"symbol"`
	Value                float64                          `json:"value"`
}

func (d DV_SCALE) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_SCALE_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_SCALE",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_SCALE _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_SCALE_META_TYPE),
		})
	}

	return errors
}

func (d *DV_SCALE) SetMetaType() {
	d.MetaType = util.Some(DV_SCALE_META_TYPE)
}

const DV_QUANTITY_META_TYPE string = "DV_QUANTITY"

type DV_QUANTITY struct {
	MetaType             util.Optional[string]            `json:"_type,omitzero"`
	NormalStatus         util.Optional[CODE_PHRASE]       `json:"normal_status,omitzero"`
	MagnitudeStatus      util.Optional[string]            `json:"magnitude_status,omitzero"`
	AccuracyIsPercent    util.Optional[bool]              `json:"accuracy_is_percent,omitzero"`
	Accuracy             util.Optional[float64]           `json:"accuracy,omitzero"`
	Magnitude            float64                          `json:"magnitude"`
	Precision            util.Optional[int64]             `json:"precision,omitzero"`
	Units                string                           `json:"units"`
	UnitsSystem          util.Optional[string]            `json:"units_system,omitzero"`
	UnitsDisplayName     util.Optional[string]            `json:"units_display_name,omitzero"`
	NormalRange          util.Optional[DV_INTERVAL]       `json:"normal_range,omitzero"`
	OtherReferenceRanges util.Optional[[]REFERENCE_RANGE] `json:"other_reference_ranges,omitzero"`
}

func (d DV_QUANTITY) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_QUANTITY_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_QUANTITY",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_QUANTITY _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_QUANTITY_META_TYPE),
		})
	}

	return errors
}

func (d *DV_QUANTITY) SetMetaType() {
	d.MetaType = util.Some(DV_QUANTITY_META_TYPE)
}

const DV_COUNT_META_TYPE string = "DV_COUNT"

type DV_COUNT struct {
	MetaType             util.Optional[string]            `json:"_type,omitzero"`
	NormalStatus         util.Optional[CODE_PHRASE]       `json:"normal_status,omitzero"`
	MagnitudeStatus      util.Optional[string]            `json:"magnitude_status,omitzero"`
	AccuracyIsPercent    util.Optional[bool]              `json:"accuracy_is_percent,omitzero"`
	Accuracy             util.Optional[float64]           `json:"accuracy,omitzero"`
	Magnitude            int64                            `json:"magnitude"`
	NormalRange          util.Optional[DV_INTERVAL]       `json:"normal_range,omitzero"`
	OtherReferenceRanges util.Optional[[]REFERENCE_RANGE] `json:"other_reference_ranges,omitzero"`
}

func (d DV_COUNT) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_COUNT_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_COUNT",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_COUNT _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_COUNT_META_TYPE),
		})
	}

	return errors
}

func (d *DV_COUNT) SetMetaType() {
	d.MetaType = util.Some(DV_COUNT_META_TYPE)
}

const DV_PROPORTION_META_TYPE string = "DV_PROPORTION"

type DV_PROPORTION struct {
	MetaType             util.Optional[string]            `json:"_type,omitzero"`
	NormalStatus         util.Optional[CODE_PHRASE]       `json:"normal_status,omitzero"`
	MagnitudeStatus      util.Optional[string]            `json:"magnitude_status,omitzero"`
	AccuracyIsPercent    util.Optional[bool]              `json:"accuracy_is_percent,omitzero"`
	Accuracy             util.Optional[float64]           `json:"accuracy,omitzero"`
	Numerator            float64                          `json:"numerator"`
	Denominator          float64                          `json:"denominator"`
	Type                 int64                            `json:"type"`
	Precision            util.Optional[int64]             `json:"precision,omitzero"`
	NormalRange          util.Optional[DV_INTERVAL]       `json:"normal_range,omitzero"`
	OtherReferenceRanges util.Optional[[]REFERENCE_RANGE] `json:"other_reference_ranges,omitzero"`
}

func (d DV_PROPORTION) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_PROPORTION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_PROPORTION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_PROPORTION _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_PROPORTION_META_TYPE),
		})
	}

	return errors
}

func (d *DV_PROPORTION) SetMetaType() {
	d.MetaType = util.Some(DV_PROPORTION_META_TYPE)
}

const DV_DATE_META_TYPE string = "DV_DATE"

type DV_DATE struct {
	MetaType             util.Optional[string]            `json:"_type,omitzero"`
	NormalStatus         util.Optional[CODE_PHRASE]       `json:"normal_status,omitzero"`
	NormalRange          util.Optional[DV_INTERVAL]       `json:"normal_range,omitzero"`
	OtherReferenceRanges util.Optional[[]REFERENCE_RANGE] `json:"other_reference_ranges,omitzero"`
	MagnitudeStatus      util.Optional[string]            `json:"magnitude_status,omitzero"`
	AccuracyIsPercent    util.Optional[bool]              `json:"accuracy_is_percent,omitzero"`
	Accuracy             util.Optional[DV_DURATION]       `json:"accuracy,omitzero"`
	Value                string                           `json:"value"`
}

func (d DV_DATE) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_DATE_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_DATE",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_DATE _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_DATE_META_TYPE),
		})
	}

	return errors
}

func (d *DV_DATE) SetMetaType() {
	d.MetaType = util.Some(DV_DATE_META_TYPE)
}

const DV_TIME_META_TYPE string = "DV_TIME"

type DV_TIME struct {
	MetaType             util.Optional[string]            `json:"_type,omitzero"`
	NormalStatus         util.Optional[CODE_PHRASE]       `json:"normal_status,omitzero"`
	NormalRange          util.Optional[DV_INTERVAL]       `json:"normal_range,omitzero"`
	OtherReferenceRanges util.Optional[[]REFERENCE_RANGE] `json:"other_reference_ranges,omitzero"`
	MagnitudeStatus      util.Optional[string]            `json:"magnitude_status,omitzero"`
	AccuracyIsPercent    util.Optional[bool]              `json:"accuracy_is_percent,omitzero"`
	Accuracy             util.Optional[DV_DURATION]       `json:"accuracy,omitzero"`
	Value                string                           `json:"value"`
}

func (d DV_TIME) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_TIME_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_TIME",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_TIME _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_TIME_META_TYPE),
		})
	}

	return errors
}

func (d *DV_TIME) SetMetaType() {
	d.MetaType = util.Some(DV_TIME_META_TYPE)
}

const DV_DATE_TIME_META_TYPE string = "DV_DATE_TIME"

type DV_DATE_TIME struct {
	MetaType             util.Optional[string]            `json:"_type,omitzero"`
	NormalStatus         util.Optional[CODE_PHRASE]       `json:"normal_status,omitzero"`
	NormalRange          util.Optional[DV_INTERVAL]       `json:"normal_range,omitzero"`
	OtherReferenceRanges util.Optional[[]REFERENCE_RANGE] `json:"other_reference_ranges,omitzero"`
	MagnitudeStatus      util.Optional[string]            `json:"magnitude_status,omitzero"`
	AccuracyIsPercent    util.Optional[bool]              `json:"accuracy_is_percent,omitzero"`
	Accuracy             util.Optional[DV_DURATION]       `json:"accuracy,omitzero"`
	Value                string                           `json:"value"`
}

func (d DV_DATE_TIME) Validate() []ValidationError {
	var errors []ValidationError

	// Validate _type field
	if d.MetaType.IsSet() && d.MetaType.V != DV_DATE_TIME_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          DV_DATE_TIME_META_TYPE,
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_DATE_TIME _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_DATE_TIME_META_TYPE),
		})
	}

	// Validate value field
	if d.Value == "" {
		errors = append(errors, ValidationError{
			Model:   DV_DATE_TIME_META_TYPE,
			Path:    ".value",
			Message: "DV_DATE_TIME value field is required",
		})
	}

	if !strings.HasSuffix(d.Value, "Z") {
		errors = append(errors, ValidationError{
			Model:          DV_DATE_TIME_META_TYPE,
			Path:           ".value",
			Message:        fmt.Sprintf("invalid DV_DATE_TIME value field: %s", d.Value),
			Recommendation: "Ensure value field is of format YYYY-MM-DDTHH:MM:SSZ",
		})
	} else {
		if _, err := time.Parse(time.RFC3339Nano, d.Value); err != nil {
			errors = append(errors, ValidationError{
				Model:          DV_DATE_TIME_META_TYPE,
				Path:           ".value",
				Message:        fmt.Sprintf("invalid DV_DATE_TIME value field: %s", d.Value),
				Recommendation: "Ensure value field is of format YYYY-MM-DDTHH:MM:SSZ",
			})
		}
	}

	return errors
}

func (d *DV_DATE_TIME) SetMetaType() {
	d.MetaType = util.Some(DV_DATE_TIME_META_TYPE)
}

const DV_DURATION_META_TYPE string = "DV_DURATION"

type DV_DURATION struct {
	MetaType             util.Optional[string]            `json:"_type,omitzero"`
	NormalStatus         util.Optional[CODE_PHRASE]       `json:"normal_status,omitzero"`
	NormalRange          util.Optional[DV_INTERVAL]       `json:"normal_range,omitzero"`
	OtherReferenceRanges util.Optional[[]REFERENCE_RANGE] `json:"other_reference_ranges,omitzero"`
	MagnitudeStatus      util.Optional[string]            `json:"magnitude_status,omitzero"`
	AccuracyIsPercent    util.Optional[bool]              `json:"accuracy_is_percent,omitzero"`
	Accuracy             util.Optional[bool]              `json:"accuracy,omitzero"`
	Value                string                           `json:"value"`
}

func (d DV_DURATION) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_DURATION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_DURATION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_DURATION _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_DURATION_META_TYPE),
		})
	}

	return errors
}

func (d *DV_DURATION) SetMetaType() {
	d.MetaType = util.Some(DV_DURATION_META_TYPE)
}

const DV_PERIODIC_TIME_SPECIFICATION_META_TYPE string = "DV_PERIODIC_TIME_SPECIFICATION"

type DV_PERIODIC_TIME_SPECIFICATION struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    DV_PARSABLE           `json:"value"`
}

func (d DV_PERIODIC_TIME_SPECIFICATION) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_PERIODIC_TIME_SPECIFICATION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_PERIODIC_TIME_SPECIFICATION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_PERIODIC_TIME_SPECIFICATION _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_PERIODIC_TIME_SPECIFICATION_META_TYPE),
		})
	}

	return errors
}

func (d *DV_PERIODIC_TIME_SPECIFICATION) SetMetaType() {
	d.MetaType = util.Some(DV_PERIODIC_TIME_SPECIFICATION_META_TYPE)
}

const DV_GENERAL_TIME_SPECIFICATION_META_TYPE string = "DV_GENERAL_TIME_SPECIFICATION"

type DV_GENERAL_TIME_SPECIFICATION struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    DV_PARSABLE           `json:"value"`
}

func (d DV_GENERAL_TIME_SPECIFICATION) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_GENERAL_TIME_SPECIFICATION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_GENERAL_TIME_SPECIFICATION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_GENERAL_TIME_SPECIFICATION _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_GENERAL_TIME_SPECIFICATION_META_TYPE),
		})
	}

	return errors
}

func (d *DV_GENERAL_TIME_SPECIFICATION) SetMetaType() {
	d.MetaType = util.Some(DV_GENERAL_TIME_SPECIFICATION_META_TYPE)
}

const DV_ENCAPSULATED_META_TYPE string = "DV_ENCAPSULATED"

// Abstract type
type DV_ENCAPSULATED struct {
	MetaType util.Optional[string]      `json:"_type,omitzero"`
	Charset  util.Optional[CODE_PHRASE] `json:"charset,omitzero"`
	Language util.Optional[CODE_PHRASE] `json:"language,omitzero"`
}

func (d DV_ENCAPSULATED) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract DV_ENCAPSULATED type")
}

func (d *DV_ENCAPSULATED) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract DV_ENCAPSULATED type")
}

type ANY_DV_ENCAPSULATED struct {
	Value any
}

// Implement UnionType interface
func (d ANY_DV_ENCAPSULATED) GetBaseType() reflect.Type {
	return reflect.TypeFor[DV_ENCAPSULATED]()
}

func (d ANY_DV_ENCAPSULATED) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Value)
}

func (d *ANY_DV_ENCAPSULATED) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case DV_MULTIMEDIA_META_TYPE:
		d.Value = new(DV_MULTIMEDIA)
	case DV_PARSABLE_META_TYPE:
		d.Value = new(DV_PARSABLE)
	case "":
		return fmt.Errorf("missing DV_ENCAPSULATED _type field")
	default:
		return fmt.Errorf("DV_ENCAPSULATED unexpected _type %s", t)
	}

	return json.Unmarshal(data, d.Value)
}

const DV_MULTIMEDIA_META_TYPE string = "DV_MULTIMEDIA"

type DV_MULTIMEDIA struct {
	MetaType                util.Optional[string]         `json:"_type,omitzero"`
	Charset                 util.Optional[CODE_PHRASE]    `json:"charset,omitzero"`
	Language                util.Optional[CODE_PHRASE]    `json:"language,omitzero"`
	AlternateText           util.Optional[string]         `json:"alternate_text,omitzero"`
	Uri                     util.Optional[DV_URI]         `json:"uri,omitzero"`
	Data                    util.Optional[string]         `json:"data,omitzero"`
	MediaType               CODE_PHRASE                   `json:"media_type"`
	CompressionAlgorithm    util.Optional[CODE_PHRASE]    `json:"compression_algorithm,omitzero"`
	IntegrityCheck          util.Optional[string]         `json:"integrity_check,omitzero"`
	IntegrityCheckAlgorithm util.Optional[CODE_PHRASE]    `json:"integrity_check_algorithm,omitzero"`
	Thumbnail               util.Optional[*DV_MULTIMEDIA] `json:"thumbnail,omitzero"`
	Size                    int64                         `json:"size"`
}

func (d DV_MULTIMEDIA) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_MULTIMEDIA_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_MULTIMEDIA",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_MULTIMEDIA _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_MULTIMEDIA_META_TYPE),
		})
	}

	return errors
}

func (d *DV_MULTIMEDIA) SetMetaType() {
	d.MetaType = util.Some(DV_MULTIMEDIA_META_TYPE)
}

const DV_PARSABLE_META_TYPE string = "DV_PARSABLE"

type DV_PARSABLE struct {
	MetaType  util.Optional[string]      `json:"_type,omitzero"`
	Charset   util.Optional[CODE_PHRASE] `json:"charset,omitzero"`
	Language  util.Optional[CODE_PHRASE] `json:"language,omitzero"`
	Value     string                     `json:"value"`
	Formalism string                     `json:"formalism"`
}

func (d DV_PARSABLE) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_PARSABLE_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_PARSABLE",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_PARSABLE _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_PARSABLE_META_TYPE),
		})
	}

	return errors
}

func (d *DV_PARSABLE) SetMetaType() {
	d.MetaType = util.Some(DV_PARSABLE_META_TYPE)
}

const DV_URI_META_TYPE string = "DV_URI"

type DV_URI struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    string                `json:"value"`
}

func (d DV_URI) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_URI_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_URI",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_URI _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_URI_META_TYPE),
		})
	}

	return errors
}

func (d *DV_URI) SetMetaType() {
	d.MetaType = util.Some(DV_URI_META_TYPE)
}

const DV_EHR_URI_META_TYPE string = "DV_EHR_URI"

type DV_EHR_URI struct {
	MetaType           util.Optional[string] `json:"_type,omitzero"`
	Value              string                `json:"value"`
	LocalTerminologyId string                `json:"local_terminology_id"`
}

func (d DV_EHR_URI) Validate() []ValidationError {
	var errors []ValidationError

	if d.MetaType.IsSet() && d.MetaType.V != DV_EHR_URI_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "DV_EHR_URI",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid DV_EHR_URI _type field: %s", d.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", DV_EHR_URI_META_TYPE),
		})
	}

	return errors
}

func (d *DV_EHR_URI) SetMetaType() {
	d.MetaType = util.Some(DV_EHR_URI_META_TYPE)
}

// -----------------------------------
// BASE_TYPES
// -----------------------------------

const UID_META_TYPE string = "UID"

// Abstract type
type UID struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    string                `json:"value"`
}

func (u UID) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract UID type")
}

func (u *UID) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract UID type")
}

type ANY_UID struct {
	Value any
}

// Implement UnionType interface
func (u ANY_UID) GetBaseType() reflect.Type {
	return reflect.TypeFor[UID]()
}

func (u ANY_UID) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Value)
}

func (u *ANY_UID) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case ISO_OID_META_TYPE:
		u.Value = new(ISO_OID)
	case UUID_META_TYPE:
		u.Value = new(UUID)
	case INTERNET_ID_META_TYPE:
		u.Value = new(INTERNET_ID)
	case "":
		return fmt.Errorf("missing UID _type field")
	default:
		return fmt.Errorf("UID unexpected _type %s", t)
	}

	return json.Unmarshal(data, u.Value)
}

const ISO_OID_META_TYPE string = "ISO_OID"

type ISO_OID struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    string                `json:"value"`
}

func (i ISO_OID) Validate() []ValidationError {
	var errors []ValidationError

	if i.MetaType.IsSet() && i.MetaType.V != ISO_OID_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ISO_OID",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ISO_OID _type field: %s", i.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ISO_OID_META_TYPE),
		})
	}

	return errors
}

func (i *ISO_OID) SetMetaType() {
	i.MetaType = util.Some(ISO_OID_META_TYPE)
}

const UUID_META_TYPE string = "UUID"

type UUID struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    string                `json:"value"`
}

func (u UUID) Validate() []ValidationError {
	var errors []ValidationError

	if u.MetaType.IsSet() && u.MetaType.V != UUID_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "UUID",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid UUID _type field: %s", u.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", UUID_META_TYPE),
		})
	}

	return errors
}

func (u *UUID) SetMetaType() {
	u.MetaType = util.Some(UUID_META_TYPE)
}

const INTERNET_ID_META_TYPE string = "INTERNET_ID"

type INTERNET_ID struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    string                `json:"value"`
}

func (i INTERNET_ID) Validate() []ValidationError {
	var errors []ValidationError

	if i.MetaType.IsSet() && i.MetaType.V != INTERNET_ID_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "INTERNET_ID",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid INTERNET_ID _type field: %s", i.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", INTERNET_ID_META_TYPE),
		})
	}

	if i.Value == "" {
		errors = append(errors, ValidationError{
			Model:          "INTERNET_ID",
			Path:           ".value",
			Message:        "value cannot be empty",
			Recommendation: "Ensure value field is set",
		})
	} else {
		if !ValidateInternetID(i.Value) {
			errors = append(errors, ValidationError{
				Model:          "INTERNET_ID",
				Path:           ".value",
				Message:        fmt.Sprintf("invalid INTERNET_ID format: %s", i.Value),
				Recommendation: "Ensure value field is a valid URL",
			})
		}
	}

	return errors
}

func (i *INTERNET_ID) SetMetaType() {
	i.MetaType = util.Some(INTERNET_ID_META_TYPE)
}

const OBJECT_ID_META_TYPE string = "OBJECT_ID"

// Abstract type
type OBJECT_ID struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    string                `json:"value"`
}

func (o OBJECT_ID) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract OBJECT_ID type")
}

func (o *OBJECT_ID) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract OBJECT_ID type")
}

type ANY_OBJECT_ID struct {
	Value any
}

// Implement UnionType interface
func (o ANY_OBJECT_ID) GetBaseType() reflect.Type {
	return reflect.TypeFor[OBJECT_ID]()
}

func (o ANY_OBJECT_ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Value)
}

func (o *ANY_OBJECT_ID) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case HIER_OBJECT_ID_META_TYPE:
		o.Value = new(HIER_OBJECT_ID)
	case OBJECT_VERSION_ID_META_TYPE:
		o.Value = new(OBJECT_VERSION_ID)
	case ARCHETYPE_ID_META_TYPE:
		o.Value = new(ARCHETYPE_ID)
	case TEMPLATE_ID_META_TYPE:
		o.Value = new(TEMPLATE_ID)
	case GENERIC_ID_META_TYPE:
		o.Value = new(GENERIC_ID)
	case "":
		return fmt.Errorf("missing OBJECT_ID _type field")
	default:
		return fmt.Errorf("OBJECT_ID unexpected _type %s", t)
	}

	return json.Unmarshal(data, o.Value)
}

const UID_BASED_ID_META_TYPE string = "UID_BASED_ID"

// Abstract type
type UID_BASED_ID struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    string                `json:"value"`
}

func (u UID_BASED_ID) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract UID_BASED_ID type")
}

func (u *UID_BASED_ID) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract UID_BASED_ID type")
}

type ANY_UID_BASED_ID struct {
	Value any
}

// Implement UnionType interface
func (u ANY_UID_BASED_ID) GetBaseType() reflect.Type {
	return reflect.TypeFor[UID_BASED_ID]()
}

func (u ANY_UID_BASED_ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Value)
}

func (u *ANY_UID_BASED_ID) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case HIER_OBJECT_ID_META_TYPE:
		u.Value = new(HIER_OBJECT_ID)
	case OBJECT_VERSION_ID_META_TYPE:
		u.Value = new(OBJECT_VERSION_ID)
	case "":
		return fmt.Errorf("missing UID_BASED_ID _type field")
	default:
		return fmt.Errorf("UID_BASED_ID unexpected _type %s", t)
	}

	return json.Unmarshal(data, u.Value)
}

const HIER_OBJECT_ID_META_TYPE string = "HIER_OBJECT_ID"

type HIER_OBJECT_ID struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    string                `json:"value"`
}

func (h HIER_OBJECT_ID) Validate() []ValidationError {
	var errors []ValidationError

	if h.MetaType.IsSet() && h.MetaType.V != HIER_OBJECT_ID_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "HIER_OBJECT_ID",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid HIER_OBJECT_ID _type field: %s", h.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", HIER_OBJECT_ID_META_TYPE),
		})
	}

	// Validate UID-based identifier format: root '::' extension (extension is optional)
	if h.Value == "" {
		errors = append(errors, ValidationError{
			Model:          "HIER_OBJECT_ID",
			Path:           ".value",
			Message:        "HIER_OBJECT_ID value cannot be empty",
			Recommendation: "Ensure HIER_OBJECT_ID value is set",
		})
	} else {
		// Split by '::' separator
		parts := strings.Split(h.Value, "::")

		// Must have 1 (root only) or 2 parts (root + extension)
		if len(parts) > 2 {
			errors = append(errors, ValidationError{
				Model:          "HIER_OBJECT_ID",
				Path:           ".value",
				Message:        "HIER_OBJECT_ID invalid format: too many '::'",
				Recommendation: "Ensure HIER_OBJECT_ID value is in the format 'root::extension'",
			})
			return errors
		}

		// Validate root part (first part)
		root := parts[0]
		if root == "" {
			errors = append(errors, ValidationError{
				Model:          "HIER_OBJECT_ID",
				Path:           ".value",
				Message:        fmt.Sprintf("HIER_OBJECT_ID root part cannot be empty in '%s'", h.Value),
				Recommendation: "Ensure HIER_OBJECT_ID value has a non-empty root part",
			})
			return errors
		}

		// Root should be a valid UID (UUID, ISO_OID, or INTERNET_ID format)
		if err := ValidateUID(root); err != nil {
			errors = append(errors, ValidationError{
				Model:          "HIER_OBJECT_ID",
				Path:           ".value",
				Message:        fmt.Sprintf("HIER_OBJECT_ID invalid root UID '%s': %v", root, err),
				Recommendation: "Ensure HIER_OBJECT_ID root part is a valid UUID, ISO_OID, or INTERNET_ID",
			})
		}

		// If extension exists, validate it's not empty
		if len(parts) == 2 {
			extension := parts[1]
			if extension == "" {
				errors = append(errors, ValidationError{
					Model:          "HIER_OBJECT_ID",
					Path:           ".value",
					Message:        "HIER_OBJECT_ID extension cannot be empty when '::' is present",
					Recommendation: "Ensure HIER_OBJECT_ID value has a non-empty extension part",
				})
			}
		}
	}

	return errors
}

func (h *HIER_OBJECT_ID) SetMetaType() {
	h.MetaType = util.Some(HIER_OBJECT_ID_META_TYPE)
}

const OBJECT_VERSION_ID_META_TYPE string = "OBJECT_VERSION_ID"

type OBJECT_VERSION_ID struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    string                `json:"value"`
}

func (o OBJECT_VERSION_ID) Validate() []ValidationError {
	var errors []ValidationError

	if o.MetaType.IsSet() && o.MetaType.V != OBJECT_VERSION_ID_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "OBJECT_VERSION_ID",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid OBJECT_VERSION_ID _type field: %s", o.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", OBJECT_VERSION_ID_META_TYPE),
		})
	}

	return errors
}

func (o *OBJECT_VERSION_ID) SetMetaType() {
	o.MetaType = util.Some(OBJECT_VERSION_ID_META_TYPE)
}

const ARCHETYPE_ID_META_TYPE string = "ARCHETYPE_ID"

type ARCHETYPE_ID struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    string                `json:"value"`
}

func (a ARCHETYPE_ID) Validate() []ValidationError {
	var errors []ValidationError

	if a.MetaType.IsSet() && a.MetaType.V != ARCHETYPE_ID_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ARCHETYPE_ID",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ARCHETYPE_ID _type field: %s", a.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ARCHETYPE_ID_META_TYPE),
		})
	}

	return errors
}

func (a *ARCHETYPE_ID) SetMetaType() {
	a.MetaType = util.Some(ARCHETYPE_ID_META_TYPE)
}

const TEMPLATE_ID_META_TYPE string = "TEMPLATE_ID"

type TEMPLATE_ID struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    string                `json:"value"`
}

func (t TEMPLATE_ID) Validate() []ValidationError {
	var errors []ValidationError

	if t.MetaType.IsSet() && t.MetaType.V != TEMPLATE_ID_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "TEMPLATE_ID",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid TEMPLATE_ID _type field: %s", t.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", TEMPLATE_ID_META_TYPE),
		})
	}

	return errors
}

func (t *TEMPLATE_ID) SetMetaType() {
	t.MetaType = util.Some(TEMPLATE_ID_META_TYPE)
}

const TERMINOLOGY_ID_META_TYPE string = "TERMINOLOGY_ID"

type TERMINOLOGY_ID struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    string                `json:"value"`
}

func (t TERMINOLOGY_ID) Validate() []ValidationError {
	var errors []ValidationError

	if t.MetaType.IsSet() && t.MetaType.V != TERMINOLOGY_ID_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "TERMINOLOGY_ID",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid TERMINOLOGY_ID _type field: %s", t.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", TERMINOLOGY_ID_META_TYPE),
		})
	}

	return errors
}

func (t *TERMINOLOGY_ID) SetMetaType() {
	t.MetaType = util.Some(TERMINOLOGY_ID_META_TYPE)
}

const GENERIC_ID_META_TYPE string = "GENERIC_ID"

type GENERIC_ID struct {
	MetaType util.Optional[string] `json:"_type,omitzero"`
	Value    string                `json:"value"`
	Scheme   string                `json:"scheme"`
}

func (g GENERIC_ID) Validate() []ValidationError {
	var errors []ValidationError

	if g.MetaType.IsSet() && g.MetaType.V != GENERIC_ID_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "GENERIC_ID",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid generic id _type field: %s", g.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", GENERIC_ID_META_TYPE),
		})
	}

	return errors
}

func (g *GENERIC_ID) SetMetaType() {
	g.MetaType = util.Some(GENERIC_ID_META_TYPE)
}

const OBJECT_REF_META_TYPE string = "OBJECT_REF"

type OBJECT_REF struct {
	MetaType  util.Optional[string] `json:"_type,omitzero"`
	Namespace string                `json:"namespace"`
	Type      string                `json:"type"`
	ID        ANY_OBJECT_ID         `json:"id"`
}

func (o OBJECT_REF) Validate() []ValidationError {
	var errors []ValidationError

	if o.MetaType.IsSet() && o.MetaType.V != OBJECT_REF_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "OBJECT_REF",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid object ref _type field: %s", o.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", OBJECT_REF_META_TYPE),
		})
	}

	return errors
}

func (o *OBJECT_REF) SetMetaType() {
	o.MetaType = util.Some(OBJECT_REF_META_TYPE)
}

const PARTY_REF_META_TYPE string = "PARTY_REF"

type PARTY_REF struct {
	MetaType  util.Optional[string] `json:"_type,omitzero"`
	Namespace string                `json:"namespace"`
	Type      string                `json:"type"`
	ID        ANY_OBJECT_ID         `json:"id"`
}

func (p PARTY_REF) Validate() []ValidationError {
	var errors []ValidationError

	if p.MetaType.IsSet() && p.MetaType.V != PARTY_REF_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "PARTY_REF",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid party ref _type field: %s", p.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", PARTY_REF_META_TYPE),
		})
	}

	return errors
}

func (p *PARTY_REF) SetMetaType() {
	p.MetaType = util.Some(PARTY_REF_META_TYPE)
}

const LOCATABLE_REF_META_TYPE string = "LOCATABLE_REF"

type LOCATABLE_REF struct {
	MetaType  util.Optional[string] `json:"_type,omitzero"`
	Namespace string                `json:"namespace"`
	Type      string                `json:"type"`
	Path      util.Optional[string] `json:"path,omitzero"`
	ID        ANY_UID_BASED_ID      `json:"id"`
}

func (l LOCATABLE_REF) Validate() []ValidationError {
	var errors []ValidationError

	if l.MetaType.IsSet() && l.MetaType.V != LOCATABLE_REF_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "LOCATABLE_REF",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid LOCATABLE_REF _type field: %s", l.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", LOCATABLE_REF_META_TYPE),
		})
	}

	return errors
}

func (l *LOCATABLE_REF) SetMetaType() {
	l.MetaType = util.Some(LOCATABLE_REF_META_TYPE)
}

const PARTY_META_TYPE string = "PARTY"

// Abstract type
type PARTY struct {
	MetaType             util.Optional[string]               `json:"_type,omitzero"`
	Name                 ANY_DV_TEXT                         `json:"name"`
	ArchetypeNodeID      string                              `json:"archetype_node_id"`
	UID                  util.Optional[ANY_UID_BASED_ID]     `json:"uid,omitzero"`
	Links                util.Optional[[]LINK]               `json:"links,omitzero"`
	ArchetypeDetails     util.Optional[ARCHETYPED]           `json:"archetype_details,omitzero"`
	FeederAudit          util.Optional[FEEDER_AUDIT]         `json:"feeder_audit,omitzero"`
	Identities           []PARTY_IDENTITY                    `json:"identities"`
	Contacts             util.Optional[[]CONTACT]            `json:"contacts,omitzero"`
	Details              util.Optional[ANY_ITEM_STRUCTURE]   `json:"details,omitzero"`
	ReverseRelationships util.Optional[[]LOCATABLE_REF]      `json:"reverse_relationships,omitzero"`
	Relationships        util.Optional[[]PARTY_RELATIONSHIP] `json:"relationships,omitzero"`
}

func (p PARTY) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract PARTY type")
}

func (p *PARTY) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract PARTY type")
}

type ANY_PARTY struct {
	Value any
}

// Implement UnionType interface
func (p ANY_PARTY) GetBaseType() reflect.Type {
	return reflect.TypeFor[PARTY]()
}

func (p ANY_PARTY) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Value)
}

func (p *ANY_PARTY) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case ROLE_META_TYPE:
		p.Value = new(ROLE)
	case ORGANISATION_META_TYPE:
		p.Value = new(ORGANISATION)
	case PERSON_META_TYPE:
		p.Value = new(PERSON)
	case AGENT_META_TYPE:
		p.Value = new(AGENT)
	case GROUP_META_TYPE:
		p.Value = new(GROUP)
	case "":
		return fmt.Errorf("missing PARTY _type field")
	default:
		return fmt.Errorf("PARTY unexpected _type %s", t)
	}

	return json.Unmarshal(data, p.Value)
}

const VERSIONED_PARTY_META_TYPE string = "VERSIONED_PARTY"

type VERSIONED_PARTY struct {
	MetaType    util.Optional[string] `json:"_type,omitzero"`
	UID         HIER_OBJECT_ID        `json:"uid"`
	OwnerID     OBJECT_REF            `json:"owner_id"`
	TimeCreated DV_DATE_TIME          `json:"time_created"`
}

func (v VERSIONED_PARTY) Validate() []ValidationError {
	var errors []ValidationError

	if v.MetaType.IsSet() && v.MetaType.V != VERSIONED_PARTY_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "VERSIONED_PARTY",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid VERSIONED_PARTY _type field: %s", v.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", VERSIONED_PARTY_META_TYPE),
		})
	}

	return errors
}

func (v *VERSIONED_PARTY) SetMetaType() {
	v.MetaType = util.Some(VERSIONED_PARTY_META_TYPE)
}

const ROLE_META_TYPE string = "ROLE"

type ROLE struct {
	MetaType             util.Optional[string]               `json:"_type,omitzero"`
	Name                 ANY_DV_TEXT                         `json:"name"`
	ArchetypeNodeID      string                              `json:"archetype_node_id"`
	UID                  util.Optional[ANY_UID_BASED_ID]     `json:"uid,omitzero"`
	Links                util.Optional[[]LINK]               `json:"links,omitzero"`
	ArchetypeDetails     util.Optional[ARCHETYPED]           `json:"archetype_details,omitzero"`
	FeederAudit          util.Optional[FEEDER_AUDIT]         `json:"feeder_audit,omitzero"`
	Identities           []PARTY_IDENTITY                    `json:"identities"`
	Contacts             util.Optional[[]CONTACT]            `json:"contacts,omitzero"`
	Details              util.Optional[ANY_ITEM_STRUCTURE]   `json:"details,omitzero"`
	ReverseRelationships util.Optional[[]LOCATABLE_REF]      `json:"reverse_relationships,omitzero"`
	Relationships        util.Optional[[]PARTY_RELATIONSHIP] `json:"relationships,omitzero"`
	TimeValidity         util.Optional[DV_INTERVAL]          `json:"time_validity,omitzero"`
	Performer            PARTY_REF                           `json:"performer"`
	Capabilities         util.Optional[[]CAPABILITY]         `json:"capabilities,omitzero"`
}

func (r ROLE) Validate() []ValidationError {
	var errors []ValidationError

	if r.MetaType.IsSet() && r.MetaType.V != ROLE_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ROLE",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ROLE _type field: %s", r.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ROLE_META_TYPE),
		})
	}

	return errors
}

func (r *ROLE) SetMetaType() {
	r.MetaType = util.Some(ROLE_META_TYPE)
}

const PARTY_RELATIONSHIP_META_TYPE string = "PARTY_RELATIONSHIP"

type PARTY_RELATIONSHIP struct {
	MetaType         util.Optional[string]             `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                       `json:"name"`
	ArchetypeNodeID  string                            `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID]   `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]             `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]         `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]       `json:"feeder_audit,omitzero"`
	Details          util.Optional[ANY_ITEM_STRUCTURE] `json:"details,omitzero"`
	Target           PARTY_REF                         `json:"target"`
	TimeValidity     util.Optional[DV_INTERVAL]        `json:"time_validity,omitzero"`
	Source           PARTY_REF                         `json:"source"`
}

func (p PARTY_RELATIONSHIP) Validate() []ValidationError {
	var errors []ValidationError

	if p.MetaType.IsSet() && p.MetaType.V != PARTY_RELATIONSHIP_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "PARTY_RELATIONSHIP",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid PARTY_RELATIONSHIP _type field: %s", p.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", PARTY_RELATIONSHIP_META_TYPE),
		})
	}

	return errors
}

func (p *PARTY_RELATIONSHIP) SetMetaType() {
	p.MetaType = util.Some(PARTY_RELATIONSHIP_META_TYPE)
}

const PARTY_IDENTITY_META_TYPE string = "PARTY_IDENTITY"

type PARTY_IDENTITY struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	Details          ANY_ITEM_STRUCTURE              `json:"details"`
}

func (p PARTY_IDENTITY) Validate() []ValidationError {
	var errors []ValidationError

	if p.MetaType.IsSet() && p.MetaType.V != PARTY_IDENTITY_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "PARTY_IDENTITY",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid PARTY_IDENTITY _type field: %s", p.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", PARTY_IDENTITY_META_TYPE),
		})
	}

	return errors
}

func (p *PARTY_IDENTITY) SetMetaType() {
	p.MetaType = util.Some(PARTY_IDENTITY_META_TYPE)
}

const CONTACT_META_TYPE string = "CONTACT"

type CONTACT struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	Addresses        []ADDRESS                       `json:"addresses"`
	TimeValidity     util.Optional[DV_INTERVAL]      `json:"time_validity,omitzero"`
}

func (c CONTACT) Validate() []ValidationError {
	var errors []ValidationError

	if c.MetaType.IsSet() && c.MetaType.V != CONTACT_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "CONTACT",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid contact _type field: %s", c.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", CONTACT_META_TYPE),
		})
	}

	return errors
}

func (c *CONTACT) SetMetaType() {
	c.MetaType = util.Some(CONTACT_META_TYPE)
}

const ADDRESS_META_TYPE string = "ADDRESS"

type ADDRESS struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	Details          ANY_ITEM_STRUCTURE              `json:"details"`
}

func (a ADDRESS) Validate() []ValidationError {
	var errors []ValidationError

	if a.MetaType.IsSet() && a.MetaType.V != ADDRESS_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ADDRESS",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid address _type field: %s", a.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ADDRESS_META_TYPE),
		})
	}

	return errors
}

func (a *ADDRESS) SetMetaType() {
	a.MetaType = util.Some(ADDRESS_META_TYPE)
}

const CAPABILITY_META_TYPE string = "CAPABILITY"

type CAPABILITY struct {
	MetaType         util.Optional[string]           `json:"_type,omitzero"`
	Name             ANY_DV_TEXT                     `json:"name"`
	ArchetypeNodeID  string                          `json:"archetype_node_id"`
	UID              util.Optional[ANY_UID_BASED_ID] `json:"uid,omitzero"`
	Links            util.Optional[[]LINK]           `json:"links,omitzero"`
	ArchetypeDetails util.Optional[ARCHETYPED]       `json:"archetype_details,omitzero"`
	FeederAudit      util.Optional[FEEDER_AUDIT]     `json:"feeder_audit,omitzero"`
	Credentials      ANY_ITEM_STRUCTURE              `json:"credentials"`
	TimeValidity     util.Optional[DV_INTERVAL]      `json:"time_validity,omitzero"`
}

func (c CAPABILITY) Validate() []ValidationError {
	var errors []ValidationError

	if c.MetaType.IsSet() && c.MetaType.V != CAPABILITY_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "CAPABILITY",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid CAPABILITY _type field: %s", c.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", CAPABILITY_META_TYPE),
		})
	}

	return errors
}

func (c *CAPABILITY) SetMetaType() {
	c.MetaType = util.Some(CAPABILITY_META_TYPE)
}

const ACTOR_META_TYPE string = "ACTOR"

// Abstract type
type ACTOR struct {
	MetaType             util.Optional[string]               `json:"_type,omitzero"`
	Name                 ANY_DV_TEXT                         `json:"name"`
	ArchetypeNodeID      string                              `json:"archetype_node_id"`
	UID                  util.Optional[ANY_UID_BASED_ID]     `json:"uid,omitzero"`
	Links                util.Optional[[]LINK]               `json:"links,omitzero"`
	ArchetypeDetails     util.Optional[ARCHETYPED]           `json:"archetype_details,omitzero"`
	FeederAudit          util.Optional[FEEDER_AUDIT]         `json:"feeder_audit,omitzero"`
	Identities           []PARTY_IDENTITY                    `json:"identities"`
	Contacts             util.Optional[[]CONTACT]            `json:"contacts,omitzero"`
	Details              util.Optional[ANY_ITEM_STRUCTURE]   `json:"details,omitzero"`
	ReverseRelationships util.Optional[[]LOCATABLE_REF]      `json:"reverse_relationships,omitzero"`
	Relationships        util.Optional[[]PARTY_RELATIONSHIP] `json:"relationships,omitzero"`
	Languages            util.Optional[[]ANY_DV_TEXT]        `json:"languages,omitzero"`
	Roles                util.Optional[PARTY_REF]            `json:"roles,omitzero"`
}

func (a ACTOR) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal abstract ACTOR type")
}

func (a *ACTOR) UnmarshalJSON(data []byte) error {
	return fmt.Errorf("cannot unmarshal abstract ACTOR type")
}

type ANY_ACTOR struct {
	Value any
}

// Implement UnionType interface
func (a ANY_ACTOR) GetBaseType() reflect.Type {
	return reflect.TypeFor[ACTOR]()
}

func (a ANY_ACTOR) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Value)
}

func (a *ANY_ACTOR) UnmarshalJSON(data []byte) error {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return err
	}

	t := extractor.MetaType
	switch t {
	case PERSON_META_TYPE:
		a.Value = new(PERSON)
	case AGENT_META_TYPE:
		a.Value = new(AGENT)
	case GROUP_META_TYPE:
		a.Value = new(GROUP)
	case "":
		return fmt.Errorf("missing ACTOR _type field")
	default:
		return fmt.Errorf("ACTOR unexpected _type %s", t)
	}

	return json.Unmarshal(data, a.Value)
}

const PERSON_META_TYPE string = "PERSON"

type PERSON struct {
	MetaType             util.Optional[string]               `json:"_type,omitzero"`
	Name                 ANY_DV_TEXT                         `json:"name"`
	ArchetypeNodeID      string                              `json:"archetype_node_id"`
	UID                  util.Optional[ANY_UID_BASED_ID]     `json:"uid,omitzero"`
	Links                util.Optional[[]LINK]               `json:"links,omitzero"`
	ArchetypeDetails     util.Optional[ARCHETYPED]           `json:"archetype_details,omitzero"`
	FeederAudit          util.Optional[FEEDER_AUDIT]         `json:"feeder_audit,omitzero"`
	Identities           []PARTY_IDENTITY                    `json:"identities"`
	Contacts             util.Optional[[]CONTACT]            `json:"contacts,omitzero"`
	Details              util.Optional[ANY_ITEM_STRUCTURE]   `json:"details,omitzero"`
	ReverseRelationships util.Optional[[]LOCATABLE_REF]      `json:"reverse_relationships,omitzero"`
	Relationships        util.Optional[[]PARTY_RELATIONSHIP] `json:"relationships,omitzero"`
	Languages            util.Optional[[]ANY_DV_TEXT]        `json:"languages,omitzero"`
	Roles                util.Optional[PARTY_REF]            `json:"roles,omitzero"`
}

func (p PERSON) Validate() []ValidationError {
	var errors []ValidationError

	if p.MetaType.IsSet() && p.MetaType.V != PERSON_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "PERSON",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid PERSON _type field: %s", p.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", PERSON_META_TYPE),
		})
	}

	return errors
}

func (p *PERSON) SetMetaType() {
	p.MetaType = util.Some(PERSON_META_TYPE)
}

const ORGANISATION_META_TYPE string = "ORGANISATION"

type ORGANISATION struct {
	MetaType             util.Optional[string]               `json:"_type,omitzero"`
	Name                 ANY_DV_TEXT                         `json:"name"`
	ArchetypeNodeID      string                              `json:"archetype_node_id"`
	UID                  util.Optional[ANY_UID_BASED_ID]     `json:"uid,omitzero"`
	Links                util.Optional[[]LINK]               `json:"links,omitzero"`
	ArchetypeDetails     util.Optional[ARCHETYPED]           `json:"archetype_details,omitzero"`
	FeederAudit          util.Optional[FEEDER_AUDIT]         `json:"feeder_audit,omitzero"`
	Identities           []PARTY_IDENTITY                    `json:"identities"`
	Contacts             util.Optional[[]CONTACT]            `json:"contacts,omitzero"`
	Details              util.Optional[ANY_ITEM_STRUCTURE]   `json:"details,omitzero"`
	ReverseRelationships util.Optional[[]LOCATABLE_REF]      `json:"reverse_relationships,omitzero"`
	Relationships        util.Optional[[]PARTY_RELATIONSHIP] `json:"relationships,omitzero"`
	Languages            util.Optional[[]ANY_DV_TEXT]        `json:"languages,omitzero"`
	Roles                util.Optional[PARTY_REF]            `json:"roles,omitzero"`
}

func (o ORGANISATION) Validate() []ValidationError {
	var errors []ValidationError

	if o.MetaType.IsSet() && o.MetaType.V != ORGANISATION_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "ORGANISATION",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid ORGANISATION _type field: %s", o.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", ORGANISATION_META_TYPE),
		})
	}

	return errors
}

func (o *ORGANISATION) SetMetaType() {
	o.MetaType = util.Some(ORGANISATION_META_TYPE)
}

const GROUP_META_TYPE string = "GROUP"

type GROUP struct {
	MetaType             util.Optional[string]               `json:"_type,omitzero"`
	Name                 ANY_DV_TEXT                         `json:"name"`
	ArchetypeNodeID      string                              `json:"archetype_node_id"`
	UID                  util.Optional[ANY_UID_BASED_ID]     `json:"uid,omitzero"`
	Links                util.Optional[[]LINK]               `json:"links,omitzero"`
	ArchetypeDetails     util.Optional[ARCHETYPED]           `json:"archetype_details,omitzero"`
	FeederAudit          util.Optional[FEEDER_AUDIT]         `json:"feeder_audit,omitzero"`
	Identities           []PARTY_IDENTITY                    `json:"identities"`
	Contacts             util.Optional[[]CONTACT]            `json:"contacts,omitzero"`
	Details              util.Optional[ANY_ITEM_STRUCTURE]   `json:"details,omitzero"`
	ReverseRelationships util.Optional[[]LOCATABLE_REF]      `json:"reverse_relationships,omitzero"`
	Relationships        util.Optional[[]PARTY_RELATIONSHIP] `json:"relationships,omitzero"`
	Languages            util.Optional[[]ANY_DV_TEXT]        `json:"languages,omitzero"`
	Roles                util.Optional[PARTY_REF]            `json:"roles,omitzero"`
}

func (g GROUP) Validate() []ValidationError {
	var errors []ValidationError

	if g.MetaType.IsSet() && g.MetaType.V != GROUP_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "GROUP",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid GROUP _type field: %s", g.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", GROUP_META_TYPE),
		})
	}

	return errors
}

func (g *GROUP) SetMetaType() {
	g.MetaType = util.Some(GROUP_META_TYPE)
}

const AGENT_META_TYPE string = "AGENT"

type AGENT struct {
	MetaType             util.Optional[string]               `json:"_type,omitzero"`
	Name                 ANY_DV_TEXT                         `json:"name"`
	ArchetypeNodeID      string                              `json:"archetype_node_id"`
	UID                  util.Optional[ANY_UID_BASED_ID]     `json:"uid,omitzero"`
	Links                util.Optional[[]LINK]               `json:"links,omitzero"`
	ArchetypeDetails     util.Optional[ARCHETYPED]           `json:"archetype_details,omitzero"`
	FeederAudit          util.Optional[FEEDER_AUDIT]         `json:"feeder_audit,omitzero"`
	Identities           []PARTY_IDENTITY                    `json:"identities"`
	Contacts             util.Optional[[]CONTACT]            `json:"contacts,omitzero"`
	Details              util.Optional[ANY_ITEM_STRUCTURE]   `json:"details,omitzero"`
	ReverseRelationships util.Optional[[]PARTY_RELATIONSHIP] `json:"reverse_relationships,omitzero"`
	Relationships        util.Optional[[]PARTY_RELATIONSHIP] `json:"relationships,omitzero"`
	Languages            util.Optional[[]ANY_DV_TEXT]        `json:"languages,omitzero"`
	Roles                util.Optional[PARTY_REF]            `json:"roles,omitzero"`
}

func (a AGENT) Validate() []ValidationError {
	var errors []ValidationError

	if a.MetaType.IsSet() && a.MetaType.V != AGENT_META_TYPE {
		errors = append(errors, ValidationError{
			Model:          "AGENT",
			Path:           "._type",
			Message:        fmt.Sprintf("invalid AGENT _type field: %s", a.MetaType.V),
			Recommendation: fmt.Sprintf("Ensure _type field is set to '%s'", AGENT_META_TYPE),
		})
	}

	return errors
}

func (a *AGENT) SetMetaType() {
	a.MetaType = util.Some(AGENT_META_TYPE)
}
