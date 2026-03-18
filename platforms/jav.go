package platforms

import (
	"dramawave/models"
)

type Jav struct {
	client *WrapperClient
}

func NewJav(language string) *Jav {
	return &Jav{
		client: NewWrapperClient("bjav", language),
	}
}

func (j *Jav) SetLanguage(language string) {
	j.client.SetLanguage(language)
}

func (j *Jav) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		j.SetLanguage(language)
	}
	return j.client.GetTabList()
}

func (j *Jav) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		j.SetLanguage(language)
	}
	return j.client.GetTabData(tabType, tabKey, positionIndex)
}

func (j *Jav) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		j.SetLanguage(language)
	}
	return j.client.GetTabFeed(pageInfo)
}

func (j *Jav) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		j.SetLanguage(language)
	}
	return j.client.GetSeriesDetail(id, quality)
}

func (j *Jav) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		j.SetLanguage(language)
	}
	return j.client.SearchSeries(keyword, pageNo, pageSize)
}
