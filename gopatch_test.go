package gopatch

import(
  "strings"
  "testing"
)

func TestDefault(t *testing.T) {

  t.Run("default-has-default-config", func(t *testing.T) {

    patcher := Default()

    if patcher.config.EmbedPath != defaultConfig.EmbedPath {
      t.Errorf("Expected default embed path %q but got %q", defaultConfig.EmbedPath, patcher.config.EmbedPath)
    }
    if patcher.config.PermittedFields != nil && defaultConfig.PermittedFields != nil {
      for _, g := range(patcher.config.PermittedFields) {

        found := false
        for _, d := range(defaultConfig.PermittedFields) {
          if g == d {
            found = true
            break
          }
        }
        if !found {
          t.Errorf("Expected default permitted fields [%q] but got [%q]", strings.Join(defaultConfig.PermittedFields, ", "), strings.Join(patcher.config.PermittedFields, ", "))
          break
        }
      }
      for _, d := range(defaultConfig.PermittedFields) {

        found := false
        for _, g := range(patcher.config.PermittedFields) {
          if g == d {
            found = true
            break
          }
        }
        if !found {
          t.Errorf("Expected default permitted fields [%q] but got [%q]", strings.Join(defaultConfig.PermittedFields, ", "), strings.Join(patcher.config.PermittedFields, ", "))
          break
        }
      }
    } else if (patcher.config.PermittedFields == nil || defaultConfig.PermittedFields == nil) && !(patcher.config.PermittedFields == nil && defaultConfig.PermittedFields == nil) {
      t.Errorf("Expected default permitted fields [%v] but got [%v]", defaultConfig.PermittedFields, patcher.config.PermittedFields)
    }
    if patcher.config.UnpermittedErrors != defaultConfig.UnpermittedErrors {
      t.Errorf("Expected default unpermitted-causes-errors %v but got %v", defaultConfig.UnpermittedErrors, patcher.config.UnpermittedErrors)
    }
    if patcher.config.UpdatedFieldErrors != defaultConfig.UpdatedFieldErrors {
      t.Errorf("Expected default missing-field-tag-causes-errors (fields) %v but got %v", defaultConfig.UpdatedFieldErrors, patcher.config.UpdatedFieldErrors)
    }
    if patcher.config.UpdatedFieldSource != defaultConfig.UpdatedFieldSource {
      t.Errorf("Expected default field source %q but got %q", defaultConfig.UpdatedFieldSource, patcher.config.UpdatedFieldSource)
    }
    if patcher.config.UpdatedMapErrors != defaultConfig.UpdatedMapErrors {
      t.Errorf("Expected default missing-field-tag-causes-errors (map) %v but got %v", defaultConfig.UpdatedMapErrors, patcher.config.UpdatedMapErrors)
    }
    if patcher.config.UpdatedMapSource != defaultConfig.UpdatedMapSource {
      t.Errorf("Expected default map source %q but got %q", defaultConfig.UpdatedMapSource, patcher.config.UpdatedMapSource)
    }
  })

  t.Run("default-patches", func(t *testing.T) {

    type TestStruct struct {
      Field1  string
      Field2  int
      Field3  bool
    }

    testInstance := TestStruct{}

    result, err := Default().Patch(&testInstance, map[string]interface{}{
      "Field2": 255,
    })

    // Test for unexpected errors.
    if err != nil {
      t.Errorf("Unexpected patch error: %q", err.Error())
      return
    }

    // Test to see if the instance was patched.
    if testInstance.Field1 != "" || testInstance.Field2 != 255 || testInstance.Field3 {
      t.Errorf("Expected patch to only patch Field2. Patch affected struct so: %v", testInstance)
      return
    }

    // Test to see if the instance was patched.
    if len(result.Fields) != 1 || result.Fields[0] != "Field2" {
      t.Errorf("Expected patch result fields to contain exactly \"Field2\". Contained [%v]", strings.Join(result.Fields, ", "))
      return
    }
  })
}