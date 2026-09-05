// Package pusher sends serialized data to HTTP endpoints.
package pusher

import (
	"bytes"
	"context"
	"encoding/gob"
	"log/slog"
	"net/http"
	"time"
)

type Pusher struct {
}

// SendTo - send given data to the given url
func (p *Pusher) SendTo(ctx context.Context, url string, data any) error {

	byteData, err := p.getBytes(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(byteData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.WarnContext(ctx, "close push response", "error", err)
		}
	}()

	slog.DebugContext(ctx, "push response", "status", resp.StatusCode)
	return nil
}

func (p *Pusher) getBytes(key any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
