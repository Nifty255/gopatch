package gopatch

import(
  "errors"
)

var errDestinationMustBeStructType = errors.New("dest interface must be a struct type")
var errDestinationMustBePointerType = errors.New("dest interface must be pointer to struct")

func errFieldMissingTag(field, tag string) error { return errors.New("field `"+field+"` is missing tag `"+tag+"`")}
func errFieldUnpermitted(field, cause string) error { return errors.New("field `"+field+"` is not permitted due to `"+cause+"`")}