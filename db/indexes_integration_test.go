package db

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	arango "github.com/arangodb/go-driver/v2/arangodb"
)

func TestEnsureIndexesPreservesRecordsAndReusesIndexes(t *testing.T) {
	endpoint := os.Getenv("ARANGO_TEST_URL")
	if endpoint == "" {
		t.Skip("set ARANGO_TEST_URL for an isolated ArangoDB integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	conn, err := Connect(ctx, endpoint, "", "", fmt.Sprintf("index_test_%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := conn.DB.Remove(ctx); err != nil {
			t.Errorf("remove test database: %v", err)
		}
	})
	if err := conn.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	// A maintenance run must leave existing feed summaries untouched.
	if err := conn.Insert(ctx, "update_groups", map[string]any{"_key": "sentinel", "name": "Keep me"}); err != nil {
		t.Fatal(err)
	}
	readIndexes := func() map[string][]arango.IndexResponse {
		t.Helper()
		result := map[string][]arango.IndexResponse{}
		for name, fields := range collectionIndexes {
			col, err := conn.DB.GetCollection(ctx, name, nil)
			if err != nil {
				t.Fatal(err)
			}
			indexes, err := col.Indexes(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(indexes) != len(fields)+1 {
				t.Fatalf("%s: got %d indexes, want %d including primary", name, len(indexes), len(fields)+1)
			}
			result[name] = indexes
		}
		return result
	}
	before := readIndexes()
	if err := conn.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	if after := readIndexes(); !reflect.DeepEqual(before, after) {
		t.Fatal("rerun changed existing indexes")
	}
	var name string
	if _, err := conn.Query(ctx, `RETURN DOCUMENT("update_groups/sentinel").name`, nil, &name); err != nil || name != "Keep me" {
		t.Fatalf("sentinel: %q, %v", name, err)
	}
}
