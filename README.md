# errors

This library is based on the amazing https://commandcenter.blogspot.mx/2017/12/error-handling-in-upspin.html post

## Install
```
go get github.com/mishudark/errors
```

# Usage
`errors` implements standard errors package, with useful extras, at least it requires the original error, and a description, if a nil error is provided, it will returns a nil

```go
errors.E(err, "validation failed", errors.Invalid)
errors.E(errors.Errorf("input must be lowercase: %s", topic), "validation failed", errors.Invalid)

// you can add metadata
meta := errors.MetaData{"foo": "bar"}
errors.E(err, "error saving user profile", errors.IO, meta)
```

This produces the following output:

```json
{
    "detail": {
        "foo": "bar",
    },
    "type": "I/O error",
    "error": "error saving user profile",
    "code": 3
}
```
