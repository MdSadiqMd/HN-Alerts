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
		top10, err := internal.FetchHNTop10(req)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println(top10)
		w.Write([]byte(fmt.Sprintf("Top 10 HN Feeds: %v", top10)))
	})
	workers.Serve(nil) // use http.DefaultServeMux
}
