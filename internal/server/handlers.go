package server

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) getBalance(c *fiber.Ctx) error {
	var (
		balance interface{}
		err     error
	)

	coin := c.Query("coin")
	addr := c.Query("addr")

	switch coin {
	case "DOGE":
		balance, err = s.DogeNode.GetBalance(addr)
	case "BCH":
		balance, err = s.BCHNode.GetBalance(addr)
	case "LTC":
		balance, err = s.LTCNode.GetBalance(addr)
	default:
		err = errors.New("ENDEFINED_COIN")
	}

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "msg_txt": err.Error()})
	} else {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "balance": balance})
	}
}

func (s *Server) getFee(c *fiber.Ctx) error {
	var (
		fee interface{}
		err error
	)

	coin := c.Query("coin")

	switch coin {
	case "DOGE":
		fee, err = s.DogeNode.GetFee()
	case "BCH":
		fee, err = s.BCHNode.GetFee()
	case "LTC":
		fee, err = s.LTCNode.GetFee()
	default:
		err = errors.New("ENDEFINED_COIN")
	}

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "msg_txt": err.Error()})
	} else {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "fee": fee})
	}
}

type BuildGraphStruct struct {
	Coin string `json:"coin"`
	Addr string `json:"address"`
}

func (s *Server) BuildGraph(c *fiber.Ctx) error {
	data := new(BuildGraphStruct)
	if err := json.Unmarshal(c.Body(), data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": "Invalid JSON data",
		})
	}

	trxs, err := s.BCHNode.GetTransactions(data.Addr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "msg_txt": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true, "msg_txt": "The request has been processed",
		"transactions": trxs,
	})
}
