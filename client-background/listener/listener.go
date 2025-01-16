package listener

import (
	"client-background/backup"
	"client-background/common"
	"client-background/delete"
	"client-background/retrieve"
	"client-background/types"
	"client-background/user"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

// TODO move to common

var origin string

func Listen(scheduleChan chan struct{}) {

	origin = "TCP Listener on " + common.BG_TCP_HOST + ":" + common.BG_TCP_PORT
	server, err := net.Listen(common.BG_TCP_TYPE, common.BG_TCP_HOST+":"+common.BG_TCP_PORT)

	if err != nil {
		common.GlobalLogger.Error("Error while listening",
			zap.String("Origin", origin),
			zap.Error(err))

	}

	common.GlobalLogger.Info("Listening",
		zap.String("Origin", origin))

	for {
		connection, err := server.Accept()
		if err != nil {
			common.GlobalLogger.Error("Error while accepting connection ",
				zap.String("Origin", origin),
				zap.Error(err))
		}
		common.GlobalLogger.Info("Client connection accepted",
			zap.String("Origin", origin))

		go processClient(connection, scheduleChan)
	}
}

func processClient(connection net.Conn, scheduleChan chan struct{}) {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		common.GlobalLogger.Error("Error while reading from connection",
			zap.String("Origin", origin),
			zap.Error(err))
	}

	var request types.InterfaceRequest
	err = json.Unmarshal(buffer[:mLen], &request)
	if err != nil {
		common.GlobalLogger.Error("Error while unmarshalling data from connection buffer",
			zap.String("Origin", origin),
			zap.Error(err))
	}

	var params map[string]interface{}
	err = json.Unmarshal(request.Parameters, &params)
	if err != nil {
		common.GlobalLogger.Error("Error while unmarshalling parameters",
			zap.String("Origin", origin),
			zap.Error(err))
	}

	common.GlobalLogger.Info("Received UI request of type "+request.Type,
		zap.String("Origin", origin))

	switch request.Type {
	case "schedule":
		scheduleChan <- struct{}{}
	case "backup":
		handleBackup(params, connection)
	case "retrieve":
		handleRetrieve(params, connection)
	case "delete":
		fmt.Println(time.Now())
		resp := handleDelete(params)
		connection.Write(resp)
		fmt.Println(time.Now())
	case "register":
		resp := handleRegister(params)
		connection.Write(resp)
	case "login":
		resp := handleLogin(params)
		connection.Write(resp)
	case "verify":
		resp := handleVerify(params)
		connection.Write(resp)
	case "update":
		resp := handleUpdate(params)
		connection.Write(resp)
	case "logout":
		resp := handleLogout(params)
		connection.Write(resp)
	}
	connection.Close()
}

func handleRegister(params map[string]interface{}) []byte {
	op := "Register"
	var ifaceResp types.InterfaceResponse
	var respParams types.UserResponseParams
	var paramBytes []byte
	var err error

	var userData types.UserData
	err = mapstructure.Decode(params["UserData"], &userData)
	if err != nil {
		common.GlobalLogger.Error("Error decoding parameters into struct",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	respParams, err = user.RegisterUser(userData)
	if err != nil {
		common.GlobalLogger.Error("Failed to register user",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	paramBytes, err = json.Marshal(respParams)
	if err != nil {
		common.GlobalLogger.Error("Failed to marshal response parameters",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	ifaceResp.Code = 0
	ifaceResp.Parameters = paramBytes

MARSHAL:
	ifaceRespBytes, err := json.Marshal(ifaceResp)
	if err != nil {
		common.GlobalLogger.Error("Failed to marshal response",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		return nil
	}

	return ifaceRespBytes
}

func handleUpdate(params map[string]interface{}) []byte {
	op := "Update"
	var ifaceResp types.InterfaceResponse
	var respParams types.UserResponseParams
	var paramBytes []byte
	var err error

	var userData types.UserData
	err = mapstructure.Decode(params["UserData"], &userData)
	if err != nil {
		common.GlobalLogger.Error("Error decoding parameters into struct",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	respParams, err = user.UpdateUser(userData)
	if err != nil {
		common.GlobalLogger.Error("Failed to perform user data update",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	paramBytes, err = json.Marshal(respParams)
	if err != nil {
		common.GlobalLogger.Error("Failed to marshal response parameters",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	ifaceResp.Code = 0
	ifaceResp.Parameters = paramBytes

MARSHAL:
	ifaceRespBytes, err := json.Marshal(ifaceResp)
	if err != nil {
		common.GlobalLogger.Error("Failed to marshal response",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		return nil
	}

	return ifaceRespBytes
}

func handleLogin(params map[string]interface{}) []byte {
	op := "Login"
	var ifaceResp types.InterfaceResponse
	var respParams types.UserResponseParams
	var paramBytes []byte
	var err error

	var userCredentials types.UserData
	err = mapstructure.Decode(params["userCredentials"], &userCredentials)
	if err != nil {
		common.GlobalLogger.Error("Error decoding parameters into struct",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	respParams, err = user.LoginUser(userCredentials)
	// here
	if err != nil {
		common.GlobalLogger.Error("Failed to login user",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	paramBytes, err = json.Marshal(respParams)
	if err != nil {
		common.GlobalLogger.Error("Failed to marshal response parameters",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	ifaceResp.Code = 0
	ifaceResp.Parameters = paramBytes

	common.SetLoginState(true)
	common.LoadPersistedState()

MARSHAL:
	ifaceRespBytes, err := json.Marshal(ifaceResp)
	if err != nil {
		common.GlobalLogger.Error("Failed to marshal response",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		return nil
	}

	return ifaceRespBytes
}

func handleLogout(params map[string]interface{}) []byte {
	op := "Logout"
	var ifaceResp types.InterfaceResponse
	var err error

	active.mu.Lock()
	processOngoing := len(active.processes) > 0
	active.mu.Unlock()

	if processOngoing {
		ifaceResp.Code = -1
	} else {
		ifaceResp.Code = 0
		common.SetLoginState(false)
	}

	ifaceRespBytes, err := json.Marshal(ifaceResp)
	if err != nil {
		common.GlobalLogger.Error("Failed to marshal response",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		return nil
	}

	return ifaceRespBytes
}

func handleVerify(params map[string]interface{}) []byte {
	op := "Verify"
	var ifaceResp types.InterfaceResponse
	var respParams types.UserResponseParams
	var paramBytes []byte
	var err error

	var userCredentials types.UserData
	err = mapstructure.Decode(params["userCredentials"], &userCredentials)
	if err != nil {
		common.GlobalLogger.Error("Error decoding parameters into struct",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	respParams, err = user.VerifyUser(userCredentials)
	if err != nil {
		common.GlobalLogger.Error("Failed to verify user",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	paramBytes, err = json.Marshal(respParams)
	if err != nil {
		common.GlobalLogger.Error("Failed to marshal response parameters",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	ifaceResp.Code = 0
	ifaceResp.Parameters = paramBytes

MARSHAL:
	ifaceRespBytes, err := json.Marshal(ifaceResp)
	if err != nil {
		common.GlobalLogger.Error("Failed to marshal response",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		return nil
	}

	return ifaceRespBytes
}

func handleBackup(params map[string]interface{}, connection net.Conn) {
	op := "Backup"

	exit := make(chan struct{})
	go handleBackupProgress(connection, exit)

	var ifaceResp types.InterfaceResponse
	var respParams types.ResponseParam
	var paramBytes []byte
	var totsize, size uint64
	var err error = nil

	backupStart := time.Now().Round(time.Second)
	timestampStr := backupStart.Format(common.TIME_FORMAT)
	added := AddToActiveProcesses(op, timestampStr)
	if added {
		_, err = backup.BackUp(backupStart)
		RemoveFromActiveProcesses(op, timestampStr)
		exit <- struct{}{}

		totsize, size = backup.GetDataStatus()

		if err != nil {
			common.GlobalLogger.Error("Backup failed",
				zap.String("Origin", origin), zap.String("Operation", op),
				zap.Error(err))
			ifaceResp.Code = -1
			goto MARSHAL
		}
	}

	respParams.ProcessedData = size
	respParams.TotalData = totsize
	paramBytes, err = json.Marshal(respParams)
	if err != nil {
		common.GlobalLogger.Error("Failed to marshal response parameters",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	ifaceResp.Code = 0
	ifaceResp.Parameters = paramBytes

MARSHAL:
	ifaceRespBytes, err := json.Marshal(ifaceResp)
	if err != nil {
		common.GlobalLogger.Error("Failed to marshal response",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		connection.Write(nil)
	}

	connection.Write(ifaceRespBytes)
}

func handleBackupProgress(connection net.Conn, exit chan struct{}) {
	op := "Backup Progress"

	var ifaceResp types.InterfaceResponse
	var respParams types.ResponseParam
	var paramBytes []byte
	var err error

	stopFlag := false
	for !stopFlag {
		select {
		case <-exit:
			stopFlag = true
		default:

			time.Sleep(1 * time.Second)
			totsize, size := backup.GetDataStatus()
			respParams.ProcessedData = size
			respParams.TotalData = totsize
			paramBytes, err = json.Marshal(respParams)
			if err != nil {
				common.GlobalLogger.Error("Failed to marshal response parameters",
					zap.String("Origin", origin), zap.String("Operation", op),
					zap.Error(err))
				ifaceResp.Code = -1
				goto MARSHAL
			}

			ifaceResp.Code = 1
			ifaceResp.Parameters = paramBytes

		MARSHAL:
			ifaceRespBytes, err := json.Marshal(ifaceResp)
			if err != nil {
				common.GlobalLogger.Error("Failed to marshal response",
					zap.String("Origin", origin), zap.String("Operation", op),
					zap.Error(err))
				connection.Write(nil)
			}

			connection.Write(ifaceRespBytes)
		}
	}
}

func handleDelete(params map[string]interface{}) []byte {
	op := "Delete"

	var ifaceResp types.InterfaceResponse
	var err error
	var added bool
	var timestamp time.Time

	timestampStr, ok := params["Timestamp"].(string)
	if !ok {
		common.GlobalLogger.Error("Error parsing parameters",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}
	timestamp, err = time.Parse(common.TIME_FORMAT, timestampStr)
	if err != nil {
		common.GlobalLogger.Error("Error parsing timestamp",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}
	added = AddToActiveProcesses(op, timestampStr)
	if added {
		err = delete.Delete(timestamp)
		RemoveFromActiveProcesses(op, timestampStr)
		if err != nil {
			common.GlobalLogger.Error("Failed to delete backup",
				zap.String("Origin", origin), zap.String("Operation", op),
				zap.Error(err))
			ifaceResp.Code = -1
			goto MARSHAL
		} else {
			ifaceResp.Code = 0
			goto MARSHAL
		}

	} else {
		ifaceResp.Code = -1
		goto MARSHAL
	}

MARSHAL:
	ifaceRespBytes, err := json.Marshal(ifaceResp)
	if err != nil {
		common.GlobalLogger.Error("Failed to marshal response",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.Error(err))
		return nil
	}

	return ifaceRespBytes
}

func handleRetrieve(params map[string]interface{}, connection net.Conn) {
	op := "Retrieve"

	exit := make(chan struct{})
	go handleRetrieveProgress(connection, exit)

	var ifaceResp types.InterfaceResponse
	var respParams types.ResponseParam
	var paramBytes []byte
	var size uint64
	var totalSize uint64
	var added bool
	var timestamp time.Time
	var err error

	timestampStr, ok := params["Timestamp"].(string)
	if !ok {
		common.GlobalLogger.Error("Error parsing parameters",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}
	timestamp, err = time.Parse(common.TIME_FORMAT, timestampStr)
	if err != nil {
		common.GlobalLogger.Error("Error parsing timestamp",
			zap.String("Origin", origin), zap.String("Operation", op),
			zap.String("Operation", op),
			zap.Error(err))
		ifaceResp.Code = -1
		goto MARSHAL
	}

	added = AddToActiveProcesses(op, timestampStr)
	if added {
		err = retrieve.Retrieve(timestamp)
		if err != nil {
			common.GlobalLogger.Error("Failed to retrieve backup",
				zap.String("Origin", origin), zap.String("Operation", op),
				zap.String("Operation", op),
				zap.Error(err))
			ifaceResp.Code = -1
			goto MARSHAL
		}
		RemoveFromActiveProcesses(op, timestampStr)
		exit <- struct{}{}
		totalSize, size = retrieve.GetDataStatus()

		respParams.ProcessedData = size
		respParams.TotalData = totalSize
		paramBytes, err = json.Marshal(ifaceResp)
		if err != nil {
			common.GlobalLogger.Error("Failed to marshal response parameters",
				zap.String("Origin", origin), zap.String("Operation", op),
				zap.String("Operation", op),
				zap.Error(err))
			ifaceResp.Code = -1
			goto MARSHAL
		}
		ifaceResp.Parameters = paramBytes
		ifaceResp.Code = 0
	} else {
		ifaceResp.Code = -1
		goto MARSHAL
	}

MARSHAL:
	ifaceRespBytes, err := json.Marshal(ifaceResp)
	if err != nil {
		common.GlobalLogger.Error("Failed to marshal response",
			zap.String("Origin", origin), zap.String("Operation", op), zap.String("Operation", op),
			zap.Error(err))
		return
	}

	connection.Write(ifaceRespBytes)
}

func handleRetrieveProgress(connection net.Conn, exit chan struct{}) {
	op := "Retrieve Progress"

	var ifaceResp types.InterfaceResponse
	var respParams types.ResponseParam
	var paramBytes []byte
	var err error

	stopFlag := false
	for !stopFlag {
		select {
		case <-exit:
			stopFlag = true
		default:

			time.Sleep(1 * time.Second)
			totalSize, size := retrieve.GetDataStatus()
			respParams.ProcessedData = size
			respParams.TotalData = totalSize
			paramBytes, err = json.Marshal(respParams)
			if err != nil {
				common.GlobalLogger.Error("Failed to marshal response parameters",
					zap.String("Origin", origin), zap.String("Operation", op), zap.String("Operation", op),
					zap.Error(err))
				ifaceResp.Code = -1
				goto MARSHAL
			}

			ifaceResp.Code = 1
			ifaceResp.Parameters = paramBytes

		MARSHAL:
			ifaceRespBytes, err := json.Marshal(ifaceResp)
			if err != nil {
				common.GlobalLogger.Error("Failed to marshal response",
					zap.String("Origin", origin), zap.String("Operation", op), zap.String("Operation", op),
					zap.Error(err))
				connection.Write(nil)
			}

			connection.Write(ifaceRespBytes)
		}
	}
}
