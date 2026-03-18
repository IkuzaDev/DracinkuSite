package models

import "encoding/json"

type DramaItem struct {
	ID           interface{} `json:"id"`
	Name         string      `json:"name"`
	Cover        interface{} `json:"cover"`
	Badge        interface{} `json:"badge,omitempty"`
	ChapterCount interface{} `json:"chapterCount"`
	Introduction interface{} `json:"introduction"`
	Tags         interface{} `json:"tags"`
	PlayCount    interface{} `json:"playCount"`
}

type Subtitle struct {
	Subtitle    string `json:"subtitle"`
	DisplayName string `json:"display_name"`
	Language    string `json:"language"`
	Type        string `json:"type"`
}

type Chapter struct {
	Eps       string     `json:"eps"`
	Index     int        `json:"index"`
	VideoPath string     `json:"videoPath"`
	Subtitle  []Subtitle `json:"subtitle"`
}

type PageInfo struct {
	HasMore       bool   `json:"has_more"`
	PageNo        int    `json:"pageNo,omitempty"`
	PageSize      int    `json:"pageSize,omitempty"`
	TabKey        any    `json:"tabKey,omitempty"`
	PositionIndex any    `json:"positionIndex,omitempty"`
	TabKey2       any    `json:"tab_key,omitempty"`
	PositionIndex2 any   `json:"position_index,omitempty"`
	Next          any    `json:"next,omitempty"`
	Current       int    `json:"current,omitempty"`
	Total         int    `json:"total,omitempty"`
	ModuleKey     string `json:"module_key,omitempty"`
	BusinessID    any    `json:"businessId,omitempty"`
	Type          string `json:"type,omitempty"`
	Key           any    `json:"key,omitempty"`
	Keyword       string `json:"keyword,omitempty"`
}

type TabItem struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	TabKey        any    `json:"tab_key"`
	PositionIndex any    `json:"position_index"`
}

type BookList struct {
	List      []DramaItem `json:"list"`
	Recommend []DramaItem `json:"recommend,omitempty"`
	Preview   []DramaItem `json:"preview,omitempty"`
}

// UnmarshalJSON implements custom unmarshaling for BookList
// It handles both object {"list": [], ...} and array [item1, item2] formats
func (b *BookList) UnmarshalJSON(data []byte) error {
	// Try parsing as array first
	var list []DramaItem
	if err := json.Unmarshal(data, &list); err == nil {
		b.List = list
		return nil
	}

	// If not array, try parsing as struct
	// We define a local alias to avoid recursive call to UnmarshalJSON
	type Alias BookList
	var aux Alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*b = BookList(aux)
	return nil
}

type TabDataResponse struct {
	IsMore   bool     `json:"isMore"`
	Book     BookList `json:"book"`
	PageInfo PageInfo `json:"page_info"`
}

type TabFeedResponse struct {
	IsMore   bool        `json:"isMore"`
	Book     []DramaItem `json:"book"`
	PageInfo PageInfo    `json:"page_info"`
}

type SeriesDetailResponse struct {
	Book     *DramaItem `json:"book"`
	Chapters []Chapter  `json:"chapters"`
}

type SearchResponse struct {
	IsMore   bool        `json:"isMore,omitempty"`
	Book     []DramaItem `json:"book"`
	PageInfo PageInfo    `json:"page_info,omitempty"`
}

type DramaListResponse struct {
	IsMore bool        `json:"isMore"`
	Book   []DramaItem `json:"book"`
}
