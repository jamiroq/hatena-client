package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"
)

const (
	defaultBaseURL = "http://b.hatena.ne.jp/hotentry/"
	userAgent      = "go-hotentry"
)

type Client struct {
	URL        *url.URL
	HTTPClient *http.Client
	UserAgent  string
	Logger     *log.Logger
}

type HatenaFeed struct {
	XMLName         xml.Name         `xml:"RDF"`
	Title           string           `xml:"channel>title"`
	Link            string           `xml:"channel>link"`
	Description     string           `xml:"channel>description"`
	HatenaBookmarks []HatenaBookmark `xml:"item"`
}

type HatenaBookmark struct {
	XMLName       xml.Name `xml:"item"`
	Title         string   `xml:"title"`
	Link          string   `xml:"link"`
	Description   string   `xml:"description"`
	Content       string   `xml:"content"`
	Date          string   `xml:"date"`
	Subject       string   `xml:"subject"`
	BookmarkCount int64    `xml:"bookmarkcount"`
}

func NewClient(baseURL string, logger *log.Logger) (*Client, error) {
	// varidate
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s", baseURL)
	}
	var discardLogger = log.New(ioutil.Discard, "", log.LstdFlags)
	if logger == nil {
		logger = discardLogger
	}

	return &Client{
		URL:        parsedURL,
		HTTPClient: http.DefaultClient,
		UserAgent:  userAgent,
		Logger:     logger,
	}, nil
}

func (c *Client) newRequest(ctx context.Context, method, spath string, body io.Reader) (*http.Request, error) {
	u := *c.URL
	u.Path = path.Join(c.URL.Path, spath)

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	req.Header.Set("User-Agent", c.UserAgent)
	return req, nil
}

func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)
	return decoder.Decode(out)
}

func (c *Client) GetHotentryIT(ctx context.Context) (*HatenaFeed, error) {
	req, err := c.newRequest(ctx, "GET", "it.rss", nil)
	if err != nil {
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	// XXX Check status code

	var f HatenaFeed
	if err := decodeBody(res, &f); err != nil {
		return nil, err
	}

	return &f, nil
}

func main() {
	c, err := NewClient(defaultBaseURL, nil)
	if err != nil {
		fmt.Println(err)
	}
	ctx := context.Background()
	f, _ := c.GetHotentryIT(ctx)
	for _, bookmark := range f.HatenaBookmarks {
		fmt.Println(bookmark.Title)
		fmt.Println()
	}
}
