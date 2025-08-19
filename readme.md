# xerr

`xerr` is a Go package that displays detailed error pages for web applications. It captures errors and panics and renders an HTML page with stack traces, code snippets, and request information.

![Error Page Example](assets/images/screenshot1.png)

## Features

- Captures panics in HTTP handlers
- Middleware for integration with Go's `http.Handler` or `http.HandlerFunc`
- Displays stack frames and code snippets (configurable)
- Shows Go version, OS, architecture, and request details
- Customizable configuration:
  - `ShowSourceCode` (bool)
  - `MaxFrames` (int)
  - `Environment` (string)
  - `DebugMode` (bool)
  - `SkipFrames` (int)

## Installation

```bash
go get github.com/iMohamedSheta/xerr
````

## Usage

### Handle a panic in an HTTP handler

```go
package main

import (
    "net/http"
    "github.com/iMohamedSheta/xerr"
)

func main() {
    eh := xerr.New(nil) // default configuration

    http.HandleFunc("/", eh.MiddlewareFunc(func(w http.ResponseWriter, r *http.Request) {
        panic("Something went wrong!")
    }))

    http.ListenAndServe(":8080", nil)
}
```

### Use Middleware with a router

```go
mux := http.NewServeMux()
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    panic("Oops!")
})

eh := xerr.New(nil)
http.ListenAndServe(":8080", eh.Middleware(mux))
```

### Custom Configuration

```go
config := &xerr.Config{
    ShowSourceCode: false,
    MaxFrames:      10,
    Environment:    "production",
    DebugMode:      false,
    SkipFrames:     3,
}
eh := xerr.New(config)
```

## Functions

* `xerr.New(config *Config) *ErrorHandler` – Create a new error handler
* `(*ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err interface{})` – Render an error page
* `(*ErrorHandler) HandlePanic(w http.ResponseWriter, r *http.Request)` – Recover from panic in HTTP handlers
* `(*ErrorHandler) Middleware(next http.Handler) http.Handler` – HTTP middleware
* `(*ErrorHandler) MiddlewareFunc(next http.HandlerFunc) http.HandlerFunc` – HTTP middleware function

## Template

The default template is `assets/templates/error.html`. You can customize it to fit your needs.

## License

MIT License
