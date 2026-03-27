package wayback

import "testing"

func TestParseResponseSkipsHeaderAndUsesOriginalURLColumn(t *testing.T) {
	resp := []byte(`[["timestamp","original"],["20240101120000","https://example.com/"],["20240101120001","https://example.com/login"]]`)

	urls, err := parseResponse(resp)
	if err != nil {
		t.Fatalf("parseResponse returned error: %v", err)
	}

	if len(urls) != 2 {
		t.Fatalf("expected 2 urls, got %d", len(urls))
	}

	if urls[0] != "https://example.com/" || urls[1] != "https://example.com/login" {
		t.Fatalf("unexpected urls: %#v", urls)
	}
}

func TestParseResponseWithoutHeaderKeepsFirstResult(t *testing.T) {
	resp := []byte(`[["20240101120000","https://example.com/"]]`)

	urls, err := parseResponse(resp)
	if err != nil {
		t.Fatalf("parseResponse returned error: %v", err)
	}

	if len(urls) != 1 || urls[0] != "https://example.com/" {
		t.Fatalf("unexpected urls: %#v", urls)
	}
}
