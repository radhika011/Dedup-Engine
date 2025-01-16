package backup

import (
	"dedup-server/logging"
	"dedup-server/session"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

const rootPath = "temp" // eventually env variable
const logpath = "temp"

type BackUpSession struct {
	Core                 *session.SessionCore
	Chunks               *session.SessionFile
	FileMeta             *session.SessionFile
	InvalidChunks        *session.SessionFile
	RequestTracker       *RequestCounts
	LastRequestTimeStamp *session.TimeStamp
}

type RequestCounts struct {
	Mu          sync.RWMutex
	Started     int
	Completed   int
	ClientCount int
}

func (bkupsess *BackUpSession) CloseFiles() {
	funcName := "CloseFiles"
	err := bkupsess.Chunks.Close()
	if err != nil {
		bkupsess.Core.Logger.Error("Failed to close file",
			zap.String("handler", funcName),
			zap.String("file", "chunks.dat"),
			zap.Error(err))
	}
	err = bkupsess.FileMeta.Close()
	if err != nil {
		bkupsess.Core.Logger.Error("Failed to close file",
			zap.String("handler", funcName),
			zap.String("file", "filemeta.dat"),
			zap.Error(err))
	}
	err = bkupsess.InvalidChunks.Close()
	if err != nil {
		bkupsess.Core.Logger.Error("Failed to close file",
			zap.String("handler", funcName),
			zap.String("file", "invalidchunks.dat"),
			zap.Error(err))
	}

	bkupsess.Core.Logger.Info("Session files closed",
		zap.String("handler", funcName))
}

func ReadHashesFromFile(sessFile *session.SessionFile) ([][32]byte, error) {
	hashes := [][32]byte{}

	for {
		hash := make([]byte, 32)
		n, err := sessFile.File.Read(hash)

		if err == io.EOF {
			break
		}

		if err != nil {
			// fmt.Println("ReadHashesFromFile :: error reading invalid chunk hashes from file")
			return hashes, err
			// FATAL
		}

		if n != 32 {
			// fmt.Println("ReadHashesFromFile :: error reading invalid chunk hashes from file")
			return hashes, errors.New("bad hashes")
			// FATAL
		}

		hashArray := (*[32]byte)(hash)
		hashes = append(hashes, *hashArray)
	}

	return hashes, nil
}

func RefreshTimeStamp(bkup *BackUpSession) {
	bkup.LastRequestTimeStamp.Mu.Lock()
	bkup.LastRequestTimeStamp.Stamp = time.Now()
	bkup.LastRequestTimeStamp.Mu.Unlock()
}

func IsCompleted(bkup *BackUpSession) bool {
	bkup.RequestTracker.Mu.RLock()
	defer bkup.RequestTracker.Mu.RUnlock()
	return bkup.RequestTracker.Completed == bkup.RequestTracker.ClientCount
}

func GetCompletedCount(bkup *BackUpSession) int {
	bkup.RequestTracker.Mu.RLock()
	defer bkup.RequestTracker.Mu.RUnlock()
	return bkup.RequestTracker.Completed
}

func GetClientCount(bkup *BackUpSession) int {
	bkup.RequestTracker.Mu.RLock()
	defer bkup.RequestTracker.Mu.RUnlock()
	return bkup.RequestTracker.ClientCount
}

func SetClientCount(bkup *BackUpSession, count int) {
	// funcName := "SetClientCount"
	bkup.RequestTracker.Mu.Lock()
	bkup.RequestTracker.ClientCount = count
	bkup.RequestTracker.Mu.Unlock()
}

func startedReqProcessing(bkup *BackUpSession) {
	bkup.RequestTracker.Mu.Lock()
	bkup.RequestTracker.Started++
	bkup.RequestTracker.Mu.Unlock()
}

func finishedReqProcessing(bkup *BackUpSession) {
	bkup.RequestTracker.Mu.Lock()
	bkup.RequestTracker.Completed++
	bkup.RequestTracker.Mu.Unlock()
}

func (bkupData *BackUpSession) IsExpired() bool {
	bkupData.LastRequestTimeStamp.Mu.RLock()
	defer bkupData.LastRequestTimeStamp.Mu.RUnlock()
	return time.Now().After(bkupData.LastRequestTimeStamp.Stamp.Add(1 * time.Hour)) // duration may change
}

//------------------------------------------------------------------------------------------------------------

func CreateBackUpSession() (string, error) {
	funcName := "CreateBackUpSession"

	core, err := session.CreateSession()

	if err != nil {
		logging.GlobalLogger.Error("Failed to create session",
			zap.String("handler", funcName),
			zap.Error(err))
	}

	sessionID := core.SessionID

	chunkFile, err := os.OpenFile(filepath.Join(core.Path, "chunks.dat"), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logging.GlobalLogger.Error("Failed to create file for session",
			zap.String("sessionID", sessionID),
			zap.String("handler", funcName),
			zap.String("file", "chunks.dat"),
			zap.Error(err))
		return "", err
	}

	fileMetaFile, err := os.OpenFile(filepath.Join(core.Path, "filemeta.dat"), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logging.GlobalLogger.Error("Failed to create file for session",
			zap.String("sessionID", sessionID),
			zap.String("handler", funcName),
			zap.String("file", "filemeta.dat"),
			zap.Error(err))
		return "", err
	}

	invalidChunksFile, err := os.OpenFile(filepath.Join(core.Path, "invalidchunks.dat"), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logging.GlobalLogger.Error("Failed to create file for session",
			zap.String("sessionID", sessionID),
			zap.String("handler", funcName),
			zap.String("file", "invalidchunks.dat"),
			zap.Error(err))
		return "", err
	}

	backUpSess := BackUpSession{}

	backUpSess.Core = core

	backUpSess.Chunks = &session.SessionFile{
		Mu:   sync.RWMutex{},
		File: chunkFile,
	}

	backUpSess.FileMeta = &session.SessionFile{
		Mu:   sync.RWMutex{},
		File: fileMetaFile,
	}

	backUpSess.InvalidChunks = &session.SessionFile{
		Mu:   sync.RWMutex{},
		File: invalidChunksFile,
	}

	backUpSess.LastRequestTimeStamp = &session.TimeStamp{
		Stamp: time.Now(),
		Mu:    sync.RWMutex{},
	}

	backUpSess.RequestTracker = &RequestCounts{
		Started:     0,
		Completed:   0,
		ClientCount: -1,
		Mu:          sync.RWMutex{},
	}

	//sess.Data = &backUpData

	backUpSess.AddToValidSessions()

	logging.GlobalLogger.Info("New backup session initialized and added to valid sessions",
		zap.String("sessionID", sessionID),
		zap.String("handler", funcName))

	return sessionID, nil
}

func (bkupsess *BackUpSession) AddToValidSessions() {
	session.ValidSessions.Mu.Lock()
	session.ValidSessions.SessionsMap[bkupsess.Core.SessionID] = bkupsess
	session.ValidSessions.Mu.Unlock()
}

func GetBackUpSession(sessionID string) (*BackUpSession, error) {
	sess, valid := session.ValidSessions.SessionsMap[sessionID]
	if !valid {
		return nil, errors.New("invalid session ID")
	}
	bkupsess, ok := sess.(*BackUpSession)
	if !ok {
		return nil, errors.New("invalid session ID")
	}

	return bkupsess, nil
}

func (bkupsess *BackUpSession) RollBack() { // should use session logger instead?
	//Do rollback of chunks and files
	funcName := "RollBack"
	fileHashes, err := ReadHashesFromFile(bkupsess.FileMeta)
	if err != nil {
		bkupsess.Core.Logger.Error("Failed to read hashes",
			zap.String("handler", funcName),
			zap.String("file", "filemeta.dat"),
			zap.Error(err))
		bkupsess.Core.SetError(err)
		return
	}

	err = RollBackFilesDBHandler(fileHashes)
	if err != nil {
		bkupsess.Core.Logger.Error("Database operations during roll back failed",
			zap.String("handler", funcName),
			zap.String("type", "filemeta"),
			zap.Error(err))
		bkupsess.Core.SetError(err)
		return
	}

	hashes, err := ReadHashesFromFile(bkupsess.Chunks)
	if err != nil {
		bkupsess.Core.Logger.Error("Failed to read hashes",
			zap.String("handler", funcName),
			zap.String("file", "chunks.dat"),
			zap.Error(err))
		bkupsess.Core.SetError(err)
		return
	}

	err = RollBackChunksDBHandler(hashes)
	if err != nil {
		bkupsess.Core.Logger.Error("Database operations during roll back failed",
			zap.String("handler", funcName),
			zap.String("type", "chunks"),
			zap.Error(err))
		bkupsess.Core.SetError(err)
		return
	}
	bkupsess.Core.SetError(nil)
	bkupsess.Core.Logger.Info("Session rolled back successfully",
		zap.String("handler", funcName))
}

func (bkupsess *BackUpSession) HandleExpiry() {
	funcName := "handleExpiry"
	core := bkupsess.Core
	core.Deactivate()

	err := bkupsess.Core.CheckError()
	if err != nil {
		bkupsess.Core.Logger.Info("Session closed with error",
			zap.String("handler", funcName),
			zap.Error(err))
		bkupsess.RollBack() /// do
	}

	bkupsess.CloseFiles()

	err = bkupsess.Core.CheckError()
	if err != nil {
		bkupsess.Core.Logger.Info("Session rollback failed with error",
			zap.String("handler", funcName),
			zap.Error(err))
		return
	}

	core.DeleteData()

	logging.GlobalLogger.Info("Session cleaned up",
		zap.String("sessionID", bkupsess.Core.SessionID),
		zap.String("handler", funcName))

	bkupsess.Core.Logger.Info("Session cleaned up",
		zap.String("handler", funcName))

}
