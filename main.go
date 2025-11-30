package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/MdSadiqMd/HN-Alerts/internal"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare"
)

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
		top20Stories, err := internal.FetchHNTopN(req, 20)
		if err != nil {
			fmt.Println("Error fetching from HN:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Printf("Fetched %d stories from HN\n", len(top20Stories))

		uniqueTop10, err := internal.GetHNTop10FromKV(req, top20Stories)
		if err != nil {
			fmt.Println("Error getting top 10 from KV:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		respStr, err := internal.MakeBotMessage(req, uniqueTop10, cloudflare.Getenv("GREEN_API_URL"))
		if err != nil {
			fmt.Println("Failed to Make Bot Message", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println(respStr)

		w.Write([]byte("Msg sent to whatsapp.\n"))
		w.Write([]byte(fmt.Sprintf("%s\n", respStr)))
	})
	workers.Serve(nil) // use http.DefaultServeMux
}
