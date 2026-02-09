package platforms

import (
	"dramawave/models"
)

type ShortMax struct {
	client *WrapperClient
}

func NewShortMax(language string) *ShortMax {
	return &ShortMax{
		client: NewWrapperClient("shortmax", language),
	}
}

func (s *ShortMax) SetLanguage(language string) {
	s.client.SetLanguage(language)
}

func (s *ShortMax) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		s.SetLanguage(language)
	}
	return s.client.GetTabList()
}

func (s *ShortMax) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		s.SetLanguage(language)
	}
	// wrapper expects interface{} for positionIndex, int is fine.
	return s.client.GetTabData(tabType, tabKey, positionIndex)
}

func (s *ShortMax) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		s.SetLanguage(language)
	}
	return s.client.GetTabFeed(pageInfo)
}

func (s *ShortMax) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		s.SetLanguage(language)
	}
	return s.client.GetSeriesDetail(id, quality)
}

func (s *ShortMax) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		s.SetLanguage(language)
	}
	return s.client.SearchSeries(keyword, pageNo, pageSize)
}
