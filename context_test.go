// +build go1.7

package way

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type IContextGroup interface {
	Handle(method, path string, handler http.Handler)
	NewGroup(path string) *Group
}

func TestContextParams(t *testing.T) {
	ctx := context.WithValue(context.Background(), wayContextKey("id"), "123")

	if v := Param(ctx, "id"); v != "123" {
		t.Errorf("expected '123', but got '%#v'", Param(ctx, "id"))
	}
}

func TestContextGroupMethods(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		testContextGroupMethods(t, scenario.RequestCreator, true, false)
		testContextGroupMethods(t, scenario.RequestCreator, false, false)
		testContextGroupMethods(t, scenario.RequestCreator, true, true)
		testContextGroupMethods(t, scenario.RequestCreator, false, true)
	}
}

func testContextGroupMethods(t *testing.T, reqGen RequestCreator, headCanUseGet bool, useContextRouter bool) {
	t.Logf("Running test: headCanUseGet %v, useContextRouter %v", headCanUseGet, useContextRouter)

	var result string
	makeHandler := func(method string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			result = method

			v := Param(r.Context(), "param")
			if v == "" {
				t.Error("missing key 'param' in context")
			}

			if headCanUseGet && (method == "GET" || v == "HEAD") {
				return
			}

			if v != method {
				t.Errorf("invalid key 'param' in context; expected '%s' but got '%s'", method, v)
			}
		})
	}

	var router http.Handler
	var rootGroup IContextGroup

	root := New()
	root.HeadCanUseGet = headCanUseGet
	t.Log(root.HeadCanUseGet)
	router = root
	rootGroup = root

	cg := rootGroup.NewGroup("/base").NewGroup("/user")
	cg.Handle("GET", "/:param", makeHandler("GET"))
	cg.Handle("POST", "/:param", makeHandler("POST"))
	cg.Handle("PATCH", "/:param", makeHandler("PATCH"))
	cg.Handle("PUT", "/:param", makeHandler("PUT"))
	cg.Handle("DELETE", "/:param", makeHandler("DELETE"))

	testMethod := func(method, expect string) {
		result = ""
		w := httptest.NewRecorder()
		r, _ := reqGen(method, "/base/user/"+method, nil)
		router.ServeHTTP(w, r)
		if expect == "" && w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Method %s not expected to match but saw code %d", method, w.Code)
		}

		if result != expect {
			t.Errorf("Method %s got result %s", method, result)
		}
	}

	testMethod("GET", "GET")
	testMethod("POST", "POST")
	testMethod("PATCH", "PATCH")
	testMethod("PUT", "PUT")
	testMethod("DELETE", "DELETE")

	if headCanUseGet {
		t.Log("Test implicit HEAD with HeadCanUseGet = true")
		testMethod("HEAD", "GET")
	} else {
		t.Log("Test implicit HEAD with HeadCanUseGet = false")
		testMethod("HEAD", "")
	}

	cg.Handle("HEAD", "/:param", makeHandler("HEAD"))
	testMethod("HEAD", "HEAD")
}

func TestNewContextGroup(t *testing.T) {
	router := New()
	group := router.NewGroup("/api")

	group.Handle("GET", "/v1", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`200 OK GET /api/v1`))
	}))

	group.Handle("GET", "/v2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`200 OK GET /api/v2`))
	}))

	tests := []struct {
		uri, expected string
	}{
		{"/api/v1", "200 OK GET /api/v1"},
		{"/api/v2", "200 OK GET /api/v2"},
	}

	for _, tc := range tests {
		r, err := http.NewRequest("GET", tc.uri, nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("GET %s: expected %d, but got %d", tc.uri, http.StatusOK, w.Code)
		}
		if got := w.Body.String(); got != tc.expected {
			t.Errorf("GET %s : expected %q, but got %q", tc.uri, tc.expected, got)
		}

	}
}

type ContextGroupHandler struct{}

//	adhere to the http.Handler interface
func (f ContextGroupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Write([]byte(`200 OK GET /api/v1`))
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func TestNewContextGroupHandler(t *testing.T) {
	router := New()
	group := router.NewGroup("/api")

	group.Handler("GET", "/v1", ContextGroupHandler{})

	tests := []struct {
		uri, expected string
	}{
		{"/api/v1", "200 OK GET /api/v1"},
	}

	for _, tc := range tests {
		r, err := http.NewRequest("GET", tc.uri, nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("GET %s: expected %d, but got %d", tc.uri, http.StatusOK, w.Code)
		}
		if got := w.Body.String(); got != tc.expected {
			t.Errorf("GET %s : expected %q, but got %q", tc.uri, tc.expected, got)
		}
	}
}

func TestDefaultContext(t *testing.T) {
	router := New()
	ctx := context.WithValue(context.Background(), wayContextKey("abc"), "def")
	expectContext := false

	router.Handle("GET", "/abc", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextValue := Param(r.Context(), "abc")
		if expectContext {
			if contextValue != "def" {
				t.Errorf("Unexpected context key value: %+v", contextValue)
			}
		} else {
			if contextValue != "" {
				t.Errorf("Expected blank context but key had value %+v", contextValue)
			}
		}
	}))

	r, err := http.NewRequest("GET", "/abc", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	t.Log("Testing without DefaultContext")
	router.ServeHTTP(w, r)

	router.DefaultContext = ctx
	expectContext = true
	w = httptest.NewRecorder()
	t.Log("Testing with DefaultContext")
	router.ServeHTTP(w, r)
}

func TestContextMuxSimple(t *testing.T) {
	router := NewContextMux()
	ctx := context.WithValue(context.Background(), wayContextKey("abc"), "def")
	expectContext := false

	router.Handle("GET", "/abc", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextValue := Param(r.Context(), "abc")
		if expectContext {
			if contextValue != "def" {
				t.Errorf("Unexpected context key value: %+v", contextValue)
			}
		} else {
			if contextValue != "" {
				t.Errorf("Expected blank context but key had value %+v", contextValue)
			}
		}
	}))

	r, err := http.NewRequest("GET", "/abc", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	t.Log("Testing without DefaultContext")
	router.ServeHTTP(w, r)

	router.DefaultContext = ctx
	expectContext = true
	w = httptest.NewRecorder()
	t.Log("Testing with DefaultContext")
	router.ServeHTTP(w, r)
}
