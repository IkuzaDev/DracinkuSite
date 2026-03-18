package platforms

import (
	"dramawave/models"
)

type Movie struct {
	client *WrapperClient
}

func NewMovie(language string) *Movie {
	return &Movie{
		client: NewWrapperClient("movie", language),
	}
}

func (m *Movie) SetLanguage(language string) {
	m.client.SetLanguage(language)
}

func (m *Movie) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		m.SetLanguage(language)
	}
	return m.client.GetTabList()
}

func (m *Movie) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		m.SetLanguage(language)
	}
	return m.client.GetTabData(tabType, tabKey, positionIndex)
}

func (m *Movie) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		m.SetLanguage(language)
	}
	return m.client.GetTabFeed(pageInfo)
}

func (m *Movie) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		m.SetLanguage(language)
	}
	return m.client.GetSeriesDetail(id, quality)
}

func (m *Movie) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		m.SetLanguage(language)
	}
	return m.client.SearchSeries(keyword, pageNo, pageSize)
}
