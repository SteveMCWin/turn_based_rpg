package handlers

import (
	"html/template"
	"math/rand"
	"strconv"

	// "log"
	"net/http"
	"sync"

	// "time"

	"tbrpg/database"
	"tbrpg/game"
	"tbrpg/models"

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
	db     *database.DataBase
	games  map[int]*game.Game
	mu     sync.RWMutex
	router *gin.Engine
}

func NewServer(config *game.GameConfig, db *database.DataBase) http.Handler {
	s := &Server{
		config: config,
		db:     db,
		games:  make(map[int]*game.Game),
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
	r.GET("/event", s.handleGetEvent)
	r.GET("/post-battle", s.handleGetAfterBattle)
	r.GET("/moves", s.handleGetMoves)

	// Game state (read-only)
	r.GET("/game/state", s.handleGetGameState)
	r.GET("/game/heroes", s.handleGetHeroes)

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
	c.SetCookie(gameSessionCookieName, strconv.Itoa(g.ID), one_week, "/", "", false, true)
}

// Get a specific game based on the id of the game stored in a cookie
func (s *Server) gameFromRequest(c *gin.Context) (*game.Game, bool) {
	cookie, err := c.Cookie(gameSessionCookieName)
	if err != nil {
		return nil, false
	}

	id, err := strconv.Atoi(cookie)
	if err != nil {
		return nil, false
	}

	s.mu.RLock()
	g, ok := s.games[id]
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
	c.HTML(http.StatusOK, "post_battle.html", gin.H{})
}

func (s *Server) handleGetEvent(c *gin.Context) {
	c.HTML(http.StatusOK, "event.html", gin.H{})
}

func (s *Server) handleGetMoves(c *gin.Context) {
	c.HTML(http.StatusOK, "moves.html", gin.H{})
}

// ====================================================
// ================= UTILITY HANDLERS =================
// ====================================================

func (s *Server) handleGetHeroes(c *gin.Context) {
	c.JSON(http.StatusOK, s.config.HeroTemplates)
}

func (s *Server) handleGetGameState(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}
	c.JSON(http.StatusOK, g)
}

// ======================================================
// ================= GAME FLOW HANDLERS =================
// ======================================================

func (s *Server) handlePostNewGame(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database unavailable"})
		return
	}

	var req struct {
		HeroID string `json:"hero_id"`
	}
	c.ShouldBindJSON(&req)

	hero := s.config.HeroTemplates[0]
	for _, h := range s.config.HeroTemplates {
		if h.ID == req.HeroID {
			hero = h
			break
		}
	}

	g := game.NewGame(s.config, hero)
	id, err := s.db.CreateGame(g)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	g.ID = id
	s.setSession(c, g)
	c.JSON(http.StatusOK, g)
}

func (s *Server) handlePostEnterRoom(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		return
	}

	var req struct { RoomID string `json:"room_id" binding:"required"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id is required"})
		return
	}

	if err := g.EnterRoom(req.RoomID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, g)
}

func (s *Server) handlePostBattleMove(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		return
	}
	var req struct { MoveID string `json:"move_id" binding:"required"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "move_id is required"})
		return
	}
	result, err := g.SubmitPlayerMove(req.MoveID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) handlePostEquipMove(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		return
	}

	var req struct { MoveID string `json:"move_id" binding:"required"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "move_id is required"})
		return
	}
	hero := &g.Player

	maxEquipped := s.config.Settings.MaxEquippedMoves

	if len(hero.EquippedMoves) >= maxEquipped {
		c.JSON(http.StatusBadRequest, gin.H{"error": "equipped moves at maximum"})
		return
	}

	for _, id := range hero.EquippedMoves {
		if id == req.MoveID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "move already equipped"})
			return
		}
	}

	if hero.GetMoveLevel(req.MoveID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "move not learned"})
		return
	}

	hero.EquippedMoves = append(hero.EquippedMoves, req.MoveID)
	c.JSON(http.StatusOK, g)
}

func (s *Server) handlePostUnequipMove(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		return
	}
	var req struct { MoveID string `json:"move_id" binding:"required"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "move_id is required"})
		return
	}

	hero := &g.Player
	if len(hero.EquippedMoves) <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "must keep at least one move equipped"})
		return
	}

	updated := hero.EquippedMoves[:0:0]
	for _, id := range hero.EquippedMoves {
		if id != req.MoveID {
			updated = append(updated, id)
		}
	}

	hero.EquippedMoves = updated
	c.JSON(http.StatusOK, g)
}

func (s *Server) handlePostLevelUp(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		return
	}
	if g.PendingLevelUp == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no pending level up"})
		return
	}

	var req struct { Allocations map[string]int `json:"allocations"` }
	c.ShouldBindJSON(&req)
	if req.Allocations == nil {
		req.Allocations = map[string]int{}
	}

	pending := g.PendingLevelUp
	hero := &g.Player

	spent := 0
	for stat, pts := range req.Allocations {
		if pts <= 0 {
			continue
		}
		available := pending.ManualPoints - spent
		if available <= 0 {
			break
		}
		if pts > available {
			pts = available
		}
		hero.LevelUpStat(models.StatType(stat), pts)
		spent += pts
	}

	// Randomize remaining manual points + the automatic random points
	total := (pending.ManualPoints - spent) + pending.RandomPoints
	statPool := []models.StatType{models.HealthStat, models.AttackStat, models.DefenseStat, models.MagicStat, models.ManaStat}
	for i := 0; i < total; i++ {
		hero.LevelUpStat(statPool[rand.Intn(len(statPool))], 1)
	}

	g.PendingLevelUp = nil
	c.JSON(http.StatusOK, g)
}

// ==========================================
// ================= SAVING =================
// ==========================================

func (s *Server) handlePostSave(c *gin.Context) {

}

func (s *Server) handleGetListSaves(c *gin.Context) {

}

func (s *Server) handlePostLoadSave(c *gin.Context) {

}

func (s *Server) handleDeleteDeleteSave(c *gin.Context) {

}

