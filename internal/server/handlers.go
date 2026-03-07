package server

import (
	"encoding/json"
	"errors"
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

	var (
		balance interface{}
		err     error
	)

	switch coin {
	// ── Electrum ──────────────────────────────────────────────────────────────
	case "BTC":
		if s.BTCNode == nil {
			return errNodeDisabled(c, coin)
		}
		balance, err = s.BTCNode.GetBalance(addr)
	case "DOGE":
		if s.DogeNode == nil {
			return errNodeDisabled(c, coin)
		}
		balance, err = s.DogeNode.GetBalance(addr)
	case "BCH":
		if s.BCHNode == nil {
			return errNodeDisabled(c, coin)
		}
		balance, err = s.BCHNode.GetBalance(addr)
	case "LTC":
		if s.LTCNode == nil {
			return errNodeDisabled(c, coin)
		}
		balance, err = s.LTCNode.GetBalance(addr)

	// ── EVM ───────────────────────────────────────────────────────────────────
	case "ETH":
		balance, err = getETHBalance(addr)
	case "BNB":
		balance, err = getBNBBalance(addr)
	case "ARB":
		balance, err = getARBBalance(addr)
	case "AVAX":
		balance, err = getAVAXBalance(addr)
	case "ETC":
		balance, err = getETCBalance(addr)

	// ── Other ─────────────────────────────────────────────────────────────────
	case "SOL":
		balance, err = getSOLBalance(addr)
	case "TRX":
		balance, err = getTRXBalance(addr)
	case "ALGO":
		balance, err = getALGOBalance(addr)
	case "ADA":
		balance, err = getADABalance(addr)

	// ── DB-stored (XMR и другие через chain.GET_BALANCE) ─────────────────────
	case "XMR":
		viewKey := c.Query("view_key")
		var vk *string
		if viewKey != "" {
			vk = &viewKey
		}
		balance, err = s.DB.GetBalance(coin, addr, vk)

	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": "UNDEFINED_COIN: " + coin,
		})
	}

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

	var (
		fee interface{}
		err error
	)

	switch coin {
	// Electrum nodes возвращают fee через estimatefee
	case "BTC":
		if s.BTCNode != nil {
			fee, err = s.BTCNode.GetFee()
		} else {
			fee, err = getBTCFee()
		}
	case "DOGE":
		if s.DogeNode != nil {
			fee, err = s.DogeNode.GetFee()
		} else {
			err = errors.New("DOGE node not initialized")
		}
	case "BCH":
		if s.BCHNode != nil {
			fee, err = s.BCHNode.GetFee()
		} else {
			err = errors.New("BCH node not initialized")
		}
	case "LTC":
		if s.LTCNode != nil {
			fee, err = s.LTCNode.GetFee()
		} else {
			err = errors.New("LTC node not initialized")
		}
	case "XMR":
		fee, err = getXMRFee()
	case "ETH":
		fee, err = getETHFee()
	case "SOL":
		fee, err = getSOLFee()
	case "TRX":
		fee, err = getTRXFee()
	case "AVAX":
		fee, err = getAVAXFee()
	case "ARB":
		fee, err = getARBFee()
	case "BNB":
		fee, err = getBNBFee()
	case "ETC":
		fee, err = getETCFee()
	case "ALGO":
		fee, err = getALGOFee()
	case "ADA":
		fee, err = getADAFee()
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": "UNDEFINED_COIN: " + coin,
		})
	}

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

	switch coin {
	case "BTC":
		if s.BTCNode == nil {
			return errNodeDisabled(c, coin)
		}
		txs, err := s.BTCNode.GetAddressTransactions(addr, limit)
		if err != nil {
			return errInternal(c, err)
		}
		return c.JSON(fiber.Map{"success": true, "coin": coin, "address": addr, "transactions": txs})

	case "DOGE":
		if s.DogeNode == nil {
			return errNodeDisabled(c, coin)
		}
		txs, err := s.DogeNode.GetAddressTransactions(addr, limit)
		if err != nil {
			return errInternal(c, err)
		}
		return c.JSON(fiber.Map{"success": true, "coin": coin, "address": addr, "transactions": txs})

	case "BCH":
		if s.BCHNode == nil {
			return errNodeDisabled(c, coin)
		}
		txs, err := s.BCHNode.GetAddressTransactions(addr, limit)
		if err != nil {
			return errInternal(c, err)
		}
		return c.JSON(fiber.Map{"success": true, "coin": coin, "address": addr, "transactions": txs})

	case "LTC":
		if s.LTCNode == nil {
			return errNodeDisabled(c, coin)
		}
		txs, err := s.LTCNode.GetAddressTransactions(addr, limit)
		if err != nil {
			return errInternal(c, err)
		}
		return c.JSON(fiber.Map{"success": true, "coin": coin, "address": addr, "transactions": txs})

	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": "getTransactions not supported for " + coin + " yet",
		})
	}
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

	switch req.Coin {
	case "BTC":
		if s.BTCNode == nil {
			return errNodeDisabled(c, req.Coin)
		}
		txs, err := s.BTCNode.GetAddressTransactions(req.Address, req.Limit)
		if err != nil {
			return errInternal(c, err)
		}
		return c.JSON(fiber.Map{"success": true, "transactions": txs})

	case "DOGE":
		if s.DogeNode == nil {
			return errNodeDisabled(c, req.Coin)
		}
		txs, err := s.DogeNode.GetAddressTransactions(req.Address, req.Limit)
		if err != nil {
			return errInternal(c, err)
		}
		return c.JSON(fiber.Map{"success": true, "transactions": txs})

	case "BCH":
		if s.BCHNode == nil {
			return errNodeDisabled(c, req.Coin)
		}
		txs, err := s.BCHNode.GetAddressTransactions(req.Address, req.Limit)
		if err != nil {
			return errInternal(c, err)
		}
		return c.JSON(fiber.Map{"success": true, "transactions": txs})

	case "LTC":
		if s.LTCNode == nil {
			return errNodeDisabled(c, req.Coin)
		}
		txs, err := s.LTCNode.GetAddressTransactions(req.Address, req.Limit)
		if err != nil {
			return errInternal(c, err)
		}
		return c.JSON(fiber.Map{"success": true, "transactions": txs})

	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": "buildGraph not supported for " + req.Coin,
		})
	}
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
