package dnsever

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const BaseUrl = "https://api.dnsever.com"

type Client struct {
	HttpClient *http.Client

	ClientId     string
	ClientSecret string
}

func NewClient(clientId string, clientSecret string) *Client {
	return &Client{
		HttpClient:   http.DefaultClient,
		ClientId:     clientId,
		ClientSecret: clientSecret,
	}
}

func (c *Client) applyAuth(req *http.Request) {
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.ClientId+":"+c.ClientSecret)))
	req.Header.Set("User-Agent", "DNSEverXml-Client/1.1.2")
}

func (c *Client) Request(ctx context.Context, url string, data *url.Values) (*DNSEverXml, error) {
	requestBody := []byte(data.Encode())

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		BaseUrl+url,
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, err
	}
	c.applyAuth(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(requestBody)))
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	responseBody, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	responseXml, err := UnmarshalDNSEver(responseBody)
	if err != nil {
		return nil, err
	}

	return responseXml, nil
}

func (c *Client) GetRecord(ctx context.Context, Zone string, Type string) (*DNSEverXml, error) {
	requestBody := &url.Values{}
	requestBody.Set("zone", Zone)
	requestBody.Set("type", Type)
	return c.Request(ctx, "/record/getrecord.php", requestBody)
}

func (c *Client) AddRecord(ctx context.Context, Name string, Type string, Value string, Rank string, Memo string) (*DNSEverXml, error) {
	requestBody := &url.Values{}
	requestBody.Set("name", Name)
	requestBody.Set("value", Value)
	requestBody.Set("type", Type)
	if Rank != "" {
		requestBody.Set("rank", Rank)
	}
	if Memo != "" {
		requestBody.Set("memo", Memo)
	}
	return c.Request(ctx, "/record/add.php", requestBody)
}

func (c *Client) UpdateRecord(ctx context.Context, Id string, Type string, Value string, Rank string, Memo string) (*DNSEverXml, error) {
	requestBody := &url.Values{}
	requestBody.Set("id", Id)
	requestBody.Set("type", Type)
	requestBody.Set("value", Value)
	if Rank != "" {
		requestBody.Set("rank", Rank)
	}
	if Memo != "" {
		requestBody.Set("memo", Memo)
	}
	return c.Request(ctx, "/record/update.php", requestBody)
}

func (c *Client) DeleteRecord(ctx context.Context, Id string, Type string) (*DNSEverXml, error) {
	requestBody := &url.Values{}
	requestBody.Set("id", Id)
	requestBody.Set("type", Type)
	return c.Request(ctx, "/record/delete.php", requestBody)
}
