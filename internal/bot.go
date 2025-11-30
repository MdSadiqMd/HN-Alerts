package internal

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"

    "github.com/syumai/workers/cloudflare/fetch"
)

func makeStoryToStr(uniqueTop10 []HNStory) string {
    finalMsg := ""
    for i, story := range uniqueTop10 {
        finalMsg += fmt.Sprintf("%d. %s\n   %s\n\n", i+1, story.Title, story.URL)
    }
    return finalMsg
}

type MyRequestBody struct {
    ChatID  string `json:"chatId"`
    Message string `json:"message"`
}

type GreenAPIResponse struct {
    IDMessage string `json:"idMessage"`
    ChatID    string `json:"chatId,omitempty"`
}

func MakeBotMessage(req *http.Request, uniqueTop10 []HNStory, botApiString string) (string, error) {
    finalMsg := ""
    if len(uniqueTop10) == 0 {
        finalMsg = "No new feeds found. All top 20 stories have already been seen."
    } else {
        finalMsg = makeStoryToStr(uniqueTop10)
    }

    data := MyRequestBody{
        ChatID:  "120363169536263534@g.us", 
        Message: finalMsg,
    }
    
    jsonData, err := json.Marshal(data)
    if err != nil {
        fmt.Println("Error marshaling JSON:", err)
        return "", err
    }

    cli := fetch.NewClient()
    r, err := fetch.NewRequest(req.Context(), http.MethodPost, botApiString, bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Println("Error creating request:", err)
        return "", err
    }
    
    r.Header.Set("Content-Type", "application/json")
    r.Header.Set("User-Agent", "GREEN-API_POSTMAN/1.0")

    res, err := cli.Do(r, nil)
    if err != nil {
        fmt.Println("Error making request:", err)
        return "", err
    }
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    if err != nil {
        fmt.Println("Read error while making Bot Msg:", err)
        return "", err
    }

    fmt.Printf("Raw API Response: %s\n", string(body))

    var respMsg GreenAPIResponse
    if err := json.Unmarshal(body, &respMsg); err != nil {
        fmt.Println("Got Error for extracting response while making Bot Msg:", err)
        fmt.Println("Response body was:", string(body))
        return "", err
    }

    return respMsg.IDMessage, nil
}
