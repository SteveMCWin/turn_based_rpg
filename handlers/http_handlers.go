package handlers

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"tbrpg/models"
	
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
)

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"rangeN": func(n int) []int {
			out := make([]int, n)
			for i := range n {
				out[i] = i
			}
			return out
		},
		"mul": func(a, b float64) float64 {
			return a * b
		},
		"div": func(a, b float64) float64 {
			return a / b
		},
		"float64": func(a int) float64 {
			return float64(a)
		},
		"substr": func(s string, start, length int) string {
			if start < 0 || start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
	}
}

var sessionManager *scs.SessionManager

func SetUpRouter(db *models.DataBase) http.Handler {
	godotenv.Load() // loads data like client secrets from the .env file in the project root dir

	sessionManager = scs.New()
	sessionManager.Lifetime = time.Hour * 24 * 30
	sessionManager.Store = sqlite3store.New(db.Data)
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.Secure = true

	if gin.Mode() == gin.TestMode {
		log.Println("WARNING: TEST MODE")
		sessionManager.Cookie.Secure = false
		sessionManager.Cookie.SameSite = http.SameSiteDefaultMode
		sessionManager.Cookie.Name = "test_session"
	}

	router := gin.Default()

	router.Static("/css", "./css")
	router.Static("/js", "./js")

	router.GET("/", HandleGetHome())
	router.GET("/run/config")
	router.GET("/run/next-monster-move")

	router.SetFuncMap(templateFuncs())
	router.LoadHTMLGlob("templates/*") // loads all templates from the templates directory

	handler := sessionManager.LoadAndSave(router)

	return handler
}

func HandleGetHome() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	}
}

