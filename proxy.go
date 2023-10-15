package goxy

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type target struct {
	path string
	url  *url.URL
}

type proxy struct {
	client  HttpClient
	targets []*target
}

func New() *proxy {
	client := &http.Client{}
	targets := make([]*target, 0)
	return &proxy{client, targets}
}

func (pxy *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target, err := pxy.getTarget(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	r.Host = target.url.Host
	r.URL.Host = target.url.Host
	r.URL.Scheme = target.url.Scheme

	path, _ := strings.CutPrefix(r.URL.Path, target.path)
	r.URL.Path = fmt.Sprintf("%s%s", target.url.Path, path)
	r.RequestURI = ""

	res, err := pxy.client.Do(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.Copy(w, res.Body)
}

func (pxy *proxy) getTarget(prefix string) (*target, error) {
	for _, target := range pxy.targets {
		if strings.HasPrefix(prefix, target.path) {
			return target, nil
		}
	}

	return nil, errors.New("no target found")
}
