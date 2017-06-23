package way

import (
	"context"
	"net/http"
)

// ContextGroup is a wrapper around Group, with the purpose of mimicking its API, but with the use of http.HandlerFunc-based handlers.
// Instead of passing a parameter map via the handler (i.e. httptreemux.HandlerFunc), the path parameters are accessed via the request
// object's context.
//type ContextGroup struct {
//	group *Group
//}

// wayContextKey is the context key type for storing
// parameters in context.Context.
type wayContextKey string

// wayCatchAllKey is the context key for storing
// catch-all values in context.Context.
const wayCatchAllKey = wayCatchAllType("...")

type wayCatchAllType string

// Param gets the path parameter from the specified Context.
// Returns an empty string if the parameter was not found.
func Param(ctx context.Context, param string) string {
	return getParam(ctx, wayContextKey(param))
}

// CatchAll gets the catch-all value from the specified Context.
// Returns an empty string if the parameter was not found.
func CatchAll(ctx context.Context) string {
	return getParam(ctx, wayCatchAllKey)
}
func getParam(ctx context.Context, key interface{}) string {
	v := ctx.Value(key)
	if v == nil {
		return ""
	}
	vStr, ok := v.(string)
	if !ok {
		return ""
	}
	return vStr
}
func setParam(ctx context.Context, key interface{}, value string) context.Context {
	return context.WithValue(ctx, key, value)
}

// UsingContext wraps the receiver to return a new instance of a ContextGroup.
// The returned ContextGroup is a sibling to its wrapped Group, within the parent TreeMux.
// The choice of using a *Group as the receiver, as opposed to a function parameter, allows chaining
// while method calls between a TreeMux, Group, and ContextGroup. For example:
//
//              tree := httptreemux.New()
//              group := tree.NewGroup("/api")
//
//              group.GET("/v1", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
//                  w.Write([]byte(`GET /api/v1`))
//              })
//
//              group.UsingContext().GET("/v2", func(w http.ResponseWriter, r *http.Request) {
//                  w.Write([]byte(`GET /api/v2`))
//              })
//
//              http.ListenAndServe(":8080", tree)
//
//func (g *Group) UsingContext() *ContextGroup {
//	return &ContextGroup{g}
//}

// NewContextGroup adds a child context group to its path.
//func (cg *ContextGroup) NewContextGroup(path string) *ContextGroup {
//	return &ContextGroup{cg.group.NewGroup(path)}
//}

// NewGroup creates a new group for path.
//func (cg *ContextGroup) NewGroup(path string) *ContextGroup {
//	return cg.NewContextGroup(path)
//}

// Handle allows handling HTTP requests via an http.HandlerFunc, as opposed to an httptreemux.HandlerFunc.
// Any parameters from the request URL are stored in a map[string]string in the request's context.
//func (cg *Group) Handle(method, path string, handler http.HandlerFunc) {
//	cg.Handle(method, path, func(w http.ResponseWriter, r *http.Request, params map[string]string) {
//		if params != nil {
//			r = r.WithContext(context.WithValue(r.Context(), paramsContextKey, params))
//		}
//		handler(w, r)
//	})
//}

// Handler allows handling HTTP requests via an http.Handler interface, as opposed to an httptreemux.HandlerFunc.
// Any parameters from the request URL are stored in a map[string]string in the request's context.
func (cg *Group) Handler(method, path string, handler http.Handler) {
	cg.Handle(method, path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}))
}

// ContextParams returns the params map associated with the given context if one exists. Otherwise, an empty map is returned.
//func ContextParams(ctx context.Context) map[string]string {
//	if p, ok := ctx.Value(paramsContextKey).(map[string]string); ok {
//		return p
//	}
//	return map[string]string{}
//}

//type contextKey int

// paramsContextKey is used to retrieve a path's params map from a request's context.
//const paramsContextKey contextKey = 0
