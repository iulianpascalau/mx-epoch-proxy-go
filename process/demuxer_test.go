package process

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewDemuxer(t *testing.T) {
	t.Parallel()

	t.Run("can work with a nil or empty map", func(t *testing.T) {
		t.Parallel()

		instance := NewDemuxer(nil, nil)
		assert.NotNil(t, instance)
		assert.NotNil(t, instance.handlers)
		assert.Equal(t, 0, len(instance.handlers))

		instance = NewDemuxer(make(map[string]http.Handler), nil)
		assert.NotNil(t, instance)
		assert.NotNil(t, instance.handlers)
		assert.Equal(t, 0, len(instance.handlers))
	})
	t.Run("can work with defined handlers", func(t *testing.T) {
		t.Parallel()

		handler1 := &testscommon.HttpHandlerStub{}
		handler2 := &testscommon.HttpHandlerStub{}

		handlers := map[string]http.Handler{
			"route1": handler1,
			"route2": handler2,
		}

		instance := NewDemuxer(handlers, nil)
		assert.NotNil(t, instance)
		assert.NotNil(t, instance.handlers)
		assert.NotEqual(t, fmt.Sprintf("%p", instance.handlers), fmt.Sprintf("%p", handlers)) // pointers should not be equal
		assert.Equal(t, len(handlers), len(instance.handlers))
		assert.Equal(t, handlers, instance.handlers)
	})
}

func TestDemuxer_ServeHTTP(t *testing.T) {
	t.Parallel()

	t.Run("a not defined route without default should return 404", func(t *testing.T) {
		t.Parallel()

		handler1 := &testscommon.HttpHandlerStub{}
		handler2 := &testscommon.HttpHandlerStub{}

		handlers := map[string]http.Handler{
			"route1": handler1,
			"route2": handler2,
		}
		instance := NewDemuxer(handlers, nil)

		recorder := httptest.NewRecorder()
		instance.ServeHTTP(recorder, &http.Request{RequestURI: "unknown"})
		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})
	t.Run("a defined route without default should call the handler", func(t *testing.T) {
		t.Parallel()

		handler1 := &testscommon.HttpHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				_, _ = writer.Write([]byte("response"))
				writer.WriteHeader(http.StatusOK)
			},
		}
		handler2 := &testscommon.HttpHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				assert.Fail(t, "should have not called this handler")
			},
		}

		handlers := map[string]http.Handler{
			"/route1": handler1,
			"/route2": handler2,
		}
		instance := NewDemuxer(handlers, nil)

		recorder := httptest.NewRecorder()
		instance.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/route1", nil))
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "response", recorder.Body.String())
	})
	t.Run("a defined route with default should call the handler", func(t *testing.T) {
		t.Parallel()

		handler1 := &testscommon.HttpHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				_, _ = writer.Write([]byte("response"))
				writer.WriteHeader(http.StatusOK)
			},
		}
		handler2 := &testscommon.HttpHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				assert.Fail(t, "should have not called this handler")
			},
		}

		handlers := map[string]http.Handler{
			"/route1": handler1,
			"*":       handler2,
		}
		instance := NewDemuxer(handlers, nil)

		recorder := httptest.NewRecorder()
		instance.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/route1", nil))
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "response", recorder.Body.String())
	})
	t.Run("a not defined route with default should call the default handler", func(t *testing.T) {
		t.Parallel()

		handler1 := &testscommon.HttpHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				assert.Fail(t, "should have not called this handler")
			},
		}
		handler2 := &testscommon.HttpHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				_, _ = writer.Write([]byte("response"))
				writer.WriteHeader(http.StatusOK)
			},
		}

		handlers := map[string]http.Handler{
			"route1": handler1,
			"*":      handler2,
		}
		instance := NewDemuxer(handlers, nil)

		recorder := httptest.NewRecorder()
		instance.ServeHTTP(recorder, &http.Request{RequestURI: "route2"})
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "response", recorder.Body.String())
	})
	t.Run("should work with a root handler", func(t *testing.T) {
		t.Parallel()

		handler := &testscommon.HttpHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				assert.Fail(t, "should have not called this handler")
			},
		}
		rootHandler := &testscommon.HttpHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				_, _ = writer.Write([]byte("response"))
				writer.WriteHeader(http.StatusOK)
			},
		}

		handlers := map[string]http.Handler{
			"*": handler,
		}
		instance := NewDemuxer(handlers, rootHandler)

		recorder := httptest.NewRecorder()
		instance.ServeHTTP(recorder, &http.Request{RequestURI: "/"})
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "response", recorder.Body.String())

		recorder = httptest.NewRecorder()
		instance.ServeHTTP(recorder, &http.Request{RequestURI: "/index.html"})
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "response", recorder.Body.String())
	})
}
