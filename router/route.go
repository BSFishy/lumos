package router

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/BSFishy/lumos/util"
)

type HandlerFunc func(http.ResponseWriter, *http.Request) error

type statusCodeResponseWriter struct {
	inner      http.ResponseWriter
	statusCode int
}

func (s *statusCodeResponseWriter) Header() http.Header {
	return s.inner.Header()
}

func (s *statusCodeResponseWriter) Write(data []byte) (int, error) {
	return s.inner.Write(data)
}

func (s *statusCodeResponseWriter) WriteHeader(statusCode int) {
	s.statusCode = statusCode
	s.inner.WriteHeader(statusCode)
}

func (e HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := util.LogFromCtx(ctx)

	logger = logger.With(slog.Group("request", "method", r.Method, "url", r.URL.String()))
	r = r.WithContext(util.WithLogger(ctx, logger))

	writer := &statusCodeResponseWriter{
		inner:      w,
		statusCode: 200,
	}

	logger.Info("handling request")

	err := e(writer, r)
	if err != nil {
		logger.Error("failed to handle route", "err", err)
		writer.WriteHeader(http.StatusInternalServerError)
	}

	logger.Info("completed request", slog.Group("response", "statusCode", writer.statusCode))
}

func (r *Router) register(method, path string, handler http.Handler) {
	util.Assert(len(path) >= 1, "path must not be empty")
	util.Assert(strings.HasPrefix(path, "/"), "path must start with /")

	routePath := fmt.Sprintf("%s/%s", r.path, path[1:])
	if path == "/" {
		routePath = r.path
	}

	route := route{
		patternType: pattern,
		method:      &method,
		pattern:     routePath,
		handler:     handler,
	}

	*r.routes = append(*r.routes, route)
}

func (r *Router) Get(path string, handler HandlerFunc) {
	r.register(http.MethodGet, path, handler)
}

func (r *Router) Post(path string, handler HandlerFunc) {
	r.register(http.MethodPost, path, handler)
}

func (r *Router) Router(path string) *Router {
	util.Assert(len(path) >= 2, "path must not be empty")
	util.Assert(strings.HasPrefix(path, "/"), "path must start with /")

	return &Router{
		routes: r.routes,
		path:   fmt.Sprintf("%s/%s", r.path, path[1:]),
	}
}

func (r *Router) Route(path string, f func(*Router)) {
	f(r.Router(path))
}
