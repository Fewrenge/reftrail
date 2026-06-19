package store

import (
	"context"
	"fmt"
	"math"
	"strings"
)

type ReferralPhysician struct {
	ID             string  `json:"id"`
	CPSONumber     *string `json:"cpsoNumber"`
	FirstName      string  `json:"firstName"`
	LastName       string  `json:"lastName"`
	EMRPhysicianID *string `json:"emrPhysicianId"`
}

type CreateReferralPhysician struct {
	CPSONumber     *string `json:"cpsoNumber"`
	FirstName      string  `json:"firstName"`
	LastName       string  `json:"lastName"`
	EMRPhysicianID *string `json:"emrPhysicianId"`
}

type UpdateReferralPhysician struct {
	ID             string  `json:"id"`
	CPSONumber     *string `json:"cpsoNumber"`
	FirstName      *string `json:"firstName"`
	LastName       *string `json:"lastName"`
	EMRPhysicianID *string `json:"emrPhysicianId"`
}

type FindReferralPhysician struct {
	ID             *string `json:"id"`
	CPSONumber     *string `json:"cpsoNumber"`
	FirstName      *string `json:"firstName"`
	LastName       *string `json:"lastName"`
	EMRPhysicianID *string `json:"emrPhysicianId"`
	GeneralSearch  *string `json:"generalSearch"`

	Limit  *int `json:"limit"`
	Offset *int `json:"offset"`
}

type DeleteReferralPhysician struct {
	ID string `json:"id"`
}

type PaginatedReferralPhysicians struct {
	ReferralPhysicians []*ReferralPhysician `json:"referralPhysicians"`
	TotalCount         int                  `json:"totalCount"`
}

func (s *Store) CreateReferralPhysician(ctx context.Context, create *CreateReferralPhysician) (*ReferralPhysician, error) {
	if strings.TrimSpace(create.FirstName) == "" || strings.TrimSpace(create.LastName) == "" {
		return nil, fmt.Errorf("business validation failed: physician first and last names are mandatory fields")
	}
	return s.driver.CreateReferralPhysician(ctx, create)
}

func (s *Store) ListReferralPhysicians(ctx context.Context, find *FindReferralPhysician) (*PaginatedReferralPhysicians, error) {
	// 1. Get the un-paginated total count for your frontend pagination math
	totalCount, err := s.driver.GetReferralPhysiciansCount(ctx, find)
	if err != nil {
		return nil, fmt.Errorf("failed fetching referral physicians count: %w", err)
	}

	// 2. Fetch the windowed list of physicians from your driver layer
	physicians, err := s.driver.ListReferralPhysicians(ctx, find)
	if err != nil {
		return nil, fmt.Errorf("failed fetching list of referral physicians: %w", err)
	}

	// 3. Prevent null values in JSON array returns by outputting an empty initialized slice
	if len(physicians) == 0 {
		return &PaginatedReferralPhysicians{
			ReferralPhysicians: []*ReferralPhysician{},
			TotalCount:         totalCount,
		}, nil
	}

	// 4. Return the packed response mapping payload back to the handler endpoint
	return &PaginatedReferralPhysicians{
		ReferralPhysicians: physicians,
		TotalCount:         totalCount,
	}, nil
}

func (s *Store) GetReferralPhysicianByID(ctx context.Context, id string) (*ReferralPhysician, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("business validation failed: query target id cannot be blank")
	}
	return s.driver.GetReferralPhysicianByID(ctx, id)
}

func (s *Store) UpdateReferralPhysician(ctx context.Context, update *UpdateReferralPhysician) error {
	if update == nil || strings.TrimSpace(update.ID) == "" {
		return fmt.Errorf("business validation failed: a valid physician structure containing an ID is required for updates")
	}
	return s.driver.UpdateReferralPhysician(ctx, update)
}

func (s *Store) DeleteReferralPhysician(ctx context.Context, delete *DeleteReferralPhysician) error {
	if delete == nil || strings.TrimSpace(delete.ID) == "" {
		return fmt.Errorf("business validation failed: valid delete request structure containing an ID required")
	}

	err := s.driver.DeleteReferralPhysician(ctx, delete)
	if err != nil {
		if strings.Contains(strings.ToUpper(err.Error()), "FOREIGN KEY") {
			return fmt.Errorf("cannot remove physician: doctor is actively linked to ongoing patient referral records")
		}
		return err
	}

	return nil
}

// CalculateSimilarity returns a score between 0.0 (unrelated) and 1.0 (identical)
func CalculateSimilarity(s1, s2 string) float64 {
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	if s1 == s2 {
		return 1.0
	}

	len1, len2 := len(s1), len(s2)
	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	matchDistance := int(math.Max(float64(len1), float64(len2))/2) - 1
	if matchDistance < 0 {
		matchDistance = 0
	}

	matches1 := make([]bool, len1)
	matches2 := make([]bool, len2)

	matchCount := 0
	transpositions := 0

	for i := 0; i < len1; i++ {
		start := int(math.Max(0, float64(i-matchDistance)))
		end := int(math.Min(float64(i+matchDistance+1), float64(len2)))

		for j := start; j < end; j++ {
			if matches2[j] {
				continue
			}
			if s1[i] == s2[j] {
				matches1[i] = true
				matches2[j] = true
				matchCount++
				break
			}
		}
	}

	if matchCount == 0 {
		return 0.0
	}

	k := 0
	for i := 0; i < len1; i++ {
		if !matches1[i] {
			continue
		}
		for !matches2[k] {
			k++
		}
		if s1[i] != s2[k] {
			transpositions++
		}
		k++
	}

	m := float64(matchCount)
	jaro := (m/float64(len1) + m/float64(len2) + (m-float64(transpositions)/2.0)/m) / 3.0

	// Winkler adjustment for common prefixes (up to 4 characters)
	prefixLength := 0
	maxPrefix := int(math.Min(4, math.Min(float64(len1), float64(len2))))
	for i := 0; i < maxPrefix; i++ {
		if s1[i] == s2[i] {
			prefixLength++
		} else {
			break
		}
	}

	return jaro + (float64(prefixLength) * 0.1 * (1.0 - jaro))
}

// ResolvePhysicianID evaluates a messy plain text name against an in-memory master list.
// It returns a valid UUID string if an 95% match is found, or an empty string if it's a new doctor.
func (s *Store) ResolvePhysicianID(rawName string, masterList []*ReferralPhysician) string {
	cleanInput := strings.ToLower(strings.TrimSpace(rawName))
	cleanInput = strings.ReplaceAll(cleanInput, "dr.", "")
	cleanInput = strings.ReplaceAll(cleanInput, "dr", "")
	cleanInput = strings.ReplaceAll(cleanInput, ",", "")
	cleanInput = strings.ReplaceAll(strings.ReplaceAll(cleanInput, " ", ""), "-", "")

	if cleanInput == "" {
		return ""
	}

	var bestMatchID string
	highestScore := 0.0

	for _, phys := range masterList {
		// Compare against both combinations to capture "First Last" and "Last First" formats
		cleanDBName1 := strings.ToLower(phys.FirstName + phys.LastName)
		cleanDBName2 := strings.ToLower(phys.LastName + phys.FirstName)

		score1 := CalculateSimilarity(cleanInput, cleanDBName1)
		score2 := CalculateSimilarity(cleanInput, cleanDBName2)

		currentMax := score1
		if score2 > score1 {
			currentMax = score2
		}

		// Keep track of the closest matched doctor in the array loop
		if currentMax > highestScore {
			highestScore = currentMax
			bestMatchID = phys.ID
		}
	}

	// 85% Jaro-Winkler Confidence Threshold limit.
	// Anything below this is considered a unique new physician record.
	if highestScore >= 0.95 {
		return bestMatchID
	}

	return ""
}
