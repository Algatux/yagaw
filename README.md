# yagaw

Lightweight Go HTTP router and small server helper.

## Overview

yagaw provides a minimal routing layer and a `Server` wrapper built on Go's standard `net/http`. It focuses on simple route registration, method-based dispatch, and basic parameterized route matching using a `{name}` syntax.

## Highlights

- Register routes per HTTP method: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, etc.
- Parameterized paths such as `/users/{id}` (supports alphanumeric, hyphen and underscore).
- `Server` helper to run an `http.Server` backed by the `Router`.
- Small dependency: uses `github.com/pho3b/tiny-logger` for logging.

## API Summary

- `yagaw.NewServer(addr string, port int) *Server` — create a new server.
- `(*Server).Run()` — start the HTTP server (blocking).
- `(*Server).GetRouter() *Router` — access the router to register routes.
- `(*Router).RegisterRoute(method HttpRequestMethod, path string, handler RequestHandler)` — register a route.
- `(*Router).RegisteredRoutes() *RequestHandlerMap` — inspect registered routes.

## Behavior notes

- Exact path matches are attempted first. If not found, parameterized route patterns (converted into regex at registration time) are tried.
- Parameter patterns are defined with `{name}` and are converted to `([a-z0-9-_]+)` when registered. The router does not automatically inject parameter values into the `http.Request` — handlers can extract values from `req.URL.Path` using string-splitting or regex extraction.
- Unmatched requests return a plain `404 - Page not found` response.

## Quick example

```go
package main

import (
    "fmt"
    "net/http"

    "github.com/Algatux/yagaw"
    "github.com/pho3b/tiny-logger/logs/log_level"
)

func main() {
    // Optional: configure logger level
    yagaw.Log = yagaw.InitLogger(log_level.ErrorLvlName)

    s := yagaw.NewServer("localhost", 8080)
    r := s.GetRouter()

    r.RegisterRoute(yagaw.GET, "/hello", func(w http.ResponseWriter, req *http.Request) {
        w.Header().Set("Content-Type", "text/plain")
        fmt.Fprint(w, "Hello, yagaw!")
    })

    // Parameterized route example — extract parameter manually inside handler
    r.RegisterRoute(yagaw.GET, "/users/{id}", func(w http.ResponseWriter, req *http.Request) {
        // simple extraction: split path ("/users/123" -> ["","users","123"])
        parts := strings.Split(req.URL.Path, "/")
        if len(parts) >= 3 {
            id := parts[2]
            fmt.Fprintf(w, "user id: %s", id)
            return
        }
        http.NotFound(w, req)
    })

    s.Run()
}
```

## Tests

Run unit tests and benchmarks with:

```bash
go test ./...
go test -bench=.
```

## Files of interest

- `server.go` — `Server` wrapper and `InitLogger` helper.
- `router.go` — route registration and pattern matching implementation.
- `router_test.go` — tests and benchmarks for the router behavior.

## Contributing

Issues and pull requests are welcome. Please run the test suite before submitting changes.

## License

See the `LICENSE` file in this repository.
