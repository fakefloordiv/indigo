package inbuilt

import (
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/router/inbuilt/types"
)

/*
This file is responsible for registering both ordinary and error handlers
*/

// Route is a base method for registering handlers
func (r *Router) Route(
	method method.Method, path string, handlerFunc types.Handler,
	middlewares ...types.Middleware,
) *Router {
	err := r.registrar.Add(r.prefix+path, method, combine(handlerFunc, middlewares))
	if err != nil {
		panic(err)
	}

	return r
}

// RouteError adds an error handler. You can handle next errors:
// - status.ErrBadRequest
// - status.ErrNotFound
// - status.ErrMethodNotAllowed
// - status.ErrTooLarge
// - status.ErrCloseConnection
// - status.ErrURITooLong
// - status.ErrHeaderFieldsTooLarge
// - status.ErrTooManyHeaders
// - status.ErrUnsupportedProtocol
// - status.ErrUnsupportedEncoding
// - status.ErrMethodNotImplemented
// - status.ErrConnectionTimeout
//
// You can set your own handler and override default response
// WARNING: calling this method from groups will affect ALL routers, including root
func (r *Router) RouteError(err error, handler types.Handler) {
	r.errHandlers[err] = handler
}
