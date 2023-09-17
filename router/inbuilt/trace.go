package inbuilt

import (
	"bytes"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/internal/httpchars"
)

/*
This file is responsible for rendering http requests. Prime use case is rendering
http requests back as a response to a trace request
*/

func traceResponse(respond *http.Builder, messageBody []byte) http.Response {
	return respond.
		WithHeader("Content-Type", "message/http").
		WithBodyByte(messageBody)
}

func renderHTTPRequest(request *http.Request, buff []byte) []byte {
	buff = append(buff, method.ToString(request.Method)...)
	buff = append(buff, httpchars.SP...)
	buff = requestURI(request, buff)
	buff = append(buff, httpchars.SP...)
	buff = append(buff, bytes.TrimSpace(proto.ToBytes(request.Proto))...)
	buff = append(buff, httpchars.CRLF...)
	buff = requestHeaders(request, buff)
	buff = append(buff, "Content-Length: 0\r\n\r\n"...)

	return buff
}

func requestURI(request *http.Request, buff []byte) []byte {
	buff = append(buff, request.Path...)

	if query := request.Query.Raw(); len(query) > 0 {
		buff = append(buff, '?')
		buff = append(buff, query...)
	}

	return buff
}

func requestHeaders(request *http.Request, buff []byte) []byte {
	unwrapped := request.Headers.Unwrap()

	for i := 0; i < len(unwrapped); i += 2 {
		buff = append(append(buff, unwrapped[i]...), httpchars.COLONSP...)
		buff = append(append(buff, unwrapped[i+1]...), httpchars.CRLF...)
	}

	return buff
}
