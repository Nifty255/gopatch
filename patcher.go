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

// Patch performs a patch operation on "dest", using the data in "patch". Patch returns a PatchResult if sucessful, or an error if not. Patch can
// also patch embedded structs and pointers to embedded structs. If a patch exists for a nil embedded struct pointer, the pointer will be assigned a
// new zero-value struct before it is patched.
func (p Patcher) Patch(dest interface{}, patch map[string]interface{}) (*PatchResult, error) {

  // Error on invalid dest.
  if reflect.ValueOf(dest).Kind() != reflect.Ptr ||
    reflect.ValueOf(dest).IsNil() ||
    reflect.ValueOf(dest).Elem().Kind() != reflect.Struct {
    
    return nil, errDestInvalid
  }
  return p.patch(dest, patch, p.config.PermittedFields, p.config.EmbedPath)
}

func (p Patcher) patch(dest interface{}, patch map[string]interface{}, permitted []string, path string) (*PatchResult, error) {
  
  // Get the actual struct data from the pointer and its type data.
  valueOfDest := reflect.ValueOf(dest).Elem()
  typeOfDest := valueOfDest.Type()
  
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
      if fieldV.Kind() == v.Kind() && fieldV.Kind() != reflect.Map {
        fieldV.Set(v)
        
        // Add data about the successful update to the results.
        if err := p.saveToResults(&results, fieldT, val, path); err != nil { return nil, err }

        // Next!
        continue
      }

      // Check updater functions for a match.
      updateSuccess := false
      for _, updater := range Updaters {

        // Try to update, breaking if successful
        if updateSuccess = updater(fieldV, v); updateSuccess { break }
      }
      if updateSuccess {

        // Add data about the successful update to the results.
        if err := p.saveToResults(&results, fieldT, val, path); err != nil { return nil, err }

        // Next!
        continue
      }

      // Dereference the value if it's a pointer.
      if fieldV.Kind() == reflect.Ptr {

        // Ensure it's not nil, initializing to zero-value if needed.
        if fieldV.IsNil() {
          fieldV.Set(reflect.New(fieldV.Type().Elem()))
        }

        fieldV = fieldV.Elem()
      }

      // If the value is a struct, attempt to deep-patch it.
      if fieldV.Kind() == reflect.Struct {

        // If the map field's kind isn't map[string]interface{}, skip it.
        if v.Kind() != reflect.Map || v.Type().Key().Kind() != reflect.String || v.Type().Elem().Kind() != reflect.Interface { continue }
        
        // If the gopatch tag specifies "replace", reset the current field value to its zero value.
        replace := fieldT.Tag.Get("gopatch") == "replace"
        if replace {
          fieldV.Set(reflect.Zero(fieldT.Type))
        }

        // Patch the field, even if it was reset, by recursion.
        if !fieldV.CanAddr() { continue }
        full := path
        if full != "" { full += "."+fieldName } else { full = fieldName }
        deep, err := p.patch(fieldV.Addr().Interface(), val.(map[string]interface{}), getPermittedAtPath(permitted, full), path)

        // If an error occurred while deep-patching, bubble up immediately.
        if err != nil { return nil, err }

        // Merge deep-patched results into the current results.
        if err := p.mergeResults(&results, deep, fieldT, replace, path); err != nil { return nil, err }
      }
    }
  }

  return &results, nil
}

func (p *Patcher) saveToResults(r *PatchResult, dest reflect.StructField, patch interface{}, path string) error {

  // Get a field name for the fields array.
  fieldName := dest.Name
  if p.config.UpdatedFieldSource != "" && p.config.UpdatedFieldSource != "struct" {
    
    testFieldName := dest.Tag.Get(p.config.UpdatedFieldSource)
    if fieldName != "" {
      fieldName = testFieldName
    } else if p.config.UpdatedFieldErrors {
      return errFieldMissingTag(fieldName, p.config.UpdatedFieldSource)
    }
  }

  // Append.
  r.Fields = append(r.Fields, fieldName)

  // Get a field name for the map.
  fieldName = dest.Name
  if p.config.UpdatedMapSource != "" && p.config.UpdatedMapSource != "struct" {
    
    testFieldName := dest.Tag.Get(p.config.UpdatedMapSource)
    if fieldName != "" {
      fieldName = testFieldName
    } else if p.config.UpdatedMapErrors {
      return errFieldMissingTag(fieldName, p.config.UpdatedMapSource)
    }
  }

  // Add to map, prepended with a path if specified, otherwise with only the field name.
  if path == "" { r.Map[fieldName] = patch } else { r.Map[path+"."+fieldName] = patch }

  return nil
}

func (p *Patcher) mergeResults(top, deep *PatchResult, dest reflect.StructField, replace bool, path string) error {

  // Get a field name for the fields array.
  fieldName := dest.Name
  if p.config.UpdatedFieldSource != "" && p.config.UpdatedFieldSource != "struct" {
    
    testFieldName := dest.Tag.Get(p.config.UpdatedFieldSource)
    if fieldName != "" {
      fieldName = testFieldName
    } else if p.config.UpdatedFieldErrors {
      return errFieldMissingTag(fieldName, p.config.UpdatedFieldSource)
    }
  }

  full := path
  if full != "" { full += "."+fieldName } else { full = fieldName }

  // Map field names to path and append.
  if replace {
    top.Fields = append(top.Fields, path+"."+fieldName)
  } else {
    for _, field := range(deep.Fields) {
      top.Fields = append(top.Fields, full+"."+field)
    }
  }

  // Get a field name for the map.
  fieldName = dest.Name
  if p.config.UpdatedMapSource != "" && p.config.UpdatedMapSource != "struct" {
    
    testFieldName := dest.Tag.Get(p.config.UpdatedMapSource)
    if fieldName != "" {
      fieldName = testFieldName
    } else if p.config.UpdatedMapErrors {
      return errFieldMissingTag(fieldName, p.config.UpdatedMapSource)
    }
  }

  full = path
  if full != "" { full += "."+fieldName } else { full = fieldName }

  if replace {
    top.Map[full] = deep.Map
  } else {
    for k, v := range(deep.Map) {
      top.Map[full+"."+k] = v
    }
  }

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