package api

import (
	"context"
	e "dota_league/error"
	"encoding/json"
	"io"
	"log/slog"
)

type SourceRecorder interface {
	Record(context.Context, string, int, json.RawMessage, string) error
}

// Configure once before starting workers. Captures the response from the same
// request used for domain imports, never a second API request.
var sourceRecorder SourceRecorder

func SetSourceRecorder(recorder SourceRecorder) { sourceRecorder = recorder }

func loadSource(ctx context.Context, kind string, id int, url string, target any) error {
	body, err := doRequest(ctx, url)
	var raw []byte
	failure := ""
	if err != nil {
		failure = "request_failed"
	} else {
		defer closeResponse(body)
		raw, err = io.ReadAll(body)
		if err != nil {
			failure = "read_failed"
		} else {
			var identity struct {
				TeamID int `json:"team_id"`
				Info   struct {
					LeagueID int `json:"league_id"`
				} `json:"info"`
			}
			err = json.Unmarshal(raw, &identity)
			if err != nil {
				failure = "invalid_json"
			} else {
				actual := identity.TeamID
				if kind == "league" {
					actual = identity.Info.LeagueID
				}
				if actual <= 0 || actual != id {
					failure = "invalid_entity"
					err = &e.Error{Code: e.ENOTFOUND, Message: "Source returned no matching entity"}
				}
			}
		}
	}
	if sourceRecorder != nil && ctx.Err() == nil {
		if recordErr := sourceRecorder.Record(ctx, kind, id, raw, failure); recordErr != nil {
			slog.ErrorContext(ctx, "save source snapshot", "kind", kind, "entity_id", id, "error", recordErr)
		}
	}
	if err != nil {
		return err
	}
	// Domain decoding may fail after a schema change; the valid source response
	// has already been archived for investigation.
	return json.Unmarshal(raw, target)
}
