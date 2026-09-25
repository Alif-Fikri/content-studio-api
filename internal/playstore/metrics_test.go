package playstore

import (
	"testing"

	"google.golang.org/api/playdeveloperreporting/v1beta1"
)

func TestFirstDecimalMetric_Found(t *testing.T) {
	rows := []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{
		{
			Metrics: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{
				{Metric: "crashRate", DecimalValue: &playdeveloperreporting.GoogleTypeDecimal{Value: "1.25"}},
			},
		},
	}

	got := firstDecimalMetric(rows, "crashRate")
	if got != 1.25 {
		t.Errorf("firstDecimalMetric = %v, want 1.25", got)
	}
}

func TestFirstDecimalMetric_NotFound(t *testing.T) {
	rows := []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{
		{
			Metrics: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{
				{Metric: "anrRate", DecimalValue: &playdeveloperreporting.GoogleTypeDecimal{Value: "0.5"}},
			},
		},
	}

	got := firstDecimalMetric(rows, "crashRate")
	if got != 0 {
		t.Errorf("firstDecimalMetric = %v, want 0 for missing metric", got)
	}
}

func TestDailyTimelineForYesterday_OneDayApart(t *testing.T) {
	timeline := dailyTimelineForYesterday()

	if timeline.AggregationPeriod != "DAILY" {
		t.Errorf("AggregationPeriod = %q, want DAILY", timeline.AggregationPeriod)
	}
	if timeline.StartTime == nil || timeline.EndTime == nil {
		t.Fatal("expected start and end time to be set")
	}

	start := timeline.StartTime
	end := timeline.EndTime
	sameDay := start.Year == end.Year && start.Month == end.Month && start.Day == end.Day-1
	monthRollover := start.Year == end.Year && start.Month == end.Month-1
	yearRollover := start.Year == end.Year-1
	if !sameDay && !monthRollover && !yearRollover {
		t.Errorf("expected end to be exactly one day after start, got start=%+v end=%+v", start, end)
	}
}
