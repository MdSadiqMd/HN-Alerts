package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/MdSadiqMd/HN-Alerts/internal"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare"
	"github.com/syumai/workers/cloudflare/cron"
)

func processHNAlerts(ctx context.Context) error {
	// Create a dummy request for the context
	req, err := http.NewRequestWithContext(ctx, "GET", "/hn-alerts", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	top20Stories, err := internal.FetchHNTopN(req, 20)
	if err != nil {
		return fmt.Errorf("error fetching from HN: %w", err)
	}
	fmt.Printf("Fetched %d stories from HN\n", len(top20Stories))

	uniqueTop10, err := internal.GetHNTop10FromKV(req, top20Stories)
	if err != nil {
		return fmt.Errorf("error getting top 10 from KV: %w", err)
	}

	respStr, err := internal.MakeBotMessage(req, uniqueTop10, cloudflare.Getenv("GREEN_API_URL"))
	if err != nil {
		return fmt.Errorf("failed to make bot message: %w", err)
	}
	fmt.Printf("Message sent to WhatsApp. ID: %s\n", respStr)
	return nil
}

func main() {
	// Register scheduled task for cron trigger
	cron.ScheduleTask(func(ctx context.Context) error {
		fmt.Println("Cron trigger fired - processing HN alerts...")
		return processHNAlerts(ctx)
	})

	http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		msg := "Hello!"
		w.Write([]byte(msg))
	})
	http.HandleFunc("/echo", func(w http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		io.Copy(w, bytes.NewReader(b))
	})

	http.HandleFunc("/hn-alerts", func(w http.ResponseWriter, req *http.Request) {
		err := processHNAlerts(req.Context())
		if err != nil {
			fmt.Println("Error processing HN alerts:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
			return
		}

		w.Write([]byte("Msg sent to whatsapp.\n"))
	})
	workers.Serve(nil) // use http.DefaultServeMux
}
