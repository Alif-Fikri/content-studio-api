package ads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type MetaClient struct {
	httpClient *http.Client
	token      string
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

func (m *MetaClient) FetchInsights(ctx context.Context, externalAdID, since string) ([]InsightRow, error) {
	url := fmt.Sprintf(
		"https://graph.facebook.com/v21.0/%s/insights?fields=impressions,reach,clicks,spend&time_range={\"since\":\"%s\",\"until\":\"today\"}&time_increment=1&access_token=%s",
		externalAdID, since, m.token,
	)

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

	var body struct {
		Data []InsightRow `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.Data, nil
}
