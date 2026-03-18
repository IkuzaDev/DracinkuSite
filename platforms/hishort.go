package platforms

import (
	"dramawave/models"
)

type HiShort struct {
	client *WrapperClient
}

func NewHiShort(language string) *HiShort {
	return &HiShort{
		client: NewWrapperClient("hishort", language),
	}
}

func (h *HiShort) SetLanguage(language string) {
	h.client.SetLanguage(language)
}

func (h *HiShort) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		h.SetLanguage(language)
	}
	return h.client.GetTabList()
}

func (h *HiShort) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		h.SetLanguage(language)
	}
	return h.client.GetTabData(tabType, tabKey, positionIndex)
}

func (h *HiShort) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		h.SetLanguage(language)
	}
	return h.client.GetTabFeed(pageInfo)
}

func (h *HiShort) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		h.SetLanguage(language)
	}
	return h.client.GetSeriesDetail(id, quality)
}

func (h *HiShort) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		h.SetLanguage(language)
	}
	return h.client.SearchSeries(keyword, pageNo, pageSize)
}
