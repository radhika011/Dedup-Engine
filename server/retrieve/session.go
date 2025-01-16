package retrieve

import (
	"dedup-server/logging"
	"dedup-server/session"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
)

type RetrieveSession struct {
	Core                      *session.SessionCore
	CreationTimeStamp         *session.TimeStamp
	Files                     chan FileStruct
	FileWorkerCompletionCount int
	CountMu                   sync.RWMutex
}

type FileStruct struct {
	ServerPath string
	DirPath    string
}

func CreateRetrieveSession() (string, error) {
	funcName := "CreateRetrieveSession"

	core, err := session.CreateSession()

	if err != nil {
		logging.GlobalLogger.Error("Failed to create session",
			zap.String("handler", funcName),
			zap.Error(err))
	}

	sess := RetrieveSession{}

	sess.Core = core
	sess.CreationTimeStamp = &session.TimeStamp{
		Stamp: time.Now(),
		Mu:    sync.RWMutex{},
	}

	sess.FileWorkerCompletionCount = 0
	sess.CountMu = sync.RWMutex{}

	//fileSend := make(chan FileStruct)
	sess.Files = make(chan FileStruct, 1)

	sess.AddToValidSessions()

	logging.GlobalLogger.Info("New retrieve session initialized and added to valid sessions",
		zap.String("sessionID", sess.Core.SessionID),
		zap.String("handler", funcName))

	return sess.Core.SessionID, nil
}

func (retSess *RetrieveSession) IsExpired() bool {
	retSess.CountMu.RLock()
	defer retSess.CountMu.RUnlock()
	return retSess.FileWorkerCompletionCount == MaxFileWorkers
}

func (retSess *RetrieveSession) HandleExpiry() {
	funcName := "handleExpiry"
	core := retSess.Core
	core.Deactivate()

	err := retSess.Core.CheckError()
	if err != nil {
		retSess.Core.Logger.Info("Session closed with error",
			zap.String("handler", funcName),
			zap.Error(err))
	}

	core.DeleteData()

	logging.GlobalLogger.Info("Session cleaned up",
		zap.String("sessionID", retSess.Core.SessionID),
		zap.String("handler", funcName))

	retSess.Core.Logger.Info("Session cleaned up",
		zap.String("handler", funcName))

}

func (retSess *RetrieveSession) AddToValidSessions() {
	session.ValidSessions.Mu.Lock()
	session.ValidSessions.SessionsMap[retSess.Core.SessionID] = retSess
	session.ValidSessions.Mu.Unlock()
}

func GetRetrieveSession(sessionID string) (*RetrieveSession, error) {
	sess, valid := session.ValidSessions.SessionsMap[sessionID]
	if !valid {
		return nil, errors.New("invalid session ID")
	}
	retSess, ok := sess.(*RetrieveSession)
	if !ok {
		return nil, errors.New("invalid session ID")
	}

	return retSess, nil
}
