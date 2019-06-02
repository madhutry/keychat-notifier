package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLog() {

	_, err := os.OpenFile("/var/log/frieze/foo.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}
	log.SetOutput(&lumberjack.Logger{
		Filename:   "/var/log/frieze/foo.log",
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
func GetFriezeChatAPIUrl() string {
	return viper.GetString("FRIEZE_CHAT_API_HOST")
}
func GetMatrixAdminCode() string {
	return viper.GetString("MATRIX_ADMIN_ACCESS_CODE")
}
func GetFilterId() string {
	return viper.GetString("FILTER_ID")
}
