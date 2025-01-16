package delete

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

func ServeInit(w http.ResponseWriter, r *http.Request) {
	// TODO: user authentication

	sessionID, err := CreateDeleteSession()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logging.GlobalLogger.Error("Creating new delete session", zap.Error(err))
	}

	cookie := &http.Cookie{
		Name:  "sessionID",
		Value: sessionID,
	}

	http.SetCookie(w, cookie)
	// logging.GlobalLogger.Info("Created new delete session", zap.String("sessionID", sessionID))
}

func ServeStatus(resWriter http.ResponseWriter, req *http.Request) {
	funcName := "ServeStatus"
	sess, code := ValidateSession(req)
	if code != 0 {
		resWriter.WriteHeader(code)
		logging.GlobalLogger.Info("Invalid sessionID", zap.String("handler", funcName), zap.Int("httpCode", code))
		return
	}

	var status clienttypes.Status
	err := sess.Core.CheckError()
	IsCompleted := sess.IsCompleted

	if IsCompleted {
		status = clienttypes.Status{
			Code:         0,
			StatusString: "Completed",
		}
	} else if err == nil {
		status = clienttypes.Status{
			Code:         1,
			StatusString: "Processing",
		}
	} else {
		status = clienttypes.Status{
			Code:         -1,
			StatusString: err.Error(),
		}
	}
	statusJSON, err := json.Marshal(status)
	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		sess.Core.Logger.Warn("Unable to process request body to JSON",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName))
		return
	}
	sess.Core.Logger.Info("",
		zap.Bool("IsCompleted", IsCompleted))
	resWriter.Write(statusJSON)
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

	if dirStruct.DirectoryArray != nil {
		resWriter.WriteHeader(http.StatusBadRequest)
		sess.Core.Logger.Warn("Bad Request",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName))
		return
	}

	go deleteHandler(sess, dirStruct)

}

func ValidateSession(req *http.Request) (*DeleteSession, int) {
	sessionCookie, err := req.Cookie("sessionID")

	if err != nil {
		return nil, http.StatusBadRequest
	}

	sess, err := GetDeleteSession(sessionCookie.Value)

	if err != nil {
		return nil, http.StatusUnauthorized
	}

	return sess, 0
}
