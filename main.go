package main

import (
	"log"
	"processor/internal/electrum"
	"slices"

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

var (
	doge *electrum.ElectrumNode
	bch  *electrum.ElectrumNode
	ltc  *electrum.ElectrumNode
)

func main() {
	log.Println("[--+] CoinPay [+--]")

	/*--- Инициализация конфгурации ---*/
	conf := Config{}
	err := l_env.FullFillStruct(&conf)
	initalize("Configuration", err)

	if conf.App.Debug {
		l, err := l_env.InitLogsWithDir("C:\\Users\\Karim\\Documents\\Sources\\CoinPay-PaymentSystem\\logs")
		initalize("Configuration(Logs)", err)
		conf.Logs = *l
		defer l_env.Close(&conf.Logs)
	}

	/*--- Инициализация БД PostgreSQL ---*/
	db, err := database.New(conf.Database.Driver, conf.Database.ConnStr, conf.Logs.DB)
	initalize("PostgreSQL", err)
	defer db.Close()

	/*--- Инициализация Electrum(DOGE) нод ---*/
	if slices.Contains(conf.App.Coins, "BTC") {
		doge, err = electrum.New("btc", conf.Logs.Common)
		initalize("Electrum(BTC)", err)
		defer doge.Close()
		doge.RunLoops()
	}

	/*--- Инициализация Electrum(DOGE) нод ---*/
	if slices.Contains(conf.App.Coins, "DOGE") {
		doge, err = electrum.New("doge", conf.Logs.Common)
		initalize("Electrum(DOGE)", err)
		defer doge.Close()
		doge.RunLoops()
	}

	/*--- Инициализация Electrum(BCH) нод ---*/
	if slices.Contains(conf.App.Coins, "BCH") {
		bch, err = electrum.New("bch", conf.Logs.Common)
		initalize("Electrum(BCH)", err)
		defer bch.Close()
		bch.RunLoops()
	}

	/*--- Инициализация Electrum(LTC) нод ---*/
	if slices.Contains(conf.App.Coins, "LTC") {
		ltc, err = electrum.New("ltc", conf.Logs.Common)
		initalize("Electrum(LTC)", err)
		defer ltc.Close()
		ltc.RunLoops()
	}

	/*--- Создание и запуск вебсервера ---*/
	srv := server.New(conf.App.Port, db, doge, bch, ltc)
	srv.SetHandlers()
	initalize("Web Server", nil)
	srv.Listen()
}
