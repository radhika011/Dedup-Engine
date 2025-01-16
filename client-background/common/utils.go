package common

import (
	"bytes"
	"client-background/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

func ToJSON(data interface{}) ([]byte, error) { // should be a common util
	jsonData, err := json.Marshal(data)
	return jsonData, err
}

func ReadCurrentUserData() (string, error) {
	funcName := "ReadCurrentUserData"

	file, err := os.OpenFile(GetCurrentUserFile(), os.O_RDONLY, 0644)
	if err != nil {
		return "", errors.New(funcName + " :: " + "Error occurred while opening file: " + err.Error())
	}

	decoder := json.NewDecoder(file)

	var currentUser types.CurrentUser
	err = decoder.Decode(&currentUser)
	if err != nil {
		return "", errors.New(funcName + " :: " + "Error occurred while decoding JSON: " + err.Error())
	}

	// if currentUser.UserName == "" {
	// 	return "", errors.New(funcName + " :: " + "No username found")
	// }

	SetCurrentUser(currentUser.UserName)
	return currentUser.UserName, nil
}

func GetClientID() int {
	return 1234
}

func UpdateSysHistoryFile(entry types.SysHistoryEntry) error {
	file, err := os.OpenFile(GetSysHistoryFile(), os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)

	err = encoder.Encode(entry)
	if err != nil {
		return err
	}

	return nil
}

func PersistState() {
	funcName := "PersistState"
	persistFile, err := os.OpenFile(GetPersistFile(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		GlobalLogger.Error(funcName + " :: " + "Error occurred while opening file: " + err.Error())
		return
	}

	encoder := json.NewEncoder(persistFile)
	err = encoder.Encode(State.data)
	if err != nil {
		GlobalLogger.Error(funcName + " :: " + "Error occurred while encoding state to file: " + err.Error())
		return
	}
}

func SetCacheFlag(input bool) {
	State.mu.Lock()
	State.data.CacheIsValid = input
	PersistState()
	State.mu.Unlock()
}

func GetCacheFlag() bool {
	State.mu.Lock()
	defer State.mu.Unlock()
	return State.data.CacheIsValid

}

func SetLastBackUpTime(lastTime time.Time) {
	State.mu.Lock()
	State.data.LastBackUpTime = lastTime
	PersistState()
	State.mu.Unlock()
}

func GetLastBackUpTime() time.Time {
	State.mu.Lock()
	defer State.mu.Unlock()
	return State.data.LastBackUpTime
}

func LoadPersistedState() {
	funcName := "LoadPersistedState"

	userIsLoggedIn := GetLoginState()
	if !userIsLoggedIn {
		return
	}

	var data types.Persist
	State.mu = sync.Mutex{}

	defer func() {
		State.data = data
	}()

	persistFile, err := os.OpenFile(GetPersistFile(), os.O_RDONLY, 0644)
	if err != nil {
		GlobalLogger.Error(funcName + " :: " + "Error occurred while opening file: " + err.Error())
		data.CacheIsValid = false
		return
	}

	decoder := json.NewDecoder(persistFile)

	err = decoder.Decode(&data)
	if err != nil {
		GlobalLogger.Error(funcName + " :: " + "Error occurred while decoding JSON: " + err.Error())
		data.CacheIsValid = false
		return
	}

}

func SetLoginState(input bool) {
	LoginState.mu.Lock()
	LoginState.UserIsLoggedIn = input
	LoginState.mu.Unlock()
}

func GetLoginState() bool {
	LoginState.mu.Lock()
	defer LoginState.mu.Unlock()
	return LoginState.UserIsLoggedIn
}

func LoadLoginState() {
	username, err := ReadCurrentUserData()

	LoginState.mu.Lock()
	defer LoginState.mu.Unlock()

	if err != nil || (username == "") {
		LoginState.UserIsLoggedIn = false
	} else {
		LoginState.UserIsLoggedIn = true
	}
}

func makeUserDataDirs() error {
	err := os.MkdirAll(GetBackUpLogsDir(), 0777)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = os.MkdirAll(GetRetrieveLogsDir(), 0777)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = os.MkdirAll(GetDeleteLogsDir(), 0777)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = os.MkdirAll(GetBackUpsDir(), 0777)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func MakeGlobalDirs() error {

	err := os.MkdirAll(DATA_PATH, 0777)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = os.MkdirAll(GetClientLogsDir(), 0777)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = os.MkdirAll(GetRestoreDir(), 0777)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

//--------------------- NETWORK OPERATIONS ----------------------------------------------

func Send(ctx context.Context, sendPacket types.SendPacket) error { // move err handling to backupSend
	funcName := "Send"
	sessionDetails := ctx.Value(KEY).(types.SessionDetails)
	bodyReader := bytes.NewReader(sendPacket.JsonBody)

	requestURL := fmt.Sprintf("%s://%s:%s/%s/%s", SERVER_TYPE, SERVER_HOST, SERVER_PORT, sessionDetails.Type, sendPacket.Endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bodyReader)
	req.Close = true

	if err != nil { // log func failure
		err = errors.New(funcName + " :: " + "Error occurred while creating HTTP request: " + err.Error())
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	sessionIDCookie := &http.Cookie{
		Name:  "sessionID",
		Value: sessionDetails.SessionID,
	}
	req.AddCookie(sessionIDCookie)

	timeoutIntervals := []int64{10, 20, 20, 40, 60}
	for i := 0; i < 5; i++ {
		res, err := BoundedRequest(ctx, req)

		if err != nil {
			d := time.Duration(timeoutIntervals[i]) * time.Second
			time.Sleep(d)
			continue
		}

		if res.StatusCode == http.StatusOK {
			return nil
		}

		if res.StatusCode == http.StatusInternalServerError {
			return errors.New(funcName + " :: " + "Internal server error occurred: " + err.Error())
		}

		if res.StatusCode == http.StatusUnprocessableEntity || res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusUnauthorized {
			return errors.New(funcName + " :: " + "Request was malformed: " + err.Error())
		}

		res.Body.Close()

		d := time.Duration(timeoutIntervals[i]) * time.Second
		time.Sleep(d)
	}

	return errors.New(funcName + " :: " + "Error occurred while sending HTTP request: " + err.Error())
}

func InitSession(sessType string) (string, error) {
	funcName := "InitSession"

	requestURL := fmt.Sprintf("%s://%s:%s/%s/%s", SERVER_TYPE, SERVER_HOST, SERVER_PORT, sessType, "init")
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		err = errors.New(funcName + " :: " + "Error occurred while creating HTTP request: " + err.Error())
		return "", err
	}

	res, err := BoundedRequest(context.TODO(), req)
	if err != nil {
		err = errors.New(funcName + " :: " + "Error occurred while sending HTTP request: " + err.Error())
		return "", err
	}

	for _, cookie := range res.Cookies() {
		if cookie.Name == "sessionID" {
			sessionID := cookie.Value
			return sessionID, nil
		}
	}

	err = errors.New(funcName + " :: " + "Session ID cookie is missing in HTTP response")
	return "", err

}

func GetStatus(ctx context.Context) (types.Status, error) {
	funcName := "GetStatus"
	sessionDetails := ctx.Value(KEY).(types.SessionDetails)

	requestURL := fmt.Sprintf("%s://%s:%s/%s/%s", SERVER_TYPE, SERVER_HOST, SERVER_PORT, sessionDetails.Type, "status")
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)

	if err != nil { // log failure
		err = errors.New(funcName + " :: " + "Error occurred while creating HTTP request: " + err.Error())
		return types.Status{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	sessionIDCookie := &http.Cookie{
		Name:  "sessionID",
		Value: sessionDetails.SessionID,
	}
	req.AddCookie(sessionIDCookie)

	res, err := BoundedRequest(ctx, req)
	if err != nil { // log failure
		err = errors.New(funcName + " :: " + "Error occurred while sending HTTP request: " + err.Error())
		return types.Status{}, err
	}

	if res.StatusCode == http.StatusOK {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			err = errors.New(funcName + " :: " + "Error occurred while reading response body: " + err.Error())
			return types.Status{}, err
		}

		var status types.Status
		err = json.Unmarshal(resBody, &status)

		if err != nil {
			err = errors.New(funcName + " :: " + "Error occurred while unmarshalling data: " + err.Error())
			return types.Status{}, err
		}

		return status, nil

	}

	err = errors.New(funcName + " :: " + "Error occurred while getting status")
	return types.Status{}, err
}

func AwaitServerCompletion(ctx context.Context, logger *zap.Logger) bool {
	timeoutIntervals := []int64{10, 20, 20, 40, 60}
	for i := 0; i < 5; {
		status, err := GetStatus(ctx)
		if err != nil {
			logger.Warn("Failed to get server status, retrying", zap.Error(err))
			d := time.Duration(timeoutIntervals[i]) * time.Second
			time.Sleep(d)
			i++
			continue
		}
		switch status.Code {
		case 1: // Processing
			time.Sleep(20 * time.Second) // may change
			logger.Info("Received server status", zap.String("ServerStatus", status.StatusString))
		case -1: // error
			logger.Info("Received server status", zap.String("ServerStatus", status.StatusString))
			return false
		case 0: // completed backup
			logger.Info("Received server status", zap.String("ServerStatus", status.StatusString))
			return true
		default:
			logger.Warn("Received unknown server status code, retrying", zap.String("ServerStatus", status.StatusString))
			d := time.Duration(timeoutIntervals[i]) * time.Second
			time.Sleep(d)
			i++
		}
	}
	return false
}

func SendAndReceive(ctx context.Context, endpoint string, data []byte) ([]byte, error) {

	funcName := "SendAndReceive"
	bodyReader := bytes.NewReader(data)
	sessionDetails := ctx.Value(KEY).(types.SessionDetails)
	requestURL := fmt.Sprintf("%s://%s:%s/%s/%s", SERVER_TYPE, SERVER_HOST, SERVER_PORT, sessionDetails.Type, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bodyReader)
	req.Close = true
	if err != nil { // log func failure
		err = errors.New(funcName + " :: " + "Error occurred while creating HTTP request: " + err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	sessionIDCookie := &http.Cookie{
		Name:  "sessionID",
		Value: sessionDetails.SessionID,
	}
	req.AddCookie(sessionIDCookie)

	timeoutIntervals := []int64{10, 20, 20, 40, 60}
	for i := 0; i < 5; i++ {
		res, err := BoundedRequest(context.TODO(), req)

		if err != nil { // logged, not failure
			d := time.Duration(timeoutIntervals[i]) * time.Second
			time.Sleep(d)
			continue
		}

		if res.StatusCode == http.StatusOK {
			respBody, err := io.ReadAll(res.Body)
			if err != nil {
				return nil, err
			}
			return respBody, nil
		}

		if res.StatusCode == http.StatusInternalServerError {
			err = errors.New(funcName + " :: " + "Internal server error occurred")
			return nil, err
		}

		if res.StatusCode == http.StatusUnprocessableEntity || res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusUnauthorized {
			err = errors.New(funcName + " :: " + "Request was malformed")
			return nil, err
		}

		res.Body.Close()

		d := time.Duration(timeoutIntervals[i]) * time.Second
		time.Sleep(d)
	}

	err = errors.New(funcName + " :: " + "Failed to send HTTP request")
	return nil, err
}

func BoundedRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	netwkSem.Acquire(ctx, 1)
	defer netwkSem.Release(1)
	resp, err := Client.Do(req)

	return resp, err

}
