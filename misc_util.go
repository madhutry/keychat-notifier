package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func InitConfig() {
	home, _ := os.UserHomeDir()
	/* 	if os.Getenv("ENVIRONMENT") == "DEV" {
	 */fmt.Println("loafing dev env")
	viper.SetConfigName("server")
	viper.SetConfigType("json")
	viper.AddConfigPath(filepath.Dir(home + "/"))
	viper.ReadInConfig()
	/* 	} else {
		viper.AutomaticEnv()
	} */

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
