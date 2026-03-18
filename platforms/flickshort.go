package platforms

import (
	"dramawave/models"
)

type FlickShort struct {
	client *WrapperClient
}

func NewFlickShort(language string) *FlickShort {
	return &FlickShort{
		client: NewWrapperClient("flickshort", language),
	}
}

func (f *FlickShort) SetLanguage(language string) {
	f.client.SetLanguage(language)
}

func (f *FlickShort) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.GetTabList()
}

func (f *FlickShort) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.GetTabData(tabType, tabKey, positionIndex)
}

func (f *FlickShort) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.GetTabFeed(pageInfo)
}

func (f *FlickShort) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.GetSeriesDetail(id, quality)
}

func (f *FlickShort) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		f.SetLanguage(language)
	}
	return f.client.SearchSeries(keyword, pageNo, pageSize)
}
