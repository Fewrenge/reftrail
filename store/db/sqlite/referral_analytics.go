package sqlite

import (
	"context"
	"fmt"
	"reftrail/store"
	"strings"
	"time"
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

func (d *Driver) GetReferralVolume(ctx context.Context, find *store.FindReferralEntry) (*store.ReferralVolumeResponse, error) {
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

	return &store.ReferralVolumeResponse{
		Data:       data,
		TotalCount: totalCount,
	}, nil
}

func (d *Driver) GetDirectBookingWaitingTime(ctx context.Context, find *store.FindReferralEntry) (*store.WaitingTimeTrendResponse, error) {
	// ✅ 1. Swapped r.created_ts for r.referral_date as your analytics starting point
	query := `
		SELECT 
			strftime('%Y-%m', r.referral_date) as period,
			r.referral_date,
			l.created_ts as booked_ts
		FROM referral_entry r
		JOIN referral_log l ON r.id = l.referral_id
		WHERE r.status = 'BOOKED' 
		  AND l.new_status = 'BOOKED'
		  AND r.referral_date IS NOT NULL AND r.referral_date != ''
		   AND r.id NOT IN (
			  SELECT referral_id FROM referral_log 
			  WHERE new_status IN ('1ST_CALL_COMPLETE', '2ND_CALL_COMPLETE', '3RD_CALL_COMPLETE', 'SUSPENDED', 'PATIENT_TO_CALL_BACK')
		  )
		ORDER BY period ASC`

	rows, err := d.conn(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type totalTracker struct {
		TotalDays float64
		Count     int
	}
	aggregates := make(map[string]*totalTracker)

	for rows.Next() {
		var period, referralDateStr, bookedStr string
		if err := rows.Scan(&period, &referralDateStr, &bookedStr); err != nil {
			return nil, err
		}

		// FIX: Parse referral_date using standard YYYY-MM-DD layout string syntax
		referralTime, errRef := time.Parse("2006-01-02", referralDateStr)
		bookedTime, errBook := time.Parse(time.RFC3339, bookedStr) // Keeps full datetime precision for completion

		// Skip rows safely if database date strings are corrupted or unparseable
		if errRef != nil || errBook != nil {
			continue
		}

		// Calculate total days elapsed from the clinical letter date to booking execution
		daysElapsed := bookedTime.Sub(referralTime).Hours() / 24.0

		if aggregates[period] == nil {
			aggregates[period] = &totalTracker{}
		}
		aggregates[period].TotalDays += daysElapsed
		aggregates[period].Count++
	}

	var trendData []store.WaitingTimeTrendMetric
	for period, tracker := range aggregates {
		avgDays := 0.0
		if tracker.Count > 0 {
			avgDays = tracker.TotalDays / float64(tracker.Count)
		}

		trendData = append(trendData, store.WaitingTimeTrendMetric{
			Period: period,
			// Clean rounding block to 1 decimal place signature output
			AverageDays: float64(int(avgDays*10+0.5)) / 10.0,
		})
	}

	return &store.WaitingTimeTrendResponse{Data: trendData}, nil
}
