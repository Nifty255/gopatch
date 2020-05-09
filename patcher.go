package gopatch

import(
  "reflect"
  "strings"
)

// Patcher is a configurable structure patcher.
type Patcher struct {
  config  PatcherConfig
}

// NewPatcher creates a new Patcher instance with the specified configuration. See `patcher_config.go`.
func NewPatcher(config PatcherConfig) *Patcher {

  return &Patcher{
    config: config,
  }
}

// Patch performs a patch operation on "dest", using the data in "patch". Patch returns a PatchResult if sucessful, or an error if not.
func (p Patcher) Patch(dest interface{}, patch map[string]interface{}) (*PatchResult, error) {

  return p.patch(dest, patch, p.config.PermittedFields, "")
}

func (p Patcher) patch(dest interface{}, patch map[string]interface{}, permitted []string, path string) (*PatchResult, error) {

  // Get the reflect value of `dest` and test that is a pointer.
  valueOfDest := reflect.ValueOf(dest)
  if valueOfDest.Kind() != reflect.Ptr {
    return nil, errDestinationMustBePointerType
  }
  
  // Get the actual struct data from the pointer and its type data
  valueOfDest = valueOfDest.Elem()
  typeOfDest := valueOfDest.Type()
  
  // Ensure the type is in fact a struct.
  if typeOfDest.Kind() != reflect.Struct {
    return nil, errDestinationMustBeStructType
  }
  
  // Initialize and allocate space for the results.
  results := PatchResult{
    Fields: make([]string, 0, len(patch)*100),
    Unpermitted: make([]string, 0, len(patch)*100),
    Map: make(map[string]interface{}, len(patch)*100),
  }

  // For each field in the destination struct,
  for i := 0; i < typeOfDest.NumField(); i++ {

    fieldT := typeOfDest.Field(i)
    fieldV := valueOfDest.Field(i)

    // Skip this field if it can't be set.
    if !valueOfDest.Field(i).CanSet() {
      continue
    }

    // Get the name of the field to check for in the patch map, defaulting to the field's struct field name.
    fieldName := fieldT.Name
    if p.config.PatchSource != "" && p.config.PatchSource != "struct" {
      
      testFieldName := fieldT.Tag.Get(p.config.PatchSource)
      if fieldName != "" {
        fieldName = testFieldName
      } else if p.config.PatchErrors {
        return nil, errFieldMissingTag(fieldName, p.config.PatchSource)
      }
    }

    // Check that the field isn't unpermitted by tag. Doing this before checking the permitted list placed priority on the tag.
    if fieldT.Tag.Get("gopatch") == "-" {
      if p.config.UnpermittedErrors { return nil, errFieldUnpermitted(fieldName, "permitted array") }
      continue
    }

    // Check that the field is permitted by the array, or the permitted array is empty.
    if p.config.PermittedFields != nil && len(p.config.PermittedFields) > 0 {

      allowed := false
      for _, permit := range(permitted) {

        // Break if it's an asterisk. It's auto-permitted.
        if permit == "*" {
          allowed = true
          break
        }

        // Permit if exact match or "match.*".
        if permit == fieldName ||  permit == fieldName+".*" {
          allowed = true
          break
        }
      }

      // Skip field or error if it wasn't permitted.
      if !allowed {
        if p.config.UnpermittedErrors { return nil, errFieldUnpermitted(fieldName, "permitted array") }
        continue
      }
    }

    // Get the patch value based on the fieldName.
    if val, ok := patch[fieldName]; ok {

      v := reflect.ValueOf(val)

      // Easily assign the value if both ends' kinds are the same
      if fieldV.Kind() == v.Kind() {
        fieldV.Set(v)
        
        // TODO: Add data about the successful update to the results.

      // Else, if the kind is struct,
      } else if fieldV.Kind() == reflect.Struct {

        // If the map field's kind isn't map[string]interface{}, skip it.
        if v.Kind() != reflect.Map || v.Type().Key().Kind() != reflect.String || v.Type().Elem().Kind() != reflect.Interface { continue }
        
        // If the gopatch tag specifies "replace", reset the current field value to its zero value.
        if fieldT.Tag.Get("gopatch") == "replace" {
          fieldV.Set(reflect.Zero(fieldT.Type))
        }

        // Patch the field, even if it was reset, by recursion.
        if !fieldV.CanAddr() { continue }
        prepend := path
        if prepend != "" { prepend += "."+fieldName } else { prepend = fieldName }
        /*results, err :=*/ p.patch(fieldV.Addr(), val.(map[string]interface{}), getPermittedAtPath(permitted, prepend), prepend)

        // TODO: Add data about the successful update to the results.

      // Else, if the kind is ptr and the Elem type is struct,
      } else if fieldV.Kind() == reflect.Ptr && fieldV.Elem().Kind() == reflect.Struct {

        // If the map field's kind isn't map[string]interface{}, skip it.
        if v.Kind() != reflect.Map || v.Type().Key().Kind() != reflect.String || v.Type().Elem().Kind() != reflect.Interface { continue }
        
        // If the gopatch tag specifies "replace", reset the current field value to its zero value.
        if fieldT.Tag.Get("gopatch") == "replace" {
          fieldV.Set(reflect.Zero(fieldT.Type))
        }

        // Patch the field, even if it was reset, by recursion.
        if !fieldV.CanAddr() { continue }
        prepend := path
        if prepend != "" { prepend += "."+fieldName } else { prepend = fieldName }
        /*results, err :=*/ p.patch(fieldV, val.(map[string]interface{}), getPermittedAtPath(permitted, prepend), prepend)

        // TODO: Add data about the successful update to the results.
        
      // Else, try to match the kind via the updaters.
      } else {

        updateSuccess := false
        for _, updater := range Updaters {

          // Try to update, breaking if successful
          if updateSuccess = updater(fieldV, v); updateSuccess { break }
        }

        // TODO: Add data about the successful update to the results.
      }
    }
  }

  return &results, nil
}

func (p *Patcher) saveToResults(r *PatchResult, val reflect.Value) error {

  return nil
}

func getPermittedAtPath(permitted []string, path string) []string {

  if len(permitted) == 0 { return []string{ "*" } }

  if path == "" { return permitted }

  out := make([]string, 0, len(permitted))

  for _, permitted := range(permitted) {

    if permitted == path+".*" { return []string{ "*" } }

    if strings.HasPrefix(permitted, path+".") { out = append(out, permitted[len(path)+1:])}
  }

  return out
}