package virustotal

import (
	"context"
	"fmt"
	neturl "net/url"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/lc/gau/v2/pkg/httpclient"
	"github.com/lc/gau/v2/pkg/providers"
	"github.com/sirupsen/logrus"
)

const (
	Name = "virustotal"
)

var _ providers.Provider = (*Client)(nil)

type Client struct {
	config *providers.Config
}

func New(c *providers.Config) *Client {
	if c.VirusTotal.Host != "" {
		setBaseURL(c.VirusTotal.Host)
	}

	return &Client{config: c}
}

func (c *Client) Name() string {
	return Name
}

func (c *Client) Fetch(ctx context.Context, domain string, results chan string) error {
	if c.config.VirusTotal.APIKey == "" {
		return nil
	}

	select {
	case <-ctx.Done():
		return nil
	default:
		logrus.WithFields(logrus.Fields{"provider": Name}).Infof("fetching %s", domain)
		resp, err := httpclient.MakeRequest(c.config.Client, c.formatURL(domain), c.config.MaxRetries, c.config.Timeout)
		if err != nil {
			return fmt.Errorf("failed to fetch virustotal results: %s", err)
		}

		var result apiResponse
		if err := jsoniter.Unmarshal(resp, &result); err != nil {
			return fmt.Errorf("failed to decode virustotal results: %s", err)
		}

		if result.ResponseCode == 0 {
			return nil
		}

		for _, entry := range result.DetectedURLs {
			if entry.URL == "" || !matchesDomain(entry.URL, domain, c.config.IncludeSubdomains) {
				continue
			}
			results <- entry.URL
		}
	}

	return nil
}

func (c *Client) formatURL(domain string) string {
	query := neturl.Values{}
	query.Set("apikey", c.config.VirusTotal.APIKey)
	query.Set("domain", domain)
	return fmt.Sprintf("%svtapi/v2/domain/report?%s", _BaseURL, query.Encode())
}

func matchesDomain(rawURL string, domain string, includeSubdomains bool) bool {
	parsed, err := neturl.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	domain = strings.ToLower(domain)
	if host == domain {
		return true
	}

	return includeSubdomains && strings.HasSuffix(host, "."+domain)
}

func setBaseURL(baseURL string) {
	_BaseURL = baseURL
}
