package user

import (
	"dedup-server/logging"
	"dedup-server/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func ServeRegister(resWriter http.ResponseWriter, req *http.Request) {
	funcName := "ServeRegister"

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		logging.GlobalLogger.Error("Could not read request body",
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	var user types.UserData
	err = json.Unmarshal(reqBody, &user)

	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		logging.GlobalLogger.Error("Unable to process request body to JSON",
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	response, err := RegisterDBHandler(user)
	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		logging.GlobalLogger.Error("Could not handle database operations",
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		logging.GlobalLogger.Error("Unable to process response body to JSON",
			zap.Error(err),
			zap.String("handler", funcName))
		return
	}
	logging.GlobalLogger.Info("Registered user",
		zap.String("user", user.EmailID),
		zap.String("handler", funcName))

	resWriter.Write(responseJSON)

}

func ServeVerify(resWriter http.ResponseWriter, req *http.Request) {
	funcName := "ServeVerify"

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		logging.GlobalLogger.Error("Could not read request body",
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	var user types.UserData
	err = json.Unmarshal(reqBody, &user)

	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		logging.GlobalLogger.Error("Unable to process request body to JSON",
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	response, err := VerifyDBHandler(user)
	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		logging.GlobalLogger.Error("Could not handle database operations",
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		logging.GlobalLogger.Error("Unable to process response body to JSON",
			zap.Error(err),
			zap.String("handler", funcName))
		return
	}
	logging.GlobalLogger.Info("Verified user",
		zap.String("user", user.EmailID),
		zap.String("handler", funcName))

	resWriter.Write(responseJSON)
}

func ServeLogin(resWriter http.ResponseWriter, req *http.Request) {
	funcName := "ServeLogin"

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		logging.GlobalLogger.Error("Could not read request body",
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	var user types.UserData
	err = json.Unmarshal(reqBody, &user)

	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		logging.GlobalLogger.Error("Unable to process request body to JSON",
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	response, err := LoginDBHandler(user)
	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		logging.GlobalLogger.Error("Could not handle database operations",
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		logging.GlobalLogger.Error("Unable to process response body to JSON",
			zap.Error(err),
			zap.String("handler", funcName))
		return
	}
	logging.GlobalLogger.Info("Verified user",
		zap.String("user", user.EmailID),
		zap.String("handler", funcName))

	resWriter.Write(responseJSON)
}

func ServeUpdate(resWriter http.ResponseWriter, req *http.Request) {
	funcName := "ServeUpdate"

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		logging.GlobalLogger.Error("Could not read request body",
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	var user types.UserData
	err = json.Unmarshal(reqBody, &user)

	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		logging.GlobalLogger.Error("Unable to process request body to JSON",
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	response, err := UpdateDBHandler(user)
	if err != nil {
		resWriter.WriteHeader(http.StatusInternalServerError)
		logging.GlobalLogger.Error("Could not handle database operations",
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	fmt.Println("update: ", response)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		resWriter.WriteHeader(http.StatusUnprocessableEntity)
		logging.GlobalLogger.Error("Unable to process response body to JSON",
			zap.Error(err),
			zap.String("handler", funcName))
		return
	}
	logging.GlobalLogger.Info("Updated user data",
		zap.String("user", user.EmailID),
		zap.String("handler", funcName))

	resWriter.Write(responseJSON)

}
