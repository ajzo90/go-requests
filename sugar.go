package requests

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
)

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

func (req *Request) BasicAuth(user, password string) *Request {
	return req.Header("Authorization", "Basic "+basicAuth(user, password))
}

func (req *Request) ContentType(contentType string) *Request {
	return req.Header("Content-Type", contentType)
}

func NewPost(url interface{}) *Request {
	return New(url).Method(http.MethodPost)
}

func NewGet(url interface{}) *Request {
	return New(url).Method(http.MethodGet)
}
