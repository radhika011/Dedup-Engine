package main

import (
	"dedup-server/backup"
	"dedup-server/common"
	"dedup-server/delete"
	"dedup-server/logging"
	"dedup-server/mongoutil"
	"dedup-server/retrieve"
	"dedup-server/session"
	"dedup-server/user"
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"
)

func main() {

	err := os.MkdirAll(common.SessionLogsPath, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = os.MkdirAll(common.ServerLogsPath, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = os.MkdirAll(common.SessionsPath, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}

	logging.InitGlobalLogger()
	logging.GlobalLogger.Info("Session maps launching")

	session.InitSessionMap()
	go session.ManageSessionExpiry()

	mongoutil.ConnectClient()
	defer mongoutil.DisconnectClient()

	go common.GarbageCollector(logging.GlobalLogger)
	go mongoutil.LogDeDupRatio(logging.GlobalLogger)

	http.HandleFunc("/backup/chunk", backup.ServeChunk)
	http.HandleFunc("/backup/filemetadata", backup.ServeFileMeta)
	http.HandleFunc("/backup/backupdirstruct", backup.ServeDirStruct)
	http.HandleFunc("/backup/unmodifiedchunks", backup.ServeUnmodifiedChunks)
	http.HandleFunc("/backup/invalidchunks", backup.ServeInvalidChunks)
	http.HandleFunc("/backup/status", backup.ServeStatus)
	http.HandleFunc("/backup/init", backup.ServeInit)

	http.HandleFunc("/delete/init", delete.ServeInit)
	http.HandleFunc("/delete/status", delete.ServeStatus)
	http.HandleFunc("/delete/backupstruct", delete.ServeDirStruct)

	http.HandleFunc("/retrieve/init", retrieve.ServeInit)
	http.HandleFunc("/retrieve/file", retrieve.ServeFile)
	http.HandleFunc("/retrieve/backupstruct", retrieve.ServeDirStruct)

	http.HandleFunc("/user/register", user.ServeRegister)
	http.HandleFunc("/user/verify", user.ServeVerify)
	http.HandleFunc("/user/login", user.ServeLogin)
	http.HandleFunc("/user/update", user.ServeUpdate)

	logging.GlobalLogger.Info("Server starting", zap.Int("port", 8080))
	http.ListenAndServe(":8080", nil)

}
