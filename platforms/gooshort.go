package platforms

import (
	"dramawave/models"
)

type GooShort struct {
	client *WrapperClient
}

func NewGooShort(language string) *GooShort {
	return &GooShort{
		client: NewWrapperClient("gooshort", language),
	}
}

func (g *GooShort) SetLanguage(language string) {
	g.client.SetLanguage(language)
}

func (g *GooShort) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		g.SetLanguage(language)
	}
	return g.client.GetTabList()
}

func (g *GooShort) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		g.SetLanguage(language)
	}
	return g.client.GetTabData(tabType, tabKey, positionIndex)
}

func (g *GooShort) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		g.SetLanguage(language)
	}
	return g.client.GetTabFeed(pageInfo)
}

func (g *GooShort) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		g.SetLanguage(language)
	}
	return g.client.GetSeriesDetail(id, quality)
}

func (g *GooShort) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		g.SetLanguage(language)
	}
	return g.client.SearchSeries(keyword, pageNo, pageSize)
}
