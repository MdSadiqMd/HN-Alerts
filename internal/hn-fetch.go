package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/syumai/workers/cloudflare/fetch"
)

func FetchHNTop10(req *http.Request) ([]string, error) {
	cli := fetch.NewClient()

	r, err := fetch.NewRequest(req.Context(), http.MethodGet, "https://hacker-news.firebaseio.com/v0/topstories.json?print=pretty", nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	r.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/111.0")

	res, err := cli.Do(r, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("read error:", err)
		return nil, err
	}

	var ids []int
	if err := json.Unmarshal(body, &ids); err != nil {
		fmt.Println("decode error:", err)
		return nil, err
	}
	top10 := ids
	if len(ids) > 10 {
		top10 = ids[:10]
	}
	fmt.Println("Top 10 feed IDs:", top10)

	titles := make([]string, 0, len(top10))
	for i, id := range top10 {
		hn_item, err := fetch.NewRequest(req.Context(), http.MethodGet, fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json?print=pretty", id), nil)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		hn_item_res, err := cli.Do(hn_item, nil)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		defer hn_item_res.Body.Close()

		hn_item_body, err := io.ReadAll(hn_item_res.Body)
		if err != nil {
			fmt.Println("read error:", err)
			return nil, err
		}

		var hn_item_data map[string]interface{}
		if err := json.Unmarshal(hn_item_body, &hn_item_data); err != nil {
			fmt.Println("decode error:", err)
			return nil, err
		}

		title, ok := hn_item_data["title"].(string)
		if !ok {
			title = ""
		}
		fmt.Printf("HN Item %d: %s\n", i+1, title)
		if i > 0 {
			titles = append(titles, ",")
		}
		titles = append(titles, title)
	}
	return titles, nil
}
