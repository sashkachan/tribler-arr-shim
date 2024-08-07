package language

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"tribler-arr-shim/pkg/storage"
	"tribler-arr-shim/pkg/tribler"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Structs
// Torrent struct
type Torrent struct {
	Tags          string  `json:"tags"`
	State         string  `json:"state"`
	ContentPath   string  `json:"content_path"`
	Name          string  `json:"name"`
	Hash          string  `json:"hash"`
	Category      string  `json:"category"`
	NumLeechs     int     `json:"num_leechs"`
	Progress      float64 `json:"progress"`
	NumComplete   int     `json:"num_complete"`
	NumIncomplete int     `json:"num_incomplete"`
	Dlspeed       int     `json:"dlspeed"`
	NumSeeds      int     `json:"num_seeds"`
	Priority      int     `json:"priority"`
	Upspeed       int     `json:"upspeed"`
	Ratio         int     `json:"ratio"`
	Eta           float64 `json:"eta"`
	Size          int     `json:"size"`
	FLPiecePrio   bool    `json:"f_l_piece_prio"`
	SeqDL         bool    `json:"seq_dl"`
	SuperSeeding  bool    `json:"super_seeding"`
	ForceStart    bool    `json:"force_start"`
}

// TorrentFiles
type TorrentFiles struct {
	Name         string  `json:"name"`
	PieceRange   []int   `json:"piece_range"`
	Index        int     `json:"index"`
	Size         int     `json:"size"`
	Progress     float64 `json:"progress"`
	Priority     int     `json:"priority"`
	Availability float64 `json:"availability"`
	IsSeed       bool    `json:"is_seed"`
}

type TorrentProperties struct {
	Comment                string  `json:"comment"`
	Name                   string  `json:"name"`
	CreatedBy              string  `json:"created_by"`
	SavePath               string  `json:"save_path"`
	ShareRatio             float64 `json:"share_ratio"`
	CreationDate           int     `json:"creation_date"`
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
	PieceSize              int     `json:"piece_size"`
	AdditionDate           int     `json:"addition_date"`
	CompletionDate         int     `json:"completion_date"`
	TotalWasted            int     `json:"total_wasted"`
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

type Handler struct {
	DB storage.Database
}

func NewHandler(db storage.Database) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a random string of 32 characters
		sid := "01234678910111213141617"

		// Store the session ID in the database
		err := h.storeSessionID(sid)
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
func (h *Handler) GetWebApiVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "2.2.8")
	}
}

// GetVersion retrieves api version
func (h *Handler) GetVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "4.1.3")
	}
}

// GetInfo retrieves information about a torrent
func (h *Handler) GetInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		// filter torrents by category
		filtered_torrents := []Torrent{}
		// get category
		category := c.Query("category")
		// get category_torrents by category
		category_torrents, err := h.DB.GetTorrentsByCategory(category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err})
			return
		}
		downloads, _ := tribler.GetDownloads()
		converted_downloads := h.ConvertTriblerDownloadstoTorrent(downloads.Downloads)

		// if torrents is empty, return empty json
		if len(category_torrents) == 0 {
			c.JSON(http.StatusOK, filtered_torrents)
			return
		}

		// convert torrents to a map so we can get by hash
		torrents_map := make(map[string]Torrent)
		for _, torrent := range converted_downloads {
			torrents_map[torrent.Hash] = torrent
		}

		for _, category_torrent := range category_torrents {
			if torrent, ok := torrents_map[category_torrent.Hash]; ok {
				torrent.Category = category_torrent.Category
				filtered_torrents = append(filtered_torrents, torrent)
			}
		}

		c.JSON(http.StatusOK, filtered_torrents)
	}
}

// GetAppPreferences retrieves app preferences
func (h *Handler) GetAppPreferences() gin.HandlerFunc {
	return func(c *gin.Context) {
		DummyAppPreferences.SavePath = os.Getenv("TRIBLER_DOWNLOAD_DIR")
		c.JSON(http.StatusOK, DummyAppPreferences)
	}
}

// GetProperties retrieves properties of a torrent
func (h *Handler) GetProperties() gin.HandlerFunc {
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
func (h *Handler) GetTorrentsContents() gin.HandlerFunc {
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

func categoryExists(category string, categories []storage.Category) bool {
	for _, v := range categories {
		if v.Name == category {
			return true
		}
	}
	return false
}

// Add adds a new torrent
func (h *Handler) Add() gin.HandlerFunc {
	return func(c *gin.Context) {
		urls := c.PostForm("urls")
		category := c.PostForm("category")
		log.Println("torrent.Add urls: ", urls)
		log.Println("torrent.Add category: ", category)

		urlsLines := strings.Split(urls, "\n")
		firstURL := urlsLines[0]

		infohash, err := tribler.AddDownload(firstURL)
		if err != nil {
			handleInternalError(c, "Error adding torrent", err)
			return
		}
		log.Printf("Response: %s", string(infohash))

		categories, err := h.DB.GetCategories()
		if err != nil {
			handleInternalError(c, "Failed to get categories", err)
			return
		}

		if !categoryExists(category, categories) {
			err = h.DB.AddCategory(category, os.Getenv("TRIBLER_DOWNLOAD_DIR"))
			if err != nil {
				handleInternalError(c, "Failed to add category "+category, err)
				return
			}
		}

		torrent := storage.Torrent{
			Hash:     infohash,
			Category: category,
		}
		h.DB.AddTorrent(torrent)

		c.JSON(http.StatusOK, gin.H{"message": "Torrent added"})
	}
}

// Delete deletes a torrent
func (h *Handler) Delete() gin.HandlerFunc {
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
		for _, hash := range strings.Split(hashes, ",") {
			h.DB.DeleteTorrent(hash)
		}
	}
}

// SetCategory sets the category of a torrent
func (h *Handler) SetCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: implement setting the category of a torrent in DB
		c.JSON(http.StatusOK, gin.H{"message": "Torrent category set"})
	}
}

// GetCategories retrieves all torrent categories
func (h *Handler) GetCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, err := h.DB.GetCategories()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err})
			return
		}

		// if categories is empty, return empty json
		if len(categories) == 0 {

			c.JSON(http.StatusOK, gin.H{})
			return
		}

		//
		// Return json in format
		// {
		//   {
		//     "CategoryName": {
		//       "savePath": "path",
		//       "name": "CategoryName"
		//     }
		//   }
		// }

		categoryMap := make(map[string]map[string]string)
		for _, category := range categories {
			categoryMap[category.Name] = map[string]string{
				"savePath": category.SavePath,
				"name":     category.Name,
			}
		}
		c.JSON(http.StatusOK, categoryMap)
	}
}

// SetShareLimits sets the share limits of a torrent
func (h *Handler) SetShareLimits() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: implement setting the share limits of a torrent in DB
		c.JSON(http.StatusOK, gin.H{"message": "Torrent share limits set"})
	}
}

// SetTopPriority sets the priority of a torrent to top
func (h *Handler) SetTopPriority() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: implement setting the priority of a torrent to top in DB
		c.JSON(http.StatusOK, gin.H{"message": "Torrent priority set to top"})
	}
}

// PauseTorrent pauses a torrent
func (h *Handler) PauseTorrent() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get hashes
		hashes := c.PostForm("hashes")
		tribler.UpdateDownload(hashes, "stop")
		c.JSON(http.StatusOK, gin.H{"message": "Torrent paused"})
	}
}

// ResumeTorrent resumes a torrent
func (h *Handler) ResumeTorrent() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get hashes
		hashes := c.PostForm("hashes")
		tribler.UpdateDownload(hashes, "resume")
		c.JSON(http.StatusOK, gin.H{"message": "Torrent resumed"})
	}
}

// SetForceStartTorrent sets a torrent to force start
func (h *Handler) SetForceStartTorrent() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: implement setting a torrent to force start in DB
		c.JSON(http.StatusOK, gin.H{"message": "Torrent set to force start"})
	}
}

// CreateCategory creates a new category
func (h *Handler) CreateCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get category
		category := c.PostForm("category")
		// savePath is the same as the default download directory
		savePath := os.Getenv("TRIBLER_DOWNLOAD_DIR")
		// check if category already exists
		categories, err := h.DB.GetCategories()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err})
			return
		}
		if c != nil {
			for _, v := range categories {
				if v.Name == category {
					// dump v to log
					log.Println("Category already exists: ", v)
					c.JSON(http.StatusOK, gin.H{"message": "Category already exists"})
					return
				}
			}
		}
		err = h.DB.AddCategory(category, savePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Category created"})
	}
}

func (h *Handler) storeSessionID(_ string) error {
	// TODO: implement storing the session ID in the database
	return nil
}

func (h *Handler) ConvertTriblerDownloadstoTorrent(downloads []tribler.Download) []Torrent {
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

func (h *Handler) containsFileExtensionSuffix(s string) bool {
	re := regexp.MustCompile(`\.[a-z0-9]{3}$`)
	return re.MatchString(s)
}

func ConvertTriblerDownloadtoTorrentProperties(download tribler.Download) TorrentProperties {
	destination := download.Destination
	// if !containsFileExtensionSuffix(download.Name) {
	// 	log.Println("Does not contain extension = ", download.Name)
	// 	destination = download.Destination + "/" + download.Name
	// }

	return TorrentProperties{
		SavePath:               destination,
		Name:                   download.Name,
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
			Name:         file.Name,
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

func handleInternalError(c *gin.Context, msg string, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"message": msg})
	log.Printf("Error: %+v", err)
}

func ImportNonCategorisedTorrents(db storage.Database, category string) error {
	// get all torrents from the database
	// then get all torrents from tribler
	// compare the two lists and import any torrents that are not in the database

	// get all torrents from the database
	all_torrents, err := db.GetAllTorrents()
	if err != nil {
		log.Printf("db.GetAllTorrents returned error")
		log.Printf("Error: %+v", err)
	}

	// get all downloads_response from tribler
	downloads_response, err := tribler.GetDownloads()
	if err != nil {
		log.Println("tribler.GetDownloads() returned error")
		log.Printf("Error: %+v", err)
	}

	log.Printf("Download response len: %d", len(downloads_response.Downloads))

	hashes := map[string]bool{}
	for _, storage_torrent := range all_torrents {
		hashes[storage_torrent.Hash] = true
	}

	for _, download := range downloads_response.Downloads {
		log.Println("Processing", download.Infohash, download.Name)
		if _, exists := hashes[download.Infohash]; !exists {
			log.Printf("Importing torrent %s", download.Infohash)
			new_torrent := storage.Torrent{
				Hash:     download.Infohash,
				Category: category,
			}
			err := db.AddTorrent(new_torrent)
			if err != nil {
				log.Printf("Error adding torrent: %+v", err)
			}
		}
	}

	return err
}
