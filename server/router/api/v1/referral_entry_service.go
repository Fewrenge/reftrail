package v1

import (
	"encoding/csv"
	"io"
	"log/slog"
	"net/http"
	"reftrail/internal/domain"
	"reftrail/store"
	"regexp"
	"strings"

	echo "github.com/labstack/echo/v5"
)

// Get all referrals
func (s *APIV1Service) ListReferralEntriesHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	list, err := s.Store.ListReferralEntries(ctx, &store.FindReferralEntry{})
	if err != nil {
		slog.Error("failed to get referral entries list", "error", err.Error())

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve referral records",
		})
	}
	if list == nil {
		list = []*store.ReferralEntry{}
	}

	return c.JSON(http.StatusOK, list)
}

// GetReferralEntryHandler handles GET /api/v1/referrals/:id
// Focuses on one referral
func (s *APIV1Service) GetReferralEntryHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	// 1. Extract the "id" from the URL path parameter
	idStr := c.Param("id")
	refID := domain.ReferralID(idStr)

	// 2. Ask the Manager (Store) to find this specific entry
	// We use our 'Find' blueprint here
	entry, err := s.Store.GetReferralEntry(ctx, &store.FindReferralEntry{
		ID: &refID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// 3. If no patient was found, return a 404
	if entry == nil {
		return c.JSON(http.StatusNotFound, map[string]any{"message": "Patient entry not found"})
	}

	// 4. Return the patient data as JSON
	return c.JSON(http.StatusOK, entry)
}

func (s *APIV1Service) CreateReferralEntryHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	create := &store.CreateReferralEntry{}

	if err := c.Bind(create); err != nil {
		slog.Warn("Failed to bind malformed JSON request body", "error", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload syntax"})
	}

	if err := domain.ValidateStruct(create); err != nil {
		slog.Warn("Referral payload structural validation failed",
			"patient_last_name", create.PatientLastName,
			"patient_first_name", create.PatientFirstName,
			"error", err.Error(),
		)

		return c.JSON(http.StatusUnprocessableEntity, map[string]string{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	entry, err := s.Store.CreateReferralEntry(ctx, create)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create referral"})
	}
	return c.JSON(http.StatusOK, entry)
}

func (s *APIV1Service) BatchCreateReferralEntriesHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	// Fetch valid system tags ONCE right here before ANY transactions start
	definitions, err := s.Store.ListReferralTags(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load master tags"})
	}

	// Build the in-memory map
	validTagMap := make(map[string]int64)
	for _, def := range definitions {
		if def != nil && def.Name != "" {
			validTagMap[strings.ToUpper(strings.TrimSpace(def.Name))] = def.ID
		}
	}

	// 1. Extract the raw file from the multi-part form data
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No file found in upload form data"})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to open uploaded file stream"})
	}
	defer src.Close()

	// 2. Initialize Go's streaming reader and configure it for tab-separated tokens (TSV)
	reader := csv.NewReader(src)
	reader.LazyQuotes = true

	// Check file extension to switch between comma and tab separation (Is this robust enough?)
	if strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".csv") {
		reader.Comma = ','
	} else {
		reader.Comma = '\t' // Default fallback for .tsv or .txt file paths
	}

	// 3. Parse the Header Row to dynamically map columns to position indices
	headers, err := reader.Read()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to read file spreadsheet headers"})
	}

	headerMap := make(map[string]int)
	for idx, name := range headers {
		cleanHeader := strings.TrimSpace(strings.ToLower(name))

		// Fix hidden character issue: Remove UTF-8 Byte Order Marks (BOM) if exported from Excel
		cleanHeader = strings.TrimPrefix(cleanHeader, "\xef\xbb\xbf")

		headerMap[cleanHeader] = idx
	}

	// Fail-fast verification check for required schema columns
	requiredFields := []string{"last name", "first name", "complaint", "complaint side", "urgency", "referral date"}
	for _, field := range requiredFields {
		if _, exists := headerMap[field]; !exists {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid file format: missing column header '" + field + "'"})
		}
	}

	var batch store.BatchCreateReferralEntries
	var healthCardRegex = regexp.MustCompile(`^(\d{10})([A-Za-z]{2})?$`)

	// 4. Stream rows sequentially (One row = One referral entry)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break // Gracefully reached end of file stream
		}
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "Data row line parsing corruption detected"})
		}

		rawHealthCard := strings.TrimSpace(row[headerMap["health card"]])
		// Strip common user input styling artifacts like spaces or dashes
		cleanedHealthCard := strings.ReplaceAll(strings.ReplaceAll(rawHealthCard, " ", ""), "-", "")

		var healthCardNum string
		var versionCode string

		if healthCardRegex.MatchString(cleanedHealthCard) {
			matches := healthCardRegex.FindStringSubmatch(cleanedHealthCard)
			if len(matches) > 1 {
				healthCardNum = matches[1] // The 10 digits

				// If the optional 2 letters exist, capture and capitalize them
				if len(matches) > 2 && matches[2] != "" {
					versionCode = strings.ToUpper(matches[2])
				}
			}
		}

		// Parse semicolon-separated text values inside matching cells
		rawComplaints := strings.Split(row[headerMap["complaint"]], ";")
		rawSides := strings.Split(row[headerMap["complaint side"]], ";")

		var complaints []store.ReferralComplaint
		for i, part := range rawComplaints {
			cleanPart := strings.TrimSpace(strings.ToUpper(part))
			if cleanPart == "" {
				continue
			}

			// Core zip pattern: resolve matching sides index, fallback to BILATERAL if data length mismatches
			sideVal := "BILATERAL"
			if i < len(rawSides) && strings.TrimSpace(rawSides[i]) != "" {
				sideVal = strings.TrimSpace(strings.ToUpper(rawSides[i]))
			}

			complaints = append(complaints, store.ReferralComplaint{
				BodyPart: cleanPart,
				Side:     sideVal,
				Details:  "",
			})
		}

		// Parse semicolon-separated optional Tags column
		var tags []string
		if tagIdx, exists := headerMap["tag"]; exists {

			rawTags := strings.SplitSeq(row[tagIdx], ";")
			for t := range rawTags {
				cleanTag := strings.TrimSpace(strings.ToUpper(t))
				if cleanTag != "" {
					// Uncomment below to enforce snake_case naming standard
					// cleanTag = strings.ReplaceAll(cleanTag, " ", "_")

					// Build the structure type matching store's expectations
					tags = append(tags, strings.TrimSpace(cleanTag))
				}
			}
		} else {
			// Uses global slog to ensure this warning shows up in production logs too
			slog.Warn("Skipping batch tokenization phase: Column lookup key 'tags' missing from spreadsheet template structure")
		}

		// Map spreadsheet elements into your exact structural schema
		entry := store.CreateReferralEntry{
			PatientLastName:              strings.TrimSpace(row[headerMap["last name"]]),
			PatientFirstName:             strings.TrimSpace(row[headerMap["first name"]]),
			PatientDOB:                   "1990-01-01", // Default placeholder since template column is missing // TODO: Add DOB column
			PatientHealthcardNumber:      healthCardNum,
			PatientHealthcardVersionCode: versionCode,
			ReferringPhysician:           strings.TrimSpace(row[headerMap["referring physician"]]),
			Urgency:                      domain.ReferralUrgency(strings.TrimSpace(row[headerMap["urgency"]])),
			Status:                       domain.ReferralStatus("READY_TO_BOOK"), // Workflow entry state default
			Source:                       domain.ReferralSource("REGULAR"),
			Complaints:                   complaints,
			Tags:                         tags,
		}

		// Run validator on this entry
		if err := domain.ValidateStruct(entry); err != nil {
			slog.Warn("Batch row validation failed",
				"patient", entry.PatientFirstName+" "+entry.PatientLastName,
				"error", err.Error(),
			)

			// Stop execution and tell the user exactly which row broke the batch import
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"error": "Batch import rejected: Duplicate body parts or invalid fields detected for patient " +
					entry.PatientFirstName + " " + entry.PatientLastName + ".",
			})
		}

		batch.ReferralEntries = append(batch.ReferralEntries, entry)
	}

	// 5. Run standard empty set check
	if len(batch.ReferralEntries) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "The uploaded file contains no valid data rows"})
	}

	// 6. Direct transaction call execution into your existing Storage layer engine
	err = s.Store.BatchCreateReferralEntries(c.Request().Context(), &batch)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Transactional database batch error: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Batch file import successful"})
}

func (s *APIV1Service) UpdateReferralEntryHandler(c *echo.Context) error {
	// 1. Get the ID from the URL
	idStr := c.Param("id")
	refID := domain.ReferralID(idStr)

	update := &store.UpdateReferralEntry{ID: refID}
	if err := c.Bind(update); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := s.Store.UpdateReferralEntry(c.Request().Context(), update); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, true)
}

func (s *APIV1Service) UpdateReferralEntryStatusHandler(c *echo.Context) error {
	// 1. Get ID from URL
	idStr := c.Param("id")
	refID := domain.ReferralID(idStr)

	// 2. Bind Request (Only the status)
	var req store.UpdateReferralEntryStatus
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}
	req.ID = refID

	// 3. Update the DB
	// The Store now handles: Transaction, Old Status Check, Role Logic, and Logging
	err := s.Store.UpdateReferralEntryStatus(c.Request().Context(), &req)
	if err != nil {
		// You can check the error type here to return 403 vs 500
		if err.Error() == "illegal status transition" {
			return c.JSON(http.StatusForbidden, err.Error())
		}
		//return c.JSON(http.StatusInternalServerError, "Failed to update status")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Internal Error",
			"debug":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, true)
}

func (s *APIV1Service) DeleteReferralEntryHandler(c *echo.Context) error {
	// 1. Get the ID from the URL
	idStr := c.Param("id")
	refID := domain.ReferralID(idStr)

	// 2. Call the "Janitor" (Store.DeleteReferralEntry)
	// We wrap the ID into the struct your store expects
	err := s.Store.DeleteReferralEntry(c.Request().Context(), &store.DeleteReferralEntry{
		ID: refID,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// 3. Return "No Content" (Status 204) to say "It's gone!"
	return c.NoContent(http.StatusNoContent)
}
