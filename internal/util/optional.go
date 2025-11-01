package util

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

type OptionalValue interface {
	GetInnerType() reflect.Type
	IsSet() bool
}

type Optional[T any] struct {
	V T    `json:"-"` // Value
	E bool `json:"-"` // Exists
}

func Some[T any](v T) Optional[T] {
	return Optional[T]{V: v, E: true}
}

func None[T any]() Optional[T] {
	return Optional[T]{}
}

func (o Optional[T]) IsSet() bool {
	return o.E
}

func (o Optional[T]) Unwrap() T {
	if !o.E {
		panic("called Unwrap on a None value")
	}
	return o.V
}

func (o Optional[T]) UnwrapOr(defaultVal T) T {
	if !o.E {
		return defaultVal
	}
	return o.V
}

func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if !o.E {
		return []byte("null"), nil
	}
	return json.Marshal(o.V)
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		o.E = false
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	o.E = true
	o.V = v
	return nil
}

// Scan implements the SQL driver.Scanner interface.
func (o *Optional[T]) Scan(value any) error {
	if value == nil {
		o.E = false
		return nil
	}

	var v T
	switch t := any(&v).(type) {
	case interface{ Scan(any) error }:
		if err := t.Scan(value); err != nil {
			return err
		}
	default:
		v = value.(T)
	}

	o.V = v
	o.E = true

	return nil
}

// Value implements the driver Valuer interface.
func (o Optional[T]) Value() (driver.Value, error) {
	if !o.E {
		return nil, nil
	}
	switch t := any(o.V).(type) {
	case interface{ Value() (any, error) }:
		return t.Value()
	default:
		return o.V, nil
	}
}

func (o Optional[T]) String() string {
	if !o.E {
		return ""
	}

	return fmt.Sprintf("%v", o.V)
}

// Then in your Optional[T] implementation
func (o Optional[T]) GetInnerType() reflect.Type {
	var zero T
	return reflect.TypeOf(zero)
}
