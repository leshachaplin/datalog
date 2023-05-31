package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	url  string
	http HTTPClient
}

func NewClient(url string, httpClient HTTPClient) *Client {
	return &Client{
		url:  url,
		http: httpClient,
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.url+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Request-ID", uuid.NewString())

	return req, nil
}

type eventReq struct {
	ClientTime string `json:"client_time"`
	DeviceID   string `json:"device_id"`
	DeviceOS   string `json:"device_os"`
	Session    string `json:"session"`
	Event      string `json:"event"`
	ParamStr   string `json:"param_str"`
	Sequence   int    `json:"sequence"`
	ParamInt   int    `json:"param_int"`
}

func (c *Client) SendEvents(ctx context.Context, params []eventReq) error {
	path := "/v1/event"

	b := make([]byte, 0)
	buf := bytes.NewBuffer(b)

	for _, event := range params {
		body, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		buf.Write(body)
		buf.WriteString("\n")
	}

	req, err := c.newRequest(ctx, http.MethodPost, path, buf)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	return nil
}
