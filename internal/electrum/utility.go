package electrum

import (
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strings"
)

type Node struct {
	Host     string
	Port     string
	Proto    int
	Version  string
	Protocol string
}

const (
	NODELIST_ADDR = "https://1209k.com/bitcoin-eye/ele.php?chain="
	EXPRESSION    = `<tr><td>(.*?)</td><td>(\d+)</td><td>(\w+)</td><td>([0-9a-f]+)</td><td>(\d+)</td><td>([\d\-: ]+)</td><td>(.*?)</td><td>([\d.]+)</td><td>(\w+)</td><td>([\d\-: ]+)</td><td>(\w+)</td><td>([\d. ]+)</td></tr>`
)

func getNodes(coin string) (*[]Node, error) {
	response, err := http.Get(NODELIST_ADDR +  strings.ToLower(coin))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(EXPRESSION)
	matches := re.FindAllStringSubmatch(string(body), -1)

	var nodelist []Node
	for _, match := range matches {

		if match[9] != "open" || match[11] != "OK" || strings.Contains(match[1], ".onion") {
			continue
		}

		if _, err := net.LookupHost(match[1]); err != nil {
			continue
		}

		nodelist = append(nodelist, Node{
			Host:     match[1],
			Port:     match[2],
			Proto:    protoStrToInt(match[3]),
			Version:  match[7],
			Protocol: match[8],
		})

	}

	return &nodelist, nil
}

func protoStrToInt(proto string) int {
	switch proto {
	case "ssl":
		return SSL
	case "tcp":
		return TCP
	}
	return -1
}

func GenID() int {
	return rand.Intn(99999999999)
}

func replaceNRs(input string) string {
	output := input
	output = strings.Replace(output, " ", "", -1)
	output = strings.Replace(output, "\n", "", -1)
	output = strings.Replace(output, "\r", "", -1)
	return output
}
