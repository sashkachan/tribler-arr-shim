package tribler

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// Create net/http client
var client = &http.Client{
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	},
}

// This package requests tribler api instance and returns the response.

// Tribler torrent struct
/*
{
  "downloads": {
    "all_time_upload": 0,
    "hops": 0,
    "files": "string",
    "destination": "string",
    "total_pieces": 0,
    "all_time_ratio": 0,
    "status_code": 0,
    "num_peers": 0,
    "speed_down": 0,
    "time_added": 0,
    "size": 0,
    "all_time_download": 0,
    "availability": 0,
    "safe_seeding": true,
    "name": "string",
    "max_download_speed": 0,
    "status": "string",
    "infohash": "string",
    "max_upload_speed": 0,
    "peers": "string",
    "trackers": "string",
    "anon_download": true,
    "error": "string",
    "num_seeds": 0,
    "progress": 0,
    "eta": 0,
    "speed_up": 0
  },
  "checkpoints": {
    "loaded": 0,
    "all_loaded": true,
    "total": 0
  }
}
*/
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

// /downloads response
type DownloadsResponse struct {
	Downloads   []Download  `json:"downloads"`
	Checkpoints Checkpoints `json:"checkpoints"`
}

// All methods are requested with TRIBLER_API_ENDPOINT environment variable.
// All requests include X-Api-Key header that is set to TRIBLER_API_KEY environment variable.
// GetDownloads retrieves torrents and returns them as a slice of Downloads structs
// URI: /downloads
// Method: GET
// Request: None
// Response: Downloads
func GetDownloads() DownloadsResponse {

	req, err := http.NewRequest("GET", os.Getenv("TRIBLER_API_ENDPOINT")+"/downloads", nil)
	if err != nil {
		log.Print(err)
		return DownloadsResponse{}
	}
	// set X-Api-Key header
	req.Header.Set("X-Api-Key", os.Getenv("TRIBLER_API_KEY"))

	// send request
	response, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return DownloadsResponse{}
	}
	// pretty print response in the log
	log.Printf("%+v", response.Body)
	var downloadsResponse DownloadsResponse
	// pretty print response in the log
	err = json.NewDecoder(response.Body).Decode(&downloadsResponse)
	if err != nil {
		log.Print(err)
		return DownloadsResponse{}
	}

	return downloadsResponse
}

// GetDownload retrieves a torrent
// URI: /downloads
// Method: GET
// Params: hash=string
// Request: None
// Response: Downloads
func GetDownload(hash string) Download {

	req, err := http.NewRequest("GET", os.Getenv("TRIBLER_API_ENDPOINT")+"/downloads?hash="+hash, nil)
	if err != nil {
		log.Print(err)
		return Download{}
	}
	// set X-Api-Key header
	req.Header.Set("X-Api-Key", os.Getenv("TRIBLER_API_KEY"))

	// send request
	response, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return Download{}
	}
	// pretty print response in the log
	var downloadResponse DownloadsResponse
	err = json.NewDecoder(response.Body).Decode(&downloadResponse)
	if err != nil {
		log.Print(err)
		return Download{}
	}
	if len(downloadResponse.Downloads) == 0 {
		return Download{}
	}

	return downloadResponse.Downloads[0]
}

// AddDownload adds a new torrent to tribler
// URI: /downloads
// Method: PUT
//
//	Body: {
//	  "anon_hops": 2,
//	  "safe_seeding": true,
//	  "uri": "
//	  "destination": "/downloads/prowlarr"
//	}
func AddDownload(uri string) {

	// create request body
	body := map[string]interface{}{
		"anon_hops":    2,
		"safe_seeding": true,
		"uri":          uri,
		"destination":  os.Getenv("TRIBLER_DOWNLOAD_DIR"),
	}
	// convert body to json
	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Print(err)
		return
	}

	req, err := http.NewRequest("PUT", os.Getenv("TRIBLER_API_ENDPOINT")+"/downloads", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Print(err)
		return
	}
	// set X-Api-Key header
	req.Header.Set("X-Api-Key", os.Getenv("TRIBLER_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	// send request
	response, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return
	}
	// pretty print response in the log
	log.Printf("%+v", response.Body)
}

// GetDownloadsFiles retrieves files of a torrent
// URI: /downloads/:hash:/files
// Method: GET
// Request: None
// Response: TorrentFiles
func GetDownloadsFiles(hash string) TorrentFiles {

	req, err := http.NewRequest("GET", os.Getenv("TRIBLER_API_ENDPOINT")+"/downloads/"+hash+"/files", nil)
	if err != nil {
		log.Print(err)
		return TorrentFiles{}
	}
	// set X-Api-Key header
	req.Header.Set("X-Api-Key", os.Getenv("TRIBLER_API_KEY"))

	// send request
	response, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return TorrentFiles{}
	}
	// pretty print response in the log
	log.Printf("%+v", response.Body)
	var torrentFiles TorrentFiles
	// pretty print response in the log
	err = json.NewDecoder(response.Body).Decode(&torrentFiles)
	if err != nil {
		log.Print(err)
		return TorrentFiles{}
	}

	return torrentFiles

}

// DeleteDownload deletes a torrent
// URI: /downloads/:hash:
// Method: DELETE
// Request:
//
//	{
//		"remove_data": bool
//	}
//
// Response: None
func DeleteDownload(hash string, remove_data bool) {

	// create request body
	body := map[string]interface{}{
		"remove_data": remove_data,
	}
	// convert body to json
	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Print(err)
		return
	}

	req, err := http.NewRequest("DELETE", os.Getenv("TRIBLER_API_ENDPOINT")+"/downloads/"+hash, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Print(err)
		return
	}
	// set X-Api-Key header
	req.Header.Set("X-Api-Key", os.Getenv("TRIBLER_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	// send request
	response, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return
	}
	// pretty print response in the log
	log.Printf("%+v", response.Body)
}

// Update a download
// URI: /downloads/:hash:
// Method: PATCH
//
//	Body: {
//		state string [resume/stop/recheck]
//	}
func UpdateDownload(hash string, state string) {

	// create request body
	body := map[string]interface{}{
		"state": state,
	}
	// convert body to json
	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Print(err)
		return
	}

	req, err := http.NewRequest("PATCH", os.Getenv("TRIBLER_API_ENDPOINT")+"/downloads/"+hash, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Print(err)
		return
	}
	// set X-Api-Key header
	req.Header.Set("X-Api-Key", os.Getenv("TRIBLER_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	// send request
	response, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return
	}
	// pretty print response in the log
	log.Printf("%+v", response.Body)
}
