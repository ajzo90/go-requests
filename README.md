# go-requests
Request builder for flexible and composable requests.


[![CICD](https://github.com/ajzo90/go-requests/actions/workflows/ci.yml/badge.svg)](https://github.com/ajzo90/go-requests/actions/workflows/ci.yml)
[![CICD](https://github.com/ajzo90/go-requests/actions/workflows/go.yml/badge.svg)](https://github.com/ajzo90/go-requests/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ajzo90/go-requests)](https://goreportcard.com/report/github.com/ajzo90/go-requests)
[![GoDoc](https://godoc.org/github.com/ajzo90/go-requests?status.svg)](https://godoc.org/github.com/ajzo90/go-requests)
[![License](https://shields.io/github/license/ajzo90/go-requests)](LICENSE)
[![Latest Version](https://shields.io/github/v/release/ajzo90/go-requests?display_name=tag&sort=semver)](https://github.com/ajzo90/go-requests/releases)
[![codecov](https://codecov.io/gh/ajzo90/go-requests/branch/main/graph/badge.svg?token=BDKHJVZCUY)](https://codecov.io/gh/ajzo90/go-requests)

## Usage
```go
var token = "secret"

var builder = requests.NewPost("example.com/test").
    Query("key", "val").
    Header("token", &token)

jsonResp, err := builder.ExecJSON()	

token = "super-secret"// update token		

jsonResp, err := builder.ExecJSON()

```
