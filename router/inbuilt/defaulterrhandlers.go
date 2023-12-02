package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt/types"
)

/*
This file is responsible for default responses for errors. Currently, this
functionality is not needed because response builder already has WithError()
method that automatically determines error and returns a corresponding code
and body (as a text of the error, just like http wants). The only exception
is 405 Method Not Allowed, because http requires Allow header, that is why
we have to customize a behaviour of response with such a code
*/

func newErrorHandlers() *types.ErrHandlers {
	errHandlers := types.NewErrHandlers()
	errHandlers.Set(status.MethodNotAllowed, defaultMethodNotAllowedHandler)

	return errHandlers
}

func defaultMethodNotAllowedHandler(request *http.Request) *http.Response {
	response := http.Error(request, status.ErrMethodNotAllowed)

	if allow, ok := request.Ctx.Value("allow").(string); ok {
		return response.Header("Allow", allow)
	}

	return response
}
