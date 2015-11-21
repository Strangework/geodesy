package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type InMessage struct {
	Channel   string
	User      string
	Text      string
	Timestamp string
	Team      string
}

type OutMessage struct {
	Id      int
	Channel string
	Text    string
}

func (m OutMessage) MarshalJSON() ([]byte, error) {
	someMap := make(map[string]interface{})
	someMap["id"] = m.Id
	someMap["type"] = "message"
	someMap["channel"] = m.Channel
	someMap["text"] = m.Text

	someMsg, err := json.Marshal(someMap)
	return someMsg, err
}

//Connects with Slack's Real-Time Messaging interface
func rtmConnect(token string) (ws *websocket.Conn, ok bool) {
	resp, err := http.PostForm("https://slack.com/api/rtm.start", url.Values{"token": {token}})
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	respJson := make(map[string]interface{})
	err = decoder.Decode(&respJson)
	if err != nil {
		panic(err)
	}
	someVal, ok := respJson["ok"]
	if ok && someVal.(bool) {
		connUrl, _ := respJson["url"]
		fmt.Printf("%s\n", connUrl.(string))
		ws, err := websocket.Dial(connUrl.(string), "", "http://localhost/")
		if err != nil {
			log.Fatal(err)
		}
		return ws, true
	}
	return nil, false
}

func main() {
	var token string
	if len(os.Args) >= 2 {
		filename := os.Args[1]
		if tokenBytes, err := ioutil.ReadFile(filename); err != nil {
			panic(err)
		} else {
			token = strings.TrimSpace(string(tokenBytes))
		}
	} else {
		os.Exit(1)
	}

	ws, ok := rtmConnect(token)
	if ok {
		stdoutJson := json.NewEncoder(os.Stdout)
		for true {
			var msgJson map[string]interface{}
			websocket.JSON.Receive(ws, &msgJson)
			stdoutJson.Encode(msgJson)
			if _, ok := msgJson["type"]; ok && msgJson["type"].(string) == "message" {
				msg := InMessage{
					Channel:   msgJson["channel"].(string),
					User:      msgJson["user"].(string),
					Text:      msgJson["text"].(string),
					Timestamp: msgJson["ts"].(string),
					Team:      msgJson["team"].(string),
				}
				fmt.Println(msg.User)
				retMsg := OutMessage{
					Id:   1,
					Text: ":green_heart:",
				}
				retMsg.Channel = msg.Channel
				retRaw, err := json.Marshal(retMsg)
				if err != nil {
					panic(err)
				}
				fmt.Println(string(retRaw))

				if err := websocket.JSON.Send(ws, retMsg); err != nil {
					panic(err)
				}
			}
		}
	}
}
