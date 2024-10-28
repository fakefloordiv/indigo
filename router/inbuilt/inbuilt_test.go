package inbuilt

import (
	"errors"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/transport/dummy"
	"testing"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo/http/status"

	"github.com/indigo-web/indigo/http/method"
	"github.com/stretchr/testify/require"
)

func TestRoute(t *testing.T) {
	raw := New()
	raw.Route(method.GET, "/", http.Respond)
	raw.Route(method.POST, "/", http.Respond)
	raw.Route(method.POST, "/hello", http.Respond)
	r := raw.Initialize()

	t.Run("GET /", func(t *testing.T) {
		request := getRequest(method.GET, "/")
		resp := r.OnRequest(request)
		require.Equal(t, status.OK, resp.Reveal().Code)
	})

	t.Run("POST /", func(t *testing.T) {
		request := getRequest(method.POST, "/")
		resp := r.OnRequest(request)
		require.Equal(t, status.OK, resp.Reveal().Code)
	})

	t.Run("POST /hello", func(t *testing.T) {
		request := getRequest(method.POST, "/hello")
		resp := r.OnRequest(request)
		require.Equal(t, status.OK, resp.Reveal().Code)
	})

	t.Run("HEAD /", func(t *testing.T) {
		request := getRequest(method.HEAD, "/")
		resp := r.OnRequest(request)
		// we have not registered any HEAD-method handler yet, so GET method
		// is expected to be called (but without body)
		require.Equal(t, status.OK, resp.Reveal().Code)
		require.Empty(t, string(resp.Reveal().Body))
	})
}

func testMethodShorthand(
	t *testing.T, router *Router,
	route func(string, Handler, ...Middleware) *Router,
	method method.Method,
) {
	route("/", http.Respond)
	require.Contains(t, router.registrar.routes, "/")
	require.NotNil(t, router.registrar.routes["/"][method])
}

func TestMethodShorthands(t *testing.T) {
	r := New()

	t.Run("GET", func(t *testing.T) {
		testMethodShorthand(t, r, r.Get, method.GET)
	})
	t.Run("HEAD", func(t *testing.T) {
		testMethodShorthand(t, r, r.Head, method.HEAD)
	})
	t.Run("POST", func(t *testing.T) {
		testMethodShorthand(t, r, r.Post, method.POST)
	})
	t.Run("PUT", func(t *testing.T) {
		testMethodShorthand(t, r, r.Put, method.PUT)
	})
	t.Run("DELETE", func(t *testing.T) {
		testMethodShorthand(t, r, r.Delete, method.DELETE)
	})
	t.Run("CONNECT", func(t *testing.T) {
		testMethodShorthand(t, r, r.Connect, method.CONNECT)
	})
	t.Run("OPTIONS", func(t *testing.T) {
		testMethodShorthand(t, r, r.Options, method.OPTIONS)
	})
	t.Run("TRACE", func(t *testing.T) {
		testMethodShorthand(t, r, r.Trace, method.TRACE)
	})
	t.Run("PATCH", func(t *testing.T) {
		testMethodShorthand(t, r, r.Patch, method.PATCH)
	})
}

func TestGroups(t *testing.T) {
	raw := New().
		Get("/", http.Respond)

	api := raw.Group("/api")

	api.Group("/v1").
		Get("/hello", http.Respond)

	api.Group("/v2").
		Get("/world", http.Respond)

	r := raw.Initialize()

	require.Equal(t, status.OK, r.OnRequest(getRequest(method.GET, "/")).Reveal().Code)
	require.Equal(t, status.OK, r.OnRequest(getRequest(method.GET, "/api/v1/hello")).Reveal().Code)
	require.Equal(t, status.OK, r.OnRequest(getRequest(method.GET, "/api/v2/world")).Reveal().Code)
}

func TestResource(t *testing.T) {
	raw := New()
	raw.Resource("/").
		Get(http.Respond).
		Post(http.Respond)

	api := raw.Group("/api")
	api.Resource("/stat").
		Get(http.Respond).
		Post(http.Respond)

	r := raw.Initialize()

	t.Run("Root", func(t *testing.T) {
		require.Equal(t, status.OK, r.OnRequest(getRequest(method.GET, "/")).Reveal().Code)
		require.Equal(t, status.OK, r.OnRequest(getRequest(method.POST, "/")).Reveal().Code)
	})

	t.Run("Group", func(t *testing.T) {
		require.Equal(t, status.OK, r.OnRequest(getRequest(method.GET, "/api/stat")).Reveal().Code)
		require.Equal(t, status.OK, r.OnRequest(getRequest(method.POST, "/api/stat")).Reveal().Code)
	})
}

func TestResource_Methods(t *testing.T) {
	raw := New()
	raw.Resource("/").
		Get(http.Respond).
		Head(http.Respond).
		Post(http.Respond).
		Put(http.Respond).
		Delete(http.Respond).
		Connect(http.Respond).
		Options(http.Respond).
		Trace(http.Respond).
		Patch(http.Respond)

	r := raw.Initialize()

	for _, m := range method.List {
		require.Equal(t, status.OK, r.OnRequest(getRequest(m, "/")).Reveal().Code)
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	r := New().
		Get("/", http.Respond).
		Initialize()

	request := getRequest(method.POST, "/")
	response := r.OnRequest(request)
	require.Equal(t, status.MethodNotAllowed, response.Reveal().Code)
}

func TestRouter_RouteError(t *testing.T) {
	r := New().
		RouteError(func(req *http.Request) *http.Response {
			return req.Respond().
				Code(status.Teapot).
				String(req.Env.Error.Error())
		}, status.BadRequest).
		Initialize()

	t.Run("status.ErrBadRequest", func(t *testing.T) {
		request := getRequest(method.GET, "/")
		resp := r.OnError(request, status.ErrBadRequest)
		require.Equal(t, status.Teapot, resp.Reveal().Code)
		require.Equal(t, status.ErrBadRequest.Error(), string(resp.Reveal().Body))
	})

	t.Run("status.ErrURIDecoding (also bad request)", func(t *testing.T) {
		request := getRequest(method.GET, "/")
		resp := r.OnError(request, status.ErrURLDecoding)
		require.Equal(t, status.Teapot, resp.Reveal().Code)
		require.Equal(t, status.ErrURLDecoding.Error(), string(resp.Reveal().Body))
	})

	t.Run("unregistered http error", func(t *testing.T) {
		request := getRequest(method.GET, "/")
		resp := r.OnError(request, status.ErrNotImplemented)
		require.Equal(t, status.NotImplemented, resp.Reveal().Code)
	})

	t.Run("unregistered ordinary error", func(t *testing.T) {
		request := getRequest(method.GET, "/")
		resp := r.OnError(request, errors.New("any error text here"))
		require.Equal(t, status.InternalServerError, resp.Reveal().Code)
		require.Empty(t, string(resp.Reveal().Body))
	})

	t.Run("universal handler", func(t *testing.T) {
		const fromUniversal = "from universal handler with love"

		r := New().
			RouteError(func(req *http.Request) *http.Response {
				return req.Respond().
					Code(status.Teapot).
					String(fromUniversal)
			}, AllErrors).
			Initialize()

		request := getRequest(method.GET, "/")
		resp := r.OnError(request, status.ErrNotImplemented)
		require.Equal(t, int(status.Teapot), int(resp.Reveal().Code))
		require.Equal(t, fromUniversal, string(resp.Reveal().Body))
	})
}

func TestAliases(t *testing.T) {
	testRootAlias := func(t *testing.T, r router.Router) {
		request := getRequest(method.GET, "/")
		response := r.OnRequest(request)
		require.Equal(t, status.OK, response.Reveal().Code)
		require.Equal(t, "magic word", string(response.Reveal().Body))
	}

	t.Run("single alias", func(t *testing.T) {
		r := New().
			Get("/hello", func(req *http.Request) *http.Response {
				require.Equal(t, "/hello", req.Path)
				require.Equal(t, "/", req.Env.AliasFrom)

				return req.Respond().String("magic word")
			}).
			Alias("/", "/hello")

		testRootAlias(t, r.Initialize())
	})

	t.Run("override normal handler", func(t *testing.T) {
		r := New().
			Get("/", func(req *http.Request) *http.Response {
				require.Fail(t, "the handler must be overridden and never be called")
				return http.Respond(req)
			}).
			Get("/hello", func(req *http.Request) *http.Response {
				require.Equal(t, "/hello", req.Path)
				require.Equal(t, "/", req.Env.AliasFrom)

				return req.Respond().String("magic word")
			}).
			Alias("/", "/hello")

		testRootAlias(t, r.Initialize())
	})

	testOrdinaryRequest := func(t *testing.T, r router.Router) {
		request := getRequest(method.GET, "/hello")
		response := r.OnRequest(request)
		require.Equal(t, status.OK, response.Reveal().Code)
		require.Equal(t, "ordinary word", string(response.Reveal().Body))
	}

	t.Run("multiple calls", func(t *testing.T) {
		var i int

		r := New().
			Get("/", func(req *http.Request) *http.Response {
				require.Fail(t, "the handler must be overridden and never be called")
				return http.Respond(req)
			}).
			Get("/hello", func(req *http.Request) *http.Response {
				defer func() {
					i++
				}()

				if i%2 == 0 {
					require.Equalf(t, "/hello", req.Path, "on iteration: %d", i)
					require.Equalf(t, "/", req.Env.AliasFrom, "on iteration: %d", i)

					return req.Respond().String("magic word")
				} else {
					require.Equalf(t, "/hello", req.Path, "on iteration: %d", i)
					require.Emptyf(t, req.Env.AliasFrom, "on iteration: %d", i)

					return req.Respond().String("ordinary word")
				}
			}).
			Alias("/", "/hello")

		testRootAlias(t, r.Initialize())
		testOrdinaryRequest(t, r.Initialize())
		testRootAlias(t, r.Initialize())
		testOrdinaryRequest(t, r.Initialize())
	})

	t.Run("aliases on groups", func(t *testing.T) {
		r := New().
			Get("/heaven", http.Respond)

		r.Group("/hello").
			Alias("/world", "/heaven")

		request := getRequest(method.GET, "/hello/world")
		response := r.Initialize().OnRequest(request)
		require.Equal(t, int(status.OK), int(response.Reveal().Code))
	})
}

func TestCatchers(t *testing.T) {
	t.Run("multiple overriding endpoints", func(t *testing.T) {
		r := New().
			Get("/", http.Respond).
			Get("/hello", http.Respond).
			Catch("/", func(req *http.Request) *http.Response {
				return req.Respond().String("magic")
			}).
			Catch("/hello", func(req *http.Request) *http.Response {
				return req.Respond().String("double magic")
			}).
			Initialize()

		resp := r.OnRequest(getRequest(method.GET, "/"))
		require.Equal(t, status.OK, resp.Reveal().Code)
		require.Empty(t, resp.Reveal().Body)
		resp = r.OnRequest(getRequest(method.GET, "/hello"))
		require.Equal(t, status.OK, resp.Reveal().Code)
		require.Empty(t, resp.Reveal().Body)
		resp = r.OnRequest(getRequest(method.GET, "/anything"))
		require.Equal(t, status.OK, resp.Reveal().Code)
		require.Equal(t, "magic", string(resp.Reveal().Body))
		resp = r.OnRequest(getRequest(method.GET, "/helloworld"))
		require.Equal(t, status.OK, resp.Reveal().Code)
		require.Equal(t, "double magic", string(resp.Reveal().Body))
	})
}

func TestMutators(t *testing.T) {
	var timesCalled int

	r := New().
		Get("/", http.Respond).
		Mutator(func(request *http.Request) {
			timesCalled++
		}).
		Initialize()

	request := construct.Request(config.Default(), dummy.NewNopClient(), nil)
	request.Method = method.GET
	request.Path = "/"
	resp := r.OnRequest(request)
	require.Equal(t, status.OK, resp.Reveal().Code)

	request.Method = method.POST
	request.Path = "/"
	resp = r.OnRequest(request)
	require.Equal(t, status.MethodNotAllowed, resp.Reveal().Code)

	request.Method = method.GET
	request.Path = "/foo"
	resp = r.OnRequest(request)
	require.Equal(t, status.NotFound, resp.Reveal().Code)

	require.Equal(t, 3, timesCalled)
}
