package platforms

import (
	"dramawave/models"
)

type AnimeV2 struct {
	client *WrapperClient
}

func NewAnimeV2(language string) *AnimeV2 {
	return &AnimeV2{
		client: NewWrapperClient("animev2", language),
	}
}

func (a *AnimeV2) SetLanguage(language string) {
	a.client.SetLanguage(language)
}

func (a *AnimeV2) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		a.SetLanguage(language)
	}
	return a.client.GetTabList()
}

func (a *AnimeV2) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		a.SetLanguage(language)
	}
	return a.client.GetTabData(tabType, tabKey, positionIndex)
}

func (a *AnimeV2) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		a.SetLanguage(language)
	}
	return a.client.GetTabFeed(pageInfo)
}

func (a *AnimeV2) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		a.SetLanguage(language)
	}
	return a.client.GetSeriesDetail(id, quality)
}

func (a *AnimeV2) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		a.SetLanguage(language)
	}
	return a.client.SearchSeries(keyword, pageNo, pageSize)
}
