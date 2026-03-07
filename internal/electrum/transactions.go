package electrum

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// HistoryEntry — одна запись из blockchain.scripthash.get_history
type HistoryEntry struct {
	TxHash string `json:"tx_hash"`
	Height int64  `json:"height"`
}

// VerboseTx — декодированная транзакция из blockchain.transaction.get (verbose=true)
type VerboseTx struct {
	TxID          string   `json:"txid"`
	Confirmations int      `json:"confirmations"`
	BlockHeight   int64    `json:"blockheight"`
	Vin           []TxVin  `json:"vin"`
	Vout          []TxVout `json:"vout"`
}

type TxVin struct {
	TxID string `json:"txid"`
	Vout int    `json:"vout"`
}

type TxVout struct {
	Value        float64      `json:"value"`
	N            int          `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

type ScriptPubKey struct {
	Addresses []string `json:"addresses"`
	Address   string   `json:"address"` // новый формат Electrum (единственный адрес)
}

// TxRecord — итоговая запись транзакции для графа
type TxRecord struct {
	TxHash        string `json:"txHash"`
	Address       string `json:"address"` // адрес контрагента
	Value         int64  `json:"value"`   // в сатоши
	Block         string `json:"block"`   // строка для совместимости с фронтом
	Height        int64  `json:"height"`
	Confirmations int    `json:"confirmations"`
	Direction     string `json:"direction"` // "in" | "out"
}

// GetHistory возвращает историю транзакций адреса (только хеши + высоты).
func (e *ElectrumNode) GetHistory(addr string) ([]HistoryEntry, error) {
	rpc, err := e.execRPC(GET_TXCOUNT, []string{addr})
	if err != nil {
		return nil, err
	}

	raw, ok := rpc.Response.Result.([]interface{})
	if !ok {
		return nil, errors.New("unexpected history result type")
	}

	var entries []HistoryEntry
	for _, item := range raw {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		txHash, _ := m["tx_hash"].(string)
		height, _ := m["height"].(float64)
		if txHash != "" {
			entries = append(entries, HistoryEntry{
				TxHash: txHash,
				Height: int64(height),
			})
		}
	}
	return entries, nil
}

// GetRawTransaction получает сырую транзакцию по хешу.
// verbose=false → hex-строка
func (e *ElectrumNode) GetRawTransaction(txHash string) (interface{}, error) {
	return e.execRPC("blockchain.transaction.get", []string{txHash, "false"})
}

// GetVerboseTransaction получает транзакцию в verbose-режиме (JSON).
// Не все серверы поддерживают verbose; если нет — парсим вручную.
func (e *ElectrumNode) GetVerboseTransaction(txHash string) (*VerboseTx, error) {
	rpc, err := e.execRPC("blockchain.transaction.get", []string{txHash, "true"})
	if err != nil {
		return nil, err
	}

	m, ok := rpc.Response.Result.(map[string]interface{})
	if !ok {
		return nil, errors.New("verbose tx: unexpected result type")
	}

	tx := &VerboseTx{}
	tx.TxID, _ = m["txid"].(string)
	if c, ok := m["confirmations"].(float64); ok {
		tx.Confirmations = int(c)
	}
	if bh, ok := m["blockheight"].(float64); ok {
		tx.BlockHeight = int64(bh)
	}

	// vout
	if vouts, ok := m["vout"].([]interface{}); ok {
		for _, v := range vouts {
			vm, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			val, _ := vm["value"].(float64)
			n, _ := vm["n"].(float64)
			vout := TxVout{Value: val, N: int(n)}

			if spk, ok := vm["scriptPubKey"].(map[string]interface{}); ok {
				if addr, ok := spk["address"].(string); ok {
					vout.ScriptPubKey.Address = addr
				}
				if addrs, ok := spk["addresses"].([]interface{}); ok {
					for _, a := range addrs {
						if s, ok := a.(string); ok {
							vout.ScriptPubKey.Addresses = append(vout.ScriptPubKey.Addresses, s)
						}
					}
				}
			}
			tx.Vout = append(tx.Vout, vout)
		}
	}

	// vin
	if vins, ok := m["vin"].([]interface{}); ok {
		for _, v := range vins {
			vm, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			txid, _ := vm["txid"].(string)
			vout, _ := vm["vout"].(float64)
			tx.Vin = append(tx.Vin, TxVin{TxID: txid, Vout: int(vout)})
		}
	}

	return tx, nil
}

// GetAddressTransactions — главная функция для графа.
// Возвращает список TxRecord для заданного адреса.
// Лимит — максимальное количество транзакций для обработки (чтобы не висеть).
func (e *ElectrumNode) GetAddressTransactions(addr string, limit int) ([]TxRecord, error) {
	scripthash, err := ScripthashAddr(addr, e.net)
	if err != nil {
		return nil, fmt.Errorf("scripthash: %w", err)
	}

	// 1. Получаем историю
	history, err := e.GetHistoryByScripthash(scripthash)
	if err != nil {
		return nil, fmt.Errorf("history: %w", err)
	}

	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}

	var records []TxRecord

	for _, entry := range history {
		// 2. Пробуем verbose транзакцию
		vtx, err := e.GetVerboseTransaction(entry.TxHash)
		if err != nil {
			// Сервер не поддерживает verbose — добавляем запись без адреса контрагента
			records = append(records, TxRecord{
				TxHash:    entry.TxHash,
				Address:   addr,
				Value:     0,
				Block:     fmt.Sprintf("%d", entry.Height),
				Height:    entry.Height,
				Direction: "unknown",
			})
			continue
		}

		// 3. Первый проход: определяем роль нашего адреса в этой транзакции.
		//    addrInVout=true  → наш адрес есть среди получателей (incoming или сдача)
		//    addrInVin=true   → наш адрес был в одном из входящих tx (outgoing)
		//
		//    Правило:
		//      addrInVout && !addrInVin  → чисто входящая (мы получили)
		//      addrInVin                 → исходящая (мы отправили), даже если есть сдача
		var totalReceived int64
		var totalSent int64

		// Собираем все адреса vout с суммами
		type voutEntry struct {
			addr     string
			satoshis int64
		}
		var vouts []voutEntry

		for _, vout := range vtx.Vout {
			satoshis := int64(vout.Value * 1e8)
			addrs := make([]string, 0, 2)
			if vout.ScriptPubKey.Address != "" {
				addrs = append(addrs, vout.ScriptPubKey.Address)
			}
			addrs = append(addrs, vout.ScriptPubKey.Addresses...)
			for _, a := range addrs {
				vouts = append(vouts, voutEntry{a, satoshis})
			}
		}

		// Считаем нашу сумму в vout
		for _, v := range vouts {
			if v.addr == addr {
				totalReceived += v.satoshis
			}
		}

		// addrInVin: ищем наш адрес среди входов.
		// Для этого нужно получить родительские tx по Vin[].TxID — дорого,
		// поэтому используем эвристику: если totalReceived == 0, мы явно не получатель,
		// значит мы отправитель. Если totalReceived > 0 и нет других выходов — чисто incoming.
		addrInVout := totalReceived > 0

		// Считаем сумму чужих выходов
		for _, v := range vouts {
			if v.addr != addr {
				totalSent += v.satoshis
			}
		}

		// Определяем direction:
		//   - Есть чужие выходы AND мы в vout (сдача) → outgoing (мы платили, получили сдачу)
		//   - Есть чужие выходы AND нас нет в vout → outgoing
		//   - Нет чужих выходов AND мы в vout → incoming (coinbase или перевод только нам)
		//   - Нас нет нигде → unknown
		var txDir string
		switch {
		case !addrInVout && totalSent == 0:
			txDir = "unknown"
		case !addrInVout:
			// мы не получатель, но есть чужие выходы — значит мы отправитель
			txDir = "out"
		case addrInVout && totalSent == 0:
			// только наш адрес в выходах — входящий перевод
			txDir = "in"
		default:
			// есть и наш адрес (сдача), и чужие — исходящая
			txDir = "out"
		}

		// 4. Генерируем TxRecord для каждого контрагента
		seen := make(map[string]bool)
		for _, v := range vouts {
			if v.addr == addr {
				continue // пропускаем свой адрес (сдача)
			}
			if seen[v.addr] {
				continue // дедупликация по адресу в рамках одной tx
			}
			seen[v.addr] = true

			records = append(records, TxRecord{
				TxHash:        entry.TxHash,
				Address:       v.addr,
				Value:         v.satoshis,
				Block:         fmt.Sprintf("%d", vtx.BlockHeight),
				Height:        vtx.BlockHeight,
				Confirmations: vtx.Confirmations,
				Direction:     txDir,
			})
		}

		// Если нет контрагентов (только наш адрес в выходах — coinbase/self-transfer)
		if len(seen) == 0 && addrInVout {
			records = append(records, TxRecord{
				TxHash:        entry.TxHash,
				Address:       addr,
				Value:         totalReceived,
				Block:         fmt.Sprintf("%d", vtx.BlockHeight),
				Height:        vtx.BlockHeight,
				Confirmations: vtx.Confirmations,
				Direction:     "in",
			})
		}
	}

	return records, nil
}

// GetHistoryByScripthash — история по scripthash (уже вычисленному).
func (e *ElectrumNode) GetHistoryByScripthash(scripthash string) ([]HistoryEntry, error) {
	rpc := UniversalRPC{
		ID: GenID(), ST: Created,
		Request: Request{
			Method: GET_TXCOUNT,
			Params: []string{scripthash},
		},
	}
	e.pull.WriteInPull(&rpc)

	for slices.Contains([]Status{Created, Sended}, rpc.ST) {
		time.Sleep(time.Millisecond * 100)
	}
	if rpc.ST == Error {
		return nil, errors.New("RPC Error")
	}

	raw, ok := rpc.Response.Result.([]interface{})
	if !ok {
		return nil, errors.New("unexpected history format")
	}

	var out []HistoryEntry
	for _, item := range raw {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		txHash, _ := m["tx_hash"].(string)
		height, _ := m["height"].(float64)
		if txHash != "" {
			out = append(out, HistoryEntry{TxHash: txHash, Height: int64(height)})
		}
	}
	return out, nil
}

// execRPC — вспомогательный метод: отправляет RPC и ждёт ответа.
func (e *ElectrumNode) execRPC(method string, params []string) (*UniversalRPC, error) {
	rpc := &UniversalRPC{
		ID: GenID(), ST: Created,
		Request: Request{Method: method, Params: params},
	}
	e.pull.WriteInPull(rpc)

	for slices.Contains([]Status{Created, Sended}, rpc.ST) {
		time.Sleep(time.Millisecond * 100)
	}
	if rpc.ST == Error {
		return nil, fmt.Errorf("RPC error: method=%s", method)
	}
	return rpc, nil
}
