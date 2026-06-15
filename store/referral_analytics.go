/*
1. Urgency distribution
2. Physicians
3. Referral Trend changes
4. Average waiting times
5. Complaint distribution
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

type ReferralTrendResponse struct {
	Data       []TrendMetric `json:"data"`
	TotalCount int           `json:"totalCount"`
}

func (s *Store) GetReferralTrend(ctx context.Context, find *FindReferralEntry) (*ReferralTrendResponse, error) {
	_, ok := domain.GetUserContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	return s.driver.GetReferralTrend(ctx, find)
}
