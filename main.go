package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/MdSadiqMd/HN-Alerts/internal"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare/kv"
)

const hnAlertsNamespace = "HN_ALERTS"
const hnAlertsKey = "hn-alerts"

func main() {
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
		fmt.Println("Fetching top 20 titles from HN...")
		top20Titles, err := internal.FetchHNTopN(req, 20)
		if err != nil {
			fmt.Println("Error fetching from HN:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Printf("Fetched %d titles from HN\n", len(top20Titles))

		HNAlertsKV, err := kv.NewNamespace(hnAlertsNamespace)
		if err != nil {
			fmt.Println("KV namespace error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var oldTitles []string
		fmt.Println("Reading old titles from KV cache...")
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("%d", i)
			title, err := HNAlertsKV.GetString(key, nil)
			if err != nil {
				break
			}
			if title != "" {
				oldTitles = append(oldTitles, title)
			}
		}
		fmt.Printf("Found %d existing titles in KV cache\n", len(oldTitles))

		oldTitlesSet := make(map[string]bool)
		for _, title := range oldTitles {
			oldTitlesSet[title] = true
		}

		var uniqueTop10 []string
		for _, title := range top20Titles {
			if title == "" {
				continue
			}
			if !oldTitlesSet[title] && len(uniqueTop10) < 10 {
				uniqueTop10 = append(uniqueTop10, title)
			}
		}
		fmt.Printf("Selected %d unique titles (ignoring duplicates)\n", len(uniqueTop10))

		fmt.Println("Unique Top 10 Titles:")
		for i, title := range uniqueTop10 {
			fmt.Printf("%d. %s\n", i+1, title)
		}

		fmt.Println("Deleting old titles from KV cache...")
		for i := 0; i < len(oldTitles); i++ {
			key := fmt.Sprintf("%d", i)
			err := HNAlertsKV.Delete(key)
			if err != nil {
				fmt.Printf("KV Delete error for key=%s: %v\n", key, err)
			}
		}

		fmt.Printf("Storing %d new titles in KV cache...\n", len(uniqueTop10))
		for i, title := range uniqueTop10 {
			key := fmt.Sprintf("%d", i)
			err = HNAlertsKV.PutString(key, title, nil)
			if err != nil {
				fmt.Printf("KV PutString error for key=%s: %v\n", key, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		response := fmt.Sprintf("Top %d Unique HN Feeds:\n", len(uniqueTop10))
		for i, title := range uniqueTop10 {
			response += fmt.Sprintf("%d. %s\n", i+1, title)
		}
		w.Write([]byte(response))
	})
	workers.Serve(nil) // use http.DefaultServeMux
}
