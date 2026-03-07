package server

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ─────────────────────────────────────────────────────────────────────────────
// GET /getBalance?coin=XXX&addr=YYY[&view_key=ZZZ]
// ─────────────────────────────────────────────────────────────────────────────

func (s *Server) getBalance(c *fiber.Ctx) error {
	coin := strings.ToUpper(c.Query("coin"))
	addr := c.Query("addr")

	if coin == "" || addr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": "coin and addr are required",
		})
	}

	balance, err := s.Nodes[coin].GetBalance(addr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true, "coin": coin, "address": addr, "balance": balance,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /getFee?coin=XXX
// ─────────────────────────────────────────────────────────────────────────────

func (s *Server) getFee(c *fiber.Ctx) error {
	coin := strings.ToUpper(c.Query("coin"))
	if coin == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": "coin is required",
		})
	}

	fee, err := s.Nodes[coin].GetFee()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true, "coin": coin, "fee": fee,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /getTransactions?coin=XXX&addr=YYY[&limit=N]
// Возвращает список транзакций с адресами контрагентов (для графа).
// ─────────────────────────────────────────────────────────────────────────────

func (s *Server) getTransactions(c *fiber.Ctx) error {
	coin := strings.ToUpper(c.Query("coin"))
	addr := c.Query("addr")
	limit := c.QueryInt("limit", 50)

	if coin == "" || addr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": "coin and addr are required",
		})
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	if s.Nodes[coin] == nil {
		return errNodeDisabled(c, coin)
	}

	txs, err := s.Nodes[coin].GetAddressTransactions(addr, limit)
	if err != nil {
		return errInternal(c, err)
	}
	return c.JSON(fiber.Map{"success": true, "coin": coin, "address": addr, "transactions": txs})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /buildGraph
// Body: { "coin": "BCH", "address": "q...", "limit": 50 }
// Теперь возвращает реальные TxRecord вместо сырого UniversalRPC.
// ─────────────────────────────────────────────────────────────────────────────

type BuildGraphReq struct {
	Coin    string `json:"coin"`
	Address string `json:"address"`
	Limit   int    `json:"limit"`
}

func (s *Server) buildGraph(c *fiber.Ctx) error {
	var req BuildGraphReq
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": "invalid JSON",
		})
	}

	req.Coin = strings.ToUpper(req.Coin)
	if req.Address == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": "address is required",
		})
	}
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 50
	}

	if s.Nodes[req.Coin] == nil {
		return errNodeDisabled(c, req.Coin)
	}

	txs, err := s.Nodes[req.Coin].GetAddressTransactions(req.Address, req.Limit)
	if err != nil {
		return errInternal(c, err)
	}
	return c.JSON(fiber.Map{"success": true, "transactions": txs})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /getChainInfo?coin=XXX
// Возвращает высоту цепи из БД.
// ─────────────────────────────────────────────────────────────────────────────

func (s *Server) getChainInfo(c *fiber.Ctx) error {
	coin := strings.ToUpper(c.Query("coin"))
	if coin == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": "coin is required",
		})
	}

	height, hash, err := s.DB.GetChainHeight(coin)
	if err != nil {
		return errInternal(c, err)
	}
	return c.JSON(fiber.Map{
		"success": true, "coin": coin,
		"height": height, "last_hash": hash,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /getNodes?coin=XXX
// Список P2P нод из БД.
// ─────────────────────────────────────────────────────────────────────────────

func (s *Server) getNodes(c *fiber.Ctx) error {
	coin := strings.ToUpper(c.Query("coin"))
	if coin == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": "coin is required",
		})
	}

	nodes, err := s.DB.GetNodeAddrs(coin)
	if err != nil {
		return errInternal(c, err)
	}
	return c.JSON(fiber.Map{
		"success": true, "coin": coin, "nodes": nodes,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func errNodeDisabled(c *fiber.Ctx, coin string) error {
	return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
		"success": false, "msg_txt": coin + " node is not initialized (check APP_COINS)",
	})
}

func errInternal(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false, "msg_txt": err.Error(),
	})
}
