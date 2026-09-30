package database

import (
	"bytes"
	"fmt"
	"net/http"
)

const (
	clickHouseURL  = "http://localhost:8123"
	clickHouseUser = "ulpf_backend"
	clickHousePass = "ulpf_dev"
	clickHouseDB   = "ulpf"
)

func Query(query string) ([]byte, error) {
	url := fmt.Sprintf(
		"%s/?database=%s",
		clickHouseURL,
		clickHouseDB,
	)

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBufferString(query),
	)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(clickHouseUser, clickHousePass)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"ClickHouse returned status %s",
			resp.Status,
		)
	}

	var result []byte

	buffer := make([]byte, 4096)

	for {
		n, err := resp.Body.Read(buffer)

		if n > 0 {
			result = append(result, buffer[:n]...)
		}

		if err != nil {
			break
		}
	}

	return result, nil
}