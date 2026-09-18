package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubmitCodeReview(t *testing.T) {
	var requestMethod string
	var requestPath string
	var requestQuery map[string][]string
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestMethod = request.Method
		requestPath = request.URL.Path
		requestQuery = request.URL.Query()
		authorization = request.Header.Get("Authorization")
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := SubmitCodeReview(context.Background(), server.URL+"/", 123, "admin", "devops-token")
	if err != nil {
		t.Fatalf("SubmitCodeReview() error = %v", err)
	}

	if requestMethod != http.MethodGet {
		t.Fatalf("request method = %q, want %q", requestMethod, http.MethodGet)
	}
	if requestPath != "/ppm/submitCodeReview" {
		t.Fatalf("request path = %q, want %q", requestPath, "/ppm/submitCodeReview")
	}
	if got := requestQuery["id"]; len(got) != 1 || got[0] != "123" {
		t.Fatalf("query id = %v, want [123]", got)
	}
	if got := requestQuery["createdBy"]; len(got) != 1 || got[0] != "admin" {
		t.Fatalf("query createdBy = %v, want [admin]", got)
	}
	if authorization != "devops-token" {
		t.Fatalf("authorization = %q, want %q", authorization, "devops-token")
	}
}
