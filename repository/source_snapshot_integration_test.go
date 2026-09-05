package repository

import (
	"context"
	"dota_league/db"
	"dota_league/model"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSourceSnapshots(t *testing.T) {
	url := os.Getenv("ARANGO_TEST_URL")
	if url == "" {
		t.Skip("set ARANGO_TEST_URL")
	}
	conn, err := db.Connect(t.Context(), url, "", "", fmt.Sprintf("sources_test_%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := conn.DB.Remove(ctx); err != nil {
			t.Error(err)
		}
	})
	if err := conn.EnsureSourceCollections(t.Context()); err != nil {
		t.Fatal(err)
	}
	r := NewSourceRepository(conn)
	empty, err := r.Inspect(t.Context(), "team", 36, "")
	if err != nil || empty.Snapshot != nil {
		t.Fatalf("empty: %+v %v", empty, err)
	}
	a := json.RawMessage(`{"team_id":36,"name":"A","audit_entries":[{"audit_action":5,"account_id":123,"timestamp":100}],"unknown":{"precise":76561198000000001}}`)
	same := json.RawMessage(`{"unknown":{"precise":76561198000000001},"audit_entries":[{"timestamp":100,"account_id":123,"audit_action":5}],"name":"A","team_id":36}`)
	for _, raw := range []json.RawMessage{a, same} {
		if err := r.Record(t.Context(), "team", 36, raw, ""); err != nil {
			t.Fatal(err)
		}
	}
	one, err := r.Inspect(t.Context(), "team", 36, "")
	if err != nil || len(one.Versions) != 1 || one.Status.AuditCount != 1 {
		t.Fatalf("dedup %+v %v", one, err)
	}
	original := one.Snapshot.Key
	if err := NewPlayerRepository(conn).Store(t.Context(), &model.Player{ID: 123, Name: "Known player"}); err != nil {
		t.Fatal(err)
	}
	named, err := r.Inspect(t.Context(), "team", 36, "")
	if err != nil || named.PlayerNames["123"] != "Known player" {
		t.Fatal("player names", err)
	}
	if string(named.Snapshot.Payload) != string(one.Snapshot.Payload) {
		t.Fatal("enrichment modified raw JSON")
	}
	if err := r.Record(t.Context(), "team", 36, nil, "request_failed"); err != nil {
		t.Fatal(err)
	}
	failed, err := r.Inspect(t.Context(), "team", 36, "")
	if err != nil || failed.Snapshot.Key != original || failed.Status.LastSuccess != one.Status.LastSuccess || failed.Status.Failure != "request_failed" {
		t.Fatalf("failure %+v %v", failed, err)
	}
	for _, raw := range []json.RawMessage{json.RawMessage(`{"team_id":36,"name":"B"}`), a} {
		if err := r.Record(t.Context(), "team", 36, raw, ""); err != nil {
			t.Fatal(err)
		}
	}
	back, err := r.Inspect(t.Context(), "team", 36, "")
	if err != nil || len(back.Versions) != 3 || back.Status.Versions != 3 || back.Snapshot.Key == original || back.Status.Failure != "" {
		t.Fatalf("A-B-A %+v %v", back, err)
	}
	historical, err := r.Inspect(t.Context(), "team", 36, original)
	if err != nil || historical.Snapshot.Key != original {
		t.Fatal("historical lookup", err)
	}
	// Coverage includes every attempted source, with no cohort requirement.
	for id := 1; id <= 45; id++ {
		payload := json.RawMessage(fmt.Sprintf(`{"team_id":%d,"name":"Coverage team %d"}`, id, id))
		if err := r.Record(t.Context(), "team", id, payload, ""); err != nil {
			t.Fatal(err)
		}
	}
	rows, total, err := r.Coverage(t.Context(), 0, 20, "")
	if err != nil || len(rows) != 20 || total != 45 {
		t.Fatalf("coverage page %d/%d %v", len(rows), total, err)
	}
	second, count, err := r.Coverage(t.Context(), 20, 20, "")
	if err != nil || len(second) != 20 || count != 45 {
		t.Fatal("second page", err)
	}
	seen := map[int]bool{}
	for _, r := range rows {
		seen[r.EntityID] = true
	}
	for _, r := range second {
		if seen[r.EntityID] {
			t.Fatal("pagination overlap")
		}
	}
	found, count, err := r.Coverage(t.Context(), 0, 20, " COVERAGE TEAM 45 ")
	if err != nil || count != 1 || len(found) != 1 || found[0].EntityID != 45 {
		t.Fatal("search outside first page", err)
	}
	if err := r.Record(t.Context(), "team", 99, nil, "request_failed"); err != nil {
		t.Fatal(err)
	}
	_, count, err = r.Coverage(t.Context(), 0, 20, "99")
	if err != nil || count != 1 {
		t.Fatal("failed attempts hidden", err)
	}
	if err := r.Record(t.Context(), "team", 37, json.RawMessage(`{"team_id":37,"audit_entries":{"changed":true}}`), ""); err != nil {
		t.Fatal("schema evolution lost", err)
	}
	wrong, err := r.Inspect(t.Context(), "team", 99, original)
	if err != nil || wrong.Snapshot != nil {
		t.Fatal("cross-entity lookup", err)
	}
}
