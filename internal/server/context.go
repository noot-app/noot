package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/grantbirki/noot/internal/storage"
)

// Context provides a Gin-like interface for handling HTTP requests
// This allows us to gradually migrate from net/http to Gin patterns
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	store   storage.Store
}

// NewContext creates a new context from http.ResponseWriter and http.Request
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
		store:   getStore(r.Context()),
	}
}

// JSON sends a JSON response
func (c *Context) JSON(code int, obj interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(code)
	if err := json.NewEncoder(c.Writer).Encode(obj); err != nil {
		LogError("Failed to encode JSON response", err)
	}
}

// String sends a plain text response
func (c *Context) String(code int, text string) {
	c.Writer.WriteHeader(code)
	c.Writer.Write([]byte(text))
}

// HTML sends an HTML response
func (c *Context) HTML(code int, html string) {
	c.Writer.Header().Set("Content-Type", "text/html")
	c.Writer.WriteHeader(code)
	c.Writer.Write([]byte(html))
}

// Status sets the response status code
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// Header sets a response header
func (c *Context) Header(key, value string) {
	c.Writer.Header().Set(key, value)
}

// GetStore returns the storage store from context
func (c *Context) GetStore() storage.Store {
	return c.store
}

// GetRequestID returns the request ID
func (c *Context) GetRequestID() string {
	return getRequestID(c.Request.Context())
}

// Method returns the HTTP method
func (c *Context) Method() string {
	return c.Request.Method
}

// Path returns the request path
func (c *Context) Path() string {
	return c.Request.URL.Path
}

// Query returns a query parameter value
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// ServeFile serves a file
func (c *Context) ServeFile(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

// ErrorJSON sends a JSON error response using the standard error format
func (c *Context) ErrorJSON(code int, message string) {
	c.JSON(code, ErrorResponse{
		Error: message,
		Code:  code,
		Time:  time.Now().UTC(),
	})
}

// HandlerFunc defines a function that handles requests with our Context
type HandlerFunc func(*Context)

// WrapHandler wraps our HandlerFunc to work with net/http
func WrapHandler(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := NewContext(w, r)
		fn(c)
	}
}