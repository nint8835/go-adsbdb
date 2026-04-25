//go:build fixturegen

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	baseURL   = "https://api.adsbdb.com/v0/"
	outputDir = "testdata"
	userAgent = "go-adsbdb fixturegen (github.com/nint8835/go-adsbdb)"
)

type fixture struct {
	file     string
	endpoint string
	query    url.Values
}

func main() {
	fixtures := []fixture{
		{file: "aircraft_c0816e.json", endpoint: "aircraft/C0816E"},
		{
			file:     "aircraft_c0816e_callsign_cjt620.json",
			endpoint: "aircraft/C0816E",
			query:    url.Values{"callsign": []string{"CJT620"}},
		},
		{file: "callsign_cjt620.json", endpoint: "callsign/CJT620"},
		{file: "airline_cjt.json", endpoint: "airline/CJT"},
		{file: "stats.json", endpoint: "stats"},
		{file: "mode_s_a00001.json", endpoint: "mode-s/A00001"},
		{file: "n_number_n1.json", endpoint: "n-number/N1"},
	}

	client := &http.Client{Timeout: 30 * time.Second}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fatal(fmt.Errorf("create fixture output directory %q: %w", outputDir, err))
	}

	for _, fixture := range fixtures {
		body, err := fetch(client, fixture)
		if err != nil {
			fatal(fmt.Errorf("fetch fixture %q: %w", fixture.file, err))
		}
		if err := writeFixture(fixture.file, body); err != nil {
			fatal(fmt.Errorf("write fixture %q: %w", fixture.file, err))
		}
	}
}

func fetch(client *http.Client, fixture fixture) ([]byte, error) {
	u, err := url.Parse(baseURL + fixture.endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse URL for endpoint %q: %w", fixture.endpoint, err)
	}
	u.RawQuery = fixture.query.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request for %q: %w", u.String(), err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request to %q: %w", u.String(), err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body from %q: %w", u.String(), err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"unexpected status from %q: %d: %s",
			u.String(),
			resp.StatusCode,
			bytes.TrimSpace(body),
		)
	}

	return body, nil
}

func writeFixture(name string, body []byte) error {
	var raw any
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}

	formatted, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("format JSON: %w", err)
	}
	formatted = append(formatted, '\n')

	path := filepath.Join(outputDir, name)
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
