package platforms

import (
	"dramawave/models"
)

type Drakor struct {
	client *WrapperClient
}

func NewDrakor(language string) *Drakor {
	return &Drakor{
		client: NewWrapperClient("drakor", language),
	}
}

func (d *Drakor) SetLanguage(language string) {
	d.client.SetLanguage(language)
}

func (d *Drakor) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabList()
}

func (d *Drakor) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabData(tabType, tabKey, positionIndex)
}

func (d *Drakor) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabFeed(pageInfo)
}

func (d *Drakor) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetSeriesDetail(id, quality)
}

func (d *Drakor) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.SearchSeries(keyword, pageNo, pageSize)
}
