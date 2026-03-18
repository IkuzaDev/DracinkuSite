package platforms

import (
	"dramawave/models"
)

type Anime struct {
	client *WrapperClient
}

func NewAnime(language string) *Anime {
	return &Anime{
		client: NewWrapperClient("anime", language),
	}
}

func (a *Anime) SetLanguage(language string) {
	a.client.SetLanguage(language)
}

func (a *Anime) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		a.SetLanguage(language)
	}
	return a.client.GetTabList()
}

func (a *Anime) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		a.SetLanguage(language)
	}
	return a.client.GetTabData(tabType, tabKey, positionIndex)
}

func (a *Anime) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		a.SetLanguage(language)
	}
	return a.client.GetTabFeed(pageInfo)
}

func (a *Anime) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		a.SetLanguage(language)
	}
	return a.client.GetSeriesDetail(id, quality)
}

func (a *Anime) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		a.SetLanguage(language)
	}
	return a.client.SearchSeries(keyword, pageNo, pageSize)
}
