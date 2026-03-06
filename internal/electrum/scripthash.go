package electrum

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"

	bch_chaincfg "github.com/gcash/bchd/chaincfg"
	bch_txscript "github.com/gcash/bchd/txscript"

	ltc_chaincfg "github.com/ltcsuite/ltcd/chaincfg"
	ltc_txscript "github.com/ltcsuite/ltcd/txscript"

	"github.com/ltcsuite/ltcd/ltcutil"
	"github.com/gcash/bchutil"
)

func getChainCfg(coin string) *chaincfg.Params {
	chain := chaincfg.MainNetParams

	coinConfigs := map[string]struct {
		PubKeyHashAddrID byte
		ScriptHashAddrID byte
		PrivateKeyID     byte
		Bech32HRPSegwit  string
	}{
		"doge": {0x1e, 0x16, 0x9e, ""},
		"bch":  {0x00, 0x05, 0x80, ""},
		"ltc":  {0x30, 0x32, 0xb0, "ltc"},
		"dash": {0x4c, 0x10, 0xcc, ""},
		"qtum": {0x3a, 0x32, 0x80, ""},
		"usdt": {0x00, 0x05, 0x80, ""},
	}

	if config, exists := coinConfigs[coin]; exists {
		chain.PubKeyHashAddrID = config.PubKeyHashAddrID
		chain.ScriptHashAddrID = config.ScriptHashAddrID
		chain.PrivateKeyID = config.PrivateKeyID
		if config.Bech32HRPSegwit != "" {
			chain.Bech32HRPSegwit = config.Bech32HRPSegwit
		}
	}

	return &chain
}

func addressToScriptHex(addressStr string, net string) (string, error) {
	if net == "bch" {
		address, err := bchutil.DecodeAddress(addressStr, &bch_chaincfg.MainNetParams)
		if err != nil {
			return "", err
		}

		script, err := bch_txscript.PayToAddrScript(address)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%x", script), nil
	}

	if net == "ltc" {
		address, err := ltcutil.DecodeAddress(addressStr, &ltc_chaincfg.MainNetParams)
		if err != nil {
			return "", err
		}

		script, err := ltc_txscript.PayToAddrScript(address)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%x", script), nil
	}

	address, err := btcutil.DecodeAddress(addressStr, getChainCfg(net))
	if err != nil {
		return "", err
	}
	script, err := txscript.PayToAddrScript(address)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", script), nil
}

func ScripthashAddr(addr string, net string) (string, error) {
	scriptHex, err := addressToScriptHex(addr, net)
	if err != nil {
		return "", err
	}

	sha256 := sha256.New()
	s, err := hex.DecodeString(scriptHex)
	if err != nil {
		return "", err
	}
	sha256.Write(s)
	return reverse(fmt.Sprintf("%x", sha256.Sum(nil))), nil
}

func reverse(s string) string {
	chars := []rune(s)
	var x string
	for i := len(chars); i > 0; i -= 2 {
		x += string(chars[i-2 : i])
	}
	return x
}
