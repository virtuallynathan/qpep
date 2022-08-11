package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/parvit/qpep/shared"
)

func getClientForAPI(localAddr net.Addr) *http.Client {
	return &http.Client{
		Timeout: 500 * time.Millisecond,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				LocalAddr: localAddr,
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:    1,
			IdleConnTimeout: 10 * time.Second,
			//TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func RequestEcho(localAddress, address string, port int) *EchoResponse {
	addr := fmt.Sprintf("http://%s:%d%s", address, port, API_PREFIX_CLIENT+API_ECHO_PATH)

	log.Printf("local: %s - %s\n", localAddress, net.ParseIP(localAddress))
	client := getClientForAPI(&net.TCPAddr{
		IP: net.ParseIP(localAddress),
	})

	resp, err := client.Get(addr)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: BAD status code %d\n", resp.StatusCode)
		return nil
	}

	str := &bytes.Buffer{}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		str.WriteString(scanner.Text())
	}

	if scanner.Err() != nil {
		log.Printf("ERROR: %v\n", scanner.Err())
		return nil
	}

	if shared.QuicConfiguration.Verbose {
		log.Printf("%s\n", str.String())
	}

	respData := &EchoResponse{}
	jsonErr := json.Unmarshal(str.Bytes(), &respData)
	if jsonErr != nil {
		log.Printf("ERROR: %v\n", err)
		return nil
	}

	return respData
}

func RequestStatus(localAddress, gatewayAddress string, apiPort int, publicAddress string) *StatusReponse {
	apiPath := strings.Replace(API_PREFIX_CLIENT+API_STATUS_PATH, ":addr", publicAddress, -1)
	addr := fmt.Sprintf("http://%s:%d%s", gatewayAddress, apiPort, apiPath)

	client := getClientForAPI(&net.TCPAddr{
		IP: net.ParseIP(localAddress),
	})

	resp, err := client.Get(addr)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: BAD status code %d\n", resp.StatusCode)
		return nil
	}

	str := &bytes.Buffer{}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		str.WriteString(scanner.Text())
	}

	if scanner.Err() != nil {
		log.Printf("ERROR: %v\n", scanner.Err())
		return nil
	}

	if shared.QuicConfiguration.Verbose {
		log.Printf("%s\n", str.String())
	}

	respData := &StatusReponse{}
	jsonErr := json.Unmarshal(str.Bytes(), &respData)
	if jsonErr != nil {
		log.Printf("ERROR: %v\n", err)
		return nil
	}

	return respData
}

func RequestStatistics(localAddress, gatewayAddress string, apiPort int, publicAddress string) *StatsInfoReponse {
	apiPath := strings.Replace(API_PREFIX_SERVER+API_STATS_DATA_SRV_PATH, ":addr", publicAddress, -1)
	addr := fmt.Sprintf("http://%s:%d%s", gatewayAddress, apiPort, apiPath)

	client := getClientForAPI(&net.TCPAddr{
		IP: net.ParseIP(localAddress),
	})

	resp, err := client.Get(addr)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: BAD status code %d\n", resp.StatusCode)
		return nil
	}

	str := &bytes.Buffer{}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		str.WriteString(scanner.Text())
	}

	if scanner.Err() != nil {
		log.Printf("ERROR: %v\n", scanner.Err())
		return nil
	}

	if shared.QuicConfiguration.Verbose {
		log.Printf("%s\n", str.String())
	}

	respData := &StatsInfoReponse{}
	// collect stats

	jsonErr := json.Unmarshal(str.Bytes(), &respData)
	if jsonErr != nil {
		log.Printf("ERROR: %v\n", err)
		return nil
	}

	return respData
}
