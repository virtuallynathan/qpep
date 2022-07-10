package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/parvit/qpep/shared"
)

func RequestEcho() *EchoResponse {
	addr := fmt.Sprintf("http://%s:%d/%s", shared.QuicConfiguration.GatewayIP, shared.QuicConfiguration.GatewayAPIPort, API_ECHO_PATH)

	resp, err := http.DefaultClient.Get(addr)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	bodyData, errData := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if errData != nil {
		log.Printf("ERROR: %v\n", err)
		return nil
	}

	respData := &EchoResponse{}
	jsonErr := json.Unmarshal(bodyData, &respData)
	if jsonErr != nil {
		log.Printf("ERROR: %v\n", err)
		return nil
	}

	return respData
}
