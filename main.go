package main

import (
	"log"
	"processor/internal/electrum"
	"strings"

	database "processor/internal/database"
	server "processor/internal/server"

	l_env "github.com/0xAF4/lib_env"
)

type Config struct {
	App struct {
		Name    string   `env:"APP_NAME"`
		Version string   `env:"APP_VERSION"`
		Debug   bool     `env:"APP_DEBUG"`
		Port    string   `env:"APP_PORT"`
		Coins   []string `env:"APP_COINS"`
	}

	Database struct {
		Driver  string `env:"DB_DRIVER"`
		ConnStr string `env:"DB_CONNSTR"`
	}

	Telegram struct {
		Token  string   `env:"TG_TOKEN"`
		Admins []string `env:"TG_ADMINS"`
	}

	l_env.Logs
}

func initalize(name string, err error) {
	if err != nil {
		log.Fatalf("[-] ERROR %s: %v", name, err)
	} else {
		log.Println("[+] INITED:", name)
	}
}

func main() {
	log.Println("[--+] CoinPay Processor [+--]")

	/*--- Конфигурация ---*/
	conf := Config{}
	err := l_env.FullFillStruct(&conf)
	initalize("Configuration", err)

	if conf.App.Debug {
		l, err := l_env.InitLogsWithDir("C:\\Users\\Karim\\Documents\\Sources\\CoinPay-PaymentSystem\\logs")
		initalize("Configuration(Logs)", err)
		conf.Logs = *l
		defer l_env.Close(&conf.Logs)
	}

	/*--- PostgreSQL ---*/
	db, err := database.New(conf.Database.Driver, conf.Database.ConnStr, conf.Logs.DB)
	initalize("PostgreSQL", err)
	defer db.Close()

	nodes := map[string]*electrum.ElectrumNode{}

	for _, coin := range conf.App.Coins {
		node, err := electrum.New(strings.ToLower(coin), conf.Logs.Common)
		initalize("Electrum("+coin+")", err)
		if err == nil {
			defer node.Close()
			nodes[coin] = node
		}
	}

	/*--- Web Server ---*/
	srv := server.New(conf.App.Port, db, nodes)
	srv.SetHandlers()
	initalize("Web Server", nil)

	log.Printf("[*] Listening on %s | coins: %v\n", conf.App.Port, conf.App.Coins)
	if err := srv.Listen(); err != nil {
		log.Fatalf("[-] Server error: %v", err)
	}
}
