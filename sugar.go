package requests

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
)

// JSONBody set content-type and provide a jit marshaler to construct the body when the request is prepared
func (req *Request) JSONBody(value interface{}) *Request {
	return req.Body(applicationJSON, func() string {
		b, _ := json.Marshal(value)
		return string(b)
	})
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// BasicAuth add basic authorization
func (req *Request) BasicAuth(user, password string) *Request {
	return req.SecretHeader("Authorization", "Basic "+basicAuth(user, password))
}

// ContentType is a helper to set content type
func (req *Request) ContentType(contentType string) *Request {
	return req.Header("Content-Type", contentType)
}

// NewPost prepares a new *Request with method=POST
func NewPost(url interface{}) *Request {
	return New(url).Method(http.MethodPost)
}

// NewGet prepares a new *Request with method=GET
func NewGet(url interface{}) *Request {
	return New(url).Method(http.MethodGet)
}
