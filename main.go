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
	"strings"
	"time"

	_ "github.com/lib/pq"
	pborman "github.com/pborman/uuid"
)

type ReceivedMesg struct {
	MessageText string `json:"message"`
	Sender      string `json:"sender"`
	Timestamp   string `json:"timestamp"`
	TransId     string `json:"transid"`
	MesgType    string `json:"mesgtype"`
	Url         string `json:"url"`
	RoomId      string
	EventId     string
}

func main() {
	InitConfig()
	Init()
	loadAdminInfoEnv()
	InitLog()
	tick()
}
func tick() {
	ticker := time.NewTicker(time.Second * 1).C
	for {
		select {
		case <-ticker:
			fetchNewMessage()
		}
	}

	//time.Sleep(time.Second * 10)
}
func loadAdminInfoEnv() {
	userid, acc_cd, filterid := dbFetchAdminInfo()
	os.Setenv("MATRIX_ADMIN_USERID", userid)
	os.Setenv("MATRIX_ADMIN_ACCESS_CODE", acc_cd)
	os.Setenv("FILTER_ID", filterid)
}
func dbFetchAdminInfo() (string, string, string) {
	fetchAdminInfo := "SELECT userid,access_code,filter_id FROM public.admin_info where active='Y'"
	var userId sql.NullString
	var accessCode sql.NullString
	var filterId sql.NullString
	db := Envdb.db

	fetchBatchIdStmt, err := db.Prepare(fetchAdminInfo)
	if err != nil {
		log.Fatal(err)
	}
	fetchBatchIdStmt.QueryRow().Scan(&userId, &accessCode, &filterId)
	return userId.String, accessCode.String, filterId.String
}
func fetchNewMessage() {
	dbBatchId := fetchBatchId()
	filterId := GetFilterId()
	apiHost := "http://%s/_matrix/client/r0/sync?access_token=%s&filter=%s&limit=2%s"
	endpoint := fmt.Sprintf(apiHost, GetMatrixServerUrl(), GetMatrixAdminCode(), filterId, "")
	//fmt.Println(endpoint)
	if len(dbBatchId) > 0 {
		endpoint = fmt.Sprintf(apiHost, GetMatrixServerUrl(), GetMatrixAdminCode(), filterId, "&since="+dbBatchId)
	}
	//log.Println(endpoint)
	start := time.Now()
	newmessageRecd := false

	response, err := http.Get(endpoint)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var out1 bytes.Buffer
		json.Indent(&out1, data, "=", "\t")
		//out1.WriteTo(os.Stdout)

		var f map[string]interface{}
		json.Unmarshal([]byte(data), &f)
		nextBatch := f["next_batch"].(string)
		rooms := f["rooms"].(map[string]interface{})["join"].(map[string]interface{})
		var messagesResult = make(map[string][]ReceivedMesg)
		for k, _ := range rooms {
			var messages []ReceivedMesg
			log.Println("Room ID" + k)
			timelime := rooms[k].(map[string]interface{})["timeline"].(map[string]interface{})["events"]
			events := timelime.([]interface{})
			for _, v1 := range events {
				sender := v1.(map[string]interface{})["sender"].(string)
				eventId := v1.(map[string]interface{})["event_id"].(string)
				timeSent := v1.(map[string]interface{})["origin_server_ts"].(float64)
				mesg := v1.(map[string]interface{})["content"].(map[string]interface{})["body"].(string)
				transIdVal := v1.(map[string]interface{})["content"].(map[string]interface{})["trans_id"]
				mesgType := v1.(map[string]interface{})["content"].(map[string]interface{})["msgtype"].(string)
				urlVal, ok := v1.(map[string]interface{})["content"].(map[string]interface{})["url"]
				url := ""
				if ok {
					url = urlVal.(string)[strings.LastIndex(urlVal.(string), "/")+1:]
				}

				transId := pborman.NewRandom().String()
				if transIdVal != nil {
					transId = transIdVal.(string)
				}
				mesgStruct := ReceivedMesg{
					MessageText: mesg,
					Sender:      sender,
					Timestamp:   fmt.Sprintf("%f", timeSent),
					RoomId:      k,
					TransId:     transId,
					MesgType:    mesgType,
					Url:         url,
					EventId:     eventId,
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
		//out.WriteTo(os.Stdout)
		elapsed := time.Now()
		if newmessageRecd {
			log.Println("Message Sent to API")
			//processAndroidNotifier(messagesResult)
			saveMessages(messagesResult)
			dbNotificationProcessed(dbBatchId)
			dbInsertNotification(start, elapsed, string(data), nextBatch)
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
	updateNotification := `UPDATE notification_job	set processed=1 WHERE processed=0`
	db := Envdb.db

	updateNotificationStmt, err := db.Prepare(updateNotification)
	if err != nil {
		log.Fatal(err)
	}
	defer updateNotificationStmt.Close()
	_, err = updateNotificationStmt.Exec()
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
func setLastTimeMesgSent(sender string, timeRecvd string) {
	log.Println("Setting Last mesg sent for sender %s", sender)
	updateTime := `update mat_acc_cd_owner set  last_mesg_recd_time=$1 where userid=$2`
	db := Envdb.db

	updateTimeStmt, err := db.Prepare(updateTime)
	if err != nil {
		log.Fatal(err)
	}
	defer updateTimeStmt.Close()
	_, err = updateTimeStmt.Exec(timeRecvd, sender)
	if err != nil {
		log.Fatal(err)
		log.Printf("Error Last mesg sent for sender %s", sender)

	}
}
func saveMessages(messagesRecvd map[string][]ReceivedMesg) {
	db := Envdb.db

	saveMesg := `INSERT INTO messages	(		mesg_id,message,server_received_ts,sender,room_id,create_ts,url,mesg_type,event_id	)
	VALUES	
	(		$1,		$2,		$3,		$4,		$5,		$6,$7,$8,$9 )	
	`
	saveMesgStmt, err := db.Prepare(saveMesg)
	if err != nil {
		log.Fatal(err)
	}
	defer saveMesgStmt.Close()

	for k, mesg := range messagesRecvd {
		roomID := k
		mesgArr := mesg
		for _, val := range mesgArr {
			mesgId := val.TransId
			mesgStr := val.MessageText
			ts := val.Timestamp
			sender := val.Sender
			mesgType := val.MesgType
			url := val.Url
			eventId := val.EventId
			/* 			v := val.(map[string]interface{})
			   			mesgStr := v["message"].(string)
			   			ts := v["timestamp"].(string)
			   			sender := v["sender"].(string) */
			_, err = saveMesgStmt.Exec(mesgId, mesgStr, ts, sender, roomID, time.Now(), url, mesgType, eventId)
			setLastTimeMesgSent(sender, ts)
			if err != nil {
				panic(err)
			}
		}
	}
}
