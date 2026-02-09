package platforms

import (
	"dramawave/models"
)

type DramaDash struct {
	client *WrapperClient
}

func NewDramaDash(language string) *DramaDash {
	return &DramaDash{
		client: NewWrapperClient("dramadash", language),
	}
}

func (d *DramaDash) SetLanguage(language string) {
	d.client.SetLanguage(language)
}

func (d *DramaDash) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabList()
}

func (d *DramaDash) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabData(tabType, tabKey, positionIndex)
}

func (d *DramaDash) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabFeed(pageInfo)
}

func (d *DramaDash) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetSeriesDetail(id, quality)
}

func (d *DramaDash) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.SearchSeries(keyword, pageNo, pageSize)
}
