package v1

import (
	"encoding/csv"
	"io"
	"log/slog"
	"net/http"
	"reftrail/internal/domain"
	"reftrail/store"
	"regexp"
	"strconv"
	"strings"
	"time"

	echo "github.com/labstack/echo/v5"
)

func nullString(s string) *string {
	cleaned := strings.TrimSpace(s)
	if cleaned == "" {
		return nil
	}
	return &cleaned
}

// Get all referrals that meet a set of criteria
// GET /api/v1/referrals
func (s *APIV1Service) ListReferralEntriesHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	find := &store.FindReferralEntry{}

	// Let Echo automatically extract all query strings and arrays into the find struct
	if err := c.Bind(find); err != nil {
		slog.Warn("Failed parsing list query parameters", "error", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid query filter parameters"})
	}

	if generalTerm := c.QueryParam("generalTerm"); generalTerm != "" {
		trimmed := strings.TrimSpace(generalTerm)

		// Regex patterns to check the characteristics of the universal input string
		hasDigits := regexp.MustCompile(`\d`).MatchString(trimmed)
		hasLetters := regexp.MustCompile(`[a-zA-Z]`).MatchString(trimmed)

		if hasDigits && !hasLetters {
			// Scenario A: Input contains ONLY numbers/hyphens (e.g., Health Card)
			find.PatientHealthcardNumber = &trimmed
		} else {
			// Scenario B: Input contains letters (or mixed text like "John 123") -> Fallback to Names
			find.PatientLastName = &trimmed
			find.PatientFirstName = &trimmed
		}
	}

	// Validate and clean Date Range URL Parameters
	if find.ReferralDateFrom != nil && *find.ReferralDateFrom != "" {
		trimmedFrom := strings.TrimSpace(*find.ReferralDateFrom)
		if _, err := time.Parse("2006-01-02", trimmedFrom); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "referralDateFrom must be in YYYY-MM-DD format"})
		}
		find.ReferralDateFrom = &trimmedFrom
	} else {
		find.ReferralDateFrom = nil // Ensure empty strings don't pass down as pointers
	}

	if find.ReferralDateTo != nil && *find.ReferralDateTo != "" {
		trimmedTo := strings.TrimSpace(*find.ReferralDateTo)
		if _, err := time.Parse("2006-01-02", trimmedTo); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "referralDateTo must be in YYYY-MM-DD format"})
		}
		find.ReferralDateTo = &trimmedTo
	} else {
		find.ReferralDateTo = nil // Ensure empty strings don't pass down as pointers
	}

	if limitQuery := c.QueryParam("limit"); limitQuery != "" {
		if val, err := strconv.Atoi(limitQuery); err == nil {
			find.Limit = &val
		}
	}
	if offsetQuery := c.QueryParam("offset"); offsetQuery != "" {
		if val, err := strconv.Atoi(offsetQuery); err == nil {
			find.Offset = &val
		}
	}

	// Apply business defaults if the frontend didn't pass pagination bounds
	if find.Limit == nil {
		defaultLimit := 10
		find.Limit = &defaultLimit
	}
	if find.Offset == nil {
		defaultOffset := 0
		find.Offset = &defaultOffset
	}

	// Fetch the paginated dataset
	paginated, err := s.Store.ListReferralEntries(ctx, find)
	if err != nil {
		slog.Error("failed to get referral entries list", "error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve records"})
	}

	// 4. Ensure arrays inside the JSON object return as [] instead of null
	if paginated.ReferralEntries == nil {
		paginated.ReferralEntries = []*store.ReferralEntry{}
	}

	return c.JSON(http.StatusOK, paginated)
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
		slog.Warn("Failed to create referral entry in database", "error", err.Error())
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
	validTagMap := make(map[string]string)
	for _, def := range definitions {
		if def != nil && def.Name != "" {
			validTagMap[strings.ToUpper(strings.TrimSpace(def.Name))] = def.Name
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

	cleanRegex := regexp.MustCompile(`[\s\t\_\-]+`)

	// Getting the header map for flexible column order
	headerMap := make(map[string]int)
	for idx, name := range headers {
		cleanHeader := strings.ToLower(name)
		cleanHeader = strings.TrimPrefix(cleanHeader, "\xef\xbb\xbf")
		cleanHeader = cleanRegex.ReplaceAllString(cleanHeader, " ")
		cleanHeader = strings.TrimSpace(cleanHeader)

		headerMap[cleanHeader] = idx
	}

	// Invoke the centralized domain schema validator
	if missingFields := domain.ValidateCSVHeaders(headerMap); len(missingFields) > 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid file format: missing absolutely required column headers: " + strings.Join(missingFields, ", "),
		})
	}

	// Track which fields from the domain schema are actually present in this session
	presentFields := make(map[string]bool)
	for fieldName := range domain.ImportDocumentHeaderSchema {
		if _, exists := headerMap[fieldName]; exists {
			presentFields[fieldName] = true
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
			sideVal := "OTHER"
			if i < len(rawSides) && strings.TrimSpace(rawSides[i]) != "" {
				sideVal = strings.TrimSpace(strings.ToUpper(rawSides[i]))
			}

			complaints = append(complaints, store.ReferralComplaint{
				BodyPart: cleanPart,
				Side:     sideVal,
				Details:  "",
			})
		}

		var emrPatientID string
		if idx, exists := headerMap["emr patient id"]; exists {
			emrPatientID = strings.TrimSpace(row[idx])
		}

		var emrReferralDocID string
		if idx, exists := headerMap["emr referral doc id"]; exists {
			emrReferralDocID = strings.TrimSpace(row[idx])
		}

		// Parse semicolon-separated optional Tags column
		var tags []string
		if tagIdx, exists := headerMap["tag"]; exists {

			// NOTE: strings.SplitSeq handles streaming tokens lazily without upfront allocations.
			// It returns an iter.Seq[string], which strictly allows ONLY ONE loop variable (t).
			rawTags := strings.SplitSeq(row[tagIdx], ";")
			for t := range rawTags {
				// FIX: 't' points directly to the mutable CSV reader memory buffer.
				// Modifying 't' or appending it raw causes a memory escape loop that breaks the file reader.
				// That truncates row execution and dropping subsequent fields (like referral date).
				// strings.Clone(t) safely decouples the string data before any transformations.
				cleanTag := strings.TrimSpace(strings.ToUpper(strings.Clone(t)))
				if cleanTag != "" {
					// Build the structure type matching store's expectations
					tags = append(tags, cleanTag)
				}
			}
		} else {
			// Uses global slog to ensure this warning shows up in production logs too
			slog.Warn("Skipping batch tokenization phase: Column lookup key 'tags' missing from spreadsheet template structure")
		}

		// Map spreadsheet elements into your exact structural schema
		entry := store.CreateReferralEntry{
			PatientLastName:  strings.TrimSpace(row[headerMap["last name"]]),
			PatientFirstName: strings.TrimSpace(row[headerMap["first name"]]),
			PatientDOB:       strings.TrimSpace(row[headerMap["dob"]]),

			// Optional Fields: Wrapped to cleanly handle empty spreadsheet columns
			PatientHealthcardNumber:      nullString(healthCardNum),
			PatientHealthcardVersionCode: nullString(versionCode),
			PatientPhoneNumber:           nullString(row[headerMap["phone number"]]),
			PatientEmail:                 nullString(row[headerMap["email"]]),
			ReferringPhysician:           nullString(row[headerMap["referring physician"]]),
			ConsultTypeDetail:            nullString(row[headerMap["consult type detail"]]),
			EMRPatientID:                 nullString(emrPatientID),
			EMRReferralDocID:             nullString(emrReferralDocID),

			// Required Workflow Structural Fields
			ReferralDate: strings.TrimSpace(row[headerMap["referral date"]]),
			Urgency:      domain.ReferralUrgency(strings.TrimSpace(row[headerMap["urgency"]])),
			Status:       domain.ReferralStatus("READY_TO_BOOK"),
			Source:       domain.ReferralSource("REGULAR"),
			Complaints:   complaints,
			Tags:         tags,
			TriageNote:   strings.TrimSpace(row[headerMap["triage note"]]),
			ConsultType:  domain.ReferralConsultType(strings.TrimSpace(row[headerMap["consult type"]])),
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
	update := &store.UpdateReferralEntry{}

	if err := c.Bind(update); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	idStr := c.Param("id")
	update.ID = domain.ReferralID(idStr)

	if err := s.Store.UpdateReferralEntry(c.Request().Context(), update); err != nil {
		switch err {
		case domain.ErrUnauthorized:
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
		case domain.ErrForbidden:
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Administrative privileges required"})
		case domain.ErrReferralEntryNotFound:
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Referral entry not found"})
		default:
			//return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server processing error when updating referral entry"})
			//---DEBUG---
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
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
