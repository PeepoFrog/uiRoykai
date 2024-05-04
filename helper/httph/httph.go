package httph

import (
	"io"
	"net/http"
)

func MakeHttpRequest(url, method string) ([]byte, error) {
	client := http.DefaultClient
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
