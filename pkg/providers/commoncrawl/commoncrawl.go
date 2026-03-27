package commoncrawl

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/lc/gau/v2/pkg/httpclient"
	"github.com/lc/gau/v2/pkg/providers"
	"github.com/sirupsen/logrus"
)

const (
	Name = "commoncrawl"
)

// verify interface compliance
var _ providers.Provider = (*Client)(nil)

// Client is the structure that holds the Filters and the Client's configuration
type Client struct {
	filters providers.Filters
	config  *providers.Config

	indexes []apiIndex
}

func New(c *providers.Config, filters providers.Filters) (*Client, error) {
	// Fetch the list of available CommonCrawl Api URLs.
	resp, err := httpclient.MakeRequest(c.Client, "https://index.commoncrawl.org/collinfo.json", c.MaxRetries, c.Timeout)
	if err != nil {
		return nil, err
	}

	var r apiResult
	if err = jsoniter.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	if len(r) == 0 {
		return nil, errors.New("failed to grab latest commoncrawl index")
	}

	return &Client{config: c, filters: filters, indexes: r}, nil
}

func (c *Client) Name() string {
	return Name
}

// Fetch fetches all urls for a given domain and sends them to a channel.
// It returns an error should one occur.
func (c *Client) Fetch(ctx context.Context, domain string, results chan string) error {
	seen := mapset.NewThreadUnsafeSet[string]()
	var lastErr error
	var foundResults bool

	for _, index := range c.indexes {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		p, err := c.getPagination(index.API, domain)
		if err != nil {
			lastErr = err
			logrus.WithFields(logrus.Fields{"provider": Name, "index": index.ID}).Warnf("pagination failed for %s: %v", domain, err)
			continue
		}

		if p.Pages == 0 {
			continue
		}

		for page := uint(0); page < p.Pages; page++ {
			select {
			case <-ctx.Done():
				return nil
			default:
				logrus.WithFields(logrus.Fields{"provider": Name, "index": index.ID, "page": page}).Infof("fetching %s", domain)
				apiURL := c.formatURL(index.API, domain, page)
				resp, err := httpclient.MakeRequest(c.config.Client, apiURL, c.config.MaxRetries, c.config.Timeout)
				if err != nil {
					lastErr = err
					logrus.WithFields(logrus.Fields{"provider": Name, "index": index.ID, "page": page}).Warnf("request failed for %s: %v", domain, err)
					continue
				}

				sc := bufio.NewScanner(bytes.NewReader(resp))
				sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

				for sc.Scan() {
					var res apiResponse
					if err := jsoniter.Unmarshal(sc.Bytes(), &res); err != nil {
						lastErr = fmt.Errorf("failed to decode commoncrawl result from %s page %d: %w", index.ID, page, err)
						logrus.WithFields(logrus.Fields{"provider": Name, "index": index.ID, "page": page}).Warn(lastErr)
						break
					}
					if res.Error != "" {
						lastErr = fmt.Errorf("received an error from commoncrawl index %s: %s", index.ID, res.Error)
						logrus.WithFields(logrus.Fields{"provider": Name, "index": index.ID, "page": page}).Warn(lastErr)
						break
					}

					if res.URL == "" || seen.Contains(res.URL) {
						continue
					}

					seen.Add(res.URL)
					foundResults = true
					results <- res.URL
				}

				if err := sc.Err(); err != nil {
					lastErr = fmt.Errorf("failed to scan commoncrawl response from %s page %d: %w", index.ID, page, err)
					logrus.WithFields(logrus.Fields{"provider": Name, "index": index.ID, "page": page}).Warn(lastErr)
				}
			}
		}
	}

	if !foundResults && lastErr == nil {
		logrus.WithFields(logrus.Fields{"provider": Name}).Infof("no results for %s", domain)
	}

	return nil
}

func (c *Client) formatURL(apiURL string, domain string, page uint) string {
	if c.config.IncludeSubdomains {
		domain = "*." + domain
	}

	filterParams := c.filters.GetParameters(false)

	return fmt.Sprintf("%s?url=%s/*&output=json&fl=url&page=%d", apiURL, domain, page) + filterParams
}

// Fetch the number of pages.
func (c *Client) getPagination(apiURL string, domain string) (r paginationResult, err error) {
	url := fmt.Sprintf("%s&showNumPages=true", c.formatURL(apiURL, domain, 0))
	var resp []byte

	resp, err = httpclient.MakeRequest(c.config.Client, url, c.config.MaxRetries, c.config.Timeout)
	if err != nil {
		return
	}

	err = jsoniter.Unmarshal(resp, &r)
	return
}
