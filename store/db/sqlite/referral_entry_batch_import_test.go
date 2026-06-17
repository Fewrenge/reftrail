package sqlite

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"reftrail/internal/domain"
	v1 "reftrail/server/router/api/v1"

	"github.com/labstack/echo/v5"
)

func TestBatchCreateReferralEntriesHandler_TableDriven(t *testing.T) {
	// 1. Setup isolated memory DB container
	s := setupTestStore(t)
	service := &v1.APIV1Service{Store: s}

	// 2. Define the structural contract for an HTTP handler scenario
	type testCase struct {
		name           string
		userContext    *domain.UserContext
		fileName       string
		fileContent    string
		expectedStatus int
	}

	// 3. Map out your success and error scenarios
	tests := []testCase{
		{
			name:        "Success Path - Process valid TSV import file smoothly",
			userContext: &domain.UserContext{Username: "admin", Role: "REFTRAIL_ADMIN"},
			fileName:    "valid_import.tsv",
			fileContent: "LAST NAME\tFIRST NAME\tHEALTHCARD\tCELL\tEMAIL\tREFERRING PHYSICIAN\tREFERRAL DATE\tCOMPLAINT\tCOMPLAINT SIDE\tCONSULT TYPE\tURGENCY\tTAG\tEMR PATIENT ID\n" +
				"Test\tPatient\t1234567890AB\t(111) 222-3333\ttest.patient@email.com\tTest Test\t2026-01-01\tShoulder;Elbow\tBILATERAL;LEFT\tAPP+UE\tELECTIVE\tSAN\t12506\n" +
				"Second\tTest Patient\t9876543210XY\t(444) 555-6666\t\tTest Test\t2026-02-01\tAnkle\tBILATERAL\tAPP+LE\tASAP\tDAN\t16637",
			expectedStatus: http.StatusCreated, // 201
		},
		{
			name:        "Failure Path - Reject batch completely if database check constraints fail",
			userContext: &domain.UserContext{Username: "admin", Role: "REFTRAIL_ADMIN"},
			fileName:    "bad_urgency.tsv",
			fileContent: "LAST NAME\tFIRST NAME\tHEALTHCARD\tCELL\tEMAIL\tREFERRING PHYSICIAN\tREFERRAL DATE\tCOMPLAINT\tCOMPLAINT SIDE\tCONSULT TYPE\tURGENCY\tTAG\tEMR PATIENT ID\n" +
				"Broken\tUrgency\t1234567890AB\t(111) 222-3333\ttest.patient@email.com\tTest Test\t2026-01-01\tShoulder\tLEFT\tAPP+UE\tCRITICAL\tSAN\t12506", // 'CRITICAL' breaks SQLite CHECK constraint
			expectedStatus: http.StatusInternalServerError, // TODO: add structural validation so that the API returns a 400 Bad Request instead of letting SQLite trigger a 500 Internal Server Error
		},
		// ADD MORE SCENARIOS HERE:
		// - Empty spreadsheet files
	}

	// 4. Loop through every test case dynamically
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange: Package the payload inside a standard MIME Multipart form stream
			bodyBuffer := &bytes.Buffer{}
			writer := multipart.NewWriter(bodyBuffer)

			part, err := writer.CreateFormFile("file", tc.fileName)
			if err != nil {
				t.Fatalf("Failed to create multipart form header segment: %v", err)
			}

			_, err = part.Write([]byte(tc.fileContent))
			if err != nil {
				t.Fatalf("Failed to write mock data into byte pipe stream: %v", err)
			}
			writer.Close()

			// Setup Echo environment routing components
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/referrals/batch", bodyBuffer)
			req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Bind Auth Context state variables cleanly
			ctx := context.Background()
			if tc.userContext != nil {
				ctx = WithUserContext(ctx, tc.userContext)
			}
			c.SetRequest(req.WithContext(ctx))

			// Act: Execute the HTTP endpoint controller block
			err = service.BatchCreateReferralEntriesHandler(c)

			// Echo handlers can return errors or handle them internally and write headers directly.
			// We check both execution styles to avoid missing crashes.
			if err != nil {
				// If your handler returns an Echo HTTP error object, extract its status code
				if echoErr, ok := err.(*echo.HTTPError); ok {
					if echoErr.Code != tc.expectedStatus {
						t.Errorf("Handler returned Echo HTTP Error Code %d, but expected %d. Message: %v", echoErr.Code, tc.expectedStatus, echoErr.Message)
					}
					return // Scenario verified successfully
				}
				t.Fatalf("Handler crashed completely with unexpected standard routing error: %v", err)
			}

			// Assert: Verify recorded response header status match expectations
			if rec.Code != tc.expectedStatus {
				t.Errorf("HTTP Status Mismatch!\nScenario: %s\nExpected: %d\nReceived: %d\nResponse Body: %s",
					tc.name, tc.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
