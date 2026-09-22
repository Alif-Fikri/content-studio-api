package ads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const maxInsightsPages = 50

type MetaClient struct {
	httpClient  *http.Client
	token       string
	adAccountID string
}

func NewMetaClient(token, adAccountID string) *MetaClient {
	return &MetaClient{
		httpClient:  http.DefaultClient,
		token:       token,
		adAccountID: adAccountID,
	}
}

type InsightRow struct {
	Date        string `json:"date_start"`
	Impressions string `json:"impressions"`
	Reach       string `json:"reach"`
	Clicks      string `json:"clicks"`
	Spend       string `json:"spend"`
}

type insightsResponse struct {
	Data   []InsightRow `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

func (m *MetaClient) FetchInsights(ctx context.Context, externalAdID, since string) ([]InsightRow, error) {
	url := fmt.Sprintf(
		"https://graph.facebook.com/v21.0/%s/insights?fields=impressions,reach,clicks,spend&time_range={\"since\":\"%s\",\"until\":\"today\"}&time_increment=1&access_token=%s",
		externalAdID, since, m.token,
	)

	var rows []InsightRow

	for page := 0; page < maxInsightsPages && url != ""; page++ {
		body, err := m.fetchPage(ctx, url)
		if err != nil {
			return nil, err
		}
		rows = append(rows, body.Data...)
		url = body.Paging.Next
	}

	return rows, nil
}

func (m *MetaClient) fetchPage(ctx context.Context, url string) (*insightsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("meta insights request failed: %s", resp.Status)
	}

	var body insightsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return &body, nil
}
