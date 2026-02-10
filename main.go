package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ctx         = context.Background()
	redisClient *redis.Client
)

type FTRequest struct {
	Days               int       `json:"days"`
	DataNormalized     bool      `json:"dataNormalized"`
	DataPeriod         string    `json:"dataPeriod"`
	DataInterval       int       `json:"dataInterval"`
	Realtime           bool      `json:"realtime"`
	YFormat            string    `json:"yFormat"`
	TimeServiceFormat  string    `json:"timeServiceFormat"`
	RulerIntradayStart int       `json:"rulerIntradayStart"`
	RulerIntradayStop  int       `json:"rulerIntradayStop"`
	RulerInterdayStart int       `json:"rulerInterdayStart"`
	RulerInterdayStop  int       `json:"rulerInterdayStop"`
	ReturnDateType     string    `json:"returnDateType"`
	Elements           []Element `json:"elements"`
}

type Element struct {
	Label             string   `json:"Label"`
	Type              string   `json:"Type"`
	Symbol            string   `json:"Symbol"`
	OverlayIndicators []string `json:"OverlayIndicators"`
	Params            struct{} `json:"Params"`
}

func handleMarketData(w http.ResponseWriter, r *http.Request) {
	ftId := r.URL.Query().Get("id")
	symbol := r.URL.Query().Get("symbol")
	startDateStr := r.URL.Query().Get("start")

	// Unique cache key
	cacheKey := fmt.Sprintf("market:%s:%s:%s", ftId, symbol, startDateStr)

	// Try to retrieve data from Redis
	cachedData, err := redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		fmt.Println("Serving from cache:", cacheKey)
		w.Header().Set("Content-Type", "application/json")
		w.Write(cachedData)
		return
	}

	url := "https://markets.ft.com/data/chartapi/series"

	// Calculate days to fetch
	daysToFetch := 180
	if startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		hours := time.Since(startDate).Hours()
		daysToFetch = int(hours / 24)

		if daysToFetch < 0 {
			http.Error(w, "Start date cannot be in the future", http.StatusBadRequest)
			return
		}
	}

	// Payload request params
	payloadObj := FTRequest{
		Days:               daysToFetch,
		DataNormalized:     false,
		DataPeriod:         "Day",
		DataInterval:       1,
		Realtime:           false,
		YFormat:            "0.###",
		TimeServiceFormat:  "JSON",
		RulerIntradayStart: 26,
		RulerIntradayStop:  3,
		RulerInterdayStart: 10957,
		RulerInterdayStop:  365,
		ReturnDateType:     "ISO8601",
		Elements: []Element{
			{
				Label:  "Price",
				Type:   "price",
				Symbol: ftId,
			},
			{
				Label:  "Volume",
				Type:   "volume",
				Symbol: ftId,
			},
		},
	}

	// Create request
	jsonData, _ := json.Marshal(payloadObj)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	// Set Request headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:147.0) Gecko/20100101 Firefox/147.0")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://markets.ft.com")
	req.Header.Set("Referer", "https://markets.ft.com/data/funds/tearsheet/charts?s="+symbol)
	// req.Header.Set("Cookie", `GZIP=1; consentUUID=; FTCookieConsentGDPR=true; spoor-id=; __RequestVerificationToken_L2RhdGE1=`)

	// Call FT API
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error contacting FT API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read response
	body, _ := io.ReadAll(resp.Body)

	// Store response in Redis for 24 hours
	if resp.StatusCode == http.StatusOK {
		err := redisClient.Set(ctx, cacheKey, body, 24*time.Hour).Err()
		if err != nil {
			log.Println("Redis save error:", err)
		}
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func main() {
	// Connect to Redis
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Start server
	http.HandleFunc("/api/market-data", handleMarketData)
	fmt.Println("Server starting at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
