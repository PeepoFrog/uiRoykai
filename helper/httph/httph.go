package httph

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"

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
	log.Printf("interx status is okay")
	return &info, nil
}

func ValidatePortRange(portStr string) bool {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return false // Not an integer
	}
	if port < 1 || port > 65535 {
		return false // Out of valid port range
	}
	return true
}

func ValidateIP(input string) bool {
	ipCheck := net.ParseIP(input)
	return ipCheck != nil
}
