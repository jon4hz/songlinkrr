package plex

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/jon4hz/songlinkrr/version"
)

type Client struct {
	http *http.Client

	server string
	token  string
}

func New(server, token string, timeout int, tlsVerify bool) *Client {
	return &Client{
		http: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: !tlsVerify},
			},
		},
		server: server,
		token:  token,
	}
}

func (c *Client) GetSessions(ctx context.Context) (*Sessions, error) {
	u, err := url.Parse(c.server + "/status/sessions")
	if err != nil {
		return nil, err
	}
	v := u.Query()
	v.Set("X-Plex-Token", c.token)
	v.Set("X-Plex-Platform", runtime.GOOS)
	v.Set("X-Plex-Platform-Version", "0.0.0")
	v.Set("X-Plex-Client-Identifier", "gmotd-v"+version.Version)
	v.Set("X-Plex-Product", "gmotd")
	v.Set("X-Plex-Version", version.Version)
	v.Set("X-Plex-Device", runtime.GOOS+" "+runtime.GOARCH)
	u.RawQuery = v.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var sessions Sessions
	if err := json.Unmarshal(bodyBytes, &sessions); err != nil {
		return nil, err
	}

	return &sessions, nil
}

type Sessions struct {
	MediaContainer MediaContainer `json:"MediaContainer"`
}

type MediaContainer struct {
	Size     int        `json:"size"`
	Metadata []Metadata `json:"Metadata"`
}

type Metadata struct {
	User             User   `json:"User"`
	Player           Player `json:"Player"`
	Type             string `json:"type"`
	Title            string `json:"title"`
	GrandparentTitle string `json:"grandparentTitle"` // this is the artist
	ParentTitle      string `json:"parentTitle"`      // this is the album
}

type User struct {
	Title string `json:"title"`
}

type Player struct {
	Product string `json:"product"`
	Title   string `json:"title"`
}
