package providers

import (
	"strings"
	"testing"
)

func TestGetParametersForWayback(t *testing.T) {
	filters := Filters{
		From:               "202401",
		To:                 "202402",
		MatchStatusCodes:   []string{"200"},
		MatchMimeTypes:     []string{"text/html"},
		FilterStatusCodes:  []string{"404"},
		FilterMimeTypes:    []string{"image/png"},
		CommonCrawlFilters: []string{"=url:https://example.com/admin"},
	}

	params := filters.GetParameters(true)

	expected := []string{
		"from=202401",
		"to=202402",
		"filter=mimetype%3Atext%2Fhtml",
		"filter=statuscode%3A200",
		"filter=%21statuscode%3A404",
		"filter=%21mimetype%3Aimage%2Fpng",
	}

	for _, fragment := range expected {
		if !strings.Contains(params, fragment) {
			t.Fatalf("expected %q in %q", fragment, params)
		}
	}

	if strings.Contains(params, "url%3Ahttps%3A%2F%2Fexample.com%2Fadmin") {
		t.Fatalf("expected commoncrawl-only filter to be excluded from wayback params: %q", params)
	}
}

func TestGetParametersForCommonCrawl(t *testing.T) {
	filters := Filters{
		MatchStatusCodes:  []string{"200"},
		MatchMimeTypes:    []string{"text/html"},
		FilterStatusCodes: []string{"404"},
		FilterMimeTypes:   []string{"image/png"},
		CommonCrawlFilters: []string{
			"=url:https://example.com/admin",
			"~url:.*\\.php$",
		},
	}

	params := filters.GetParameters(false)

	expected := []string{
		"filter=status%3A200",
		"filter=mime%3Atext%2Fhtml",
		"filter=%21%3Dstatus%3A404",
		"filter=%21%3Dmime%3Aimage%2Fpng",
		"filter=%3Durl%3Ahttps%3A%2F%2Fexample.com%2Fadmin",
		"filter=~url%3A.%2A%5C.php%24",
	}

	for _, fragment := range expected {
		if !strings.Contains(params, fragment) {
			t.Fatalf("expected %q in %q", fragment, params)
		}
	}
}
