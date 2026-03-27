package wayback

import (
	"context"
	"errors"
	"fmt"

	"github.com/agmmnn/gau2/pkg/httpclient"
	"github.com/agmmnn/gau2/pkg/providers"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

const (
	Name = "wayback"
)

// verify interface compliance
var _ providers.Provider = (*Client)(nil)

// Client is the structure that holds the WaybackFilters and the Client's configuration
type Client struct {
	filters providers.Filters
	config  *providers.Config
}

func New(config *providers.Config, filters providers.Filters) *Client {
	return &Client{filters, config}
}

func (c *Client) Name() string {
	return Name
}

// waybackResult holds the response from the wayback API
type waybackResult [][]string

// Fetch fetches all urls for a given domain and sends them to a channel.
// It returns an error should one occur.
func (c *Client) Fetch(ctx context.Context, domain string, results chan string) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		logrus.WithFields(logrus.Fields{"provider": Name}).Infof("fetching %s", domain)
		apiURL := c.formatURL(domain)
		resp, err := httpclient.MakeRequest(c.config.Client, apiURL, c.config.MaxRetries, c.config.Timeout)
		if err != nil {
			if errors.Is(err, httpclient.ErrBadRequest) {
				return nil
			}
			return fmt.Errorf("failed to fetch wayback results: %s", err)
		}

		urls, err := parseResponse(resp)
		if err != nil {
			return fmt.Errorf("failed to decode wayback results: %s", err)
		}

		for _, archivedURL := range urls {
			results <- archivedURL
		}
	}

	return nil
}

// formatUrl returns a formatted URL for the Wayback API
func (c *Client) formatURL(domain string) string {
	if c.config.IncludeSubdomains {
		domain = "*." + domain
	}
	filterParams := c.filters.GetParameters(true)
	return fmt.Sprintf(
		"http://web.archive.org/cdx/search/cdx?url=%s/*&output=json&collapse=urlkey&fl=timestamp,original",
		domain,
	) + filterParams
}

func parseResponse(resp []byte) ([]string, error) {
	var result waybackResult
	if err := jsoniter.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	start := 0
	if len(result) > 0 && len(result[0]) > 0 && (result[0][0] == "original" || result[0][0] == "timestamp" || result[0][0] == "urlkey") {
		start = 1
	}

	capHint := len(result) - start
	if capHint < 0 {
		capHint = 0
	}

	urls := make([]string, 0, capHint)
	for _, entry := range result[start:] {
		if len(entry) == 0 {
			continue
		}

		archivedURL := entry[len(entry)-1]
		if archivedURL == "" {
			continue
		}

		urls = append(urls, archivedURL)
	}

	return urls, nil
}
