package indigo

import (
	"context"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/decoder"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	httpparser "github.com/indigo-web/indigo/internal/parser"
	"github.com/indigo-web/indigo/internal/parser/http1"
	"github.com/indigo-web/indigo/internal/render"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/buffer"
	"github.com/indigo-web/utils/pool"
	"net"
)

func newClient(tcpSettings settings.TCP, conn net.Conn) tcp.Client {
	readBuff := make([]byte, tcpSettings.ReadBufferSize)

	return tcp.NewClient(conn, tcpSettings.ReadTimeout, readBuff)
}

func newKeyValueArenas(s settings.Headers) (*buffer.Buffer[byte], *buffer.Buffer[byte]) {
	keyArena := buffer.NewBuffer[byte](
		s.MaxKeyLength*s.Number.Default,
		s.MaxKeyLength*s.Number.Maximal,
	)
	valArena := buffer.NewBuffer[byte](
		s.ValueSpace.Default,
		s.ValueSpace.Maximal,
	)

	return keyArena, valArena
}

func newBodyReader(client tcp.Client, body settings.Body, decoders map[string]decoder.Constructor) http.BodyReader {
	manager := decoder.NewManager(body.DecodedBufferSize)

	for token, constructor := range decoders {
		manager.Add(token, constructor)
	}

	return http1.NewBodyReader(client, http1.NewChunkedBodyParser(body), manager)
}

func newRequest(
	ctx context.Context, s settings.Settings, conn net.Conn, r http.BodyReader,
) *http.Request {
	q := query.NewQuery(headers.NewHeaders())
	hdrs := headers.NewPreallocHeaders(s.Headers.Number.Default)
	response := http.NewBuilder()
	params := make(http.Params)
	body := http.NewBody(r)

	return http.NewRequest(ctx, hdrs, q, response, conn, body, params, s.URL.Params.DisableMapClear)
}

func newRenderer(httpSettings settings.HTTP, a *Application) render.Engine {
	respBuff := make([]byte, 0, httpSettings.ResponseBuffSize)

	return render.NewEngine(respBuff, nil, a.defaultHeaders)
}

func newHTTPParser(s settings.Settings, req *http.Request) httpparser.HTTPRequestsParser {
	keyArena, valArena := newKeyValueArenas(s.Headers)
	objPool := pool.NewObjectPool[[]string](s.Headers.MaxValuesObjectPoolSize)

	startLineArena := buffer.NewBuffer[byte](
		s.URL.BufferSize.Default,
		s.URL.BufferSize.Maximal,
	)

	return http1.NewHTTPRequestsParser(
		req, *keyArena, *valArena, *startLineArena, *objPool, s.Headers,
	)
}
