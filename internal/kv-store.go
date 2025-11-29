package internal

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/syumai/workers/cloudflare/kv"
)

const hnAlertsNamespace = "HN_ALERTS"

func GetHNTop10FromKV(req *http.Request, stories []HNStory) ([]HNStory, error) {
	HNAlertsKV, err := kv.NewNamespace(hnAlertsNamespace)
	if err != nil {
		fmt.Println("KV namespace error:", err)
		return nil, err
	}

	var oldIDs []int
	fmt.Println("Reading old IDs from KV cache...")
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("%d", i)
		idStr, err := HNAlertsKV.GetString(key, nil)
		if err != nil {
			break
		}
		if idStr != "" {
			id, err := strconv.Atoi(idStr)
			if err == nil {
				oldIDs = append(oldIDs, id)
			}
		}
	}
	fmt.Printf("Found %d existing IDs in KV cache: %v\n", len(oldIDs), oldIDs)

	oldIDsSet := make(map[int]bool)
	for _, id := range oldIDs {
		oldIDsSet[id] = true
	}

	var uniqueTop10 []HNStory
	for _, story := range stories {
		if story.ID == 0 {
			continue
		}
		if !oldIDsSet[story.ID] && len(uniqueTop10) < 10 {
			uniqueTop10 = append(uniqueTop10, story)
		}
	}
	fmt.Printf("Selected %d unique stories (ignoring duplicates)\n", len(uniqueTop10))

	fmt.Println("Unique Top 10 Story IDs:")
	for i, story := range uniqueTop10 {
		fmt.Printf("%d. [%d] %s\n", i+1, story.ID, story.Title)
	}

	fmt.Println("Deleting old IDs from KV cache...")
	for i := 0; i < len(oldIDs); i++ {
		key := fmt.Sprintf("%d", i)
		err := HNAlertsKV.Delete(key)
		if err != nil {
			fmt.Printf("KV Delete error for key=%s: %v\n", key, err)
		}
	}

	fmt.Printf("Storing %d new IDs in KV cache...\n", len(uniqueTop10))
	for i, story := range uniqueTop10 {
		key := fmt.Sprintf("%d", i)
		idStr := fmt.Sprintf("%d", story.ID)
		err = HNAlertsKV.PutString(key, idStr, nil)
		if err != nil {
			fmt.Printf("KV PutString error for key=%s: %v\n", key, err)
			return nil, err
		}
	}

	return uniqueTop10, nil
}
