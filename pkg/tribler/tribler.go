package tribler

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Download struct {
	AllTimeUpload    float64    `json:"all_time_upload"`
	Hops             int        `json:"hops"`
	Files            string     `json:"files"`
	Destination      string     `json:"destination"`
	TotalPieces      int        `json:"total_pieces"`
	AllTimeRatio     float64    `json:"all_time_ratio"`
	StatusCode       int        `json:"status_code"`
	NumPeers         int        `json:"num_peers"`
	SpeedDown        int        `json:"speed_down"`
	TimeAdded        int        `json:"time_added"`
	Size             int        `json:"size"`
	AllTimeDownload  float64    `json:"all_time_download"`
	Availability     int        `json:"availability"`
	SafeSeeding      bool       `json:"safe_seeding"`
	Name             string     `json:"name"`
	MaxDownloadSpeed int        `json:"max_download_speed"`
	Status           string     `json:"status"`
	Infohash         string     `json:"infohash"`
	MaxUploadSpeed   int        `json:"max_upload_speed"`
	Peers            string     `json:"peers"`
	Trackers         []Trackers `json:"trackers"`
	AnonDownload     bool       `json:"anon_download"`
	Error            string     `json:"error"`
	NumSeeds         int        `json:"num_seeds"`
	Progress         float64    `json:"progress"`
	Eta              float64    `json:"eta"`
	SpeedUp          int        `json:"speed_up"`
}

type Trackers struct {
	Url    string `json:"url"`
	Peers  int    `json:"peers"`
	Status string `json:"status"`
}

type Checkpoints struct {
	Loaded    int  `json:"loaded"`
	AllLoaded bool `json:"all_loaded"`
	Total     int  `json:"total"`
}

type TorrentFiles struct {
	Infohash string  `json:"infohash"`
	Files    []Files `json:"files"`
}

type Files struct {
	Index    int     `json:"index"`
	Name     string  `json:"name"`
	Size     int     `json:"size"`
	Included bool    `json:"included"`
	Progress float64 `json:"progress"`
}

type DownloadsResponse struct {
	Downloads   []Download  `json:"downloads"`
	Checkpoints Checkpoints `json:"checkpoints"`
}

const (
	apiKeyHeader           = "X-Api-Key"
	triblerAPIEndpointEnv  = "TRIBLER_API_ENDPOINT"
	triblerAPIKeyEnv       = "TRIBLER_API_KEY"
	triblerDownloadDirEnv  = "TRIBLER_DOWNLOAD_DIR"
	tlsSkipVerifyEnv       = "TLS_SKIP_VERIFY"
	defaultDownloadTimeout = 5 * time.Second
)

func newHTTPClient() (*http.Client, error) {
	tlsConfig, err := getTLSConfig()
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout:   defaultDownloadTimeout,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}, nil
}

func getTLSConfig() (*tls.Config, error) {
	skipVerify := false
	if v := os.Getenv(tlsSkipVerifyEnv); v != "" {
		skipVerify, _ = strconv.ParseBool(v)
	}

	if skipVerify {
		return &tls.Config{InsecureSkipVerify: true}, nil
	}

	return nil, nil
}

func newDownloadRequest(method, path string, body interface{}) (*http.Request, error) {
	apiEndpoint := os.Getenv(triblerAPIEndpointEnv)
	if apiEndpoint == "" {
		return nil, errors.New("TRIBLER_API_ENDPOINT environment variable is not set")
	}

	apiKey := os.Getenv(triblerAPIKeyEnv)
	if apiKey == "" {
		return nil, errors.New("TRIBLER_API_KEY environment variable is not set")
	}

	u, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, err
	}
	u.Path = path

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set(apiKeyHeader, apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func executeDownloadRequest(client *http.Client, req *http.Request) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(strings.TrimSpace(resp.Status))
	}

	return io.ReadAll(resp.Body)
}

func GetDownloads() (DownloadsResponse, error) {
	client, err := newHTTPClient()
	if err != nil {
		return DownloadsResponse{}, err
	}

	req, err := newDownloadRequest("GET", "/downloads", nil)
	if err != nil {
		return DownloadsResponse{}, err
	}

	body, err := executeDownloadRequest(client, req)
	if err != nil {
		return DownloadsResponse{}, err
	}

	var dr DownloadsResponse
	if err := json.Unmarshal(body, &dr); err != nil {
		return DownloadsResponse{}, err
	}

	return dr, nil
}

func GetDownload(hash string) (Download, error) {
	client, err := newHTTPClient()
	if err != nil {
		return Download{}, err
	}

	req, err := newDownloadRequest("GET", "/downloads", map[string]string{"hash": hash})
	if err != nil {
		return Download{}, err
	}

	body, err := executeDownloadRequest(client, req)
	if err != nil {
		return Download{}, err
	}

	var dr DownloadsResponse
	if err := json.Unmarshal(body, &dr); err != nil {
		return Download{}, err
	}

	if len(dr.Downloads) == 0 {
		return Download{}, errors.New("download not found")
	}

	return dr.Downloads[0], nil
}

func AddDownload(uri string) error {
	client, err := newHTTPClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"anon_hops":    2,
		"safe_seeding": true,
		"uri":          uri,
		"destination":  os.Getenv(triblerDownloadDirEnv),
	}

	req, err := newDownloadRequest("PUT", "/downloads", body)
	if err != nil {
		return err
	}

	_, err = executeDownloadRequest(client, req)
	return err
}

func GetDownloadsFiles(hash string) (TorrentFiles, error) {
	client, err := newHTTPClient()
	if err != nil {
		return TorrentFiles{}, err
	}

	req, err := newDownloadRequest("GET", "/downloads/"+hash+"/files", nil)
	if err != nil {
		return TorrentFiles{}, err
	}

	body, err := executeDownloadRequest(client, req)
	if err != nil {
		return TorrentFiles{}, err
	}

	var tf TorrentFiles
	if err := json.Unmarshal(body, &tf); err != nil {
		return TorrentFiles{}, err
	}

	return tf, nil
}

func DeleteDownload(hash string, removeData bool) error {
	client, err := newHTTPClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"remove_data": removeData,
	}

	req, err := newDownloadRequest("DELETE", "/downloads/"+hash, body)
	if err != nil {
		return err
	}

	_, err = executeDownloadRequest(client, req)
	return err
}

func UpdateDownload(hash string, state string) error {
	client, err := newHTTPClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"state": state,
	}

	req, err := newDownloadRequest("PATCH", "/downloads/"+hash, body)
	if err != nil {
		return err
	}

	_, err = executeDownloadRequest(client, req)
	return err
}
