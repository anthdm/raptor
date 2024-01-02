package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/anthdm/ffaas/pkg/api"
	"github.com/anthdm/ffaas/pkg/types"
	"github.com/google/uuid"
)

type Config struct {
	url string
}

func NewConfig() Config {
	return Config{
		url: "http://localhost:3000",
	}
}

func (c Config) WithURL(url string) Config {
	c.url = url
	return c
}

type Client struct {
	*http.Client

	config Config
}

func New(config Config) *Client {
	return &Client{
		Client: http.DefaultClient,
		config: config,
	}
}

func (c *Client) CreateEndpoint(params api.CreateEndpointParams) (*types.Endpoint, error) {
	b, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/%s", c.config.url, "endpoint")
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api responded with a non 200 status code: %d", resp.StatusCode)
	}
	var endpoint types.Endpoint
	if err := json.NewDecoder(resp.Body).Decode(&endpoint); err != nil {
		return nil, err
	}
	return &endpoint, nil
}

func (c *Client) CreateDeploy(endpointID uuid.UUID, blob io.Reader, params api.CreateDeployParams) (*types.Deploy, error) {
	url := fmt.Sprintf("%s/endpoint/%s/deploy", c.config.url, endpointID)
	req, err := http.NewRequest("POST", url, blob)
	req.Header.Add("Content-Type", "application/octet-stream")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api responded with a non 200 status code: %d", resp.StatusCode)
	}
	var deploy types.Deploy
	if err := json.NewDecoder(resp.Body).Decode(&deploy); err != nil {
		return nil, err
	}
	return &deploy, nil
}
