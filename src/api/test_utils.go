package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"marlinraker-go/src/api/executors"
	"marlinraker-go/src/files"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func testHttp[Result any](t *testing.T, method string, endpoint string, params executors.Params,
	f func(*testing.T, *httptest.ResponseRecorder, *Result, *Error)) {
	t.Run(method+endpoint, func(t *testing.T) {

		urlQuery, err := url.Parse(endpoint)
		if err != nil {
			t.Fatal(err)
		}

		values := urlQuery.Query()
		for param, value := range params {
			values.Set(param, fmt.Sprintf("%v", value))
		}
		urlQuery.RawQuery = values.Encode()

		request, _ := http.NewRequest(method, urlQuery.String(), nil)
		recorder := httptest.NewRecorder()
		handleHttp(recorder, request)

		var errorResponse ErrorResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatal(err)
			return
		}
		if errorResponse.Error.Code > 0 {
			f(t, recorder, nil, &errorResponse.Error)
			return
		}

		var response struct {
			Result Result `json:"result"`
		}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
			return
		}
		f(t, recorder, &response.Result, nil)
	})
}

func testSocket[Result any](t *testing.T, method string, params executors.Params, f func(*testing.T, *Result, *Error)) {
	t.Run(method, func(t *testing.T) {

		server := httptest.NewServer(http.HandlerFunc(handleSocket))
		defer server.Close()
		socketUrl := "ws" + server.URL[4:]

		socket, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer func(socket *websocket.Conn) {
			_ = socket.Close()
		}(socket)

		request := RpcRequest{
			Rpc: Rpc{
				JsonRpc: "2.0",
				Id:      0,
			},
			Method: method,
			Params: params,
		}

		{
			data, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}

			err = socket.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				t.Fatal(err)
			}
		}

		{
			_, data, err := socket.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}

			var errorResponse RpcErrorResponse
			err = json.Unmarshal(data, &errorResponse)
			if err != nil {
				t.Fatal(err)
			}
			if errorResponse.Error.Code > 0 {
				f(t, nil, &errorResponse.Error)
				return
			}

			var response struct {
				Result Result `json:"result"`
			}
			err = json.Unmarshal(data, &response)
			if err != nil {
				t.Fatal(err)
			}
			f(t, &response.Result, nil)
		}
	})
}

func testAll[Result any](t *testing.T, rpcMethod string, httpMethod string, httpUrl string, params executors.Params,
	f func(*testing.T, *httptest.ResponseRecorder, *Result, *Error)) {

	testSocket(t, rpcMethod, params, func(t *testing.T, result *Result, error *Error) {
		f(t, nil, result, error)
	})
	testHttp(t, httpMethod, httpUrl, params, f)
}

func makeConnection(t *testing.T) (*websocket.Conn, int) {

	server := httptest.NewServer(http.HandlerFunc(handleSocket))
	defer server.Close()
	socketUrl := "ws" + server.URL[4:]

	socket, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = socket.WriteJSON(RpcRequest{
		Rpc:    Rpc{JsonRpc: "2.0", Id: 42},
		Method: "server.connection.identify",
		Params: executors.Params{"client_name": "", "version": "", "type": "", "url": ""},
	})
	if err != nil {
		t.Fatal(err)
	}

	var response RpcResultResponse
	err = socket.ReadJSON(&response)
	if err != nil {
		t.Fatal(err)
	}

	connectionId := response.Result.(map[string]any)["connection_id"].(float64)
	return socket, int(connectionId)
}

func testFileUpload[Result any](t *testing.T, url string, fields map[string]string, fileName string, filePath string,
	f func(*testing.T, *httptest.ResponseRecorder, *Result, *Error)) {
	t.Run("POST"+url, func(t *testing.T) {

		pipeReader, pipeWriter := io.Pipe()
		writer := multipart.NewWriter(pipeWriter)

		go func() {
			defer func() {
				if err := writer.Close(); err != nil {
					t.Fatal(err)
				}
			}()

			for name, value := range fields {
				if err := writer.WriteField(name, value); err != nil {
					t.Fatal(err)
				}
			}

			source, err := files.Fs.OpenFile(filePath, os.O_RDONLY, 0755)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := source.Close(); err != nil {
					t.Fatal(err)
				}
			}()

			uploadWriter, err := writer.CreateFormFile("file", fileName)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := io.Copy(uploadWriter, source); err != nil {
				t.Fatal(err)
			}
		}()

		request := httptest.NewRequest("POST", url, pipeReader)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		recorder := httptest.NewRecorder()
		handleHttp(recorder, request)

		var errorResponse ErrorResponse
		if err := json.Unmarshal(recorder.Body.Bytes(), &errorResponse); err != nil {
			t.Fatal(err)
			return
		}
		if errorResponse.Error.Code > 0 {
			f(t, recorder, nil, &errorResponse.Error)
			return
		}

		var result Result
		if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
			return
		}
		f(t, recorder, &result, nil)
	})
}
