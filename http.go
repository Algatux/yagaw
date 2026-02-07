package yagaw

import "net/http"

type HttpMethod string

const (
	GET     HttpMethod = `GET`
	HEAD    HttpMethod = `HEAD`
	OPTIONS HttpMethod = `OPTIONS`
	TRACE   HttpMethod = `TRACE`
	PUT     HttpMethod = `PUT`
	DELETE  HttpMethod = `DELETE`
	POST    HttpMethod = `POST`
	PATCH   HttpMethod = `PATCH`
	CONNECT HttpMethod = `CONNECT`
)

type Params map[string]any
type HttpRequestHandler func(req *http.Request, params Params) *HttpResponse

type HttpResponse struct {
	headers map[string]string
	status  int
	body    string
}

func (r *HttpResponse) SetHeader(key string, value string) *HttpResponse {
	r.headers[key] = value
	return r
}

func (r *HttpResponse) SetBody(body string) *HttpResponse {
	r.body = body
	return r
}

func NewHttpResponse(status int) *HttpResponse {
	return &HttpResponse{
		status:  status,
		headers: make(map[string]string),
	}
}
