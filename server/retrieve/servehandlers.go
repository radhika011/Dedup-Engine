package retrieve

import (
	"dedup-server/common"
	"dedup-server/logging"
	"dedup-server/types"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func ServeInit(w http.ResponseWriter, r *http.Request) {
	// TODO: user authentication
	sessionID, err := CreateRetrieveSession()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logging.GlobalLogger.Error("Creating new retrieve session", zap.Error(err))
	}

	cookie := &http.Cookie{
		Name:  "sessionID",
		Value: sessionID,
	}

	http.SetCookie(w, cookie)
	// logging.GlobalLogger.Info("Created new delete session", zap.String("sessionID", sessionID))
}

func ServeDirStruct(resWriter http.ResponseWriter, req *http.Request) {

	funcName := "ServeDirStruct"

	sess, code := ValidateSession(req)
	if code != 0 {
		resWriter.WriteHeader(code)
		logging.GlobalLogger.Info("Invalid sessionID", zap.String("handler", funcName), zap.Int("httpCode", code))
		return
	}

	err := sess.Core.CheckError()

	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		sess.Core.Logger.Info("Request failed due to bad session status",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName))
		return
	}

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	var dirStruct types.BackUpDirStruct
	err = json.Unmarshal(reqBody, &dirStruct)

	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		sess.Core.Logger.Warn("Unable to process request body to JSON",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName))
		return
	}

	sizeChan := make(chan uint64)
	go RetrieveHandler(sess, dirStruct, sizeChan)

	size := <-sizeChan

	sizeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBytes, size)

	resWriter.Write(sizeBytes)

}

func ServeFile(resWriter http.ResponseWriter, req *http.Request) {
	funcName := "ServeFile"

	sess, code := ValidateSession(req)
	if code != 0 {
		resWriter.WriteHeader(code)
		logging.GlobalLogger.Info("Invalid sessionID", zap.String("handler", funcName), zap.Int("httpCode", code))
		return
	}

	err := sess.Core.CheckError()

	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		sess.Core.Logger.Info("Request failed due to bad session status",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName))
		return
	}

	select {
	case fileStruct := <-sess.Files:
		{
			dirPath := fileStruct.DirPath

			cookie := &http.Cookie{
				Name:  "dirPath",
				Value: dirPath,
			}
			// sess.Core.Logger.Info("FYI", zap.String("dirPath", dirPath))
			// resWriter.Header().Add("dirPath", dirPath)

			http.SetCookie(resWriter, cookie)
			http.ServeFile(resWriter, req, fileStruct.ServerPath)
		}
	default:
		sess.CountMu.RLock()
		fileWorkerCompletionCount := sess.FileWorkerCompletionCount
		sess.CountMu.RUnlock()
		if fileWorkerCompletionCount == MaxFileWorkers {
			resWriter.WriteHeader(common.HTTPStatusRetrievalCompleted) // make constant
			go sess.HandleExpiry()
			return
		} else {
			resWriter.WriteHeader(http.StatusAccepted) // processing code
			return
		}
	}

}

func ValidateSession(req *http.Request) (*RetrieveSession, int) {
	sessionCookie, err := req.Cookie("sessionID")

	if err != nil {
		return nil, http.StatusBadRequest
	}

	sess, err := GetRetrieveSession(sessionCookie.Value)

	if err != nil {
		return nil, http.StatusUnauthorized
	}

	return sess, 0
}
