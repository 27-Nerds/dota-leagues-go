package delivery

import (
	"context"
	"dota_league/model"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"net/http/httptest"
	"strings"
	"testing"
)

type sourceStub struct{}

func (sourceStub) Inspect(_ context.Context, _ string, id int, version string) (*model.SourceInspection, error) {
	result := &model.SourceInspection{Versions: []model.SourceSnapshot{}}
	if id == 36 && version != "missing" {
		result.Snapshot = &model.SourceSnapshot{Payload: json.RawMessage(`{"team_id":36,"audit_entries":[]}`)}
	}
	return result, nil
}
func (sourceStub) Coverage(context.Context, int, int, string) ([]model.SourceStatus, int64, error) {
	return []model.SourceStatus{}, 0, nil
}
func TestSourceEndpoints(t *testing.T) {
	e := echo.New()
	NewSourceDelivery(e, sourceStub{})
	for _, tc := range []struct {
		path   string
		status int
	}{
		{"/source-data/team/36", 200}, {"/source-data/team/36?download=1", 200},
		{"/source-data/team/36?version=missing&download=1", 404},
		{"/source-data/team/0", 400}, {"/source-data/invalid/36", 400}, {"/source-data/coverage", 200},
	} {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest("GET", tc.path, nil))
		if rec.Code != tc.status {
			t.Fatalf("%s: %d %s", tc.path, rec.Code, rec.Body)
		}
		if tc.status == 200 && strings.Contains(tc.path, "download=1") {
			if !strings.Contains(rec.Header().Get("Content-Disposition"), "attachment") || rec.Body.String() != `{"team_id":36,"audit_entries":[]}` {
				t.Fatal("raw download", rec.Body)
			}
		}
	}
}
