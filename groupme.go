package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const botID string = "44a74f5918bb19f58e4f148d5a"

var leaderboard map[string]*lbentry = make(map[string]*lbentry)

type callback struct {
	Sender_id string `json:"sender_id"`
	Name      string `json:"name"`
	Text      string `json:"text"`
}

type botMessage struct {
	Bot_id string `json:"bot_id"`
	Text   string `json:"text"`
}

type lbentry struct {
	lift, run, throw int
}

func parseRequest(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	var m callback
	err = json.Unmarshal(body, &m)
	if err != nil {
		panic(err)
	}
	parseCallback(m)
}

func parseCallback(message callback) {
	responseText := ""
	workout := false
	if leaderboard[message.Sender_id] == nil {
		entry := lbentry{}
		leaderboard[message.Sender_id] = &entry
	}
	if strings.Contains(message.Text, "#lift") {
		workout = true
		leaderboard[message.Sender_id].lift++
	}
	if strings.Contains(message.Text, "#run") {
		workout = true
		leaderboard[message.Sender_id].run++
	}
	if strings.Contains(message.Text, "#throw") {
		workout = true
		leaderboard[message.Sender_id].throw++
	}
	if workout {
		entry := leaderboard[message.Sender_id]

		responseText += "Stats for " + message.Name + "-- |Lift: " + strconv.FormatInt(int64(entry.lift), 10) + "| Run: " + strconv.FormatInt(int64(entry.run), 10) + "| Throw: " + strconv.FormatInt(int64(entry.throw), 10) + "|"
		sendBotMessage(botMessage{Bot_id: botID, Text: responseText})
	}
}

func sendBotMessage(m botMessage) {
	url := "https://api.groupme.com/v3/bots/post"
	byt, _ := json.Marshal(m)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byt))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println("response Status:", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func main() {
	http.HandleFunc("/bot", parseRequest)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
