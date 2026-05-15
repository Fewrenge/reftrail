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
	"reftrail/store"

	"github.com/labstack/echo/v5"
)

func TestBatchCreateReferralEntriesHandler_Success(t *testing.T) {
	// 1. Setup isolated memory DB using your helper function
	s := setupTestStore(t)

	// 2. Build the domain context payload mirroring your tag integration test
	ctx := WithUserContext(context.Background(), &domain.UserContext{
		ID: 1, Role: "REFTRAIL_ADMIN",
	})

	// 3. Optional/Pre-seed Step:
	// If you want to confirm the store execution functions flawlessly before testing file streams:
	_, err := s.CreateReferralEntry(ctx, &store.CreateReferralEntry{
		PatientLastName:    "Preseed",
		PatientFirstName:   "Check",
		Urgency:            "Elective",
		ReferringPhysician: "System Init",
		Complaints: []store.ReferralComplaint{
			{BodyPart: "KNEE", Side: "LEFT"},
		},
	})
	if err != nil {
		t.Fatalf("Failed to initialize baseline database state: %v", err)
	}

	// 4. Prepare your raw TSV template document spreadsheet content lines
	tsvTemplateDraft := "LAST NAME\tFIRST NAME\tHEALTHCARD\tCELL\tEMAIL\tREFERRING PHYSICIAN\tREFERRAL DATE\tCOMPLAINT\tCOMPLAINT SIDE\tCONSULT TYPE\tURGENCY\tTAG\tJUVONNO PATIENT ID\n" +
		"Test\tPatient\t1234567890AB\t(111) 222-3333\ttest.patient@email.com\tTest Test\t2026-01-01\tShoulder;Elbow\tBILATERAL;LEFT\tAPP+UE\tElective\tSAN\t12506\n" +
		"Second\tTest Patient\t9876543210XY\t(444) 555-6666\t\tTest Test\t2026-02-01\tAnkle\tBILATERAL\tAPP+LE\tASAP\tDAN\t16637"

	// 5. Package the payload stream inside standard MIME Multipart layout blocks
	bodyBuffer := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuffer)

	part, err := writer.CreateFormFile("file", "import_template.tsv")
	if err != nil {
		t.Fatalf("Failed to create multipart form header segment: %v", err)
	}

	_, err = part.Write([]byte(tsvTemplateDraft))
	if err != nil {
		t.Fatalf("Failed to write mock data into byte pipe stream: %v", err)
	}
	writer.Close()

	// 6. Spin up Echo environment context mocks
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/referrals/batch", bodyBuffer)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Attach your pre-built user context onto the target HTTP engine routing pipeline
	c.SetRequest(req.WithContext(ctx))

	// 7. Inject your concrete SQLite test store into your handler initialization block
	service := &v1.APIV1Service{
		Store: s, // Compiles cleanly because type matches *store.Store exactly
	}

	// 8. Execute batch parsing pipeline handler logic
	err = service.BatchCreateReferralEntriesHandler(c)
	if err != nil {
		t.Fatalf("Handler returned an unexpected controller crash error: %v", err)
	}

	// 9. Assert correct structural creation codes
	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201 Created, but received: %d. Body: %s", rec.Code, rec.Body.String())
	}
}
