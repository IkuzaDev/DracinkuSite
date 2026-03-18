package platforms

import (
	"dramawave/models"
)

type DramaWave struct {
	client *WrapperClient
}

func NewDramaWave(language string) *DramaWave {
	return &DramaWave{
		client: NewWrapperClient("dramawave", language),
	}
}

func (d *DramaWave) SetLanguage(language string) {
	d.client.SetLanguage(language)
}

func (d *DramaWave) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabList()
}

func (d *DramaWave) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabData(tabType, tabKey, positionIndex)
}

func (d *DramaWave) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetTabFeed(pageInfo)
}

func (d *DramaWave) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.GetSeriesDetail(id, quality)
}

func (d *DramaWave) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		d.SetLanguage(language)
	}
	return d.client.SearchSeries(keyword, pageNo, pageSize)
}
