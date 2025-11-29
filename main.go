package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare/fetch"
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
		cli := fetch.NewClient()

		r, err := fetch.NewRequest(req.Context(), http.MethodGet, "https://hacker-news.firebaseio.com/v0/topstories.json?print=pretty", nil)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/111.0")

		res, err := cli.Do(r, nil)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("read error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var ids []int
		if err := json.Unmarshal(body, &ids); err != nil {
			fmt.Println("decode error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		top10 := ids
		if len(ids) > 10 {
			top10 = ids[:10]
		}
		fmt.Println("Top 10 feed IDs:", top10)

		w.Header().Set("content-type", "application/json")
		b, err := json.Marshal(top10)
		if err != nil {
			fmt.Println("encode error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
	})
	workers.Serve(nil) // use http.DefaultServeMux
}
