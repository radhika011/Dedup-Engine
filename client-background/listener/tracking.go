package listener

import (
	"client-background/common"
	"sync"

	"go.uber.org/zap"
)

var active struct {
	mu            sync.Mutex
	processes     map[string]bool
	backupOngoing bool
}

func InitTracker() {
	active.mu = sync.Mutex{}
	active.processes = map[string]bool{}
	active.backupOngoing = false
	common.GlobalLogger.Info("Initialized active process tracker")
}

func AddToActiveProcesses(procType string, timestamp string) bool {
	added := false
	active.mu.Lock()
	defer active.mu.Unlock()

	if procType == "Delete" && timestamp == common.GetLastBackUpTime().Format(common.TIME_FORMAT) && active.backupOngoing {
		return false
	}

	_, exists := active.processes[timestamp]
	if !exists {
		active.processes[timestamp] = true
		added = true
		common.GlobalLogger.Info("Added process to active processes",
			zap.String("Timestamp", timestamp),
			zap.String("Process Type", procType))
		if procType == "Backup" {
			active.backupOngoing = true
		}
	}

	return added
}

func RemoveFromActiveProcesses(procType string, timestamp string) {
	active.mu.Lock()
	defer active.mu.Unlock()

	_, exists := active.processes[timestamp]
	if exists {
		delete(active.processes, timestamp)
		common.GlobalLogger.Info("Removed process from active processes",
			zap.String("Timestamp", timestamp),
			zap.String("Process Type", procType))

		if procType == "Backup" {
			active.backupOngoing = false
		}
	}
}
