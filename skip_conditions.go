package gopatch

import(
	"reflect"
)

// SkipCondition is used to determine if a struct field should or should not be skipped.
type SkipCondition func(field reflect.StructField) bool