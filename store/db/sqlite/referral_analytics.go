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

	// Reuse your date range bounds logic
	if find.ReferralDateFrom != nil && *find.ReferralDateFrom != "" {
		baseQuery += " AND referral_date >= ?"
		args = append(args, *find.ReferralDateFrom)
	}
	if find.ReferralDateTo != nil && *find.ReferralDateTo != "" {
		baseQuery += " AND referral_date <= ?"
		args = append(args, *find.ReferralDateTo)
	}

	// Optional: You can reuse your existing Tag or Status filters here if your boss
	// wants to see urgency breakdown for specific clinics or tags later.
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
