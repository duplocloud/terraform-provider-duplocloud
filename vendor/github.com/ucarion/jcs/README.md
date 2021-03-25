# jcs

[![GoDev](https://img.shields.io/static/v1?label=godev&message=reference&color=00add8)][godev]
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/ucarion/jcs/tests?label=tests&logo=github)](https://github.com/ucarion/jcs/actions)

This package is a Golang implementation of [JSON Canonicalization Scheme
("JCS")][jcs], aka [RFC 8785][rfc]. This package complies with all test data
supplied by the JCS spec authors.

## Installation

To install this package, run:

```bash
go get github.com/ucarion/jcs
```

## Usage

This package works on the standard JSON-like Golang data structures, namely:

* `bool`, for JSON booleans
* `float64`, for JSON numbers
* `string`, for JSON strings
* `[]interface{}`, for JSON arrays
* `map[string]interface{}`, for JSON objects
* `nil` for JSON null

These are the types of data you get out of `json.Unmarshal` by default. So, to
encode some existing JSON data into canonical format, you can do something like:

```go
import (
  "encoding/json"
  "fmt"

  "github.com/ucarion/jcs"
)

fn main() {
  input := `{"z": [1, 2, 3], "a": "foo" }`

  var v interface{}
  if err := json.Unmarshal([]byte(input), &v); err != nil {
    panic(err)
  }

  // See note about error handling below
  out, _ := jcs.Format(v)
  fmt.Println(out)
}
```

If you have an existing buffer you'd prefer to output to instead, use
`jcs.Append` instead of `jcs.Format`:

```go
var buf []byte

// See note about error handling below
buf, _ = jcs.Append(buf, &input)
```

You can, of course, pass your own data structures to `Format` or `Append`, but
note that you will get an error if you use a type other than those described
above.

## Error Handling

See the [documentation](https://pkg.go.dev/github.com/ucarion/jcs) for more details,
but note that `Append` and `Format` may return errors if any of the following are true:

* The inputted data contains types other than those listed in "Usage".
* The inputted data contains `NaN`.
* The inputted data contains positive or negative infinity.

If your inputted data is an `interface{}` from `json.Unmarshal`, then you do not
need to worry about errors from this package. Otherwise, you will need to
somehow ensure you don't pass unsupported data, or otherwise add support for an
error from this package.

## Contributing

A note on testing in this package: the `TestFormatFloat100M` test in this
package is disabled by default, because it takes too long time and compute power
to run by default. Plus, the required test file takes 3.8G of disk space.

If you would like to thoroughly test the encoding of `float64` values in this
package, download `es6testfile100m.txt` from the link in [the `testdata` dir of
the JCS repo][testdata]. When running tests, pass `JCS_TEST_100M=1`, for
instance:

```bash
JCS_TEST_100M=1 go test ./... -v
```

Passing `-v` will show you a progress indication once for every million tests
run.

[godev]: https://pkg.go.dev/github.com/ucarion/jcs
[jcs]: https://github.com/cyberphone/json-canonicalization
[rfc]: https://www.rfc-editor.org/rfc/rfc8785.html
[testdata]: https://github.com/cyberphone/json-canonicalization/tree/master/testdata
