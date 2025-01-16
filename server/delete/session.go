package delete

import (
	"dedup-server/logging"
	"dedup-server/session"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

type DeleteSession struct {
	Core                      *session.SessionCore
	Chunks                    *session.SessionFile
	FileMeta                  *session.SessionFile
	CreationTimeStamp         *session.TimeStamp
	FileWorkerCompletionCount int
	CountMu                   sync.RWMutex
	IsCompleted               bool
}

func CreateDeleteSession() (string, error) {
	funcName := "CreateDeleteSession"

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

	delSess := DeleteSession{}

	delSess.Core = core

	delSess.Chunks = &session.SessionFile{
		Mu:   sync.RWMutex{},
		File: chunkFile,
	}

	delSess.FileMeta = &session.SessionFile{
		Mu:   sync.RWMutex{},
		File: fileMetaFile,
	}

	delSess.CreationTimeStamp = &session.TimeStamp{
		Stamp: time.Now(),
		Mu:    sync.RWMutex{},
	}

	delSess.FileWorkerCompletionCount = 0
	delSess.CountMu = sync.RWMutex{}

	delSess.IsCompleted = false

	delSess.AddToValidSessions()

	logging.GlobalLogger.Info("New delete session initialized and added to valid sessions",
		zap.String("sessionID", sessionID),
		zap.String("handler", funcName))

	return sessionID, nil
}

func (delSess *DeleteSession) AddToValidSessions() {
	session.ValidSessions.Mu.Lock()
	session.ValidSessions.SessionsMap[delSess.Core.SessionID] = delSess
	session.ValidSessions.Mu.Unlock()
}

func GetDeleteSession(sessionID string) (*DeleteSession, error) {
	sess, valid := session.ValidSessions.SessionsMap[sessionID]
	if !valid {
		return nil, errors.New("invalid session ID")
	}
	delSess, ok := sess.(*DeleteSession)
	if !ok {
		return nil, errors.New("invalid session ID")
	}

	return delSess, nil
}

func (delSess *DeleteSession) CloseFiles() {
	funcName := "CloseFiles"
	err := delSess.Chunks.Close()
	if err != nil {
		delSess.Core.Logger.Error("Failed to close file",
			zap.String("handler", funcName),
			zap.String("file", "chunks.dat"),
			zap.Error(err))
	}
	err = delSess.FileMeta.Close()
	if err != nil {
		delSess.Core.Logger.Error("Failed to close file",
			zap.String("handler", funcName),
			zap.String("file", "filemeta.dat"),
			zap.Error(err))
	}

	delSess.Core.Logger.Info("Session files closed",
		zap.String("handler", funcName))

}

func (delSess *DeleteSession) IsExpired() bool {
	delSess.CountMu.RLock()
	defer delSess.CountMu.RUnlock()
	return delSess.FileWorkerCompletionCount == MaxFileWorkers
}

//TODO

func (delSess *DeleteSession) HandleExpiry() {
	funcName := "handleExpiry"
	core := delSess.Core
	core.Deactivate()

	err := delSess.Core.CheckError()
	if err != nil {
		delSess.Core.Logger.Info("Session closed with error",
			zap.String("handler", funcName),
			zap.Error(err))
	}

	delSess.CloseFiles()

	core.DeleteData()

	logging.GlobalLogger.Info("Session cleaned up",
		zap.String("sessionID", delSess.Core.SessionID),
		zap.String("handler", funcName))

	delSess.Core.Logger.Info("Session cleaned up",
		zap.String("handler", funcName))

}
