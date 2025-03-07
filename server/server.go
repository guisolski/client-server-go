package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type QuoteResponse struct {
	Bid string `json:"bid"`
}

func main() {
	db, err := sql.Open("sqlite3", "./quotes.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := createTableSQL(db); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", getQuoteHandler(db))

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getQuoteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelAPI := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancelAPI()

		bid, err := getQuote(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = insertQuote(db, bid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bid)
	}
}

func createTableSQL(db *sql.DB) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS quotes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bid TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		return err
	}
	return nil
}

func getQuote(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	usdbrl, ok := data["USDBRL"].(map[string]interface{})
	if !ok {
		return "", errors.New("api response missing udd-brl field")
	}

	bid, ok := usdbrl["bid"].(string)
	if !ok {
		return "", errors.New("bid field not found in usd-brl")
	}

	return bid, nil
}

func insertQuote(db *sql.DB, bid string) error {
	ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancelDB()

	_, err := db.ExecContext(ctxDB, "INSERT INTO quotes(bid) VALUES(?)", bid)
	return err
}
