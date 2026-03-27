package virustotal

var _BaseURL = "https://www.virustotal.com/"

type apiResponse struct {
	ResponseCode int `json:"response_code"`
	DetectedURLs []struct {
		URL string `json:"url"`
	} `json:"detected_urls"`
}
