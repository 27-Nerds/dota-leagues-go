package delivery

import (
	"bytes"
	"dota_league/model"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestDirectoryJSONSafeEmbedding(t *testing.T) {
	c := echo.New().NewContext(httptest.NewRequest("GET", "/team?country=UA", nil), httptest.NewRecorder())
	name := "</script><script>alert(1)</script>"
	encoded, err := directoryJSON(c, teamDirectory([]model.Team{{ID: 2, Name: name}}), 101)
	if err != nil {
		t.Fatal(err)
	}
	var head bytes.Buffer
	if err := pageHead.Execute(&head, pageContent{DirectoryData: encoded}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(head.String(), name) {
		t.Fatal("unescaped script in HTML")
	}
	var data struct {
		Path    string
		Search  string
		Results []map[string]any
		Meta    struct{ Total int }
	}
	if err := json.Unmarshal([]byte(encoded), &data); err != nil {
		t.Fatal(err)
	}
	if data.Path != "/team" || data.Search != "country=UA" || data.Meta.Total != 101 || data.Results[0]["name"] != name {
		t.Fatalf("bootstrap: %+v", data)
	}
	if _, ok := data.Results[0]["members"]; ok {
		t.Fatal("directory includes roster")
	}
}

func TestDirectoryPagesIncludeInitialResults(t *testing.T) {
	server := pageServer(t, nil)
	for _, path := range []string{"/team?country=UA&offset=100", "/player?team=true&offset=100"} {
		res := getPage(server, path)
		if res.Code != 200 {
			t.Fatalf("%s: status %d", path, res.Code)
		}
		_, after, found := strings.Cut(res.Body.String(), `<script id="directory-data" type="application/json">`)
		if !found {
			t.Fatalf("%s: missing initial results", path)
		}
		encoded, _, _ := strings.Cut(after, "</script>")
		var data struct {
			Results []map[string]any
			Meta    struct{ Total int }
			Search  string
		}
		if err := json.Unmarshal([]byte(encoded), &data); err != nil {
			t.Fatal(err)
		}
		if len(data.Results) != 1 || data.Meta.Total != 101 || data.Search != strings.SplitN(path, "?", 2)[1] {
			t.Fatalf("%s: %+v", path, data)
		}
	}
}
