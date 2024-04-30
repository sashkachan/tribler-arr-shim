package server

import (
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
)

func StartServer() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v\n", err)
	}

	r := apiv2Routes()

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

func apiv2Routes() *gin.Engine {
	r := gin.Default()

	gob.Register(map[string]interface{}{})
	r.Use(sessions.Sessions("tribler-arr-shim", cookiestore.NewStore([]byte(os.Getenv("SESSION_SECRET")))))
	r.Use(loggingMiddleware)
	r.POST("/api/v2/auth/login", torrent.LoginHandler())
	// r.GET("/api/auth/callback", authentication.CallbackHandler(authenticator, ()))
	// r.GET("/api/user", isAuthenticatedFn(), user.GetUserInfoHandler())
	// user subscription
	r.GET("/api/v2/app/webapiVersion", torrent.GetWebApiVersion())
	r.GET("/api/v2/app/version", torrent.GetVersion())
	r.GET("/api/v2/app/preferences", torrent.GetAppPreferences())
	r.GET("/api/v2/torrents/info", torrent.GetInfo())
	r.GET("/api/v2/torrents/properties", torrent.GetProperties())
	r.GET("/api/v2/torrents/files", torrent.GetTorrentsContents())
	r.POST("/api/v2/torrents/add", torrent.Add())
	r.POST("/api/v2/torrents/delete", torrent.Delete())
	r.POST("/api/v2/torrents/setCategory", torrent.SetCategory())
	r.GET("/api/v2/torrents/categories", torrent.GetCategories())
	r.POST("/api/v2/torrents/setShareLimits", torrent.SetShareLimits())
	r.POST("/api/v2/torrents/topPrio", torrent.SetTopPriority())
	r.POST("/api/v2/torrents/pause", torrent.PauseTorrent())
	r.POST("/api/v2/torrents/resume", torrent.ResumeTorrent())
	r.POST("/api/v2/torrents/setForceStart", torrent.SetForceStartTorrent())

	return r
}
