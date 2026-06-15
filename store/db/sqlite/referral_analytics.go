package sqlite

import (
	"context"
	"fmt"
	"reftrail/store"
	"strings"
)

// GetUrgencyDistribution aggregates referral metrics for pie chart rendering
func (d *Driver) GetUrgencyDistribution(ctx context.Context, find *store.FindReferralEntry) (*store.UrgencyDistributionResponse, error) {
	// Base query pulling counts grouped by the urgency column
	baseQuery := `SELECT urgency, COUNT(*) as count FROM referral_entry WHERE 1 = 1`
	var args []any

	// Reuse date range bounds logic
	if find.ReferralDateFrom != nil && *find.ReferralDateFrom != "" {
		baseQuery += " AND referral_date >= ?"
		args = append(args, *find.ReferralDateFrom)
	}
	if find.ReferralDateTo != nil && *find.ReferralDateTo != "" {
		baseQuery += " AND referral_date <= ?"
		args = append(args, *find.ReferralDateTo)
	}

	// Optional: reuse your existing Tag or Status filters here
	if len(find.Statuses) > 0 {
		placeholders := make([]string, len(find.Statuses))
		for i, s := range find.Statuses {
			placeholders[i] = "?"
			args = append(args, string(s))
		}
		baseQuery += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ", "))
	}

	// Group the data by urgency
	baseQuery += " GROUP BY urgency"

	rows, err := d.conn(ctx).QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*store.UrgencyMetric
	totalCount := 0

	// Scan results from SQLite
	for rows.Next() {
		var m store.UrgencyMetric
		if err := rows.Scan(&m.Urgency, &m.Count); err != nil {
			return nil, err
		}

		// Handle empty or null urgency values gracefully
		if m.Urgency == "" {
			m.Urgency = "Unassigned"
		}

		totalCount += m.Count
		metrics = append(metrics, &m)
	}

	// Calculate percentages in Go to deliver clean data straight to the frontend
	if totalCount > 0 {
		for _, m := range metrics {
			m.Percentage = (float64(m.Count) / float64(totalCount)) * 100
		}
	}

	return &store.UrgencyDistributionResponse{
		Metrics:    metrics,
		TotalCount: totalCount,
	}, nil
}

func (d *Driver) GetReferralTrend(ctx context.Context, find *store.FindReferralEntry) (*store.ReferralTrendResponse, error) {
	// Base query formatting date string to 'YYYY-MM' for timeline continuity
	baseQuery := `
		SELECT 
			strftime('%Y-%m', referral_date) as period, 
			COUNT(*) as count 
		FROM referral_entry 
		WHERE 1 = 1 AND referral_date IS NOT NULL AND referral_date != ''`
	var args []any

	// Reuse date range bounds logic
	if find.ReferralDateFrom != nil && *find.ReferralDateFrom != "" {
		baseQuery += " AND referral_date >= ?"
		args = append(args, *find.ReferralDateFrom)
	}
	if find.ReferralDateTo != nil && *find.ReferralDateTo != "" {
		baseQuery += " AND referral_date <= ?"
		args = append(args, *find.ReferralDateTo)
	}

	// Reuse Status filter mappings
	if len(find.Statuses) > 0 {
		placeholders := make([]string, len(find.Statuses))
		for i, s := range find.Statuses {
			placeholders[i] = "?"
			args = append(args, string(s))
		}
		baseQuery += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ", "))
	}

	// Group chronologically so the LineChart renders left-to-right correctly
	baseQuery += " GROUP BY period ORDER BY period ASC"

	rows, err := d.conn(ctx).QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []store.TrendMetric
	totalCount := 0

	// Scan trend results from SQLite
	for rows.Next() {
		var t store.TrendMetric
		if err := rows.Scan(&t.Period, &t.Count); err != nil {
			return nil, err
		}

		// Fallback for corrupt or improperly formatted entries
		if t.Period == "" {
			t.Period = "Unknown"
		}

		totalCount += t.Count
		data = append(data, t)
	}

	// Check if loop encountered errors midway
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &store.ReferralTrendResponse{
		Data:       data,
		TotalCount: totalCount,
	}, nil
}
