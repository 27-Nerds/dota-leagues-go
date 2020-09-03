package api

import (
	e "dota_league/error"
	"io"
	"net/http"
	"time"
)

func doRequest(url string) (io.ReadCloser, error) {
	op := "api.doRequest"
	httpsClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	req.Header.Set("User-Agent", "Valve/Steam HTTP Client 1.0 (570)")

	res, err := httpsClient.Do(req)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return res.Body, nil
}
