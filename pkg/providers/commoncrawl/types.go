package commoncrawl

type apiResponse struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

type paginationResult struct {
	Blocks   uint `json:"blocks"`
	PageSize uint `json:"pageSize"`
	Pages    uint `json:"pages"`
}

type apiIndex struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	API  string `json:"cdx-api"`
}

type apiResult []apiIndex
