package platforms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"dramawave/models"
	"dramawave/utils"
)

type NetShort struct {
	language   string
	tokenData  *netShortToken
	util       *utils.NetShortUtil
	httpClient *http.Client
	baseURL    string
}

type netShortToken struct {
	Token      string
	DeviceCode string
	Timestamp  int64
}

func NewNetShort(language string) *NetShort {
	if language == "" {
		language = "id_ID"
	}
	// Netshort requires specific format (e.g. id_ID), ensure we map it if it's generic
	switch language {
	case "id":
		language = "id_ID"
	case "en":
		language = "en_US"
	case "th":
		language = "th_TH"
	case "vi":
		language = "vi_VN"
	case "zh":
		language = "zh_CN"
	case "ko":
		language = "kr_KR"
	case "ja":
		language = "jp_JP"
	}

	return &NetShort{
		language:   language,
		util:       utils.NewNetShortUtil(),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://appsecapi.netshort.com/prod-app-api",
	}
}

func (n *NetShort) SetLanguage(language string) {
	switch language {
	case "id":
		language = "id_ID"
	case "en":
		language = "en_US"
	case "th":
		language = "th_TH"
	case "vi":
		language = "vi_VN"
	case "zh":
		language = "zh_CN"
	case "ko":
		language = "kr_KR"
	case "ja":
		language = "jp_JP"
	}
	n.language = language
}

func (n *NetShort) generateRandomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		rand.Intn(223)+1,
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(254)+1)
}

func (n *NetShort) buildHeaders(deviceCode, encryptKey, token string) map[string]string {
	randomIP := n.generateRandomIP()
	headers := map[string]string{
		"Content-Type":     "application/json",
		"version":          "2.1.1",
		"content-language": n.language,
		"platform":         "android",
		"User-Agent":       "ozilla/5.0 (Linux; Android 9; SM-X810N Build/PQ3A.190605.11261621; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/124.0.6367.82 Mobile Safari/537.36",
		"user-agent":       "ozilla/5.0 (Linux; Android 9; SM-X810N Build/PQ3A.190605.11261621; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/124.0.6367.82 Mobile Safari/537.36",
		"Authorization":    "Bearer " + token,
		"X-Forwarded-For":  randomIP,
		"X-Real-IP":        randomIP,
		"canary":           "v2",
		"os":               "1",
		"device-Code":      deviceCode,
		"encrypt-key":      encryptKey,
		"push_switch":      "true",
		"start_type":       "cold",
		"timestamp":        fmt.Sprintf("%d", time.Now().UnixMilli()),
		"network":          "other,cold,true",
	}

	return headers
}

func (n *NetShort) Login() (*netShortToken, error) {
	fmt.Printf("[NetShort] Login called\n")
	if n.tokenData != nil && time.Now().Unix()-n.tokenData.Timestamp < 86400 {
		return n.tokenData, nil
	}

	deviceCode := n.util.GenerateDeviceCode()
	deviceBase64 := n.util.ToBase64(deviceCode)
	encryptKey, _ := n.util.RSAEncrypt(deviceBase64)

	payload := map[string]interface{}{
		"os":         "Android",
		"osVer":      "13",
		"appVer":     "1.9.2",
		"identity":   1,
		"model":      fmt.Sprintf("G%s", n.generateRandomIP()),
		"deviceCode": deviceCode,
		"source":     "visitor",
	}
	jsonPayload, _ := json.Marshal(payload)
	encryptedBody, _ := n.util.AESEncrypt(string(jsonPayload), deviceCode)

	req, _ := http.NewRequest("POST", n.baseURL+"/auth/login", bytes.NewBufferString(encryptedBody))
	for k, v := range n.buildHeaders(deviceCode, encryptKey, "") {
		req.Header.Set(k, v)
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	encKeyHeader := resp.Header.Get("encrypt-key")
	if encKeyHeader == "" {
		// Log response body if login fails unexpectedly
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("[NetShort] Login failed, no encrypt-key. Response: %s\n", string(body))
		return nil, fmt.Errorf("login failed: no encrypt-key header")
	}

	decryptedKeyBase64, err := n.util.RSADecrypt(encKeyHeader)
	if err != nil {
		return nil, err
	}
	responseKey := n.util.FromBase64(decryptedKeyBase64)

	respBody, _ := io.ReadAll(resp.Body)
	decryptedDataStr, err := n.util.AESDecrypt(string(respBody), responseKey)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(decryptedDataStr), &result)

	data, _ := result["data"].(map[string]interface{})
	if data == nil {
		return nil, fmt.Errorf("login failed: no data")
	}
	token, _ := data["token"].(string)

	n.tokenData = &netShortToken{
		Token:      token,
		DeviceCode: deviceCode,
		Timestamp:  time.Now().Unix(),
	}

	return n.tokenData, nil
}

func (n *NetShort) request(endpoint string, payload interface{}) (map[string]interface{}, error) {
	fmt.Printf("[NetShort] Requesting %s\n", endpoint)
	token, err := n.Login()
	if err != nil {
		return nil, err
	}

	deviceBase64 := n.util.ToBase64(token.DeviceCode)
	encryptKey, _ := n.util.RSAEncrypt(deviceBase64)

	jsonPayload, _ := json.Marshal(payload)
	encryptedBody, _ := n.util.AESEncrypt(string(jsonPayload), token.DeviceCode)

	req, _ := http.NewRequest("POST", n.baseURL+endpoint, bytes.NewBufferString(encryptedBody))
	for k, v := range n.buildHeaders(token.DeviceCode, encryptKey, token.Token) {
		req.Header.Set(k, v)
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	encKeyHeader := resp.Header.Get("encrypt-key")
	if encKeyHeader == "" {
		return nil, fmt.Errorf("request failed: no encrypt-key header")
	}

	decryptedKeyBase64, _ := n.util.RSADecrypt(encKeyHeader)
	responseKey := n.util.FromBase64(decryptedKeyBase64)

	respBody, _ := io.ReadAll(resp.Body)
	decryptedDataStr, err := n.util.AESDecrypt(string(respBody), responseKey)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(decryptedDataStr), &result)

	return result, nil
}

func (n *NetShort) transformItem(item map[string]interface{}) models.DramaItem {
	id := ""
	if v, ok := item["shortPlayId"]; ok {
		id = fmt.Sprintf("%v", v)
	} else if v, ok := item["id"]; ok {
		id = fmt.Sprintf("%v", v)
	}

	introduction := getSt(item, "shotIntroduce")
	if introduction == "" {
		introduction = getSt(item, "introduction")
	}

	name := getSt(item, "shortPlayName")
	if name == "" {
		name = getSt(item, "shortPlayNames")
	}

	chapterCount := 0
	if v, ok := item["totalEpisode"].(float64); ok {
		chapterCount = int(v)
	} else if v, ok := item["episode_count"].(float64); ok {
		chapterCount = int(v)
	}

	playCount := item["heatScoreShow"]
	if playCount == nil {
		playCount = item["formatHeatScore"]
	}

	var tags []string
	if v, ok := item["labelArray"].([]interface{}); ok {
		for _, t := range v {
			if s, ok := t.(string); ok {
				tags = append(tags, s)
			}
		}
	}

	return models.DramaItem{
		ID:           id,
		Name:         name,
		Cover:        getSt(item, "shortPlayCover"),
		Introduction: introduction,
		ChapterCount: chapterCount,
		PlayCount:    playCount,
		Tags:         tags,
	}
}

func (n *NetShort) GetTabList(language string) ([]models.TabItem, error) {
	fmt.Printf("[NetShort] GetTabList called with language: %s\n", language)
	if language != "" {
		n.SetLanguage(language)
	}

	// JS calls getClassesConstant first.
	// Endpoint: /video/classes/constant
	result, err := n.request("/video/classes/constant", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	var tabs []models.TabItem

	// Iterate through categories
	for categoryName, items := range data {
		itemList, ok := items.([]interface{})
		if !ok {
			continue
		}

		for _, item := range itemList {
			it, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			// Handle cascadeTag differently
			if categoryName == "cascadeTag" {
				// Process nested children structure
				children, ok := it["children"].([]interface{})
				if !ok {
					continue
				}

				for _, child := range children {
					ite, ok := child.(map[string]interface{})
					if !ok {
						continue
					}

					grandChildren, ok := ite["children"].([]interface{})
					if !ok {
						continue
					}

					for _, grandChild := range grandChildren {
						gc, ok := grandChild.(map[string]interface{})
						if !ok {
							continue
						}

						// Build tab item for cascade tag
						var positionIndex interface{}

						// Position index priority: childrenLabelIds > [labelId, ite.labelId]
						if childrenLabelIds := gc["childrenLabelIds"]; childrenLabelIds != nil {
							positionIndex = childrenLabelIds
						} else {
							// Create array [it.labelId, ite.labelId]
							gcLabelId := gc["labelId"]
							iteLabelId := ite["labelId"]

							if gcLabelId != nil && iteLabelId != nil {
								positionIndex = []interface{}{gcLabelId, iteLabelId}
							} else if gcLabelId != nil {
								positionIndex = []interface{}{gcLabelId}
							} else if iteLabelId != nil {
								positionIndex = []interface{}{iteLabelId}
							} else {
								positionIndex = ""
							}
						}

						tab := models.TabItem{
							Type:          "tag", // PERUBAHAN: hardcoded ke "tag"
							Name:          getSt(gc, "labelName"),
							TabKey:        ite["labelLevel"], // PERUBAHAN: labelId -> labelLevel
							PositionIndex: positionIndex,
						}

						tabs = append(tabs, tab)
					}
				}
			} else {
				// Handle regular categories (tag, region, audio, etc)
				var name string
				var tabKey interface{}
				var positionIndex interface{}

				if categoryName == "tag" {
					name = getSt(it, "labelName")
					tabKey = it["labelLanguageId"]
					if it["labelId"] != nil {
						tabKey = it["labelId"]
					} else {
						tabKey = -1
					}
					if extList := it["extList"]; extList != nil {
						positionIndex = extList
					} else if newLabelIdList := it["newLabelIdList"]; newLabelIdList != nil {
						positionIndex = newLabelIdList
					} else {
						positionIndex = ""
					}
				} else {
					// For other categories
					name = getSt(it, "value")
					if name == "" {
						name = getSt(it, "name")
					}
					if name == "" {
						name = getSt(it, "labelName")
					}

					if key := it["key"]; key != nil {
						tabKey = key
					} else if sort := it["sort"]; sort != nil {
						tabKey = sort
					}

					if extList := it["extList"]; extList != nil {
						positionIndex = extList
					} else if newLabelIdList := it["newLabelIdList"]; newLabelIdList != nil {
						positionIndex = newLabelIdList
					} else {
						positionIndex = ""
					}
				}

				tabs = append(tabs, models.TabItem{
					Type:          categoryName,
					Name:          name,
					TabKey:        tabKey,
					PositionIndex: positionIndex,
				})
			}
		}
	}

	return tabs, nil
}

func (n *NetShort) GetSeriesDetail(videoID string, language string) (*models.SeriesDetailResponse, error) {
	fmt.Printf("[NetShort] GetSeriesDetail called for videoID: %s\n", videoID)
	if language != "" {
		n.SetLanguage(language)
	}

	// JS: /base/checkstand/waitUnlock then /video/shortPlay/base/detail_info
	n.request("/base/checkstand/waitUnlock", map[string]interface{}{})

	result, err := n.request("/video/shortPlay/base/detail_info", map[string]interface{}{
		"codec":            "h264",
		"shortPlayId":      videoID,
		"isRequestReserve": true,
		"playClarity":      "720p",
	})
	if err != nil {
		return nil, err
	}

	data, _ := result["data"].(map[string]interface{})
	if data == nil {
		return nil, fmt.Errorf("no data")
	}

	book := n.transformItem(data)
	// Override chapter count from detail
	if v, ok := data["totalEpisode"].(float64); ok {
		book.ChapterCount = int(v)
	}

	var chapters []models.Chapter
	if shortPlayEpisodeInfos, ok := data["shortPlayEpisodeInfos"].([]interface{}); ok {
		for _, ep := range shortPlayEpisodeInfos {
			episode := ep.(map[string]interface{})
			epNum := int(episode["episodeNo"].(float64))
			videoPath := getSt(episode, "playVoucher")
			if videoPath == "" {
				videoPath = getSt(episode, "external_audio_h264_m3u8")
			}

			var subtitles []models.Subtitle
			if subtitleList, ok := episode["subtitleList"].([]interface{}); ok {
				for _, sub := range subtitleList {
					subtitle := sub.(map[string]interface{})
					subtitles = append(subtitles, models.Subtitle{
						Subtitle:    getSt(subtitle, "url"),
						DisplayName: getSt(subtitle, "subtitleLanguage"),
						Language:    getSt(subtitle, "subtitleLanguage"),
						Type:        "normal",
					})
				}
			}

			chapters = append(chapters, models.Chapter{
				Eps:       fmt.Sprintf("EP-%d", epNum),
				Index:     epNum - 1,
				VideoPath: videoPath,
				Subtitle:  subtitles,
			})
		}
	}

	return &models.SeriesDetailResponse{Book: &book, Chapters: chapters}, nil
}

func (n *NetShort) SearchSeries(keyword, language string) (*models.SearchResponse, error) {
	fmt.Printf("[NetShort] SearchSeries called with keyword: %s\n", keyword)
	if language != "" {
		n.SetLanguage(language)
	}

	result, err := n.request("/video/shortPlay/search/searchByKeyword", map[string]interface{}{
		"searchFlag": 0,
		"pageNo":     1,
		"pageSize":   20,
		"searchCode": keyword,
	})
	if err != nil {
		return nil, err
	}

	data, _ := result["data"].(map[string]interface{})
	if data == nil {
		return &models.SearchResponse{IsMore: false}, nil
	}

	searchCodeSearchResult, _ := data["searchCodeSearchResult"].([]interface{})
	var items []models.DramaItem
	for _, item := range searchCodeSearchResult {
		items = append(items, n.transformItem(item.(map[string]interface{})))
	}

	return &models.SearchResponse{IsMore: false, Book: items, PageInfo: models.PageInfo{HasMore: false}}, nil
}

func (n *NetShort) buildFilterBody(params map[string]interface{}) (map[string]interface{}, error) {
	// constant data
	result, err := n.request("/video/classes/constant", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	data, _ := result["data"].(map[string]interface{})

	// Parse params with defaults
	regionKey := 0
	if v, ok := params["regionKey"].(int); ok {
		regionKey = v
	} else if v, ok := params["regionKey"].(float64); ok {
		regionKey = int(v)
	}

	audioKey := 0
	if v, ok := params["audioKey"].(int); ok {
		audioKey = v
	} else if v, ok := params["audioKey"].(float64); ok {
		audioKey = int(v)
	}

	tagId := "semua"
	if v, ok := params["tagId"].(string); ok {
		tagId = v
	}

	offset := 0
	if v, ok := params["offset"].(int); ok {
		offset = v
	} else if v, ok := params["offset"].(float64); ok {
		offset = int(v)
	}

	limit := 15
	if v, ok := params["limit"].(int); ok {
		limit = v
	} else if v, ok := params["limit"].(float64); ok {
		limit = int(v)
	}

	orderMode := "1"
	if v, ok := params["orderMode"].(string); ok {
		orderMode = v
	}

	// Find helpers
	var region, audio, tag map[string]interface{}

	if regions, ok := data["region"].([]interface{}); ok {
		for _, r := range regions {
			rm := r.(map[string]interface{})
			rk := 0
			if k, ok := rm["key"].(float64); ok {
				rk = int(k)
			} else if k, ok := rm["key"].(int); ok {
				rk = k
			}
			if rk == regionKey {
				region = rm
				break
			}
		}
	}

	if audios, ok := data["audio"].([]interface{}); ok {
		for _, a := range audios {
			am := a.(map[string]interface{})
			ak := 0
			if k, ok := am["key"].(float64); ok {
				ak = int(k)
			} else if k, ok := am["key"].(int); ok {
				ak = k
			}
			if ak == audioKey {
				audio = am
				break
			}
		}
	}

	// Convert tagId to string (handle number input)
	tagIdString := tagId
	if tagId != "semua" {
		if tags, ok := data["tag"].([]interface{}); ok {
			for _, t := range tags {
				tm := t.(map[string]interface{})
				labelName := getSt(tm, "labelName")
				// case insensitive check: labelName?.toLowerCase().includes(tagId.toLowerCase())
				if utils.StringContainsCaseInsensitive(labelName, tagIdString) {
					tag = tm
					break
				}
			}
		}
	}

	// Build body - PERUBAHAN: tagIdList logic
	var tagIdList []interface{}
	if tag != nil && tagId != "semua" {
		if labelLanguageId, ok := tag["labelLanguageId"]; ok {
			tagIdList = []interface{}{labelLanguageId}
		} else {
			tagIdList = []interface{}{"-1"}
		}
	} else {
		tagIdList = []interface{}{"-1"}
	}

	body := map[string]interface{}{
		"offset":    offset,
		"orderMode": orderMode,
		"limit":     limit,
		"tagIdList": tagIdList,
	}

	// Logic from JS
	if tagId != "semua" {
		// PERUBAHAN: filterValue logic
		if tag != nil {
			body["filterValue"] = tag["labelLanguageId"]
		} else {
			body["filterValue"] = []interface{}{-1} // PERUBAHAN: sudah array
		}
		body["filterType"] = "categories"

		if audioKey != 0 {
			if audio != nil {
				body["libraryTypeIdList"] = audio["extList"]
			} else {
				body["libraryTypeIdList"] = []interface{}{} // TAMBAHAN: default empty array
			}
		}

		if regionKey != 0 {
			if region != nil {
				body["sourceLanguageList"] = region["extList"]
			} else {
				body["sourceLanguageList"] = []interface{}{} // TAMBAHAN: default empty array
			}
		}

	} else if audioKey != 0 {
		// audioKey != 0 && tagId == 'semua'
		var audioTypes interface{} = []interface{}{}
		if audio != nil && audio["extList"] != nil {
			audioTypes = audio["extList"]
		}
		body["libraryTypeIdList"] = audioTypes

		if regionKey == 0 {
			// Marshal to JSON string
			jsonBytes, _ := json.Marshal(audioTypes)
			body["filterValue"] = string(jsonBytes)
			body["filterType"] = "audio_categories"

		} else {
			// regionKey != 0
			var extList interface{} = []interface{}{}
			if region != nil && region["extList"] != nil {
				extList = region["extList"]
			}
			jsonBytes, _ := json.Marshal(extList)
			body["filterValue"] = string(jsonBytes)
			body["sourceLanguageList"] = extList
			body["filterType"] = "region"
		}

	} else {
		// tagId == 'semua' && audioKey == 0
		var languages interface{} = []interface{}{}
		if region != nil && region["extList"] != nil {
			languages = region["extList"]
		}

		if regionKey == 0 {
			body["filterValue"] = "[]"
		} else {
			jsonBytes, _ := json.Marshal(languages)
			body["filterValue"] = string(jsonBytes)
		}
		body["filterType"] = "region"

		if regionKey != 0 {
			body["sourceLanguageList"] = languages
		}
	}

	return body, nil
}

func (n *NetShort) filterDrama(tagId string, orderMode string, regionKey, audioKey, offset, limit int) ([]models.DramaItem, int, bool, error) {
	params := map[string]interface{}{
		"regionKey": regionKey,
		"audioKey":  audioKey,
		"tagId":     tagId,
		"offset":    offset,
		"limit":     limit,
		"orderMode": orderMode,
	}

	body, err := n.buildFilterBody(params)
	if err != nil {
		return nil, 0, false, err
	}

	// endpoint: /video/classes/page
	res, err := n.request("/video/classes/page", body)
	if err != nil {
		return nil, 0, false, err
	}

	data, _ := res["data"].(map[string]interface{})
	if data == nil {
		return []models.DramaItem{}, 0, false, nil
	}

	var list []models.DramaItem
	if dataList, ok := data["dataList"].([]interface{}); ok {
		for _, item := range dataList {
			list = append(list, n.transformItem(item.(map[string]interface{})))
		}
	}

	maxOffset := 0
	if v, ok := data["maxOffset"].(float64); ok {
		maxOffset = int(v)
	}

	completed := false
	if v, ok := data["completed"].(bool); ok {
		completed = v
	}

	return list, maxOffset, !completed, nil
}

func (n *NetShort) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		n.SetLanguage(language)
	}

	var list []models.DramaItem
	var next int
	var isMore bool
	var err error

	// key might be float64 from generic interface
	keyInt := 0
	keyStr := "" // unused
	if v, ok := tabKey.(float64); ok {
		keyInt = int(v)
		keyStr = fmt.Sprintf("%d", int(v))
	} else if v, ok := tabKey.(int); ok {
		keyInt = v
		keyStr = fmt.Sprintf("%d", v)
	} else if v, ok := tabKey.(string); ok {
		_ = v
		keyStr = v
	}

	// positionIndex might be string/int/float64/slice
	posInt := 0
	posStr := ""
	if v, ok := positionIndex.(float64); ok {
		posInt = int(v)
		posStr = fmt.Sprintf("%d", posInt)
	} else if v, ok := positionIndex.(int); ok {
		posInt = v
		posStr = fmt.Sprintf("%d", v)
	} else if v, ok := positionIndex.(string); ok {
		posStr = v
		// try parse int
		// ...
	}

	switch tabType {
	case "region":
		list, next, isMore, err = n.filterDrama("-1", "1", keyInt, 0, 0, 15)
	case "audio":
		list, next, isMore, err = n.filterDrama("-1", "1", 0, keyInt, 0, 15)
	case "tag":
		// positionIndex is used as tagId (string)
		tagId := posStr
		if tagId == "" {
			tagId = fmt.Sprintf("%v", positionIndex)
		}
		list, next, isMore, err = n.filterDrama(tagId, "1", 0, 0, 0, 15)

	case "orderMode":
		// filterDrama('-1', positionIndex.toString(), ...)
		list, next, isMore, err = n.filterDrama("-1", posStr, 0, 0, 0, 15)
	default:
		list, next, isMore, err = n.filterDrama(keyStr, posStr, 0, 0, 0, 15)
	}

	if err != nil {
		return nil, err
	}

	return &models.TabDataResponse{
		IsMore: isMore,
		Book: models.BookList{
			List: list,
		},
		PageInfo: models.PageInfo{
			HasMore:       isMore,
			Next:          next,
			Type:          tabType,
			Key:           tabKey,
			PositionIndex: positionIndex,
		},
	}, nil
}

func (n *NetShort) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		n.SetLanguage(language)
	}

	nextOffset := 0
	if v, ok := pageInfo.Next.(float64); ok {
		nextOffset = int(v)
	} else if v, ok := pageInfo.Next.(int); ok {
		nextOffset = v
	}

	// Extract keys similarly
	keyInt := 0
	if v, ok := pageInfo.Key.(float64); ok {
		keyInt = int(v)
	}

	posInt := 0
	if v, ok := pageInfo.PositionIndex.(float64); ok {
		posInt = int(v)
	} else if v, ok := pageInfo.PositionIndex.(int); ok {
		posInt = v
	}

	var list []models.DramaItem
	var next int
	var isMore bool
	var err error

	switch pageInfo.Type {
	case "region":
		list, next, isMore, err = n.filterDrama("-1", "1", keyInt, 0, nextOffset, 15)
	case "audio":
		list, next, isMore, err = n.filterDrama("-1", "1", 0, keyInt, nextOffset, 15)
	case "tag":
		list, next, isMore, err = n.filterDrama(fmt.Sprintf("%d", posInt), "1", 0, 0, nextOffset, 15)
	case "orderMode":
		list, next, isMore, err = n.filterDrama("-1", fmt.Sprintf("%d", posInt), 0, 0, nextOffset, 15)
	default:
		list = []models.DramaItem{}
	}

	if err != nil {
		return nil, err
	}

	return &models.TabFeedResponse{
		IsMore: isMore,
		Book:   list,
		PageInfo: models.PageInfo{
			HasMore:       isMore,
			Next:          next,
			Type:          pageInfo.Type,
			Key:           pageInfo.Key,
			PositionIndex: pageInfo.PositionIndex,
		},
	}, nil
}

func getSt(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
