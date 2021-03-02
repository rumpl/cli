package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const server = "http://127.0.0.1:1224"

// Client is the client for the image info server
type Client interface {
	// GetImageInfo ...
	GetImageInfo(ctx context.Context, ID string) (*ImageInfoResponse, error)
}

type client struct {
	baseURL string
}

// Opt is for configuring the Client
type Opt func(c *client) error

// WithBaseURL changes the default url of the server
func WithBaseURL(url string) Opt {
	return func(c *client) error {
		c.baseURL = url
		return nil
	}
}

// New creates a new Client
func New(opts ...Opt) (Client, error) {
	c := &client{
		baseURL: server,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *client) GetImageInfo(ctx context.Context, ID string) (*ImageInfoResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/images/newest/%s", c.baseURL, ID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode > 300 {
		return nil, errors.New("Not found")
	}

	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var imageInfoResponses ImageInfoResponse
	err = json.Unmarshal(buff, &imageInfoResponses)
	if err != nil {
		return nil, err
	}

	return &imageInfoResponses, nil
}
