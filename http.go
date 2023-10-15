package goxy

import "net/http"

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}
