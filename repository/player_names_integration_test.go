package repository

import (
	"context"
	"dota_league/db"
	"dota_league/model"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestStoredPlayerNames(t *testing.T) {
	url := os.Getenv("ARANGO_TEST_URL")
	if url == "" {
		t.Skip("set ARANGO_TEST_URL for ArangoDB integration")
	}
	conn, err := db.Connect(t.Context(), url, "", "", fmt.Sprintf("player_names_test_%d", time.Now().UnixNano()))
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
	repo := NewPlayerRepository(conn)
	if err := repo.StoreAll(t.Context(), []model.Player{{ID: 1, Name: "Player One"}, {ID: 2, Name: ""}, {ID: 3, Name: "Not requested"}}); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetNames(t.Context(), []int{1, 2, 999})
	if err != nil || !reflect.DeepEqual(got, map[int]string{1: "Player One", 2: ""}) {
		t.Fatalf("names=%v err=%v", got, err)
	}
	got, err = repo.GetNames(t.Context(), nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty lookup=%v err=%v", got, err)
	}
}
