package playstore

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/api/playdeveloperreporting/v1beta1"
)

type RatingSummary struct {
	Average float64
	Count   int
}

func (c *Client) FetchRatingSummary(ctx context.Context, packageName string) (RatingSummary, error) {
	var total, count int64

	call := c.publisher.Reviews.List(packageName).Context(ctx).MaxResults(100)
	resp, err := call.Do()
	if err != nil {
		return RatingSummary{}, fmt.Errorf("list reviews: %w", err)
	}

	for _, review := range resp.Reviews {
		if review.Comments == nil {
			continue
		}
		for _, comment := range review.Comments {
			if comment.UserComment == nil || comment.UserComment.StarRating == 0 {
				continue
			}
			total += comment.UserComment.StarRating
			count++
		}
	}

	if count == 0 {
		return RatingSummary{}, nil
	}

	return RatingSummary{
		Average: float64(total) / float64(count),
		Count:   int(count),
	}, nil
}

type VitalsRates struct {
	CrashRate float64
	AnrRate   float64
}

func (c *Client) FetchVitalsRates(ctx context.Context, packageName string) (VitalsRates, error) {
	name := fmt.Sprintf("apps/%s", packageName)
	timeline := dailyTimelineForYesterday()

	crashResp, err := c.reporting.Vitals.Crashrate.Query(name, &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		Metrics:      []string{"crashRate"},
		TimelineSpec: timeline,
	}).Context(ctx).Do()
	if err != nil {
		return VitalsRates{}, fmt.Errorf("query crash rate: %w", err)
	}

	anrResp, err := c.reporting.Vitals.Anrrate.Query(name, &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		Metrics:      []string{"anrRate"},
		TimelineSpec: timeline,
	}).Context(ctx).Do()
	if err != nil {
		return VitalsRates{}, fmt.Errorf("query anr rate: %w", err)
	}

	return VitalsRates{
		CrashRate: firstDecimalMetric(crashResp.Rows, "crashRate"),
		AnrRate:   firstDecimalMetric(anrResp.Rows, "anrRate"),
	}, nil
}

func dailyTimelineForYesterday() *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec {
	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	start := &playdeveloperreporting.GoogleTypeDateTime{
		Year:  int64(yesterday.Year()),
		Month: int64(yesterday.Month()),
		Day:   int64(yesterday.Day()),
	}
	today := yesterday.AddDate(0, 0, 1)
	end := &playdeveloperreporting.GoogleTypeDateTime{
		Year:  int64(today.Year()),
		Month: int64(today.Month()),
		Day:   int64(today.Day()),
	}

	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "DAILY",
		StartTime:         start,
		EndTime:           end,
	}
}

func firstDecimalMetric(rows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow, metricName string) float64 {
	for _, row := range rows {
		for _, metric := range row.Metrics {
			if metric.Metric != metricName || metric.DecimalValue == nil {
				continue
			}
			if v, err := strconv.ParseFloat(metric.DecimalValue.Value, 64); err == nil {
				return v
			}
		}
	}
	return 0
}
