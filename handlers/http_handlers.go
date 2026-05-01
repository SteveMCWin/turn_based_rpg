package handlers

import (
	"html/template"
	"math/rand"
	"strconv"

	"net/http"
	"sync"

	"tbrpg/database"
	"tbrpg/game"
	"tbrpg/models"

	"github.com/gin-gonic/gin"
)

// This file basically just creates the server and handles requests
// handling requests is not that interesting here, mostly just using
// logic defined in the the models and game packages

// name of the cookie storing current game session id
const gameSessionCookieName = "current_game_session"
const oneWeekSecs = 60 * 60 * 24 * 7

// funcitons used in html templates
// (I just compied all of these from another project, prolly half of them aren't even used)
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

// struct that containes all games running and what not
// oh also note that I was too lazy to delete the running games form the games map
// so eventually it'd overflow and crash the app, so don't go around
// playing 1000000 games without refresing the server! :^P
type Server struct {
	config *game.GameConfig
	db     *database.DataBase
	games  map[int]*game.Game
	mu     sync.RWMutex
	router *gin.Engine
}

// creates a new server with all the endpoints
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
	r.Static("/sprites", "./spirtes")

	// Page routes
	r.GET("/", s.handleGetMenu)
	r.GET("/map", s.handleGetMap)
	r.GET("/battle", s.handleGetBattle)
	r.GET("/event", s.handleGetEvent)
	r.GET("/post-battle", s.handleGetAfterBattle)
	r.GET("/moves", s.handleGetMoves)
	r.GET("/items", s.handleGetItemsPage)

	// Game state (read-only)
	r.GET("/game/state", s.handleGetGameState)
	r.GET("/game/heroes", s.handleGetHeroes)

	// Game lifecycle
	r.POST("/game/new", s.handlePostCreateNewGame)
	r.POST("/game/room/enter", s.handlePostEnterRoom)
	r.POST("/game/battle/move", s.handlePostBattleMove)
	r.POST("/game/battle/monster-move", s.handlePostMonsterMove)
	r.POST("/game/moves/equip", s.handlePostEquipMove)
	r.POST("/game/moves/unequip", s.handlePostUnequipMove)
	r.POST("/game/levelup", s.handlePostLevelUp)

	// Items
	r.POST("/game/items/equip", s.handlePostEquipItem)
	r.POST("/game/items/unequip", s.handlePostUnequipItem)
	r.POST("/game/items/use", s.handlePostUseItem)

	// Shop
	r.POST("/game/shop/buy", s.handlePostBuyItem)
	r.POST("/game/shop/sell", s.handlePostSellItem)

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

	c.SetCookie(gameSessionCookieName, strconv.Itoa(g.ID), oneWeekSecs, "/", "", false, true)
}

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

func (s *Server) handleGetItemsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "items.html", gin.H{})
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

func (s *Server) handlePostCreateNewGame(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database unavailable"})
		return
	}

	var req struct {
		HeroID    string `json:"hero_id"`
		IsEndless bool   `json:"is_endless"`
	}
	c.ShouldBindJSON(&req)

	// pick first hero by default as fallback
	hero := s.config.HeroTemplates[0]
	for _, h := range s.config.HeroTemplates {
		if h.ID == req.HeroID {
			hero = h
			break
		}
	}

	g := game.NewGame(s.config, hero)
	g.IsEndless = req.IsEndless
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
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}

	var req struct {
		RoomID string `json:"room_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id is required"})
		return
	}

	initialLog, err := g.EnterRoom(req.RoomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"game_state": g, "initial_log": initialLog})
}

func (s *Server) handlePostBattleMove(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}
	var req struct {
		MoveID string `json:"move_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "move_id is required"})
		return
	}
	result, err := g.SubmitPlayerMove(req.MoveID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if result.BattleOver && result.PlayerWon && result.WasBoss && g.IsEndless {
		g.AddRealm(s.config)
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) handlePostMonsterMove(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}
	result, err := g.SubmitMonsterMove()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if result.BattleOver && result.PlayerWon && result.WasBoss && g.IsEndless {
		g.AddRealm(s.config)
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) handlePostEquipMove(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}

	var req struct {
		MoveID string `json:"move_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "move_id is required"})
		return
	}

	if err := g.EquipMove(req.MoveID, s.config.Settings.MaxEquippedMoves); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, g)
}

func (s *Server) handlePostUnequipMove(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}
	var req struct {
		MoveID string `json:"move_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "move_id is required"})
		return
	}

	if err := g.UnequipMove(req.MoveID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, g)
}

// ==========================================
// ================= ITEMS ==================
// ==========================================

func (s *Server) handlePostEquipItem(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}
	if g.IsInBattle {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot equip items during battle"})
		return
	}
	var req struct {
		ItemID string `json:"item_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id is required"})
		return
	}
	if err := g.EquipItemFromPool(req.ItemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (s *Server) handlePostUnequipItem(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}
	if g.IsInBattle {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot unequip items during battle"})
		return
	}
	var req struct {
		ItemID string `json:"item_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id is required"})
		return
	}
	g.Player.UnequipItem(req.ItemID)
	g.Player.ItemPool = append(g.Player.ItemPool, req.ItemID)
	c.JSON(http.StatusOK, g)
}

func (s *Server) handlePostUseItem(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}
	if g.IsInBattle {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot use items during battle"})
		return
	}
	var req struct {
		ItemID string `json:"item_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id is required"})
		return
	}
	if err := g.UseItemFromPool(req.ItemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (s *Server) handlePostLevelUp(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}
	if g.PendingLevelUp == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no pending level up"})
		return
	}

	var req struct {
		Allocations map[string]int `json:"allocations"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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

	total := (pending.ManualPoints - spent) + pending.RandomPoints
	statPool := []models.StatType{models.HealthStat, models.AttackStat, models.DefenseStat, models.MagicStat, models.ManaStat}
	for i := 0; i < total; i++ {
		hero.LevelUpStat(statPool[rand.Intn(len(statPool))], 1)
	}

	g.PendingLevelUp = nil
	c.JSON(http.StatusOK, g)
}

// ==========================================
// ================= SHOP ===================
// ==========================================

func (s *Server) handlePostBuyItem(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}
	var req struct {
		ItemID string `json:"item_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id is required"})
		return
	}
	if err := g.BuyItem(req.ItemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (s *Server) handlePostSellItem(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}
	var req struct {
		ItemID string `json:"item_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id is required"})
		return
	}
	if err := g.SellItem(req.ItemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, g)
}

// ==========================================
// ================= SAVING =================
// ==========================================

func (s *Server) handlePostSave(c *gin.Context) {
	g, ok := s.gameFromRequest(c)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active game"})
		return
	}
	if err := s.db.SaveGame(g); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) handleGetListSaves(c *gin.Context) {
	saves, err := s.db.ListSaves()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, saves)
}

func (s *Server) handlePostLoadSave(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid save id"})
		return
	}
	g, err := s.db.LoadSave(id, s.config)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "save not found"})
		return
	}
	s.setSession(c, g)
	c.JSON(http.StatusOK, g)
}

func (s *Server) handleDeleteDeleteSave(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid save id"})
		return
	}
	if err := s.db.DeleteSave(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
