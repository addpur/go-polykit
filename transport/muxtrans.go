package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/addpur/go-polykit"
)

type HTTPDecodeRequestFunc func(context.Context, *http.Request) (interface{}, error)

type HTTPEncodeResponseFunc func(context.Context, http.ResponseWriter, interface{}) error

type HTTPEncodeErrorFunc func(context.Context, error, http.ResponseWriter)

type HTTPRequestFunc func(ctx context.Context, r *http.Request) context.Context

type HTTPResponseFunc func(ctx context.Context, w http.ResponseWriter) context.Context

type HTTPServerOption func(*HTTPServer)

type HTTPServer struct {
	e      polykit.Endpoint
	dec    HTTPDecodeRequestFunc
	enc    HTTPEncodeResponseFunc
	errEnc HTTPEncodeErrorFunc
	before []HTTPRequestFunc
	after  []HTTPResponseFunc
}

func NewHTTPServer(
	e polykit.Endpoint,
	dec HTTPDecodeRequestFunc,
	enc HTTPEncodeResponseFunc,
	options ...HTTPServerOption,
) *HTTPServer {
	s := &HTTPServer{
		e:      e,
		dec:    dec,
		enc:    enc,
		errEnc: DefaultHTTPErrorEncoder,
	}
	for _, o := range options {
		o(s)
	}
	return s
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	for _, f := range s.before {
		ctx = f(ctx, r)
	}

	request, err := s.dec(ctx, r)
	if err != nil {
		s.errEnc(ctx, err, w)
		return
	}

	response, err := s.e(ctx, request)
	if err != nil {
		s.errEnc(ctx, err, w)
		return
	}

	for _, f := range s.after {
		ctx = f(ctx, w)
	}

	if err := s.enc(ctx, w, response); err != nil {
		s.errEnc(ctx, err, w)
		return
	}
}

func HTTPServerBefore(before ...HTTPRequestFunc) HTTPServerOption {
	return func(s *HTTPServer) {
		s.before = append(s.before, before...)
	}
}

func HTTPServerAfter(after ...HTTPResponseFunc) HTTPServerOption {
	return func(s *HTTPServer) {
		s.after = append(s.after, after...)
	}
}

func HTTPServerErrorEncoder(errEnc HTTPEncodeErrorFunc) HTTPServerOption {
	return func(s *HTTPServer) {
		s.errEnc = errEnc
	}
}

func DefaultHTTPErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(polykit.StandardResponse{
		ResponseCode: "500",
		Message:      err.Error(),
	})
}
