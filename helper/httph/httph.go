package httph

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/PeepoFrog/km2UI/types"
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

func GetInterxStatus(nodeIP string) (*types.Info, error) {
	url := fmt.Sprintf("http://%v:11000/api/status", nodeIP)
	b, err := MakeHttpRequest(url, "GET")
	if err != nil {
		return nil, err
	}
	var info types.Info
	err = json.Unmarshal(b, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}
