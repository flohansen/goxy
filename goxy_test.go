package goxy

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	mocks "github.com/flohansen/goxy/_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProxy(t *testing.T) {
	ctrl := gomock.NewController(t)

	clientMock := mocks.NewMockHttpClient(ctrl)
	pxy := proxy{
		client: clientMock,
		targets: createTargetUrls(map[string]string{
			"/some/target": "http://host1/any/other/target",
		}),
	}

	t.Run("should return 404 NOT FOUND if there is no target for the request", func(t *testing.T) {
		// given
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/unknown/path", nil)

		// when
		pxy.ServeHTTP(w, r)

		// then
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 500 INTERNAL SERVER ERROR if sending request to target fails", func(t *testing.T) {
		// given
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/some/target/with/more/subroutes", nil)

		expectedRequest := httptest.NewRequest("GET", "/any/other/target/with/more/subroutes", nil)
		expectedRequest.Host = "host1"
		expectedRequest.URL.Host = "host1"
		expectedRequest.URL.Scheme = "http"
		expectedRequest.RequestURI = ""

		clientMock.
			EXPECT().
			Do(expectedRequest).
			Return(nil, errors.New("error when sending request"))

		// when
		pxy.ServeHTTP(w, r)

		// then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.True(t, ctrl.Satisfied())
	})

	t.Run("should redirect request to registered target", func(t *testing.T) {
		// given
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/some/target/with/more/subroutes", nil)

		expectedRequest := httptest.NewRequest("GET", "/any/other/target/with/more/subroutes", nil)
		expectedRequest.Host = "host1"
		expectedRequest.URL.Host = "host1"
		expectedRequest.URL.Scheme = "http"
		expectedRequest.RequestURI = ""

		clientMock.
			EXPECT().
			Do(expectedRequest).
			Return(&http.Response{
				Body: io.NopCloser(strings.NewReader("response from target")),
			}, nil)

		// when
		pxy.ServeHTTP(w, r)

		// then
		res := w.Result()
		responseText, _ := io.ReadAll(res.Body)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "response from target", string(responseText))
		assert.True(t, ctrl.Satisfied())
	})
}

func createTargetUrls(mappings map[string]string) []*target {
	var targets []*target

	for path, host := range mappings {
		targetUrl, err := url.Parse(host)
		if err != nil {
			panic(err)
		}

		targets = append(targets, &target{
			path: path,
			url:  targetUrl,
		})
	}

	return targets
}
