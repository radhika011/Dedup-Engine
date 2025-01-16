package session

import (
	"crypto/rand"
	"dedup-server/common"
	"dedup-server/logging"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Session interface {
	AddToValidSessions()
	IsExpired() bool
	HandleExpiry()
}

var ValidSessions struct {
	SessionsMap map[string]Session
	Mu          sync.RWMutex // need to lock only when inserting or removing a Session, maps are read only concurrency safe
}

func InitSessionMap() {
	ValidSessions.SessionsMap = make(map[string]Session)
	ValidSessions.Mu = sync.RWMutex{}

	logging.GlobalLogger.Info("SessionMap initialized")
}

type SessionCore struct {
	SessionID string
	Error     *SessionError
	Logger    *zap.Logger
	Path      string
}

type SessionError struct {
	Err error
	Mu  sync.RWMutex
}

type TimeStamp struct {
	Stamp time.Time
	Mu    sync.RWMutex
}

// type SessionData interface {
// 	CloseFiles() error
// 	IsExpired() bool
// 	HandleExpiry() error
// }

func (sess *SessionCore) CheckError() error {
	sess.Error.Mu.RLock()
	defer sess.Error.Mu.RUnlock()
	return sess.Error.Err
}

func (sess *SessionCore) SetError(err error) {
	funcName := "SetError"
	sess.Error.Mu.Lock()
	sess.Error.Err = err
	sess.Error.Mu.Unlock()

	sess.Logger.Warn("Error set in session",
		zap.String("handler", funcName),
		zap.Error(err))
}

func (sess *SessionCore) DeleteData() {
	funcName := "DeleteData"
	err := os.RemoveAll(filepath.Join(common.SessionsPath, sess.SessionID))

	if err != nil {
		logging.GlobalLogger.Error("Failed to delete session data",
			zap.String("handler", funcName),
			zap.String("sessionID", sess.SessionID),
			zap.Error(err))

		sess.Logger.Error("Failed to delete session data",
			zap.String("handler", funcName),
			zap.Error(err))

		return
	}

	sess.Logger.Info("Session data deleted",
		zap.String("handler", funcName))

}

func CreateSession() (*SessionCore, error) {
	funcName := "CreateSession"

	b := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return nil, err
	}

	sessionID := base64.URLEncoding.EncodeToString(b)

	err = os.Mkdir(filepath.Join(common.SessionsPath, sessionID), 0777)
	if err != nil {
		logging.GlobalLogger.Error("Failed to create directory for session",
			zap.String("sessionID", sessionID),
			zap.String("handler", funcName),
			zap.Error(err))
		return nil, err
	}

	logFile, err := os.Create(filepath.Join(common.SessionLogsPath, sessionID+".log"))
	if err != nil {
		logging.GlobalLogger.Error("Failed to create log file for session",
			zap.String("sessionID", sessionID),
			zap.String("handler", funcName),
			zap.Error(err))
		return nil, err
	}

	logger := logging.CreateLogger(logFile)

	session := SessionCore{
		SessionID: sessionID,
		Logger:    logger,
		Path:      filepath.Join(common.SessionsPath, sessionID),
	}

	session.Error = &SessionError{
		Err: nil,
		Mu:  sync.RWMutex{},
	}

	return &session, nil
}

func (sess *SessionCore) Deactivate() {
	funcName := "Deactivate"

	ValidSessions.Mu.RLock()
	_, valid := ValidSessions.SessionsMap[sess.SessionID]
	ValidSessions.Mu.RUnlock()

	if valid {

		ValidSessions.Mu.Lock()
		delete(ValidSessions.SessionsMap, sess.SessionID)
		ValidSessions.Mu.Unlock()

		sess.Logger.Info("Deactivated session",
			zap.String("handler", funcName))

		logging.GlobalLogger.Info("Removed session from valid sessions",
			zap.String("sessionID", sess.SessionID),
			zap.String("handler", funcName))

	}
}

func ManageSessionExpiry() { // go routine
	for {
		for sessName := range ValidSessions.SessionsMap {

			ValidSessions.Mu.RLock()
			sess := ValidSessions.SessionsMap[sessName]
			ValidSessions.Mu.RUnlock()

			isInvalid := sess.IsExpired()
			if isInvalid {
				go sess.HandleExpiry()
			}
		}
		time.Sleep(10 * time.Minute) // duration may change
	}
}
