package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const matrixApiHost = "13.232.162.152:8008"
const friezeChatHost = "localhost:6060"
const matAccCode = "MDAxNWxvY2F0aW9uIHByaXZhdGUKMDAxM2lkZW50aWZpZXIga2V5CjAwMTBjaWQgZ2VuID0gMQowMDI1Y2lkIHVzZXJfaWQgPSBAbWFpbmFkbWluOnByaXZhdGUKMDAxNmNpZCB0eXBlID0gYWNjZXNzCjAwMjFjaWQgbm9uY2UgPSBeeU8qSVVmXkQmb2QmQVNKCjAwMmZzaWduYXR1cmUgIZ0wsA7ywHHPQUhQ1AYPhlc-ePmVa8YPnib36bvM7_oK"

type ReceivedMesg struct {
	MessageText string `json:"message"`
	Sender      string `json:"sender"`
	Timestamp   string `json:"timestamp"`
	RoomId      string
}

func main() {
	apiHost := "http://%s/_matrix/client/r0/sync?access_token=%s&filter=7&limit=2"
	endpoint := fmt.Sprintf(apiHost, matrixApiHost, matAccCode)
	response, err := http.Get(endpoint)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var f map[string]interface{}
		json.Unmarshal([]byte(data), &f)
		nextBatch := f["next_batch"].(string)
		fmt.Println(nextBatch)
		rooms := f["rooms"].(map[string]interface{})["join"].(map[string]interface{})
		var messagesResult = make(map[string][]ReceivedMesg)
		for k, _ := range rooms {
			var messages []ReceivedMesg
			fmt.Println("Room ID" + k)
			timelime := rooms[k].(map[string]interface{})["timeline"].(map[string]interface{})["events"]
			events := timelime.([]interface{})
			for _, v1 := range events {
				sender := v1.(map[string]interface{})["sender"].(string)
				timeSent := v1.(map[string]interface{})["origin_server_ts"].(float64)
				mesg := v1.(map[string]interface{})["content"].(map[string]interface{})["body"].(string)
				mesgStruct := ReceivedMesg{
					MessageText: mesg,
					Sender:      sender,
					Timestamp:   fmt.Sprintf("%f", timeSent),
					RoomId:      k,
				}
				messages = append(messages, mesgStruct)
			}
			messagesResult[k] = messages
		}
		result := make(map[string]interface{})
		result["messages"] = messagesResult
		result["batchId"] = nextBatch
		bytes, _ := json.Marshal(result)
		fmt.Println(string(bytes))
		apiSendMessage(result)

	}
}

func apiSendMessage(jsonData map[string]interface{}) {
	apiHost := "http://%s/chat/notify"
	endpoint := fmt.Sprintf(apiHost, friezeChatHost)
	jsonValue, _ := json.Marshal(jsonData)
	_, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		fmt.Println("Succ")
		//data, _ := ioutil.ReadAll(response.Body)
		//var f interface{}
		//json.Unmarshal([]byte(data), &f)
	}

}
