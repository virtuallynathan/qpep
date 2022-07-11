package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/parvit/qpep/shared"
)

func RequestEcho(address string, port int) *EchoResponse {
	addr := fmt.Sprintf("http://%s:%d%s", address, port, API_ECHO_PATH)

	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

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

func RequestStatus(gatewayAddress string, apiPort int, address string) *StatusReponse {
	apiPath := strings.Replace(API_STATUS_PATH, ":addr", address, -1)
	addr := fmt.Sprintf("http://%s:%d%s", gatewayAddress, apiPort, apiPath)

	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

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
