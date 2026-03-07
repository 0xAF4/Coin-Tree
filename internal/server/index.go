package server

import (
	"processor/internal/database"
	elx "processor/internal/electrum"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	Port     string
	DB       *database.Database
	App      *fiber.App
	BTCNode  *elx.ElectrumNode
	DogeNode *elx.ElectrumNode
	BCHNode  *elx.ElectrumNode
	LTCNode  *elx.ElectrumNode
}

func New(
	port string,
	db *database.Database,
	btc *elx.ElectrumNode,
	doge *elx.ElectrumNode,
	bch *elx.ElectrumNode,
	ltc *elx.ElectrumNode,
) *Server {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		// Читаемые JSON-ошибки вместо HTML
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false, "msg_txt": err.Error(),
			})
		},
	})

	return &Server{
		Port:     port,
		DB:       db,
		App:      app,
		BTCNode:  btc,
		DogeNode: doge,
		BCHNode:  bch,
		LTCNode:  ltc,
	}
}

func (s *Server) SetHandlers() {
	// ── Middleware ─────────────────────────────────────────────────────────────
	s.App.Use(recover.New())
	s.App.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))
	s.App.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} ${latency} ${method} ${path}\n",
	}))

	// ── Healthcheck ────────────────────────────────────────────────────────────
	s.App.Get("/health", func(c *fiber.Ctx) error {
		nodes := fiber.Map{}
		if s.BTCNode != nil {
			nodes["BTC"] = "ok"
		}
		if s.DogeNode != nil {
			nodes["DOGE"] = "ok"
		}
		if s.BCHNode != nil {
			nodes["BCH"] = "ok"
		}
		if s.LTCNode != nil {
			nodes["LTC"] = "ok"
		}
		return c.JSON(fiber.Map{"success": true, "nodes": nodes})
	})

	// ── Chain ──────────────────────────────────────────────────────────────────
	s.App.Get("/getBalance", s.getBalance)
	s.App.Get("/getFee", s.getFee)
	s.App.Get("/getTransactions", s.getTransactions)
	s.App.Get("/getChainInfo", s.getChainInfo)
	s.App.Get("/getNodes", s.getNodes)

	// ── Graph ──────────────────────────────────────────────────────────────────
	s.App.Post("/buildGraph", s.buildGraph)
}

func (s *Server) Listen() error {
	return s.App.Listen(s.Port)
}
