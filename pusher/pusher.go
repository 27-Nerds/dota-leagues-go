// Package pusher sends serialized data to HTTP endpoints.
package pusher

import (
	"bytes"
	"encoding/gob"
	"log"
	"net/http"
)

type Pusher struct {
}

// SendTo - send given data to the given url
func (p *Pusher) SendTo(url string, data *interface{}) error {

	byteData, err := p.getBytes(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byteData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("close push response: %v", err)
		}
	}()

	log.Println("response Status:", resp.Status)
	return nil
}

func (p *Pusher) getBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
