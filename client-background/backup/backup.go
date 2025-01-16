package backup

import (
	"client-background/cache"
	"client-background/common"
	"client-background/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// partial success for backup indicate
func BackUp(backUpStart time.Time) (types.BackUpDirStruct, error) { // fix return types

	userIsLoggedIn := common.GetLoginState()

	if !userIsLoggedIn {
		common.GlobalLogger.Warn("No user is logged in currently")
		return types.BackUpDirStruct{}, errors.New("no user is logged in currently")

	}

	user, err := common.ReadCurrentUserData()
	if err != nil {
		common.GlobalLogger.Error("Could not get current user data", zap.Error(err))
		return types.BackUpDirStruct{}, err
	}

	gotLock := backUpMu.TryLock()
	if !gotLock { // backup already in progress
		return types.BackUpDirStruct{}, errors.New("a backup operation is already ongoing") // custom err, notify frontend, log in client
	}

	defer backUpMu.Unlock()

	cache.Load()
	defer cache.Release()

	backUpStruct := types.BackUpDirStruct{
		TimeStamp:    backUpStart,
		Username:     user,
		ClientUtilID: common.GetClientID(),
	}

	err = initBackUp(backUpStart)
	if err != nil {
		common.GlobalLogger.Error("Could not initialize back up logger", zap.Error(err))
		entry := types.SysHistoryEntry{
			Timestamp:   backUpStart,
			Status:      statusStrings[2],
			Description: err.Error(),
			Type:        "Backup",
		}

		common.UpdateSysHistoryFile(entry)
		return backUpStruct, err
	}

	sessionID, sessErr := common.InitSession("backup")
	if sessErr != nil {
		logger.Error("Could not initialize back up session",
			zap.Error(err))
		entry := types.SysHistoryEntry{
			Timestamp:   backUpStart,
			Status:      statusStrings[2],
			Description: sessErr.Error(),
			Type:        "Backup",
		}

		common.UpdateSysHistoryFile(entry)
		return backUpStruct, sessErr
	}

	sessionDetails := types.SessionDetails{
		SessionID: sessionID,
		Type:      "backup",
	}

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), common.KEY, sessionDetails))
	defer cancel()

	// End of initialisation
	// ---------------------------------------------------------------------------
	//  Helper goroutines

	go statusListener(ctx)
	defer writeToSysHistoryFile(backUpStart)
	go checkFatalErrors(ctx, cancel)
	go handleSends(ctx)

	// ---------------------------------------------------------------------------

	dirs, err := getRegisteredDirs()
	if err != nil {
		logger.Error("Cannot get directories", zap.Error(err))
		setErrorStatus(err.Error())
		return backUpStruct, err
	}

	var dirSucceeded int = 0

	dirStructs := []types.Dirfile{}

	results := make(chan map[string]types.Dirfile, len(dirs))

	var totalSize uint64 = 0
	for _, entry := range dirs {
		current_size, err := getDirSize(entry)
		totalSize += uint64(current_size)
		if err != nil {
			logger.Error("Could not determine directory sizes",
				zap.Error(err))
			setErrorStatus(err.Error())
			return backUpStruct, err
		}
		fmt.Println(totalSize)

		go directoryHandler(ctx, entry, results)
	}

	status.mu.Lock()
	status.totalSize = totalSize
	status.mu.Unlock()

	for i := 0; i < len(dirs); {
		select {
		case <-ctx.Done():
			return backUpStruct, errors.New("something went wrong") // return custom error

		case dirMap := <-results:
			key := ""
			for k := range dirMap { // only one key always
				key = k
			}
			dirStruct := dirMap[key]
			if dirStruct.Valid {
				dirStruct.Name = key
				dirStructs = append(dirStructs, dirStruct)
				logger.Info("Directory processed", zap.String("name", dirStructs[i].Name))
				i++
				dirSucceeded++
			}
		}
	}
	backUpStruct.DirectoryArray = dirStructs
	status.mu.RLock()
	backUpStruct.Size = status.size
	status.mu.RUnlock()

	if ctx.Err() == nil {

		if dirSucceeded == 0 {
			err = errors.New("could not backup any directories")
			logger.Error("No data backed up",
				zap.Error(err))

			setErrorStatus(err.Error())

			return backUpStruct, err
		}
		// send hashes of chunk hits
		logger.Info("Directories processed successfully",
			zap.Int("count", dirSucceeded))
		UnmodChunkHandler(ctx, types.Hash{}, true)

		backUpStructJSON, err := common.ToJSON(backUpStruct)
		if err != nil {
			logger.Error("Could not marshall backUpDirStruct into JSON",
				zap.Error(err))

			setErrorStatus(err.Error())
			return backUpStruct, err
		}

		if ctx.Err() != nil {
			return backUpStruct, ctx.Err()
		}

		sendPacket := types.SendPacket{
			JsonBody: backUpStructJSON,
			Endpoint: "backupdirstruct",
		}
		channels.SendData <- sendPacket

		logger.Info("BackUpDirStruct sent to server")
		blockTillAllDataSent(ctx)

		if ctx.Err() != nil {
			return backUpStruct, ctx.Err()
		}

		requestsTracker.mu.RLock()
		count := requestsTracker.completedCount
		requestsTracker.mu.RUnlock()

		status := types.Status{
			Code:         0,
			StatusString: "Completed",
			Count:        count,
		}

		statusJSON, err := common.ToJSON(status)
		if err != nil {
			logger.Error("Could not marshall status into JSON",
				zap.Error(err),
				zap.Int("requestCount", status.Count))

			setErrorStatus(err.Error())
			return backUpStruct, err
		}

		sendPacket = types.SendPacket{
			JsonBody: statusJSON,
			Endpoint: "status",
		}

		if ctx.Err() != nil {
			return backUpStruct, ctx.Err()
		}

		err = common.Send(ctx, sendPacket)
		if err != nil {
			logger.Error("Could not send status to server",
				zap.Error(err),
				zap.Int("requestCount", status.Count))

			setErrorStatus(err.Error())
			return backUpStruct, err
		}

		logger.Info("Back up status sent to server",
			zap.Int("requestCount", status.Count))

		if serverSuccess := common.AwaitServerCompletion(ctx, logger); !serverSuccess {
			logger.Error("Canceling back up due to server failure")
			setErrorStatus("server failed")

			return backUpStruct, errors.New("server failed")
		}

		if ctx.Err() != nil {
			return backUpStruct, ctx.Err()
		}

		cache.Persist()

		common.SetLastBackUpTime(backUpStruct.TimeStamp)

		writeToFile(backUpStructJSON, backUpStruct.TimeStamp)
		ChunkingStats.mu.Lock()
		speed := float64(ChunkingStats.Size) / float64(ChunkingStats.Duration.Seconds())
		logger.Info("Stats",
			zap.Uint64("Total Data Processed", ChunkingStats.Size),
			zap.Uint64("Total Chunks Processed", ChunkingStats.Num),
			zap.Float64("Average Chunking Speed", speed))
		ChunkingStats.mu.Unlock()
	}
	return backUpStruct, ctx.Err()
}

func checkFatalErrors(ctx context.Context, cancel context.CancelFunc) {
	done := false

	for !done {
		select {
		case err := <-channels.Err:
			logger.Error("Back up failed",
				zap.Error(err))

			setErrorStatus(err.Error())
			done = true

			cancel()
		default:
			continue
		}
	}
}

func blockTillAllDataSent(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(1 * time.Second)
			requestsTracker.mu.RLock()
			completedCount := requestsTracker.completedCount
			begunCount := requestsTracker.begunCount
			requestsTracker.mu.RUnlock()

			if completedCount == begunCount {
				return
			}
		}
	}
}

func getRegisteredDirs() ([]string, error) {
	funcName := "getRegisteredDirs"
	file, err := os.OpenFile(common.GetDirectoriesFile(), os.O_RDONLY, 0644)
	if err != nil {
		logger.Error("Failed to open file",
			zap.String("handler", funcName),
			zap.Error(err))
		return nil, err
	}

	decoder := json.NewDecoder(file)

	var ManageDirs struct {
		Dirs []string
	}

	err = decoder.Decode(&ManageDirs)
	if err != nil {
		logger.Error("Failed to decode directories",
			zap.String("handler", funcName),
			zap.Error(err))
		return nil, err
	}

	return ManageDirs.Dirs, nil
}

func writeToFile(backUpStructJSON []byte, timestamp time.Time) {
	funcName := "writeToFile"
	filename := timestamp.Format(common.TIME_FORMAT) + ".bkup"
	bkupfile, err := os.OpenFile(filepath.Join(common.GetBackUpsDir(), filename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logger.Error("Failed to open file",
			zap.String("handler", funcName),
			zap.Error(err))
	}

	_, err = bkupfile.Write(backUpStructJSON)
	if err != nil {
		logger.Error("Failed to write backupdirstruct to file",
			zap.String("handler", funcName),
			zap.Error(err))
	}

}
