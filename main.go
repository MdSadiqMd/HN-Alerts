package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/MdSadiqMd/HN-Alerts/internal"
	"github.com/syumai/workers"
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
		fmt.Println("Fetching top 20 titles from HN...")
		top20Titles, err := internal.FetchHNTopN(req, 20)
		if err != nil {
			fmt.Println("Error fetching from HN:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Printf("Fetched %d titles from HN\n", len(top20Titles))

		uniqueTop10, err := internal.GetHNTop10FromKV(req, top20Titles)
		if err != nil {
			fmt.Println("Error getting top 10 from KV:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write([]byte(fmt.Sprintf("Top %d Unique HN Feeds:\n", len(uniqueTop10))))
		for i, title := range uniqueTop10 {
			w.Write([]byte(fmt.Sprintf("%d. %s\n", i+1, title)))
		}
	})
	workers.Serve(nil) // use http.DefaultServeMux
}
