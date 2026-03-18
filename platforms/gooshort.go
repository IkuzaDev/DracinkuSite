package platforms

import (
	"dramawave/models"
)

type GoodShort struct {
	client *WrapperClient
}

func NewGoodShort(language string) *GoodShort {
	return &GoodShort{
		client: NewWrapperClient("goodshort", language),
	}
}

func (g *GoodShort) SetLanguage(language string) {
	g.client.SetLanguage(language)
}

func (g *GoodShort) GetTabList(language string) ([]models.TabItem, error) {
	if language != "" {
		g.SetLanguage(language)
	}
	return g.client.GetTabList()
}

func (g *GoodShort) GetTabData(tabType string, tabKey interface{}, positionIndex interface{}, language string) (*models.TabDataResponse, error) {
	if language != "" {
		g.SetLanguage(language)
	}
	return g.client.GetTabData(tabType, tabKey, positionIndex)
}

func (g *GoodShort) GetTabFeed(pageInfo models.PageInfo, language string) (*models.TabFeedResponse, error) {
	if language != "" {
		g.SetLanguage(language)
	}
	return g.client.GetTabFeed(pageInfo)
}

func (g *GoodShort) GetSeriesDetail(id string, quality int, language string) (*models.SeriesDetailResponse, error) {
	if language != "" {
		g.SetLanguage(language)
	}
	return g.client.GetSeriesDetail(id, quality)
}

func (g *GoodShort) SearchSeries(keyword string, pageNo, pageSize int, language string) (*models.SearchResponse, error) {
	if language != "" {
		g.SetLanguage(language)
	}
	return g.client.SearchSeries(keyword, pageNo, pageSize)
}
