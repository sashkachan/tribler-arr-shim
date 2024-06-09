package server

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"

	"tribler-arr-shim/pkg/storage"
	torrent "tribler-arr-shim/pkg/torrent"

	"github.com/gin-contrib/sessions"
	cookiestore "github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"

	"path/filepath"

)

func StartServer() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v\n", err)
	}

	db_path := os.Getenv("SQLITE_PATH")
	if db_path == "" {
		log.Fatal("SQLITE_PATH not set")
		db_path = "/data/database.db"
	}
	initsql_path, err := filepath.Abs("./scripts/init_db.sql")
	if err != nil {
		log.Fatal("Error getting absolute path for init.sql: ", err)
	}
	db, err := storage.New(db_path, initsql_path)
	defer db.Close()

	r := apiv2Routes(db)

	// scheme := os.Getenv("TRIBLER_ARR_SHIM_SCHEME")
	addr := os.Getenv("TRIBLER_ARR_SHIM_ADDR")
	port := os.Getenv("TRIBLER_ARR_SHIM_PORT")
	serverAddr := fmt.Sprintf("%s:%s", addr, port)

	log.Printf("Starting server on %s", serverAddr)
	err = http.ListenAndServe(serverAddr, r)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}

func loggingMiddleware(c *gin.Context) {
	log.Printf("Request: %s %s", c.Request.Method, c.Request.URL.Path)
	c.Next()
}

func apiv2Routes(db storage.Database) *gin.Engine {
	var handler = torrent.NewHandler(db)
	r := gin.Default()
	gob.Register(map[string]interface{}{})
	r.Use(sessions.Sessions("tribler-arr-shim", cookiestore.NewStore([]byte(os.Getenv("SESSION_SECRET")))))
	r.Use(loggingMiddleware)
	r.POST("/api/v2/auth/login", handler.LoginHandler())
	// r.GET("/api/auth/callback", authentication.CallbackHandler(authenticator, ()))
	// r.GET("/api/user", isAuthenticatedFn(), user.GetUserInfoHandler())
	// user subscription
	r.GET("/api/v2/app/webapiVersion", handler.GetWebApiVersion())
	r.GET("/api/v2/app/version", handler.GetVersion())
	r.GET("/api/v2/app/preferences", handler.GetAppPreferences())
	r.GET("/api/v2/torrents/info", handler.GetInfo())
	r.GET("/api/v2/torrents/properties", handler.GetProperties())
	r.GET("/api/v2/torrents/files", handler.GetTorrentsContents())
	r.POST("/api/v2/torrents/add", handler.Add())
	r.POST("/api/v2/torrents/delete", handler.Delete())
	r.POST("/api/v2/torrents/setCategory", handler.SetCategory())
	r.GET("/api/v2/torrents/categories", handler.GetCategories())
	r.POST("/api/v2/torrents/setShareLimits", handler.SetShareLimits())
	r.POST("/api/v2/torrents/topPrio", handler.SetTopPriority())
	r.POST("/api/v2/torrents/pause", handler.PauseTorrent())
	r.POST("/api/v2/torrents/resume", handler.ResumeTorrent())
	r.POST("/api/v2/torrents/setForceStart", handler.SetForceStartTorrent())
	r.POST("/api/v2/torrents/createCategory", handler.CreateCategory())

	return r
}
