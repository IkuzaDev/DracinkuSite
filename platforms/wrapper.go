package platforms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"dramawave/models"
)

type WrapperClient struct {
	Platform   string
	Language   string
	BaseURL    string
	HttpClient *http.Client
}

type WrapperResponse struct {
	Success  bool            `json:"success"`
	Platform string          `json:"platform"`
	Language string          `json:"language"`
	Data     json.RawMessage `json:"data"`
	Error    *WrapperError   `json:"error,omitempty"`
}

type WrapperError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func NewWrapperClient(platform, language string) *WrapperClient {
	if language == "" {
		language = "id"
	}
	return &WrapperClient{
		Platform: platform,
		Language: language,
		BaseURL:  "https://api.dracinku.site",
		HttpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *WrapperClient) SetLanguage(language string) {
	// Map generic language code to platform-specific code
	// Default to original if not found
	code := language

	switch c.Platform {
	case "dramabox", "shortmax", "melolo", "dramadash":
		// id, en, th, vi, zh, kr, jp
		// Remap ko -> kr, ja -> jp, zhtw/zhcn -> zh
		switch language {
		case "ko":
			code = "kr"
		case "ja":
			code = "jp"
		case "zhtw", "zhcn":
			code = "zh"
		}
	case "gooshort":
		// id -> in
		switch language {
		case "id":
			code = "in"
		case "ko":
			code = "kr"
		case "ja":
			code = "jp"
		case "zhtw", "zhcn":
			code = "zh"
		}
	case "netshort":
		// id_ID, en_US, th_TH, vi_VN, zh_CN, kr_KR, jp_JP
		switch language {
		case "id":
			code = "id_ID"
		case "en":
			code = "en_US"
		case "th":
			code = "th_TH"
		case "vi":
			code = "vi_VN"
		case "zh", "zhcn":
			code = "zh_CN"
		case "zhtw":
			code = "zh_TW" // or zh_CN default? User map says zh_CN for zh.
		case "ko":
			code = "kr_KR"
		case "ja":
			code = "jp_JP"
		}
	case "flickreels", "freereels":
		// id-ID, en-US, th-TH, vi-VN, zh-CN, kr-KR, jp-JP
		// Flickreels specific: kr=ko-KR, jp=ja-JP?
		// User map says:
		// freereels: kr-KR, jp-JP
		// flickreels: kr=ko-KR, jp=ja-JP
		if c.Platform == "flickreels" {
			switch language {
			case "id":
				code = "id-ID"
			case "en":
				code = "en-US"
			case "th":
				code = "th-TH"
			case "ko":
				code = "ko-KR"
			case "ja":
				code = "ja-JP"
			}
		} else { // freereels
			switch language {
			case "id":
				code = "id-ID"
			case "en":
				code = "en-US"
			case "th":
				code = "th-TH"
			case "vi":
				code = "vi-VN"
			case "zh", "zhcn":
				code = "zh-CN"
			case "ko":
				code = "kr-KR"
			case "ja":
				code = "jp-JP"
			}
		}
	}

	c.Language = code
}

var GlobalLicenseKey = "dk-demo-key-123456789"

func SetGlobalLicenseKey(key string) {
	GlobalLicenseKey = key
}

func (c *WrapperClient) Request(endpoint, method string, query map[string]string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/%s%s", c.BaseURL, c.Platform, endpoint)

	reqBody := []byte{}
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("x-api-key", GlobalLicenseKey)
	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("lang", c.Language)
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	fmt.Printf("[%sWrapper] Requesting %s %s\n", c.Platform, method, req.URL.String())

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Log response for debugging
	fmt.Printf("[%sWrapper] Response Status: %d, Body: %s\n", c.Platform, resp.StatusCode, string(respData))

	// Check HTTP status code
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respData))
	}

	var wrapperResp WrapperResponse
	if err := json.Unmarshal(respData, &wrapperResp); err != nil {
		return nil, fmt.Errorf("unmarshal wrapper response: %w. Body: %s", err, string(respData))
	}

	if !wrapperResp.Success {
		if wrapperResp.Error != nil {
			// Check for license errors
			if wrapperResp.Error.Code == "INVALID_LICENSE_KEY" ||
				wrapperResp.Error.Code == "LICENSE_EXPIRED" ||
				wrapperResp.Error.Code == "INVALID_API_KEY" ||
				wrapperResp.Error.Code == "API_KEY_EXPIRED" {
				return nil, fmt.Errorf("LICENSE_ERROR: %s", wrapperResp.Error.Message)
			}
			return nil, fmt.Errorf("api error: %s (%s)", wrapperResp.Error.Message, wrapperResp.Error.Code)
		}
		return nil, fmt.Errorf("unknown api error")
	}

	return wrapperResp.Data, nil
}

func (c *WrapperClient) GetTabList() ([]models.TabItem, error) {
	data, err := c.Request("/tablist", "GET", nil, nil)
	if err != nil {
		return nil, err
	}

	var tabs []models.TabItem
	if err := json.Unmarshal(data, &tabs); err != nil {
		return nil, err
	}
	return tabs, nil
}

func (c *WrapperClient) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}) (*models.TabDataResponse, error) {
	payload := map[string]interface{}{
		"type":     tabType,
		"key":      tabKey,
		"position": positionIndex,
	}

	data, err := c.Request("/tabdata", "POST", nil, payload)
	if err != nil {
		return nil, err
	}

	var resp models.TabDataResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *WrapperClient) GetTabFeed(pageInfo models.PageInfo) (*models.TabFeedResponse, error) {
	data, err := c.Request("/tabfeed", "POST", nil, pageInfo) // pageInfo matches structure expected by API
	if err != nil {
		return nil, err
	}

	// API returns same structure as TabData often, but let's see.
	// API.md says "Get next page of tab content". Response example not explicitly shown but implied to be similar.
	// Actually typical paged response.
	// Let's assume it parses into TabFeedResponse which is essentially List + PageInfo
	var resp models.TabFeedResponse
	// Sometimes it might return full Book object wrapper. Check API.md
	// API.md doesn't explicitly show TabFeed response. Assuming similar to TabData but maybe different root.
	// Actually, usually TabFeed returns just the list and new page info.
	// Struct models.TabFeedResponse has IsMore, Book []DramaItem, PageInfo.
	// Let's try to unmarshal into that.
	if err := json.Unmarshal(data, &resp); err != nil {

		// Fallback: maybe it returns { isMore, book: { list: [...] }, page_info } like TabData?
		var tabDataResp models.TabDataResponse
		if err2 := json.Unmarshal(data, &tabDataResp); err2 == nil {
			return &models.TabFeedResponse{
				IsMore:   tabDataResp.IsMore,
				Book:     tabDataResp.Book.List,
				PageInfo: tabDataResp.PageInfo,
			}, nil
		}
		return nil, err
	}
	return &resp, nil
}

func (c *WrapperClient) GetSeriesDetail(id string, quality int) (*models.SeriesDetailResponse, error) {
	query := map[string]string{
		"quality": fmt.Sprintf("%d", quality),
	}
	data, err := c.Request(fmt.Sprintf("/series/%s", id), "GET", query, nil)
	if err != nil {
		return nil, err
	}

	var resp models.SeriesDetailResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *WrapperClient) SearchSeries(keyword string, pageNo, pageSize int) (*models.SearchResponse, error) {
	payload := map[string]interface{}{
		"keyword":  keyword,
		"pageno":   pageNo,
		"pagesize": pageSize,
	}

	data, err := c.Request("/search", "POST", nil, payload)
	if err != nil {
		return nil, err
	}

	var resp models.SearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
