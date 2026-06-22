/*
1. Urgency distribution
2. Physicians
3. Referral Trend changes
4. Average waiting times (Referral Date to Booked Date)
5. Average triage times (Referral Date to Created ts)
6. Complaint distribution
*/

package store

import (
	"context"
	"reftrail/internal/domain"
)

// UrgencyMetric represents a single slice of your pie chart data
type UrgencyMetric struct {
	Urgency    domain.ReferralUrgency `json:"urgency"`    // "Elective", "Urgent", "ASAP"
	Count      int                    `json:"count"`      // Raw total rows
	Percentage float64                `json:"percentage"` // Calculated percentage
}

// The final payload returned to the frontend
type UrgencyDistributionResponse struct {
	Metrics    []*UrgencyMetric `json:"metrics"`
	TotalCount int              `json:"totalCount"`
}

func (s *Store) GetUrgencyDistribution(ctx context.Context, find *FindReferralEntry) (*UrgencyDistributionResponse, error) {
	_, ok := domain.GetUserContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	return s.driver.GetUrgencyDistribution(ctx, find)
}

type TrendMetric struct {
	Period string `json:"period"`
	Count  int    `json:"count"`
}

type ReferralVolumeResponse struct {
	Data       []TrendMetric `json:"data"`
	TotalCount int           `json:"totalCount"`
}

func (s *Store) GetReferralVolume(ctx context.Context, find *FindReferralEntry) (*ReferralVolumeResponse, error) {
	_, ok := domain.GetUserContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	return s.driver.GetReferralVolume(ctx, find)
}

type WaitingTimeTrendMetric struct {
	Period      string  `json:"period"`      // "2026-01", "2026-02" (X-Axis)
	AverageDays float64 `json:"averageDays"` // The calculated active working time (Y-Axis)
}

type WaitingTimeTrendResponse struct {
	Data []WaitingTimeTrendMetric `json:"data"`
}

func (s *Store) GetDirectBookingWaitingTime(ctx context.Context, find *FindReferralEntry) (*WaitingTimeTrendResponse, error) {
	// 1. Accountability Guard: Check if the user is authorized to read analytics data
	_, ok := domain.GetUserContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	// 2. Delegate the raw execution call straight down to your SQLite Driver layer
	return s.driver.GetDirectBookingWaitingTime(ctx, find)
}
