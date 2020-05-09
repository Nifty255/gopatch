// Package gopatch allows structures to be patched in a multitude of configurable ways. Patching is accomplished via Patchers, and a default is
// initialized for immediate use, found by calling `gopatch.Default()`. All initialized patchers can be used multiple times, and are thread safe.
//
// Use the default patcher...
//
//     type User struct {
//     
//       ID           int     `json:"id"`
//       Username     string  `json:"username"`
//       EmailAddress string  `json:"email_address"`
//       IsBanned     bool    `json:"is_banned"`
//     }
//     
//     user := User{ Username: "Nifty255", EmailAddress: "really_old@address.com"}
//     
//     results, err := gopatch.Default().Patch(user, map[string]interface{}{
//       "EmailAddress": "shiny_new@address.com",
//     })
//
// Or configure your own!
//
//     patcher := gopatch.NewPatcher(gopatch.PatcherConfig{
//       PermittedFields: []string{ "username", "email_address" },
//       UnpermittedErrors: true,
//       PatchSource: "json",
//     })
//     
//     // A nefarious user is trying to unban their own account.
//     nefariousPatchRequest := map[string]interface{}{
//       "username": "an_inappropriate_word",
//       "is_banned": false,
//     }
//     
//     results, err := gopatch.Default().Patch(user, nefariousPatchRequest)
//     
//     // err != nil
//
// Field Name Sources
//
// Struct field names on Go often differ from field names of counterpart object representations such as JSON and BSON. This package refers to these
// representations as "Field Name Sources". In the above examples, the User struct's EmailAddress has a "struct" field name source ( resulting in the
// field name "EmailAddress"), and a "json" field name source (resulting in the field name "email_address"). Thus, if you want to patch a struct from
// a map built from JSON bytes, you need to add json tags AND configure your Patcher to use the struct's "json" field name sources. In the second
// example above, you can see the Patcher has been configured with the "json" field name source and thuse can use a JSON-derived map to patch.
//
// Patch Results
//
// Patch operations return results or an error when they fail. These results can be used for purposes ranging from logging suspicious activity to
// persisting the changes to a database row or document. The results consist of an array of fields which were successfully patched, an array of 
// fields which were found in the patch yet not permitted, and a map of the successfully updated fields and their values. Take the above example
// of patching a nefarious user's account. If `UnpermittedErrors` were false, the patch would succeed and result would not be nil; however, the
// Unpermitted array would contain "IsBanned", because the patching of that field wasn't permitted. Meanwhile, the "Fields" array would contain
// "Username" because it was permitted, and Map would contain the same data as `nefariousPatchRequest`, but without "is_banned".
//
// Gopatch Field Tag
//
// Patching behavior can be enforced while defining the structure by using the "gopatch" tag, which overrides configuration. This way, restrictions
// on how your database model can be patched can be limited while designing the model itself, rather than while designing the endpoint that patches
// it, reducing the chance of unexpected behavior.
//
//     type User struct {
//     
//       ID           int     `json:"id"        gopatch:"-"`        // NEVER patch this field, even if permitted in configuration.
//     
//       Username string      `json:"username"`                     // No gopatch tag allows normal patching behavior, acts like "patch" for structs.
//     
//       Profile  UserProfile `json:"profile"   gopatch:"patch"`    // Patch data for Profile will patch the fields inside.
//     
//       BanData  UserBanData `json:"ban_data"  gopatch:"replace"`  // Patch data for BanData will create a new zero-value BanData and patch that.
//     }
//
// When the gopatch tag "patch" is used, the PatchResult's Map field will contain the struct field's values flattened with dot-notation keys created
// using the absolute path to the struct patched. For example, if the above User struct's `Profile.Motto` field is patched, the result's Map field
// would contain the following data: `"profile.motto": "..."`. This facilitates the patch-embedded-fields behavior of embedded objects in database
// servers such as MongoDB.
//
// When the gopatch tag "replace" is used, the PatchResult's Map field will contain the struct field's values inside the struct field's patch key,
// exactly as presented to the Patcher. For example, if the above User struct's `BanData.Length` field is patched, the result's Map field would
// contain the following data: `"ban_data": map[string]interface{}{ "length": 30 }`. This facilitates the patch-whole-object behavior of embedded
// objects in database servers such as MongoDB.
package gopatch