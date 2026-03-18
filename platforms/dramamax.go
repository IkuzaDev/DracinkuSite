package platforms

import (
	"dramawave/models"
)

type DramaMax struct {
	client *WrapperClient
}

func NewDramaMax(language string) *DramaMax {
	return &DramaMax{
		client: NewWrapperClient("dramamax", language),
	}
}

func (d *DramaMax) SetLanguage(language string) {
	d.client.SetLanguage(language)
}

func (d *DramaMax) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabList()
}

func (d *DramaMax) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabData(tabType, tabKey, positionIndex)
}

func (d *DramaMax) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabFeed(pageInfo)
}

func (d *DramaMax) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetSeriesDetail(id, quality)
}

func (d *DramaMax) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.SearchSeries(keyword, pageNo, pageSize)
}
