package openehr

import "reflect"

type UnionType interface {
	GetBaseType() reflect.Type
}
