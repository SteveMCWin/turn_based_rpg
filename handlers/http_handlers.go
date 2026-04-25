package handlers

import (
	"database/sql"
	"html/template"
	// "log"
	"net/http"
	"sync"
	// "time"

	"tbrpg/game"
	// "tbrpg/models"

	"github.com/gin-gonic/gin"
)

const gameSessionCookieName = "current_game_session"

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

type Server struct {
	config *game.GameConfig
	db     *sql.DB
	games  map[string]*game.Game
	mu     sync.RWMutex
	router *gin.Engine
}

func NewServer(config *game.GameConfig, db *sql.DB) http.Handler {
	s := &Server{
		config: config,
		db:     db,
		games:  make(map[string]*game.Game),
	}
	r := gin.Default()

	r.SetFuncMap(templateFuncs())
	r.LoadHTMLGlob("templates/*")
	r.Static("/css", "./css")
	r.Static("/js", "./js")

	// Page routes
	r.GET("/", s.handleGetMenu)
	r.GET("/map", s.handleGetMap)
	r.GET("/battle", s.handleGetBattle)
	r.GET("/post-battle", s.handleGetAfterBattle)
	r.GET("/moves", s.handleGetMoves)

	// Game state (read-only)
	r.GET("/game/state", s.handleGetGameState)

	// Game lifecycle
	r.POST("/game/new", s.handlePostNewGame)
	r.POST("/game/room/enter", s.handlePostEnterRoom)
	r.POST("/game/battle/move", s.handlePostBattleMove)
	r.POST("/game/moves/equip", s.handlePostEquipMove)
	r.POST("/game/moves/unequip", s.handlePostUnequipMove)
	r.POST("/game/levelup", s.handlePostLevelUp)

	// Saves
	r.POST("/game/save", s.handlePostSave)
	r.GET("/game/saves", s.handleGetListSaves)
	r.POST("/game/saves/:id/load", s.handlePostLoadSave)
	r.DELETE("/game/saves/:id", s.handleDeleteDeleteSave)

	s.router = r
	return r
}

func (s *Server) setSession(c *gin.Context, g *game.Game) {
	s.mu.Lock()
	s.games[g.ID] = g
	s.mu.Unlock()

	one_week := 60*60*24*7
	c.SetCookie(gameSessionCookieName, g.ID, one_week, "/", "", false, true)
}

// Get a specific game based on the id of the game stored in a cookie
func (s *Server) gameFromRequest(c *gin.Context) (*game.Game, bool) {
	cookie, err := c.Cookie(gameSessionCookieName)
	if err != nil {
		return nil, false
	}
	s.mu.RLock()
	g, ok := s.games[cookie]
	s.mu.RUnlock()
	
	
	return g, ok
}

// ================================================
// ================= STATIC PAGES =================
// ================================================

func (s *Server) handleGetMenu(c *gin.Context) {
	c.HTML(http.StatusOK, "menu.html", gin.H{})
}

func (s *Server) handleGetMap(c *gin.Context) {
	c.HTML(http.StatusOK, "map.html", gin.H{})
}

func (s *Server) handleGetBattle(c *gin.Context) {
	c.HTML(http.StatusOK, "battle.html", gin.H{})
}

func (s *Server) handleGetAfterBattle(c *gin.Context) {
	c.HTML(http.StatusOK, "menu.html", gin.H{})
}

func (s *Server) handleGetMoves(c *gin.Context) {
	c.HTML(http.StatusOK, "menu.html", gin.H{})
}

func (s *Server) handleGetGameState(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (s *Server) handlePostNewGame(c *gin.Context) {
	g := game.NewGame(s.config)
	s.setSession(c, g)
	c.JSON(http.StatusOK, g)
}

func (s *Server) handlePostEnterRoom(c *gin.Context) {

}

func (s *Server) handlePostBattleMove(c *gin.Context) {

}

func (s *Server) handlePostEquipMove(c *gin.Context) {

}

func (s *Server) handlePostUnequipMove(c *gin.Context) {

}

func (s *Server) handlePostLevelUp(c *gin.Context) {

}

func (s *Server) handlePostSave(c *gin.Context) {

}

func (s *Server) handleGetListSaves(c *gin.Context) {

}

func (s *Server) handlePostLoadSave(c *gin.Context) {

}

func (s *Server) handleDeleteDeleteSave(c *gin.Context) {

}

