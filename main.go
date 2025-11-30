package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	// "os"

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
		// godotenv.Load(".env")
		// botApiString := os.Getenv("BOT_GREEN_API")
		// if botApiString == "" {
		// 	log.Fatal("Bot Green Api is not Found")
		// 	return
		// }

		fmt.Println("Got BOT API! and Fetching top 20 stories from HN...")

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

		respStr, err := internal.MakeBotMessage(req, uniqueTop10, "https://7105.api.green-api.com/waInstance7105341467/sendMessage/d4a457b0df2a45adbf8f1dd0922f9e10d97be7e200bc4dd081")

		if err != nil {
			fmt.Println("Failed to Make Bot Message", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println(respStr)

		w.Write([]byte("Msg sent to whatsapp.\n"))
		w.Write([]byte(fmt.Sprintf("%s\n", respStr)))

		if len(uniqueTop10) == 0 {
			w.Write([]byte("No new feeds found. All top 20 stories have already been seen.\n"))
		}

	})
	workers.Serve(nil) // use http.DefaultServeMux
}
