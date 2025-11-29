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
		top10, err := internal.FetchHNTop10(req)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Printf("Fetched top10 from HN: (%d items): %v\n", len(top10), top10)

		HNAlertsKV, err := kv.NewNamespace(hnAlertsNamespace)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		count := 0
		for i, title := range top10 {
			if title == "," {
				continue
			}

			key := fmt.Sprintf("%d", count)
			fmt.Printf("Writing to KV: key=%s, value=%s\n", key, title)

			err = HNAlertsKV.PutString(key, title, nil)
			if err != nil {
				fmt.Printf("KV PutString error at i=%d, key=%s, value=%s: %v\n", i, key, title, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			count++
		}

		fmt.Println("Reading Top 10 HN Feeds from KV:")
		var results []string
		for i := 0; i < count; i++ {
			readKey := fmt.Sprintf("%d", i)
			title, err := HNAlertsKV.GetString(readKey, nil)
			if err != nil {
				fmt.Printf("KV GetString error at i=%d, key=%s: %v\n", i, readKey, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			fmt.Printf("From KV index %d: %s\n", i, title)

			results = append(results, title)
		}
		w.Write([]byte(fmt.Sprintf("Top 10 HN Feeds: %v", results)))
	})
	workers.Serve(nil) // use http.DefaultServeMux
}
