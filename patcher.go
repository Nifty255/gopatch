package gopatch

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

// Patch performs a patch operation on "in", using the data in "patch". Patch returns a PatchResult if sucessful, or an error if not.
func (p Patcher) Patch(in interface{}, patch map[string]interface{}) (PatchResult, error) {

  return PatchResult{}, nil
}