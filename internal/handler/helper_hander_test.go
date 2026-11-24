package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"todo_api/internal/apierror"
)

func decodeAPIError(t *testing.T, rec *httptest.ResponseRecorder) apierror.Response {
	t.Helper()

	var testErr apierror.Response
	if err := json.NewDecoder(rec.Body).Decode(&testErr); err != nil {
		t.Fatalf("error for decode body: %v", err)
	}

	if testErr.Code == ""{
		t.Fatalf("error code must not be empty")
	}

	rr := rec.Result()
	contentType := rr.Header.Get("Content-type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("header do not have application/json")
	}

	return testErr
}
