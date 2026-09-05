package repository

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"dota_league/db"
	"dota_league/model"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/arangodb/go-driver/v2/arangodb/shared"
	"strings"
	"sync"
	"time"
)

type SourceRepository struct {
	Conn Database
	mu   sync.Mutex
}

func NewSourceRepository(conn Database) *SourceRepository { return &SourceRepository{Conn: conn} }
func sourceKey(kind string, id int) string                { return fmt.Sprintf("%s_%d", kind, id) }
func canonicalSource(raw json.RawMessage) ([]byte, error) {
	var value any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	return json.Marshal(value)
}
func (r *SourceRepository) Record(ctx context.Context, kind string, id int, raw json.RawMessage, failure string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	now := time.Now().UnixMilli()
	key := sourceKey(kind, id)
	if failure != "" {
		return r.Conn.DoQueryBuilder(ctx, `UPSERT {_key:@key} INSERT {_key:@key,kind:@kind,entity_id:@id,last_attempt:@now,failure:@failure} UPDATE {last_attempt:@now,failure:@failure} IN source_status`, map[string]any{"key": key, "kind": kind, "id": id, "now": now, "failure": failure})
	}
	canonical, err := canonicalSource(raw)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(canonical)
	hash := hex.EncodeToString(sum[:])
	// Derive only coverage metadata; the full JSON remains available unchanged in structure.
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	var data struct {
		Name          string
		Audit, Groups []json.RawMessage
	}
	// Unexpected metadata types must not prevent archiving a valid source.
	_ = json.Unmarshal(fields["name"], &data.Name)
	_ = json.Unmarshal(fields["audit_entries"], &data.Audit)
	_ = json.Unmarshal(fields["node_groups"], &data.Groups)
	if kind == "league" {
		var info map[string]json.RawMessage
		_ = json.Unmarshal(fields["info"], &info)
		_ = json.Unmarshal(info["name"], &data.Name)
	}
	// One AQL query atomically writes snapshot + collection status. Only the latest
	// hash is compared: returning to an earlier payload creates a new version.
	return r.Conn.DoQueryBuilder(ctx, `LET latest = FIRST(FOR s IN source_snapshots FILTER s.kind == @kind && s.entity_id == @id SORT s.first_seen DESC, s._key DESC LIMIT 1 RETURN s)
LET same = latest != null && latest.hash == @hash
LET snapshot = same ? MERGE(latest,{last_seen:@now}) : {_key:@version,kind:@kind,entity_id:@id,hash:@hash,first_seen:MAX([@now,TO_NUMBER(latest.first_seen)+1]),last_seen:@now,payload:@payload}
UPSERT {_key:snapshot._key} INSERT snapshot UPDATE {last_seen:@now} IN source_snapshots
LET saved = NEW
UPSERT {_key:@key} INSERT {_key:@key,kind:@kind,entity_id:@id,name:@name,last_attempt:@now,last_success:@now,failure:"",versions:1,audit_count:@audits,group_count:@groups}
UPDATE {name:@name,last_attempt:@now,last_success:@now,failure:"",versions:TO_NUMBER(OLD.versions)+(same ? 0 : 1),audit_count:@audits,group_count:@groups} IN source_status
RETURN saved._key`, map[string]any{"key": key, "kind": kind, "id": id, "now": now, "hash": hash, "payload": raw, "version": rand.Text(), "name": data.Name, "audits": len(data.Audit), "groups": len(data.Groups)})
}
func (r *SourceRepository) Inspect(ctx context.Context, kind string, id int, version string) (*model.SourceInspection, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	var result model.SourceInspection
	_, err := r.Conn.Query(ctx, `LET status=DOCUMENT("source_status",@key)
LET versions=(FOR s IN source_snapshots FILTER s.kind==@kind && s.entity_id==@id SORT s.first_seen DESC, s._key DESC LIMIT 100 RETURN UNSET(s,"payload","_id","_rev"))
LET selected=@version=="" ? FIRST(versions)._key : @version
LET snapshot=DOCUMENT("source_snapshots",selected)
LET valid=snapshot.kind==@kind && snapshot.entity_id==@id
LET names=MERGE(FOR a IN (valid && IS_ARRAY(snapshot.payload.audit_entries) ? snapshot.payload.audit_entries : [])
 LET player=DOCUMENT("players",TO_STRING(a.account_id))
 FILTER player.name != null && player.name != ""
 RETURN {[TO_STRING(a.account_id)]:player.name})
RETURN {status,versions,player_names:names,snapshot:valid ? snapshot : null}`, map[string]any{"key": sourceKey(kind, id), "kind": kind, "id": id, "version": version}, &result)
	return &result, err
}

// Coverage lists all attempted source records, independently of the old pilot flags.
func (r *SourceRepository) Coverage(ctx context.Context, offset, limit int, search string) ([]model.SourceStatus, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	rows := []model.SourceStatus{}
	cursor, err := r.Conn.QueryAll(ctx, `FOR s IN source_status
 FILTER s.last_attempt > 0
 FILTER @search=="" || CONTAINS(LOWER(TO_STRING(s.name)),@search) || TO_STRING(s.entity_id)==@search
 SORT s.last_attempt DESC,s.kind ASC,s.entity_id ASC
 LIMIT @offset,@limit RETURN s`, map[string]any{"offset": offset, "limit": limit, "search": strings.ToLower(strings.TrimSpace(search))}, true)
	if err != nil {
		return nil, 0, err
	}
	defer db.CloseCursor(cursor)
	for {
		var row model.SourceStatus
		_, err := cursor.ReadDocument(ctx, &row)
		if shared.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		rows = append(rows, row)
	}
	return rows, int64(cursor.Statistics().FullCountInt), nil
}
