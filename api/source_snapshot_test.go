package api

import (
	"context"
	"dota_league/model"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type recordedSource struct {
	raw     json.RawMessage
	failure string
	calls   int
}

func (r *recordedSource) Record(_ context.Context, _ string, _ int, raw json.RawMessage, failure string) error {
	r.raw = append([]byte(nil), raw...)
	r.failure = failure
	r.calls++
	return nil
}
func TestSourceCaptureUsesValidatedResponse(t *testing.T) {
	original := sourceRecorder
	t.Cleanup(func() { sourceRecorder = original })
	r := &recordedSource{}
	SetSourceRecorder(r)
	calls := 0
	payload := `{"team_id":36,"unknown":{"nested":true},"audit_entries":[]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(payload))
	}))
	defer server.Close()
	var team model.Team
	if err := loadSource(t.Context(), "team", 36, server.URL, &team); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || r.calls != 1 || !json.Valid(r.raw) || r.failure != "" {
		t.Fatalf("capture calls %d/%d failure %s", calls, r.calls, r.failure)
	}
	payload = `{"team_id":36,"wins":"changed-type","audit_entries":[]}`
	if err := loadSource(t.Context(), "team", 36, server.URL, &model.Team{}); err == nil || r.failure != "" || string(r.raw) != payload {
		t.Fatal("schema change was not preserved")
	}
	if err := loadSource(t.Context(), "team", 99, server.URL, &model.Team{}); err == nil || r.failure != "invalid_entity" {
		t.Fatal("invalid identity accepted")
	}
}
