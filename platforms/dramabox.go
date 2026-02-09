package platforms

import (
	"dramawave/models"
)

type Dramabox struct {
	client *WrapperClient
}

func NewDramabox(language string) *Dramabox {
	return &Dramabox{
		client: NewWrapperClient("dramabox", language),
	}
}

func (d *Dramabox) SetLanguage(language string) {
	d.client.SetLanguage(language)
}

func (d *Dramabox) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabList()
}

func (d *Dramabox) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabData(tabType, tabKey, positionIndex)
}

func (d *Dramabox) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabFeed(pageInfo)
}

func (d *Dramabox) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetSeriesDetail(id, quality)
}

func (d *Dramabox) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.SearchSeries(keyword, pageNo, pageSize)
}
