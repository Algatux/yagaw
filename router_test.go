package yagaw

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterRoute(t *testing.T) {
	router := NewRouter()

	handler := func(req *http.Request, params Params) *HttpResponse {
		return NewHttpResponse(http.StatusOK).
			SetHeader("Content-Type", "text/plain").
			SetBody("test response")
	}

	router.RegisterRoute(GET, "/users", handler)
	routes := router.RegisteredRoutes()

	if routes == nil {
		t.Fatal("routes should not be nil")
	}

	if _, exists := (*routes)[GET]; !exists {
		t.Error("GET method should be registered")
	}
}

func TestServeHTTPExactPath(t *testing.T) {
	router := NewRouter()

	handler := func(req *http.Request, params Params) *HttpResponse {
		return NewHttpResponse(http.StatusOK).
			SetHeader("Content-Type", "text/plain").
			SetBody("exact path match")
	}

	router.RegisterRoute(GET, "/users", handler)

	req := httptest.NewRequest(string(GET), "/users", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rw.Code)
	}

	body := rw.Body.String()
	if body != "exact path match" {
		t.Errorf("expected 'exact path match', got %q", body)
	}
}

func TestServeHTTPPatternPath(t *testing.T) {
	router := NewRouter()

	handler := func(req *http.Request, params Params) *HttpResponse {
		return NewHttpResponse(http.StatusOK).
			SetHeader("Content-Type", "text/plain").
			SetBody("pattern matched")
	}

	router.RegisterRoute(GET, "/users/{id}", handler)

	tests := []struct {
		name      string
		path      string
		shouldAct bool
	}{
		{"valid id", "/users/123", true},
		{"id with hyphen", "/users/user-123", true},
		{"id with underscore", "/users/user_123", true},
		{"alphanumeric id", "/users/abc123", true},
		{"wrong path", "/posts/123", false},
		{"missing id", "/users/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(string(GET), tt.path, nil)
			rw := httptest.NewRecorder()

			router.ServeHTTP(rw, req)

			if tt.shouldAct {
				if rw.Code != http.StatusOK {
					t.Errorf("expected status 200, got %d", rw.Code)
				}
				if rw.Body.String() != "pattern matched" {
					t.Errorf("expected 'pattern matched', got %q", rw.Body.String())
				}
			} else {
				if rw.Code != http.StatusNotFound {
					t.Errorf("expected 404, got %d", rw.Code)
				}
			}
		})
	}
}

func TestServeHTTPMultipleParameters(t *testing.T) {
	router := NewRouter()

	handler := func(req *http.Request, params Params) *HttpResponse {
		return NewHttpResponse(http.StatusOK).
			SetBody("multi param matched")
	}

	router.RegisterRoute(GET, "/users/{userId}/posts/{postId}", handler)

	tests := []struct {
		name      string
		path      string
		shouldAct bool
	}{
		{"valid params", "/users/123/posts/456", true},
		{"params with hyphens", "/users/user-123/posts/post-456", true},
		{"missing second param", "/users/123/posts/", false},
		{"wrong structure", "/users/123/comments/456", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(string(GET), tt.path, nil)
			rw := httptest.NewRecorder()

			router.ServeHTTP(rw, req)

			if tt.shouldAct {
				if rw.Code != http.StatusOK {
					t.Errorf("expected status 200, got %d", rw.Code)
				}
			} else {
				if rw.Code != http.StatusNotFound {
					t.Errorf("expected 404, got %d", rw.Code)
				}
			}
		})
	}
}

func TestServeHTTPDifferentMethods(t *testing.T) {
	router := NewRouter()

	getHandler := func(req *http.Request, params Params) *HttpResponse {
		return NewHttpResponse(http.StatusOK).
			SetBody("GET")
	}
	postHandler := func(req *http.Request, params Params) *HttpResponse {
		return NewHttpResponse(http.StatusOK).
			SetBody("POST")
	}
	deleteHandler := func(req *http.Request, params Params) *HttpResponse {
		return NewHttpResponse(http.StatusOK).
			SetBody("DELETE")
	}

	router.RegisterRoute(GET, "/resource", getHandler)
	router.RegisterRoute(POST, "/resource", postHandler)
	router.RegisterRoute(DELETE, "/resource", deleteHandler)

	tests := []struct {
		method   HttpMethod
		expected string
	}{
		{GET, "GET"},
		{POST, "POST"},
		{DELETE, "DELETE"},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			req := httptest.NewRequest(string(tt.method), "/resource", nil)
			rw := httptest.NewRecorder()

			router.ServeHTTP(rw, req)

			if rw.Body.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, rw.Body.String())
			}
		})
	}
}

func TestServeHTTPMethodNotFound(t *testing.T) {
	router := NewRouter()

	handler := func(req *http.Request, params Params) *HttpResponse {
		return NewHttpResponse(http.StatusOK).
			SetBody("found")
	}

	router.RegisterRoute(GET, "/test", handler)

	// Request with unsupported method
	req := httptest.NewRequest(string(PATCH), "/test", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unsupported method, got %d", rw.Code)
	}
}

func TestServeHTTPRouteNotFound(t *testing.T) {
	router := NewRouter()

	handler := func(req *http.Request, params Params) *HttpResponse {
		return NewHttpResponse(http.StatusOK)
	}

	router.RegisterRoute(GET, "/users", handler)

	req := httptest.NewRequest(string(GET), "/posts", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rw.Code)
	}

	if rw.Body.String() != "404 - Page not found" {
		t.Errorf("expected '404 - Page not found', got %q", rw.Body.String())
	}
}

func TestNestedPathsWithParameters(t *testing.T) {
	router := NewRouter()

	handler := func(req *http.Request, params Params) *HttpResponse {
		return NewHttpResponse(http.StatusOK).
			SetBody("nested")
	}

	router.RegisterRoute(GET, "/api/v1/users/{id}/profile", handler)

	tests := []struct {
		name      string
		path      string
		shouldAct bool
	}{
		{"valid nested path", "/api/v1/users/123/profile", true},
		{"wrong version", "/api/v2/users/123/profile", false},
		{"missing profile", "/api/v1/users/123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(string(GET), tt.path, nil)
			rw := httptest.NewRecorder()

			router.ServeHTTP(rw, req)

			if tt.shouldAct {
				if rw.Code != http.StatusOK {
					t.Errorf("expected 200, got %d", rw.Code)
				}
			} else {
				if rw.Code != http.StatusNotFound {
					t.Errorf("expected 404, got %d", rw.Code)
				}
			}
		})
	}
}

func BenchmarkRegisterRoute(b *testing.B) {
	router := NewRouter()
	handler := func(req *http.Request, params Params) *HttpResponse { return NewHttpResponse(http.StatusOK) }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.RegisterRoute(GET, "/users/{id}", handler)
	}
}

func BenchmarkServeHTTPExact(b *testing.B) {
	router := NewRouter()
	handler := func(req *http.Request, params Params) *HttpResponse {
		return NewHttpResponse(http.StatusOK)
	}
	router.RegisterRoute(GET, "/users", handler)

	req := httptest.NewRequest(string(GET), "/users", nil)
	rw := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(rw, req)
	}
}

func BenchmarkServeHTTPPattern(b *testing.B) {
	router := NewRouter()
	handler := func(req *http.Request, params Params) *HttpResponse {
		return NewHttpResponse(http.StatusOK)
	}
	router.RegisterRoute(GET, "/users/{id}", handler)

	req := httptest.NewRequest(string(GET), "/users/123", nil)
	rw := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(rw, req)
	}
}
