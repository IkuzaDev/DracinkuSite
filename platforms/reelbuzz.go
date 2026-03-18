package platforms

import (
	"dramawave/models"
)

type ReelBuzz struct {
	client *WrapperClient
}

func NewReelBuzz(language string) *ReelBuzz {
	return &ReelBuzz{
		client: NewWrapperClient("reelbuzz", language),
	}
}

func (r *ReelBuzz) SetLanguage(language string) {
	r.client.SetLanguage(language)
}

func (r *ReelBuzz) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		r.SetLanguage(language)
	}
	return r.client.GetTabList()
}

func (r *ReelBuzz) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		r.SetLanguage(language)
	}
	return r.client.GetTabData(tabType, tabKey, positionIndex)
}

func (r *ReelBuzz) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		r.SetLanguage(language)
	}
	return r.client.GetTabFeed(pageInfo)
}

func (r *ReelBuzz) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		r.SetLanguage(language)
	}
	return r.client.GetSeriesDetail(id, quality)
}

func (r *ReelBuzz) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		r.SetLanguage(language)
	}
	return r.client.SearchSeries(keyword, pageNo, pageSize)
}
