package retrieve

import (
	"client-background/common"
	"client-background/types"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger
var statusErr error = nil

var dataStatus struct {
	mu        sync.RWMutex
	size      uint64
	totalSize uint64
}

func Retrieve(backupTime time.Time) error {
	var err error
	retrieveStart := time.Now().Round(time.Second)

	userIsLoggedIn := common.GetLoginState()

	if !userIsLoggedIn {
		common.GlobalLogger.Warn("No user is logged in currently")
		return errors.New("no user is logged in currently")
	}

	user, err := common.ReadCurrentUserData()
	if err != nil {
		logger.Error("Could not get current user data",
			zap.Error(err))
		statusErr = err
		return err
	}

	defer writeToSysHistoryFile(retrieveStart)

	retLogfileName := retrieveStart.Format(common.TIME_FORMAT) + ".log"
	logger, err = common.InitLogger(filepath.Join(common.GetRetrieveLogsDir(), retLogfileName))

	if err != nil {
		common.GlobalLogger.Error("Could not initialize retrieve logger", zap.Error(err))
		statusErr = err
		return err
	}

	sessionID, sessErr := common.InitSession("retrieve")
	if sessErr != nil {
		logger.Error("Could not initialize retrieve session",
			zap.Error(sessErr))
		statusErr = sessErr
		return sessErr
	}

	sessionDetails := types.SessionDetails{
		SessionID: sessionID,
		Type:      "retrieve",
	}

	ctx := context.WithValue(context.Background(), common.KEY, sessionDetails)

	backup := types.BackUpDirStruct{
		TimeStamp:    backupTime,
		Username:     user,
		ClientUtilID: common.GetClientID(),
	}

	backUpStructJSON, err := common.ToJSON(backup)
	if err != nil {
		logger.Error("Could not marshall backUpDirStruct into JSON",
			zap.Error(err))

		statusErr = err
		return err
	}

	response, err := common.SendAndReceive(ctx, "backupstruct", backUpStructJSON)

	dataStatus.mu.Lock()
	dataStatus.totalSize = binary.BigEndian.Uint64(response)
	fmt.Println(dataStatus.totalSize)
	dataStatus.mu.Unlock()

	if err != nil {
		logger.Error("Error while sending backupstruct",
			zap.Error(err))
		statusErr = err
		return err
	}

	restorePath := filepath.Join(common.GetRestoreDir(), "Retrieve at "+retrieveStart.Format(common.TIME_FORMAT), "Backup at "+backup.TimeStamp.Format(common.TIME_FORMAT))
	err = os.MkdirAll(restorePath, 0777)
	if err != nil {
		logger.Error("Error while creating restore directory",
			zap.Error(err))
		statusErr = err
		return err
	}

	err = retrieveFiles(ctx, restorePath)
	if err != nil {
		logger.Error("Could not retrieve files",
			zap.Error(err))
		statusErr = err
		return err
	}

	return nil
}

func GetDataStatus() (uint64, uint64) {
	dataStatus.mu.RLock()
	defer dataStatus.mu.RUnlock()

	return dataStatus.totalSize, dataStatus.size
}

func retrieveFiles(ctx context.Context, restorePath string) error {
	for {
		code, err := GetFile(ctx, restorePath)
		if err != nil {
			return err
		}

		switch code {
		case 1:
			continue
		case 0:
			return nil
		case -1:
			time.Sleep(2 * time.Second)
		}
	}
}

func writeToSysHistoryFile(retrieveStart time.Time) {

	entry := types.SysHistoryEntry{
		Timestamp: retrieveStart,
		Type:      "Retrieve",
	}

	if statusErr != nil {
		entry.Description = statusErr.Error()
		entry.Status = "Failure"
	} else {
		entry.Description = "Backup restored successfully at " + common.GetRestoreDir()
		entry.Status = "Success"
	}

	common.UpdateSysHistoryFile(entry)
}
