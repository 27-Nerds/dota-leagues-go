package api

import (
	"io"
	"net/http"
	"time"
)

func doRequest(url string) (io.ReadCloser, error) {
	httpsClient := http.Client{
		Timeout: time.Second * 5, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Valve/Steam HTTP Client 1.0 (570)")

	res, err := httpsClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}
