package api

import (
	e "dota_league/error"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

// DownloadImageIfNotExist Download image
func DownloadImageIfNotExist(sourceURL string, relativeDestPath string, imageName string) error {
	op := "api.DownloadImageIfNotExist"
	basePath, err := os.Getwd()
	if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	path := fmt.Sprintf("%s/%s", basePath, relativeDestPath)

	fullPathWithName := fmt.Sprintf("%s/%s", path, imageName)

	_, err = os.Stat(fullPathWithName)
	if os.IsNotExist(err) {
		log.Printf("downloading image, %s", sourceURL)

		resp, err := http.Get(sourceURL)
		if err != nil {
			return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return &e.Error{Code: e.ENOTFOUND, Op: op}
		}

		os.MkdirAll(path, os.ModePerm)
		out, err := os.Create(fullPathWithName)
		if err != nil {
			return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
		}

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
		}

	}

	return nil
}
