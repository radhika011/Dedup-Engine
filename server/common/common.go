package common

import (
	"encoding/json"
	"path/filepath"
)

const RootPath = "."
const SessionsDir = "sessions"
const SessionLogsDir = "sessionlogs"
const ServerLogsDir = "serverlogs"
const Data = "data"
const TimeFormat = "2006_01_02_15_04_05_-0700"

const HTTPStatusRetrievalCompleted = 250

var DataPath = filepath.Join(RootPath, Data)
var SessionsPath = filepath.Join(DataPath, SessionsDir)
var SessionLogsPath = filepath.Join(DataPath, SessionLogsDir)
var ServerLogsPath = filepath.Join(DataPath, ServerLogsDir)

func ToJSON(data interface{}) ([]byte, error) { // should be a common util
	jsonData, err := json.Marshal(data)
	return jsonData, err
}
