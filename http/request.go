package http

import (
	"context"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/http/status"
	json "github.com/json-iterator/go"
	"net"
)

type Params = map[string]string

// Request represents HTTP request
type Request struct {
	Method method.Method
	Path   string
	// Query is a key-value part of the Path
	Query query.Query
	// Params are dynamic segments, in case dynamic routing is enabled
	Params        Params
	Proto         proto.Proto
	Headers       *headers.Headers
	Encoding      headers.Encoding
	ContentLength int
	ContentType   string
	// Upgrade is the protocol token, which is set by default to proto.Unknown. In
	// case it is anything else, then Upgrade header was received
	Upgrade proto.Proto
	// Remote represents remote net.Addr
	Remote net.Addr
	// Ctx is a request context. It may be filled with arbitrary data across middlewares
	// and handler by itself.
	Ctx            context.Context
	defaultCtx     context.Context
	body           *Body
	conn           net.Conn
	wasHijacked    bool
	clearParamsMap bool
	builder        *Builder
}

// NewRequest returns a new instance of request object and body gateway
// Must not be used externally, this function is for internal purposes only
// HTTP/1.1 as a protocol by default is set because if first request from user
// is invalid, we need to render a response using request method, but appears
// that default method is a null-value (proto.Unknown)
func NewRequest(
	ctx context.Context, hdrs *headers.Headers, query query.Query, builder *Builder,
	conn net.Conn, body *Body, paramsMap Params, disableParamsMapClearing bool,
) *Request {
	request := &Request{
		Query:          query,
		Params:         paramsMap,
		Proto:          proto.HTTP11,
		Headers:        hdrs,
		Remote:         conn.RemoteAddr(),
		Ctx:            ctx,
		defaultCtx:     ctx,
		body:           body,
		conn:           conn,
		clearParamsMap: !disableParamsMapClearing,
		builder:        builder,
	}

	return request
}

// JSON takes a model and returns an error if occurred. Model must be a pointer to a structure.
// If Content-Type header is given, but is not "application/json", then status.ErrUnsupportedMediaType
// will be returned. If JSON is malformed, or it doesn't match the model, then custom jsoniter error
// will be returned
func (r *Request) JSON(model any) error {
	if len(r.ContentType) > 0 && r.ContentType != "application/json" {
		return status.ErrUnsupportedMediaType
	}

	data, err := r.Body().Full()
	if err != nil {
		return err
	}

	iterator := json.ConfigDefault.BorrowIterator(data)
	iterator.ReadVal(model)
	err = iterator.Error
	json.ConfigDefault.ReturnIterator(iterator)

	return err
}

// Body returns an entity representing a request's body
func (r *Request) Body() *Body {
	return r.body
}

// Respond returns Response builder, associated with the request
func (r *Request) Respond() *Builder {
	return r.builder
}

// Hijack the connection. Request body will be implicitly read (so if you need it you
// should read it before) all the body left. After handler exits, the connection will
// be closed, so the connection can be hijacked only once
func (r *Request) Hijack() (net.Conn, error) {
	if err := r.body.Reset(); err != nil {
		return nil, err
	}

	r.wasHijacked = true

	return r.conn, nil
}

// WasHijacked returns true or false, depending on whether was a connection hijacked
func (r *Request) WasHijacked() bool {
	return r.wasHijacked
}

// Clear resets request headers and reads body into nowhere until completed.
// It is implemented to clear the request object between requests
func (r *Request) Clear() (err error) {
	r.Query.Set(nil)
	r.Ctx = r.defaultCtx
	r.builder = r.builder.Clear()

	if err = r.body.Reset(); err != nil {
		return err
	}

	r.ContentLength = 0
	r.Encoding = headers.Encoding{}
	r.ContentType = ""
	r.Upgrade = proto.Unknown

	if r.clearParamsMap && len(r.Params) > 0 {
		for k := range r.Params {
			delete(r.Params, k)
		}
	}

	return nil
}

// TODO: implement FormData parsing

// Respond returns a response object of request
func Respond(request *Request) Response {
	return request.builder
}

func Error(request *Request, err error) Response {
	return request.Respond().WithError(err)
}
