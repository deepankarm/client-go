package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jina-ai/client-go/jina"
)

type WebSocketClient struct {
	Host string
	conn *websocket.Conn
	ctx  context.Context
}

func NewWebSocketClient(host string) (*WebSocketClient, error) {
	var u *url.URL
	if !strings.HasPrefix(host, "ws") {
		host = "ws://" + host
	}
	u, err := url.Parse(host)
	if err != nil {
		u = &url.URL{Scheme: "ws", Host: host, Path: "/"}
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return &WebSocketClient{}, err
	}
	client := &WebSocketClient{
		Host: host,
		conn: conn,
		ctx:  context.Background(),
	}
	return client, nil
}

func (client WebSocketClient) POST(requests <-chan *jina.DataRequestProto, onDone, onError, onAlways CallbackType) error {
	var wg sync.WaitGroup

	handleRequest := func(request *jina.DataRequestProto) {
		reqJSON, err := json.Marshal(request)
		if err != nil {
			if onError != nil {
				onError(request)
			}
			if onAlways != nil {
				onAlways(request)
			}
		}

		err = client.conn.WriteMessage(websocket.TextMessage, reqJSON)
		if err != nil {
			if onError != nil {
				onError(request)
			}
			if onAlways != nil {
				onAlways(request)
			}
		}
	}
	go func() {
		for {
			_, data, err := client.conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				break
			}
			var res jina.DataRequestProto
			if err := json.Unmarshal(data, &res); err != nil {
				// Unsure how to handle OnError here
				fmt.Println(err)
			} else if onDone != nil {
				onDone(&res)
			}
			if onAlways != nil {
				onAlways(&res)
			}
			wg.Done()
		}
	}()

	for {
		req, ok := <-requests
		if !ok {
			break
		}
		handleRequest(req)
		wg.Add(1)
	}
	wg.Wait()
	return nil
}

type WebSocketHealthCheckClient struct {
	Host string
	ctx  context.Context
}

func NewWebSocketHealthCheckClient(host string) (*WebSocketHealthCheckClient, error) {
	if strings.HasPrefix(host, "ws") {
		host = strings.Replace(host, "ws", "http", 1)
	}
	if !strings.HasPrefix(host, "http") {
		host = "http://" + host
	}
	return &WebSocketHealthCheckClient{
		Host: host,
		ctx:  context.Background(),
	}, nil
}

func (c WebSocketHealthCheckClient) HealthCheck() (bool, error) {
	httpResp, err := http.Get(c.Host)
	if err != nil {
		return false, err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode == http.StatusOK {
		return true, nil
	}
	return false, fmt.Errorf("got non 200 status code %d", httpResp.StatusCode)
}

type WebSocketInfoClient struct {
	Host string
	ctx  context.Context
}

func NewWebSocketInfoClient(host string) (WebSocketInfoClient, error) {
	if strings.HasPrefix(host, "ws") {
		host = strings.Replace(host, "ws", "http", 1)
	}
	if !strings.HasPrefix(host, "http") {
		host = "http://" + host
	}
	return WebSocketInfoClient{
		Host: host,
		ctx:  context.Background(),
	}, nil
}

func (c WebSocketInfoClient) InfoJSON() ([]byte, error) {
	httpResp, err := http.Get(c.Host + "/status")
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got non 200 status code %d", httpResp.StatusCode)
	}
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, body, "", "  "); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c WebSocketInfoClient) Info() (*jina.JinaInfoProto, error) {
	body, err := c.InfoJSON()
	if err != nil {
		return nil, err
	}

	var res jina.JinaInfoProto
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
