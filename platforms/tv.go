package platforms

import (
	"dramawave/models"
)

type TV struct {
	client *WrapperClient
}

func NewTV(language string) *TV {
	return &TV{
		client: NewWrapperClient("tv", language),
	}
}

func (t *TV) SetLanguage(language string) {
	t.client.SetLanguage(language)
}

func (t *TV) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		t.SetLanguage(language)
	}
	return t.client.GetTabList()
}

func (t *TV) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		t.SetLanguage(language)
	}
	return t.client.GetTabData(tabType, tabKey, positionIndex)
}

func (t *TV) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		t.SetLanguage(language)
	}
	return t.client.GetTabFeed(pageInfo)
}

func (t *TV) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		t.SetLanguage(language)
	}
	return t.client.GetSeriesDetail(id, quality)
}

func (t *TV) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		t.SetLanguage(language)
	}
	return t.client.SearchSeries(keyword, pageNo, pageSize)
}
