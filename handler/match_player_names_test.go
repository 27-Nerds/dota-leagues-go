package handler

import (
	"context"
	"dota_league/model"
	"errors"
	"reflect"
	"testing"
)

type matchNames struct {
	ids   []int
	names map[int]string
	err   error
	calls int
}

func (n *matchNames) GetNames(_ context.Context, ids []int) (map[int]string, error) {
	n.ids = ids
	n.calls++
	return n.names, n.err
}

func TestMatchNamesOnCachedAndFreshMatches(t *testing.T) {
	for _, cached := range []bool{true, false} {
		t.Run(map[bool]string{true: "cached", false: "fresh"}[cached], func(t *testing.T) {
			original := &model.MatchMinimal{MatchID: "123", Players: []model.MatchMinimalPlayer{
				{AccountID: 1, ProName: "Match name"}, {AccountID: 2}, {AccountID: 3, ProName: "  "}, {AccountID: 4}, {AccountID: 0}, {AccountID: 2},
			}}
			repo := &mockMatchRepo{}
			if cached {
				repo.cached = original
			}
			names := &matchNames{names: map[int]string{1: "New profile name", 2: " Resolved ", 3: " ", 0: "Anonymous"}}
			h := NewMatchMinimalHandler(repo, names, func(context.Context, int, string) (*model.MatchMinimal, error) { return original, nil })
			result, err := h.Get(t.Context(), 1, "123")
			if err != nil {
				t.Fatal(err)
			}
			if names.calls != 1 || !reflect.DeepEqual(names.ids, []int{2, 3, 4}) {
				t.Fatalf("expected one deduplicated lookup, got %+v", names)
			}
			want := []string{"Match name", "Resolved", "  ", "", "", "Resolved"}
			for i, p := range result.Players {
				if p.ProName != want[i] {
					t.Errorf("player %d: %q", i, p.ProName)
				}
			}
			if original.Players[1].ProName != "" {
				t.Fatal("mutated source or cached names")
			}
			if !cached && repo.stored != original {
				t.Fatal("raw match must still be stored")
			}
			names.names[2] = "Updated profile"
			result, err = h.Get(t.Context(), 1, "123")
			if err != nil || result.Players[1].ProName != "Updated profile" {
				t.Fatal("profile updates not reflected")
			}
		})
	}
}
func TestMatchNameLookupFailureKeepsScoreboard(t *testing.T) {
	match := &model.MatchMinimal{Players: []model.MatchMinimalPlayer{{AccountID: 1}}}
	names := &matchNames{err: errors.New("database unavailable")}
	h := NewMatchMinimalHandler(&mockMatchRepo{cached: match}, names, nil)
	got, err := h.Get(t.Context(), 1, "123")
	if err != nil || got != match {
		t.Fatal("optional names must not fail the match")
	}
	match.Players[0].ProName = "Known"
	_, err = h.Get(t.Context(), 1, "123")
	if err != nil || names.calls != 1 {
		t.Fatal("looked up already named players")
	}
}
