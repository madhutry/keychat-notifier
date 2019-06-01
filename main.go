package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/theckman/go-flock"
)

type ReceivedMesg struct {
	MessageText string `json:"message"`
	Sender      string `json:"sender"`
	Timestamp   string `json:"timestamp"`
	RoomId      string
}

func main() {
	fmt.Printf("Locking %s/%s:%s\n", os.TempDir(), "/go-lock.lock", strconv.Itoa(os.Getpid()))
	f := flock.New(os.TempDir() + "/go-lock.lock")
	f.TryLock() // unchecked errors here
	if !f.Locked() {
		fmt.Printf("Existing...%s\n", strconv.Itoa(os.Getpid()))
		os.Exit(3)
	}
	defer f.Unlock()

	InitConfig()
	Init()
	dbBatchId := fetchBatchId()
	newmessageRecd := false

	apiHost := "http://%s/_matrix/client/r0/sync?access_token=%s&filter=7&limit=2%s"
	endpoint := fmt.Sprintf(apiHost, GetMatrixServerUrl(), GetMatrixAdminCode())

	if len(dbBatchId) > 0 {
		endpoint = fmt.Sprintf(apiHost, GetMatrixServerUrl(), GetMatrixAdminCode(), "&since="+dbBatchId)
	}
	fmt.Println(endpoint)
	start := time.Now()

	response, err := http.Get(endpoint)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var out1 bytes.Buffer
		json.Indent(&out1, data, "=", "\t")
		out1.WriteTo(os.Stdout)

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
				newmessageRecd = true
				messages = append(messages, mesgStruct)
			}
			messagesResult[k] = messages
		}
		result := make(map[string]interface{})
		result["messages"] = messagesResult
		result["batchId"] = nextBatch
		bytesArr, _ := json.Marshal(result)

		var out bytes.Buffer
		json.Indent(&out, bytesArr, "=", "\t")
		out.WriteTo(os.Stdout)
		if newmessageRecd {
			log.Println("Message Sent to API")
			apiSendMessage(result)
		} else {
			log.Println("No Message Sent To API")
		}

		elapsed := time.Now()

		dbInsertNotification(start, elapsed, string(data), nextBatch)
		if len(dbBatchId) > 0 {
			dbNotificationProcessed(dbBatchId)
		}
	}
}
func fetchBatchId() string {
	fetchBatchId := "select batch_id from notification_job where processed=0"
	var batchId sql.NullString
	db := Envdb.db

	fetchBatchIdStmt, err := db.Prepare(fetchBatchId)
	if err != nil {
		log.Fatal(err)
	}
	fetchBatchIdStmt.QueryRow().Scan(&batchId)
	if batchId.Valid {
		return batchId.String
	} else {
		return ""
	}
}
func dbNotificationProcessed(batchId string) {
	updateNotification := `UPDATE notification_job	set processed=1 WHERE processed=0 and batch_id=$1`
	db := Envdb.db

	updateNotificationStmt, err := db.Prepare(updateNotification)
	if err != nil {
		log.Fatal(err)
	}
	defer updateNotificationStmt.Close()
	_, err = updateNotificationStmt.Exec(batchId)
	if err != nil {
		log.Fatal(err)
	}
}
func dbInsertNotification(startTime time.Time, endTime time.Time, payload string, batchId string) {
	insertNotification := `INSERT INTO notification_job	(	start_time,end_time,	payload,batch_id,processed
	)	VALUES 	($1,$2,$3,$4,$5)`
	db := Envdb.db

	insertNotificationStmt, err := db.Prepare(insertNotification)
	if err != nil {
		log.Fatal(err)
	}
	defer insertNotificationStmt.Close()
	_, err = insertNotificationStmt.Exec(startTime, endTime, payload, batchId, 0)
	if err != nil {
		panic(err)
	}
}
func apiSendMessage(jsonData map[string]interface{}) {
	apiHost := "http://%s/chat/notify"
	endpoint := fmt.Sprintf(apiHost, GetFriezeChatAPIUrl())
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
