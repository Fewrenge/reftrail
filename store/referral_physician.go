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

type FindReferralPhysician struct {
	ID             *string `json:"id"`
	CPSONumber     *string `json:"cpsoNumber"`
	FirstName      *string `json:"firstName"`
	LastName       *string `json:"lastName"`
	EMRPhysicianID *string `json:"emrPhysicianId"`
	GeneralSearch  *string `json:"generalSearch"`
}

type DeleteReferralPhysician struct {
	ID string `json:"id"`
}

func (s *Store) CreateReferralPhysician(ctx context.Context, p *ReferralPhysician) (*ReferralPhysician, error) {
	if strings.TrimSpace(p.FirstName) == "" || strings.TrimSpace(p.LastName) == "" {
		return nil, fmt.Errorf("business validation failed: physician first and last names are mandatory fields")
	}
	return s.driver.CreateReferralPhysician(ctx, p)
}

// GetReferralPhysicianByID fetches an isolated individual record by its primary key
func (s *Store) GetReferralPhysicianByID(ctx context.Context, id string) (*ReferralPhysician, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("business validation failed: query target id cannot be blank")
	}
	return s.driver.GetReferralPhysicianByID(ctx, id)
}

// UpdateReferralPhysician handles modification updates down to database driver
func (s *Store) UpdateReferralPhysician(ctx context.Context, p *ReferralPhysician) error {
	if p == nil || strings.TrimSpace(p.ID) == "" {
		return fmt.Errorf("business validation failed: a valid physician structure containing an ID is required for updates")
	}
	return s.driver.UpdateReferralPhysician(ctx, p)
}

func (s *Store) FindReferralPhysicians(ctx context.Context, find *FindReferralPhysician) ([]*ReferralPhysician, error) {
	if find == nil {
		find = &FindReferralPhysician{}
	}
	return s.driver.FindReferralPhysicians(ctx, find)
}

// DeleteReferralPhysician inspects foreign constraint responses to return cleaner domain errors
func (s *Store) DeleteReferralPhysician(ctx context.Context, payload *DeleteReferralPhysician) error {
	if payload == nil || strings.TrimSpace(payload.ID) == "" {
		return fmt.Errorf("business validation failed: valid delete request structure containing an ID required")
	}

	err := s.driver.DeleteReferralPhysician(ctx, payload)
	if err != nil {
		// Formats SQLite standard error strings into clean text messages for your API handlers
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
