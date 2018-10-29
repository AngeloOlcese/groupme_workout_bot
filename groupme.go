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
	"os"
	"bufio"
	"github.com/bradfitz/slice"
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
	name string
	lift, run, throw int64
}

func readLeaderboard() {
	file, _ := os.Open("lb.txt")
	scanner := bufio.NewScanner(file)
  for scanner.Scan() {
      line := strings.Split(scanner.Text(), " ")
			lift, _ := strconv.ParseInt(line[1], 10, 64)
			run, _ := strconv.ParseInt(line[2], 10, 64)
			throw, _ := strconv.ParseInt(line[3], 10, 64)
			leaderboard[line[0]] = &lbentry{lift: lift, run: run, throw: throw, name: line[4]}
  }
}

func writeLeaderboard() {
	os.Remove("lb.txt")
	file, err := os.Create("lb.txt")
	if err != nil {
			log.Fatal("Cannot open file", err)
	}
	defer file.Close()

	for key, val := range leaderboard {
		fmt.Fprintf(file, key+" "+strconv.FormatInt(val.lift, 10)+" "+strconv.FormatInt(val.run, 10)+" "+strconv.FormatInt(val.throw, 10)+" "+val.name+"\n")
	}

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
		entry := lbentry{name: message.Name}
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
		writeLeaderboard()

		responseText += "Stats for " + message.Name + "-- |Lift: " + strconv.FormatInt(int64(entry.lift), 10) + "| Run: " + strconv.FormatInt(int64(entry.run), 10) + "| Throw: " + strconv.FormatInt(int64(entry.throw), 10) + "|"
		sendBotMessage(botMessage{Bot_id: botID, Text: responseText})
	}
}

func sendStats() {
	var scores []*lbentry
	for _, vals := range leaderboard {
		scores = append(scores, vals)
	}
	slice.Sort(scores[:], func(i, j int) bool {
	    return scores[i].lift < scores[j].lift
	})
	one := scores[0]
	two:= scores[1]
	three := scores[2]
	sendBotMessage(botMessage{Bot_id: botID, Text: "Top 3 Lifters:\n"+"1."+one.name+" "+strconv.FormatInt(one.lift,10)+"\n2."+two.name+" "+strconv.FormatInt(two.lift,10)+"\n3."+three.name+" "+strconv.FormatInt(three.lift,10)})

	slice.Sort(scores[:], func(i, j int) bool {
	    return scores[i].throw < scores[j].throw
	})
	one = scores[0]
	two = scores[1]
	three = scores[2]
	sendBotMessage(botMessage{Bot_id: botID, Text: "Top 3 Throwers:\n"+"1."+one.name+" "+strconv.FormatInt(one.throw,10)+"\n2."+two.name+" "+strconv.FormatInt(two.throw,10)+"\n3."+three.name+" "+strconv.FormatInt(three.throw,10)})

	slice.Sort(scores[:], func(i, j int) bool {
	    return scores[i].run < scores[j].run
	})
	one = scores[0]
	two = scores[1]
	three = scores[2]
	sendBotMessage(botMessage{Bot_id: botID, Text: "Top 3 Runners:\n"+"1."+one.name+" "+strconv.FormatInt(one.run,10)+"\n2."+two.name+" "+strconv.FormatInt(two.run,10)+"\n3."+three.name+" "+strconv.FormatInt(three.run,10)})
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
	readLeaderboard()
	sendStats()
	http.HandleFunc("/bot", parseRequest)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
