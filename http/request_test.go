package http

import (
	"github.com/indigo-web/indigo/http/decode"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

type dummyBodyReader struct {
	client tcp.Client
}

func newDummyReader(client tcp.Client) dummyBodyReader {
	return dummyBodyReader{
		client: client,
	}
}

func (d dummyBodyReader) Init(*Request) {}

func (d dummyBodyReader) Read() ([]byte, error) {
	return d.ReadNoDecoding()
}

func (d dummyBodyReader) ReadNoDecoding() ([]byte, error) {
	return d.client.Read()
}

func TestRequest_Reader(t *testing.T) {
	t.Run("Ordinary", func(t *testing.T) {
		first, second := []byte("Hello"), []byte("World!")
		client := dummy.NewCircularClient(first, second)
		reader := newBodyIOReader()
		reader.reader = newDummyReader(client)
		buff := make([]byte, 1024)

		n, err := reader.Read(buff)
		require.Equal(t, string(first), string(buff[:len(first)]))
		require.NoError(t, err)

		n, err = reader.Read(buff)
		require.Equal(t, string(second), string(buff[:len(second)]))
		require.NoError(t, err)

		require.NoError(t, client.Close())
		n, err = reader.Read(buff)
		require.Equal(t, 0, n)
		require.EqualError(t, err, io.EOF.Error())
	})

	t.Run("Partially", func(t *testing.T) {
		first, second := []byte("Hello"), []byte("World!")
		client := dummy.NewCircularClient(first, second)
		reader := newBodyIOReader()
		reader.reader = newDummyReader(client)
		buff := make([]byte, 3)

		n, err := reader.Read(buff)
		require.Equal(t, string(first[:n]), string(buff[:n]))
		require.NoError(t, err)

		n, err = reader.Read(buff)
		require.Equal(t, string(first[len(buff):len(buff)+n]), string(buff[:n]))
		require.NoError(t, err)

		n, err = reader.Read(buff)
		require.Equal(t, string(second[:n]), string(buff[:n]))
		require.NoError(t, err)

		n, err = reader.Read(buff)
		require.Equal(t, string(second[len(buff):len(buff)+n]), string(buff[:n]))
		require.NoError(t, err)

		require.NoError(t, client.Close())

		n, err = reader.Read(buff)
		require.Equal(t, 0, n)
		require.EqualError(t, err, io.EOF.Error())
	})

	t.Run("JSON", func(t *testing.T) {
		type Model struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		body := []byte(`{"name": "John", "email": "john@smith.com"}`)
		client := dummy.NewCircularClient(body)
		client.OneTime()
		reader := newDummyReader(client)

		request := NewRequest(
			headers.NewHeaders(nil), query.Query{}, NewResponse(), dummy.NewNopConn(),
			NewBody(reader, decode.NewDecoder()), nil, true)

		model := new(Model)
		require.NoError(t, request.JSON(model))
		require.Equal(t, "John", model.Name)
		require.Equal(t, "john@smith.com", model.Email)
	})
}
