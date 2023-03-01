package router

import (
	"github.com/indigo-web/indigo/v2/http"
)

// Router is a general interface for any router compatible with indigo
// OnRequest called every time headers are parsed and ready to be processed
// OnError called once, and if it called, it means that connection will be
// closed anyway. So you can process the error, send some response,
// and when you are ready, just notify core that he can safely close
// the connection (even if it's already closed from client side).
type Router interface {
	OnRequest(request *http.Request) http.Response
	OnError(request *http.Request, err error) http.Response
}

// OnStarter is an interface that provides OnStart() method that will be called
// just once, when server is initializing.
type OnStarter interface {
	OnStart()
}
