package language

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"tribler-arr-shim/pkg/tribler"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Structs
// Torrent struct
type Torrent struct {
	Dlspeed       int     `json:"dlspeed"`
	Eta           float64 `json:"eta"`
	FLPiecePrio   bool    `json:"f_l_piece_prio"`
	ForceStart    bool    `json:"force_start"`
	Hash          string  `json:"hash"`
	Category      string  `json:"category"`
	Tags          string  `json:"tags"`
	Name          string  `json:"name"`
	NumComplete   int     `json:"num_complete"`
	NumIncomplete int     `json:"num_incomplete"`
	NumLeechs     int     `json:"num_leechs"`
	NumSeeds      int     `json:"num_seeds"`
	Priority      int     `json:"priority"`
	Progress      float64 `json:"progress"`
	Ratio         int     `json:"ratio"`
	SeqDL         bool    `json:"seq_dl"`
	ContentPath   string  `json:"content_path"`
	Size          int     `json:"size"`
	// state can be: "downloading", "uploading", "pausedUP"
	State        string `json:"state"`
	SuperSeeding bool   `json:"super_seeding"`
	Upspeed      int    `json:"upspeed"`
}

// TorrentFiles
type TorrentFiles struct {
	Index        int     `json:"index"`
	Name         string  `json:"name"`
	Size         int     `json:"size"`
	Progress     float64 `json:"progress"`
	Priority     int     `json:"priority"`
	IsSeed       bool    `json:"is_seed"`
	PieceRange   []int   `json:"piece_range"`
	Availability float64 `json:"availability"`
}

type TorrentProperties struct {
	SavePath               string  `json:"save_path"`
	CreationDate           int     `json:"creation_date"`
	PieceSize              int     `json:"piece_size"`
	Comment                string  `json:"comment"`
	TotalWasted            int     `json:"total_wasted"`
	TotalUploaded          int     `json:"total_uploaded"`
	TotalUploadedSession   int     `json:"total_uploaded_session"`
	TotalDownloaded        int     `json:"total_downloaded"`
	TotalDownloadedSession int     `json:"total_downloaded_session"`
	UpLimit                int     `json:"up_limit"`
	DlLimit                int     `json:"dl_limit"`
	TimeElapsed            int     `json:"time_elapsed"`
	SeedingTime            int     `json:"seeding_time"`
	NbConnections          int     `json:"nb_connections"`
	NbConnectionsLimit     int     `json:"nb_connections_limit"`
	ShareRatio             float64 `json:"share_ratio"`
	AdditionDate           int     `json:"addition_date"`
	CompletionDate         int     `json:"completion_date"`
	CreatedBy              string  `json:"created_by"`
	DlSpeedAvg             int     `json:"dl_speed_avg"`
	DlSpeed                int     `json:"dl_speed"`
	Eta                    int     `json:"eta"`
	LastSeen               int     `json:"last_seen"`
	Peers                  int     `json:"peers"`
	PeersTotal             int     `json:"peers_total"`
	PiecesHave             int     `json:"pieces_have"`
	PiecesNum              int     `json:"pieces_num"`
	Reannounce             int     `json:"reannounce"`
	Seeds                  int     `json:"seeds"`
	SeedsTotal             int     `json:"seeds_total"`
	TotalSize              int     `json:"total_size"`
	UpSpeedAvg             int     `json:"up_speed_avg"`
	UpSpeed                int     `json:"up_speed"`
}

type AppPreferences struct {
	SavePath               string  `json:"save_path"`
	MaxRatioAction         string  `json:"max_ratio_act"`
	MaxRatio               float64 `json:"max_ratio"`
	MaxSeedingTime         int     `json:"max_seeding_time"`
	MaxRatioEnabled        bool    `json:"max_ratio_enabled"`
	MaxSeedingTimeEnabled  bool    `json:"max_seeding_time_enabled"`
	QueueingEnabled        bool    `json:"queueing_enabled"`
	DhtEnabled             bool    `json:"dht"`
	CreateSubfolderEnabled bool    `json:"create_subfolder_enabled"`
}

var DummyAppPreferences = AppPreferences{
	SavePath:               os.Getenv("TRIBLER_DOWNLOAD_DIR"),
	MaxRatioEnabled:        false,
	MaxRatio:               0,
	MaxSeedingTimeEnabled:  false,
	MaxSeedingTime:         0,
	MaxRatioAction:         "pause",
	QueueingEnabled:        true,
	DhtEnabled:             true,
	CreateSubfolderEnabled: false,
}

func LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a random string of 32 characters
		sid := "01234678910111213141617"

		// Store the session ID in the database
		err := storeSessionID(sid)
		if err != nil {
			return
		}
		c.SetCookie("SID", sid, 0, "/", "", false, false)
		// Set the SID cookie
		session := sessions.Default(c)
		session.Set("SID", sid)
		session.Save()

		c.String(http.StatusOK, "Ok.")
	}
}

// GetApiVersion retrieves api version
func GetWebApiVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "2.2.8")
	}
}

// GetVersion retrieves api version
func GetVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "4.1.3")
	}
}

// GetInfo retrieves information about a torrent
func GetInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get category from query
		// category := c.Query("category")
		// get Downloads from tribler
		downloads, _ := tribler.GetDownloads()
		// convert downloads to the following struct
		torrents := ConvertTriblerDownloadstoTorrent(downloads.Downloads)

		// TODO: implement getting torrent info from DB
		c.JSON(http.StatusOK, torrents)
	}
}

// GetAppPreferences retrieves app preferences
func GetAppPreferences() gin.HandlerFunc {
	return func(c *gin.Context) {
		DummyAppPreferences.SavePath = os.Getenv("TRIBLER_DOWNLOAD_DIR")
		c.JSON(http.StatusOK, DummyAppPreferences)
	}
}

// GetProperties retrieves properties of a torrent
func GetProperties() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get hash
		hash := c.Query("hash")
		download, err := tribler.GetDownload(hash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err})
			return
		}
		// convert download to the following struct
		properties := ConvertTriblerDownloadtoTorrentProperties(download)
		c.JSON(http.StatusOK, properties)
	}
}

// GetTorrentsContents retrieves contents of a torrent
func GetTorrentsContents() gin.HandlerFunc {
	return func(c *gin.Context) {
		hash := c.Query("hash")
		torrentFiles, _ := tribler.GetDownloadsFiles(hash)
		// convert downloads to the following struct
		files := ConvertTriblerFilesToTorrentFiles(torrentFiles.Files)
		// if files is empty, return empty json
		if len(files) == 0 {
			c.JSON(http.StatusOK, gin.H{})
			return
		}

		c.JSON(http.StatusOK, files)
	}
}

// Add adds a new torrent
func Add() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get urls parameter, it's separated by new lines
		urls := c.PostForm("urls")
		log.Println("torrent.Add urls: ", urls)
		// split urls by new line
		urls_lines := strings.Split(urls, "\n")
		// only use the first url
		response, err := tribler.AddDownload(urls_lines[0])
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error adding torrent"})
			log.Printf("Error: %+v", err)
			return
		}
		log.Printf("Response: %+v", response)
		c.JSON(http.StatusOK, gin.H{"message": "Torrent added"})
	}
}

// Delete deletes a torrent
func Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get hashes
		hashes := c.PostForm("hashes")
		deleteFiles := false
		// get deleteFiles
		deleteFilesReq := c.PostForm("deleteFiles")
		if deleteFilesReq == "true" {
			deleteFiles = true
		}

		tribler.DeleteDownload(hashes, deleteFiles)
	}
}

// SetCategory sets the category of a torrent
func SetCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: implement setting the category of a torrent in DB
		c.JSON(http.StatusOK, gin.H{"message": "Torrent category set"})
	}
}

// GetCategories retrieves all torrent categories
func GetCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: implement getting all torrent categories from DB
		c.JSON(http.StatusOK, gin.H{
			os.Getenv("DEFAULT_CATEGORY"): gin.H{
				"name":     os.Getenv("DEFAULT_CATEGORY"),
				"savePath": os.Getenv("TRIBLER_DOWNLOAD_DIR"),
			},
		})
	}
}

// SetShareLimits sets the share limits of a torrent
func SetShareLimits() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: implement setting the share limits of a torrent in DB
		c.JSON(http.StatusOK, gin.H{"message": "Torrent share limits set"})
	}
}

// SetTopPriority sets the priority of a torrent to top
func SetTopPriority() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: implement setting the priority of a torrent to top in DB
		c.JSON(http.StatusOK, gin.H{"message": "Torrent priority set to top"})
	}
}

// PauseTorrent pauses a torrent
func PauseTorrent() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get hashes
		hashes := c.PostForm("hashes")
		tribler.UpdateDownload(hashes, "stop")
		c.JSON(http.StatusOK, gin.H{"message": "Torrent paused"})
	}
}

// ResumeTorrent resumes a torrent
func ResumeTorrent() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get hashes
		hashes := c.PostForm("hashes")
		tribler.UpdateDownload(hashes, "resume")
		c.JSON(http.StatusOK, gin.H{"message": "Torrent resumed"})
	}
}

// SetForceStartTorrent sets a torrent to force start
func SetForceStartTorrent() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: implement setting a torrent to force start in DB
		c.JSON(http.StatusOK, gin.H{"message": "Torrent set to force start"})
	}
}

func storeSessionID(sid string) error {
	// TODO: implement storing the session ID in the database
	return nil
}

func ConvertTriblerDownloadstoTorrent(downloads []tribler.Download) []Torrent {
	// Convert tribler download to torrent
	torrent := []Torrent{}
	for _, download := range downloads {
		var state string
		switch download.Status {
		case "SEEDING":
			state = "pausedUP"
		case "DOWNLOADING":
			state = "downloading"
		case "PAUSED":
			state = "pausedUP"
		}
		torrent = append(torrent, Torrent{
			Dlspeed:       download.SpeedDown,
			Eta:           download.Eta,
			FLPiecePrio:   false,
			ForceStart:    false,
			Hash:          download.Infohash,
			Category:      os.Getenv("DEFAULT_CATEGORY"),
			Tags:          "",
			Name:          download.Name,
			NumComplete:   download.NumPeers,
			NumIncomplete: download.NumPeers,
			NumLeechs:     download.NumPeers,
			NumSeeds:      download.NumPeers,
			Priority:      0,
			Progress:      download.Progress,
			Ratio:         0,
			SeqDL:         false,
			Size:          download.Size,
			ContentPath:   download.Destination + "/" + download.Name,
			State:         state,
			SuperSeeding:  false,
			Upspeed:       download.SpeedUp,
		})
	}
	return torrent
}

func containsFileExtensionSuffix(s string) bool {
	re := regexp.MustCompile(`\.{3}$`)
	return re.MatchString(s)
}

func ConvertTriblerDownloadtoTorrentProperties(download tribler.Download) TorrentProperties {
	// manipulate destination so that single file downloads destination is handled
	destination := download.Destination
	if !containsFileExtensionSuffix(download.Name) {
		destination = download.Destination + "/" + download.Name
	}
	return TorrentProperties{
		SavePath:               destination,
		CreationDate:           0,
		PieceSize:              0,
		Comment:                "",
		TotalWasted:            0,
		TotalUploaded:          0,
		TotalUploadedSession:   0,
		TotalDownloaded:        0,
		TotalDownloadedSession: 0,
		UpLimit:                0,
		DlLimit:                0,
		TimeElapsed:            0,
		SeedingTime:            0,
		NbConnections:          0,
		NbConnectionsLimit:     0,
		ShareRatio:             0,
		AdditionDate:           0,
		CompletionDate:         0,
		CreatedBy:              "",
		DlSpeedAvg:             0,
		DlSpeed:                0,
		Eta:                    0,
		LastSeen:               0,
		Peers:                  0,
		PeersTotal:             0,
		PiecesHave:             0,
		PiecesNum:              0,
		Reannounce:             0,
		Seeds:                  0,
		SeedsTotal:             0,
		TotalSize:              0,
		UpSpeedAvg:             0,
		UpSpeed:                0,
	}
}

func ConvertTriblerFilesToTorrentFiles(files []tribler.Files) []TorrentFiles {
	torrentFiles := []TorrentFiles{}
	for _, file := range files {
		torrentFiles = append(torrentFiles, TorrentFiles{
			Index:        file.Index,
			Name:         "./" + file.Name,
			Size:         file.Size,
			Progress:     file.Progress,
			Priority:     0,
			IsSeed:       false,
			PieceRange:   []int{},
			Availability: 0,
		})
	}
	return torrentFiles
}
