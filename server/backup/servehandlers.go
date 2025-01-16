package backup

import (
	"dedup-server/clienttypes"
	"dedup-server/logging"
	"dedup-server/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// need to put session id in logger?
func ServeChunk(resWriter http.ResponseWriter, req *http.Request) {
	funcName := "ServeChunk"
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
		sess.Core.Logger.Warn("Unable to read request body",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	var chunk clienttypes.Chunk
	err = json.Unmarshal(reqBody, &chunk)

	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		sess.Core.Logger.Warn("Unable to process request body to JSON",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName))
		return
	}
	RefreshTimeStamp(sess)
	go ChunkHandler(sess, chunk)

}

func ServeFileMeta(resWriter http.ResponseWriter, req *http.Request) {
	funcName := "ServeFileMeta"
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
		fmt.Println("ServeFileMeta :: could not read request body: ", err)
		return
	}

	var fileMD clienttypes.FileMetadata
	err = json.Unmarshal(reqBody, &fileMD)

	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		sess.Core.Logger.Warn("Unable to process request body to JSON",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName))
		return
	}

	RefreshTimeStamp(sess)
	go FileMetaHandler(sess, fileMD)

}

func ServeUnmodifiedChunks(resWriter http.ResponseWriter, req *http.Request) {
	funcName := "ServeUnmodifiedChunks"
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
		fmt.Println("ServeUnmodifiedChunks :: could not read request body: ", err)
		return
	}

	var chunkHashes clienttypes.ChunkHashes
	err = json.Unmarshal(reqBody, &chunkHashes)

	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		sess.Core.Logger.Warn("Unable to process request body to JSON",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName))
		return
	}

	RefreshTimeStamp(sess)
	go UnmodChunksHandler(sess, chunkHashes)
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
		fmt.Println("ServeDirStruct :: could not read request body: ", err)
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

	RefreshTimeStamp(sess)
	go DirStructHandler(sess, dirStruct)

}

func ServeInvalidChunks(resWriter http.ResponseWriter, req *http.Request) {
	funcName := "ServeInvalidChunks"
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
		fmt.Println("ServeInvalidChunks :: could not read request body: ", err)
		return
	}

	var chunkHashes clienttypes.ChunkHashes
	err = json.Unmarshal(reqBody, &chunkHashes)

	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		sess.Core.Logger.Warn("Unable to process request body to JSON",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName))
		return
	}
	RefreshTimeStamp(sess)
	go InvalidChunksHandler(sess, chunkHashes)
}

func ServeStatus(resWriter http.ResponseWriter, req *http.Request) {
	funcName := "ServeStatus"
	sess, code := ValidateSession(req)
	if code != 0 {
		resWriter.WriteHeader(code)
		logging.GlobalLogger.Info("Invalid sessionID", zap.String("handler", funcName), zap.Int("httpCode", code))
		return
	}

	if req.Method == http.MethodGet {
		var status clienttypes.Status
		err := sess.Core.CheckError()

		completionStatus := IsCompleted(sess)
		completedCount := GetCompletedCount(sess)
		clientCount := GetClientCount(sess)
		killFlag := false

		if completionStatus {
			status = clienttypes.Status{
				Code:         0,
				StatusString: "Completed",
				Count:        completedCount,
			}
			killFlag = true
		} else if err == nil {
			status = clienttypes.Status{
				Code:         1,
				StatusString: "Processing",
				Count:        completedCount,
			}
		} else {
			status = clienttypes.Status{
				Code:         -1,
				StatusString: err.Error(),
				Count:        completedCount,
			}
			killFlag = true
		}
		statusJSON, err := json.Marshal(status)
		if err != nil {
			resWriter.WriteHeader(http.StatusUnprocessableEntity)
			sess.Core.Logger.Warn("Unable to process request body to JSON",
				zap.String("sessionID", sess.Core.SessionID),
				zap.String("handler", funcName))
			return
		}
		sess.Core.Logger.Info("Completion status",
			zap.Int("Completed count", completedCount),
			zap.Int("Target count", clientCount))
		resWriter.Write(statusJSON)
		if killFlag {
			go sess.HandleExpiry()
		}
	}

	if req.Method == http.MethodPost {

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
			fmt.Println("ServeStatus :: could not read request body: ", err)
			return
		}

		var status clienttypes.Status
		err = json.Unmarshal(reqBody, &status)

		if err != nil {
			resWriter.WriteHeader(http.StatusUnprocessableEntity)
			sess.Core.Logger.Warn("Unable to process request body to JSON",
				zap.String("sessionID", sess.Core.SessionID),
				zap.String("handler", funcName))
			return
		}

		sess.Core.Logger.Info("Client status received",
			zap.Int("RequestCount", status.Count),
			zap.String("ClientStatus", status.StatusString))
		SetClientCount(sess, status.Count)
	}

}

func ServeInit(w http.ResponseWriter, r *http.Request) {
	// TODO: user authentication

	sessionID, err := CreateBackUpSession()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logging.GlobalLogger.Error("Couldn't create new backup session", zap.Error(err))
		return
	}

	cookie := &http.Cookie{
		Name:  "sessionID",
		Value: sessionID,
	}

	http.SetCookie(w, cookie)
	// logging.GlobalLogger.Info("Created new backup session", zap.String("sessionID", sessionID))
}

func ValidateSession(req *http.Request) (*BackUpSession, int) {
	sessionCookie, err := req.Cookie("sessionID")

	if err != nil {
		return nil, http.StatusBadRequest
	}

	sess, err := GetBackUpSession(sessionCookie.Value)

	if err != nil {
		return nil, http.StatusUnauthorized
	}

	return sess, 0
}
