package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLog() {
	logFile := GetLogFileName()
	_, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}
	log.SetOutput(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	})
}
func InitConfig() {
	home, _ := os.UserHomeDir()
	if os.Getenv("ENVIRONMENT") == "DEV" {
		fmt.Println("loafing dev env")
		viper.SetConfigName("server")
		viper.SetConfigType("json")
		viper.AddConfigPath(filepath.Dir(home + "/"))
		viper.ReadInConfig()
	} else {
		viper.AutomaticEnv()
	}

}

func GetDBUrl() string {
	return viper.GetString("DB_URL")
}
func GetMatrixServerUrl() string {
	return viper.GetString("MATRIX_URL")
}
func GetMatrixAdminCode() string {
	adminCd := viper.GetString("MATRIX_ADMIN_ACCESS_CODE")
	if len(adminCd) == 0 {
		loadAdminInfoEnv()
		adminCd = viper.GetString("MATRIX_ADMIN_ACCESS_CODE")
	}
	return adminCd
}
func GetFilterId() string {
	filterid := viper.GetString("FILTER_ID")
	if len(filterid) == 0 {
		loadAdminInfoEnv()
		filterid = viper.GetString("FILTER_ID")
	}
	return filterid
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
func GetFCMServerCode() string {
	return viper.GetString("FCM_SERVER_CODE")
}
func GetLogFileName() string {
	return viper.GetString("FRIEZE_NOTIFIER_LOG_FILE")
}
