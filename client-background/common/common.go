package common

import (
	"client-background/types"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

// const (
// 	BG_TCP_HOST = "localhost"
// 	BG_TCP_PORT = "1796"
// 	BG_TCP_TYPE = "tcp"
// )

// const (
// 	ROOT_PATH   = "."
// 	TIME_FORMAT = "2006_01_02_15_04_05_-0700"
// )

// const (
// 	SERVER_PORT = 8080
// 	SERVER_HOST = "127.0.0.1"
// 	SERVER_TYPE = "http"

// 	HTTP_CLIENT_TIMEOUT  = 30
// 	MAX_NETWORK_REQUESTS = 40
// )

var (
	BG_TCP_HOST string
	BG_TCP_PORT string
	BG_TCP_TYPE string

	ROOT_PATH    string
	DATA_PATH    string
	RESTORE_PATH string
	TIME_FORMAT  string
)

var (
	SERVER_PORT string
	SERVER_HOST string
	SERVER_TYPE string

	HTTP_CLIENT_TIMEOUT  int
	MAX_NETWORK_REQUESTS int
)

var Client http.Client
var netwkSem *semaphore.Weighted

func LoadEnvFile() error {

	rootPath, pathSet := os.LookupEnv("ROOT_PATH")
	if pathSet {
		ROOT_PATH = rootPath
	} else { // dev mode
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			return err
		}
		ROOT_PATH = filepath.Join(filepath.Dir(dir), "test")
	}

	DATA_PATH = filepath.Join(ROOT_PATH, "data")

	err := godotenv.Load(filepath.Join(ROOT_PATH, ".env.common"))
	if err != nil {
		fmt.Println(err)
		return err
	}

	BG_TCP_HOST = os.Getenv("BG_TCP_HOST")
	BG_TCP_PORT = os.Getenv("BG_TCP_PORT")
	BG_TCP_TYPE = os.Getenv("BG_TCP_TYPE")
	TIME_FORMAT = os.Getenv("TIME_FORMAT")

	err = godotenv.Load(filepath.Join(ROOT_PATH, ".env.background"))
	if err != nil {
		fmt.Println(err)
		return err
	}

	restorePath, pathSet := os.LookupEnv("RESTORE_PATH")

	if pathSet {
		RESTORE_PATH = restorePath
	} else {
		RESTORE_PATH = filepath.Join(ROOT_PATH, "restore")
	}

	SERVER_PORT = os.Getenv("SERVER_PORT")
	SERVER_HOST = os.Getenv("SERVER_HOST")
	SERVER_TYPE = os.Getenv("SERVER_TYPE")

	HTTP_CLIENT_TIMEOUT, err = strconv.Atoi(os.Getenv("HTTP_CLIENT_TIMEOUT"))
	if err != nil {
		fmt.Println(err)
		return err
	}

	MAX_NETWORK_REQUESTS, err = strconv.Atoi(os.Getenv("MAX_NETWORK_REQUESTS"))
	if err != nil {
		fmt.Println(err)
		return err
	}

	Client = http.Client{
		Timeout: time.Duration(HTTP_CLIENT_TIMEOUT) * time.Second,
	}

	netwkSem = semaphore.NewWeighted(int64(MAX_NETWORK_REQUESTS))

	return nil
}

var GlobalLogger *zap.Logger

// ------------------------------------------------------------------------------------

var UserDataDir string // = filepath.Join(RootPath, "data") //change
var UserDataDirMu = sync.RWMutex{}

func getUserDataDir() string {
	UserDataDirMu.RLock()
	defer UserDataDirMu.RUnlock()
	return UserDataDir
}

func SetCurrentUser(newUser string) {
	UserDataDirMu.Lock()
	UserDataDir = filepath.Join(DATA_PATH, newUser)
	UserDataDirMu.Unlock()

	makeUserDataDirs()
}

func GetCurrentUserFile() string {
	return filepath.Join(DATA_PATH, "currentUser.json")
}

func GetClientLogsDir() string {
	return filepath.Join(DATA_PATH, "backgroundlogs")
}

func GetRestoreDir() string {
	return filepath.Join(RESTORE_PATH)
}

func GetBackUpLogsDir() string {
	return filepath.Join(getUserDataDir(), "logs", "backup")
}

func GetRetrieveLogsDir() string {
	return filepath.Join(getUserDataDir(), "logs", "retrieve")
}

func GetDeleteLogsDir() string {
	return filepath.Join(getUserDataDir(), "logs", "delete")
}

func GetBackUpsDir() string {
	return filepath.Join(getUserDataDir(), "backups")
}

func GetDirectoriesFile() string {
	return filepath.Join(getUserDataDir(), "directories.json")
}

func GetCacheFile() string {
	return filepath.Join(getUserDataDir(), "hashes.cache")
}

func GetScheduleFile() string {
	return filepath.Join(getUserDataDir(), "schedule.json")
}

func GetPersistFile() string {
	return filepath.Join(getUserDataDir(), "persist.json")
}

func GetSysHistoryFile() string {
	return filepath.Join(getUserDataDir(), "sysHistory.jsonl")
}

func GetLoginStateFile() string {
	return filepath.Join(DATA_PATH, "loginState.json")
}

// ------------------------------------------------------------------------------------

const HTTP_STATUS_RETRIEVAL_COMPLETE = 250
const KEY types.SessionKey = "key"

// ------------------------------------------------------------------------------------

var State struct {
	data types.Persist
	mu   sync.Mutex
}

var LoginState struct {
	UserIsLoggedIn bool
	mu             sync.Mutex
}

func InitStateVars() {
	State.mu = sync.Mutex{}
	LoginState.mu = sync.Mutex{}
}
