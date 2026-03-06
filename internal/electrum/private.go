package electrum

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"math/rand"
	"net"
	"strings"
	"time"
)

func (e *ElectrumNode) connect() error {
	e.mus.Lock()
	defer e.mus.Unlock()

	node := e.getRandomNode()

	d := net.Dialer{Timeout: time.Second * 5}
	conn, err := d.Dial("tcp", node.Host+":"+node.Port)
	if err != nil {
		return err
	}

	if node.Proto == SSL {
		sslConn := tls.Client(conn, &tls.Config{InsecureSkipVerify: true})
		if err = sslConn.Handshake(); err != nil {
			return err
		}
		e.Conn = sslConn
	} else {
		e.Conn = conn
	}
	e.connected = true

	e.log.Println("Connected to the Node:", node)
	return nil
}

func (e *ElectrumNode) getRandomNode() *Node {
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(*e.Nodes))
	return &(*e.Nodes)[randomIndex]
}

func (e *ElectrumNode) writer() {
	for !e.destroy {
		time.Sleep(time.Millisecond * 100)
		e.pull.Iterate(func(u *UniversalRPC) {
			if u.ST == Created {
				if err := e.sendRequest(u.BuildPayload()); err != nil {
					e.log.Println("Send request error:", err.Error())
					u.ST = Error
				} else {
					e.log.Println("Sended request:", string(u.BuildPayload()))
					u.ST = Sended
				}
			}
		})
	}
}

func (e *ElectrumNode) reader() {
	for !e.destroy {
		if !e.connected {
			e.connect()
		} else {
			buf := make([]byte, 256)

			rCount, err := e.Conn.Read(buf)
			if err != nil {
				e.connected = false
				e.Conn.Close()
				e.log.Println("Error:", err.Error())
			} else {
				e.handleResponse(buf[:rCount])
			}
		}

		if e.destroy {
			e.Conn.Close()
		}
	}
}

// TODO: need to find "RPC Error"
func (e *ElectrumNode) handleResponse(data []byte) {
	strData := replaceNRs(string(data))
	e.log.Println("Response text:", strData)

	e.strBuffer += strData

	var currentVar strings.Builder
	openBraceCount := 0
	for _, char := range e.strBuffer {
		switch char {
		case '{':
			openBraceCount++
			currentVar.WriteRune(char)
		case '}':
			openBraceCount--
			currentVar.WriteRune(char)
			if openBraceCount == 0 {
				var response Response
				if err := json.Unmarshal([]byte(currentVar.String()), &response); err != nil {
					e.log.Println("Unmarshaling json error:", currentVar.String())
				} else {
					u := e.pull.ReadFromPull(response.ID)
					if u.Method == GET_BALANCE && !response.IsOkResult() {
						e.pull.ResetRecord(u.ID)
					} else {
						u.Response = response
						u.ST = Done
					}
				}
				currentVar.Reset()
			}
		default:
			if openBraceCount > 0 {
				currentVar.WriteRune(char)
			}
		}
	}
	e.strBuffer = currentVar.String()
}

func (e *ElectrumNode) sendRequest(payload []byte) error {
	e.mus.Lock()
	defer e.mus.Unlock()

	if e.Conn == nil {
		return errors.New("Socket is not connected")
	}

	if _, err := e.Conn.Write(append(payload, '\n')); err != nil {
		e.Conn.Close()
		e.Conn = nil
		return errors.New("Write Error")
	}

	return nil
}
