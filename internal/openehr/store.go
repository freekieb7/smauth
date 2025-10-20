package openehr

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/freekieb7/smauth/internal/database"
	"github.com/google/uuid"
)

type Store struct {
	DB *database.Database
}

func NewStore(db *database.Database) Store {
	return Store{
		DB: db,
	}
}

func (s *Store) NewEHR() EHR {
	return EHR{
		EHRID: HIER_OBJECT_ID{
			Value: uuid.New().String(),
		},
		EHRStatus: OBJECT_REF{
			Namespace: "local",
			Type:      "VERSIONED_EHR_STATUS",
			ID: ANY_OBJECT_ID{
				Value: HIER_OBJECT_ID{
					Value: uuid.New().String(),
				},
			},
		},
		EHRAccess: OBJECT_REF{
			Namespace: "local",
			Type:      "VERSIONED_EHR_ACCESS",
			ID: ANY_OBJECT_ID{
				Value: HIER_OBJECT_ID{
					Value: uuid.New().String(),
				},
			},
		},
		TimeCreated: DV_DATE_TIME{
			Value: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

func (s *Store) SaveEHR(ctx context.Context, ehr EHR) error {
	// Ensure meta types are set before saving
	forceSetMetaType(&ehr)

	_, err := s.DB.Conn.Exec(ctx, `INSERT INTO tbl_openehr_ehr (id, data, created_at) VALUES ($1, $2, NOW())`, ehr.EHRID.Value, ehr)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) SaveComposition(ctx context.Context, comp COMPOSITION) error {
	if !comp.UID.IsSet() {
		return errors.New("composition UID is required")
	}

	// Ensure meta types are set before saving
	forceSetMetaType(&comp)

	_, err := s.DB.Conn.Exec(ctx, `INSERT INTO tbl_openehr_composition (id, data, created_at) VALUES ($1, $2, NOW())`, comp.UID.Value, comp)
	if err != nil {
		return err
	}
	return nil
}

func forceSetMetaType(model any) {
	if model == nil {
		return
	}

	v := reflect.ValueOf(model)

	// We need a pointer to modify the original
	if v.Kind() != reflect.Pointer {
		// If not already a pointer, we can't modify the original
		// The caller needs to pass a pointer
		panic("EnforceMetaType requires a pointer to modify the original value")
	}

	forceSetMetaTypeReflect(v)
}

func forceSetMetaTypeReflect(v reflect.Value) {
	if !v.IsValid() {
		return
	}

	// Handle pointers
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}

		elem := v.Elem()

		// Check if the pointed-to value implements OpenEHRType interface
		if elem.CanAddr() && elem.Addr().CanInterface() {
			if openehrType, ok := elem.Addr().Interface().(OpenEHRType); ok {
				openehrType.SetMetaType()
			}
		}

		// Continue processing the dereferenced value
		forceSetMetaTypeReflect(elem)
		return
	}

	// Handle slices and arrays
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			// For slice/array elements, we need to check if they're addressable
			if elem.CanAddr() {
				forceSetMetaTypeReflect(elem.Addr())
			} else {
				// If not addressable, we can't modify them
				forceSetMetaTypeReflect(elem)
			}
		}
		return
	}

	// Handle maps
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			mapValue := v.MapIndex(key)
			// Map values are not addressable in Go, so we can't modify them directly
			forceSetMetaTypeReflect(mapValue)
		}
		return
	}

	// Handle structs
	if v.Kind() == reflect.Struct {
		// Check if this struct implements OpenEHRType
		if v.CanAddr() && v.Addr().CanInterface() {
			if openehrType, ok := v.Addr().Interface().(OpenEHRType); ok {
				openehrType.SetMetaType()
			}
		}

		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := v.Type().Field(i)

			// Skip unexported fields
			if !fieldType.IsExported() {
				continue
			}

			// For struct fields, check if they're addressable
			if field.CanAddr() {
				forceSetMetaTypeReflect(field.Addr())
			} else {
				forceSetMetaTypeReflect(field)
			}
		}
		return
	}
}
