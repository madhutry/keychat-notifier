package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func processAndroidNotifier(messagesRecvd map[string][]ReceivedMesg) {
	for _, mesg := range messagesRecvd {
		mesgArr := mesg
		for _, val := range mesgArr {
			sender := val.Sender
			accessCd := fetchSenderAccessCode(sender)
			pushKey := apiGetPushkey(accessCd)
			if len(pushKey) == 0 {
				continue
			}
			eventId := val.EventId
			roomId := val.RoomId
			apiSendNotification(pushKey, eventId, roomId)
		}
	}
}
func fetchSenderAccessCode(sender string) string {
	fetchAccCode := "select matrix_access_code from mat_acc_cd_owner where userid=$1"
	db := Envdb.db

	fetchAccCdStmt, err := db.Prepare(fetchAccCode)
	if err != nil {
		log.Fatal(err)
	}
	var accessCode sql.NullString
	fetchAccCdStmt.QueryRow(sender).Scan(&accessCode)
	if accessCode.Valid {
		return accessCode.String
	} else {
		return ""
	}
}
func apiGetPushkey(āccessCd string) string {
	apiHost := "http://%s/_matrix/client/r0/pushers?access_token=%s"
	endpoint := fmt.Sprintf(apiHost, GetMatrixServerUrl())
	response, err := http.Get(endpoint)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var out1 bytes.Buffer
		json.Indent(&out1, data, "=", "\t")
		//out1.WriteTo(os.Stdout)

		var f map[string]interface{}
		json.Unmarshal([]byte(data), &f)
		pushers := f["pushers"].(map[string]interface{})
		pushkey := ""
		if len(pushers) > 0 {
			pushkey = pushers["pushkey"].(string)
		}
		return pushkey
	}
	return ""
}
func apiSendNotification(pushkey string, eventId string, roomId string) interface{} {
	jsonData := map[string]interface{}{
		"priority": "high",
		"to":       pushkey,
		"data": map[string]interface{}{
			"prio":     "high",
			"event_id": eventId,
			"room_id":  roomId,
			"unread":   1,
		},
	}
	endpoint := "https://fcm.googleapis.com/fcm/send"
	client := &http.Client{}
	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonValue))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+GetFCMServerCode())

	response, err := client.Do(request)
	fmt.Print(response.StatusCode)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return nil
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var f interface{}
		json.Unmarshal([]byte(data), &f)
		var out1 bytes.Buffer
		json.Indent(&out1, data, "=", "\t")
		//out1.WriteTo(os.Stdout)
		return f
	}
}
