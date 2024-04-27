package server

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"

	torrent "tribler-arr-shim/pkg/torrent"

	"github.com/gin-contrib/sessions"
	cookiestore "github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
)

func StartServer() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v\n", err)
	}

	db, err := sql.Open("sqlite3", os.Getenv("SQLITE_PATH"))
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	initDBSQL, err := os.ReadFile("scripts/init_db.sql")
	if err != nil {
		log.Fatal("Error reading init_db.sql:", err)
	}

	_, err = db.Exec(string(initDBSQL))
	if err != nil {
		log.Fatal("Error executing init_db.sql:", err)
	}

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

func apiv2Routes(db *sql.DB) *gin.Engine {
	r := gin.Default()

	gob.Register(map[string]interface{}{})
	r.Use(sessions.Sessions("tribler-arr-shim", cookiestore.NewStore([]byte(os.Getenv("SESSION_SECRET")))))
	r.Use(loggingMiddleware)
	r.POST("/api/v2/auth/login", torrent.LoginHandler(db))
	// r.GET("/api/auth/callback", authentication.CallbackHandler(authenticator, db))
	// r.GET("/api/user", isAuthenticatedFn(db), user.GetUserInfoHandler(db))
	// user subscription
	r.GET("/api/v2/app/webapiVersion", torrent.GetWebApiVersion(db))
	r.GET("/api/v2/app/version", torrent.GetVersion(db))
	r.GET("/api/v2/app/preferences", torrent.GetAppPreferences(db))
	r.GET("/api/v2/torrents/info", torrent.GetInfo(db))
	r.GET("/api/v2/torrents/properties", torrent.GetProperties(db))
	r.GET("/api/v2/torrents/files", torrent.GetTorrentsContents(db))
	r.POST("/api/v2/torrents/add", torrent.Add(db))
	r.POST("/api/v2/torrents/delete", torrent.Delete(db))
	r.POST("/api/v2/torrents/setCategory", torrent.SetCategory(db))
	r.GET("/api/v2/torrents/categories", torrent.GetCategories(db))
	r.POST("/api/v2/torrents/setShareLimits", torrent.SetShareLimits(db))
	r.POST("/api/v2/torrents/topPrio", torrent.SetTopPriority(db))
	r.POST("/api/v2/torrents/pause", torrent.PauseTorrent(db))
	r.POST("/api/v2/torrents/resume", torrent.ResumeTorrent(db))
	r.POST("/api/v2/torrents/setForceStart", torrent.SetForceStartTorrent(db))

	return r
}
