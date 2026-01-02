package process

import (
	"net/http"
	"strings"
)

const defaultHandler = "*"
const httpDelimiter = "/"

type demuxer struct {
	handlers    map[string]http.Handler
	rootHandler http.Handler
}

// NewDemuxer can create a demuxer able to call different http handlers based on the request.RequestURI.
// the default handler should be registered with the * key
func NewDemuxer(handlers map[string]http.Handler, rootHandler http.Handler) *demuxer {
	instance := &demuxer{
		handlers:    make(map[string]http.Handler),
		rootHandler: rootHandler,
	}

	if len(handlers) > 0 {
		for route, handler := range handlers {
			instance.handlers[route] = handler
		}
	}

	return instance
}

// ServeHTTP will try to serve the http request based on the registered handlers
func (d *demuxer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	urlPath := ""
	if request.URL != nil {
		urlPath = request.URL.Path
	}
	handler := d.handlers[urlPath]
	if handler != nil {
		handler.ServeHTTP(writer, request)
		return
	}

	if strings.Count(request.RequestURI, httpDelimiter) == 1 {
		if d.rootHandler != nil {
			d.rootHandler.ServeHTTP(writer, request)
			return
		}
	}

	handler = d.handlers[defaultHandler]
	if handler != nil {
		handler.ServeHTTP(writer, request)
		return
	}

	http.NotFound(writer, request)
}
