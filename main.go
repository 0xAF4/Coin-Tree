package main

import (
	"log"
	"processor/internal/electrum"
	"slices"

	database "processor/internal/database"
	server   "processor/internal/server"

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

var (
	btc  *electrum.ElectrumNode
	doge *electrum.ElectrumNode
	bch  *electrum.ElectrumNode
	ltc  *electrum.ElectrumNode
)

func main() {
	log.Println("[--+] CoinPay Processor [+--]")

	/*--- Конфигурация ---*/
	conf := Config{}
	err  := l_env.FullFillStruct(&conf)
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

	/*--- Electrum(BTC) ---*/
	if slices.Contains(conf.App.Coins, "BTC") {
		btc, err = electrum.New("btc", conf.Logs.Common)
		initalize("Electrum(BTC)", err)
		if err == nil {
			defer btc.Close()
			btc.RunLoops()
		}
	}

	/*--- Electrum(DOGE) ---*/
	if slices.Contains(conf.App.Coins, "DOGE") {
		doge, err = electrum.New("doge", conf.Logs.Common)
		initalize("Electrum(DOGE)", err)
		if err == nil {
			defer doge.Close()
			doge.RunLoops()
		}
	}

	/*--- Electrum(BCH) ---*/
	if slices.Contains(conf.App.Coins, "BCH") {
		bch, err = electrum.New("bch", conf.Logs.Common)
		initalize("Electrum(BCH)", err)
		if err == nil {
			defer bch.Close()
			bch.RunLoops()
		}
	}

	/*--- Electrum(LTC) ---*/
	if slices.Contains(conf.App.Coins, "LTC") {
		ltc, err = electrum.New("ltc", conf.Logs.Common)
		initalize("Electrum(LTC)", err)
		if err == nil {
			defer ltc.Close()
			ltc.RunLoops()
		}
	}

	/*--- Web Server ---*/
	srv := server.New(conf.App.Port, db, btc, doge, bch, ltc)
	srv.SetHandlers()
	initalize("Web Server", nil)

	log.Printf("[*] Listening on %s | coins: %v\n", conf.App.Port, conf.App.Coins)
	if err := srv.Listen(); err != nil {
		log.Fatalf("[-] Server error: %v", err)
	}
}
