package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	serverResp, err := getQuote(ctx)
	if err != nil {
		log.Fatal("Error getting quote:", err)
	}

	err = os.WriteFile("cotacao.txt", []byte(fmt.Sprintf("Dólar: %s", serverResp)), 0644)
	if err != nil {
		log.Fatal("Error writing to file:", err)
	}

	fmt.Println("Cotação salva em cotacao.txt")
}

func getQuote(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var bid string
	if err := json.NewDecoder(resp.Body).Decode(&bid); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return bid, nil
}
