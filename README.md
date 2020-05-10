# Go PATCH
![Project status](https://img.shields.io/badge/version-1.0.2-green.svg)
[![Build Status](https://travis-ci.org/Nifty255/gopatch.svg?branch=master)](https://travis-ci.org/Nifty255/gopatch)
[![Coverage Status](https://coveralls.io/repos/github/Nifty255/gopatch/badge.svg?branch=master)](https://coveralls.io/github/Nifty255/gopatch?branch=master)
[![GoDoc](https://godoc.org/github.com/Nifty255/gopatch?status.svg)](https://godoc.org/github.com/Nifty255/gopatch)

Go PATCH is a structure patching package designed for receiving HTTP PATCH requests with a body of data, and applying it to a structure without affecting any other data in it.

Use the default patcher...

```go
type User struct {

  ID           int     `json:"id"`
  Username     string  `json:"username"`
  EmailAddress string  `json:"email_address"`
  IsBanned     bool    `json:"is_banned"`
}

user := User{ Username: "Nifty255", EmailAddress: "really_old@address.com"}

results, err := gopatch.Default().Patch(user, map[string]interface{}{
  "EmailAddress": "shiny_new@address.com",
})
```

Or configure your own!

```go
patcher := gopatch.NewPatcher(gopatch.PatcherConfig{
  PermittedFields: []string{ "username", "email_address" },
  UnpermittedErrors: true,
  PatchSource: "json",
})

// A nefarious user is trying to unban their own account.
nefariousPatchRequest := map[string]interface{}{
  "username": "an_inappropriate_word",
  "is_banned": false,
}

results, err := gopatch.Default().Patch(user, nefariousPatchRequest)

// err != nil
```

Using Go PATCH, structures can be patched by struct field name, or any name provided by any tag, including "json", and "bson". Furthermore, results are returned explaining which fields are patched, which fields weren't permitted, and even a map, sourced from the struct field names or any tag's names, which can be used to also patch a database representation.