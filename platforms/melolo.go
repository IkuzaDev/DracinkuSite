package platforms

import (
	"dramawave/models"
)

type Melolo struct {
	client *WrapperClient
}

func NewMelolo(language string) *Melolo {
	return &Melolo{
		client: NewWrapperClient("melolo", language),
	}
}

func (m *Melolo) SetLanguage(language string) {
	m.client.SetLanguage(language)
}

func (m *Melolo) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		m.SetLanguage(language)
	}
	return m.client.GetTabList()
}

func (m *Melolo) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		m.SetLanguage(language)
	}
	return m.client.GetTabData(tabType, tabKey, positionIndex)
}

func (m *Melolo) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		m.SetLanguage(language)
	}
	return m.client.GetTabFeed(pageInfo)
}

func (m *Melolo) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		m.SetLanguage(language)
	}
	return m.client.GetSeriesDetail(id, quality)
}

func (m *Melolo) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		m.SetLanguage(language)
	}
	return m.client.SearchSeries(keyword, pageNo, pageSize)
}
