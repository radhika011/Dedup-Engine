package delete

import (
	"client-background/cache"
	"client-background/common"
	"client-background/types"
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger
var statusErr error = nil

func Delete(backupTime time.Time) error {
	var err error
	deleteStart := time.Now().Round(time.Second)

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

	defer writeToSysHistoryFile(deleteStart)

	delLogFileName := deleteStart.Format(common.TIME_FORMAT) + ".log"
	logger, err = common.InitLogger(filepath.Join(common.GetDeleteLogsDir(), delLogFileName))

	if err != nil {
		common.GlobalLogger.Error("Could not initialize deletion logger", zap.Error(err))

		statusErr = err
		return err
	}

	if backupTime == common.GetLastBackUpTime() {
		cache.Invalidate()
	}

	sessionID, sessErr := common.InitSession("delete")
	if sessErr != nil {
		logger.Error("Could not initialize delete session",
			zap.Error(sessErr))
		statusErr = sessErr
		return sessErr
	}

	sessionDetails := types.SessionDetails{
		SessionID: sessionID,
		Type:      "delete",
	}

	ctx := context.WithValue(context.Background(), common.KEY, sessionDetails)

	backup := types.BackUpDirStruct{
		TimeStamp:    backupTime,
		Username:     user,
		ClientUtilID: common.GetClientID(),
	}

	backUpStructJSON, err := common.ToJSON(backup)
	if err != nil {
		logger.Error("Error while sending backupstruct",
			zap.Error(err))

		statusErr = err
		return err
	}

	sendPacket := types.SendPacket{
		JsonBody: backUpStructJSON,
		Endpoint: "backupstruct",
	}
	common.Send(ctx, sendPacket)

	if serverSuccess := common.AwaitServerCompletion(ctx, logger); !serverSuccess {
		logger.Error("Canceling delete due to server failure")
		err = errors.New("server failed")
		statusErr = err
		return err
	}
	err = deleteFile(backupTime)
	logger.Error("Failed to delete .bkup file", zap.Error(err))

	return nil
}

func writeToSysHistoryFile(deleteStart time.Time) {

	entry := types.SysHistoryEntry{
		Timestamp: deleteStart,
		Type:      "Delete",
	}

	if statusErr != nil {
		entry.Description = statusErr.Error()
		entry.Status = "Failure"
	} else {
		entry.Description = "Deleted successfully" // TODO what here?
		entry.Status = "Success"
	}

	common.UpdateSysHistoryFile(entry)
}

func deleteFile(backUpTime time.Time) error {
	filename := backUpTime.Format(common.TIME_FORMAT) + ".bkup"
	fpath := filepath.Join(common.GetBackUpsDir(), filename)
	err := os.Remove(fpath)
	return err
}
