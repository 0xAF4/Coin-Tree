package server

import (
	"processor/internal/database"
	elx "processor/internal/electrum"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	Port     string
	DB       *database.Database
	App      *fiber.App
	DogeNode *elx.ElectrumNode
	BCHNode  *elx.ElectrumNode
	LTCNode  *elx.ElectrumNode
}

func New(p string, d *database.Database, doge *elx.ElectrumNode, bch *elx.ElectrumNode, ltc *elx.ElectrumNode) *Server {
	return &Server{
		Port:     p,
		DB:       d,
		App:      fiber.New(fiber.Config{DisableStartupMessage: true}),
		DogeNode: doge,
		BCHNode:  bch,
		LTCNode:  ltc,
	}
}

func (s *Server) SetHandlers() {
	s.App.Get("/getBalance", s.getBalance)
	s.App.Get("/getFee", s.getFee)
	s.App.Post("/buildGraph", s.BuildGraph)
}

func (s *Server) Listen() {
	s.App.Listen(s.Port)
}
