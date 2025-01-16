package backup

import (
	"client-background/common"
	"client-background/types"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

const MaxSendWorkers = 40

var logger *zap.Logger
var backUpMu = sync.Mutex{}

var unmodChunks struct {
	chunkHashes types.ChunkHashes
	mu          sync.RWMutex
}

var requestsTracker struct {
	completedCount int
	begunCount     int
	mu             sync.RWMutex
}

var channels = types.Channels{
	Err:      make(chan error, 1),
	SendData: make(chan types.SendPacket),
	Status:   make(chan int),
}

var status struct {
	code        int
	description string
	size        uint64
	totalSize   uint64
	mu          sync.RWMutex
}

func GetDataStatus() (uint64, uint64) {
	status.mu.RLock()
	defer status.mu.RUnlock()

	return status.totalSize, status.size
}

var statusStrings = []string{"Success", "Partial", "Failure"}

func SendWorker(ctx context.Context, sendPacket types.SendPacket) {

	requestsTracker.mu.Lock()
	requestsTracker.begunCount++
	requestsTracker.mu.Unlock()
	err := common.Send(ctx, sendPacket)
	if err != nil {
		channels.Err <- err
	} else {
		requestsTracker.mu.Lock()
		requestsTracker.completedCount++
		requestsTracker.mu.Unlock()
	}

}

func handleSends(ctx context.Context) {
	done := false
	for !done {
		select {
		case sendPacket := <-channels.SendData:
			go SendWorker(ctx, sendPacket)
		case <-ctx.Done():
			done = true
		}
	}
}

func initBackUp(timestamp time.Time) error {
	status.code = 0
	status.description = ""
	status.size = 0
	status.totalSize = 0
	status.mu = sync.RWMutex{}

	unmodChunks.chunkHashes = types.ChunkHashes{}
	unmodChunks.mu = sync.RWMutex{}

	requestsTracker.completedCount = 0
	requestsTracker.begunCount = 0
	requestsTracker.mu = sync.RWMutex{}

	channels = types.Channels{
		Err:      make(chan error, 1),
		SendData: make(chan types.SendPacket),
		Status:   make(chan int),
	}

	ChunkingStats.mu = sync.Mutex{}
	ChunkingStats.Num = 0
	ChunkingStats.Size = 0
	ChunkingStats.Duration = 0

	var err error
	bkupLogfileName := timestamp.Format(common.TIME_FORMAT) + ".log"
	logger, err = common.InitLogger(filepath.Join(common.GetBackUpLogsDir(), bkupLogfileName))
	return err
}

func writeToSysHistoryFile(backUpStart time.Time) {
	status.mu.RLock()
	description := status.description
	code := status.code
	size := status.size
	status.mu.RUnlock()

	if code != 2 {
		if size < 1024 {
			description = fmt.Sprintf("%d", size) + " B"
		} else if size < 1024*1024 {
			description = fmt.Sprintf("%d", size/1024) + "KB"
		} else if size < 1024*1024*1024 {
			description = fmt.Sprintf("%d", size/(1024*1024)) + "MB"
		} else {
			description = fmt.Sprintf("%d", size/(1024*1024*1024)) + "GB"
		}
	}
	entry := types.SysHistoryEntry{
		Timestamp:   backUpStart,
		Status:      statusStrings[code],
		Description: description,
		Type:        "Backup",
	}

	common.UpdateSysHistoryFile(entry)
}

func statusListener(ctx context.Context) {
	curr := 0
	done := false
	for !done {
		select {
		case x := <-channels.Status:
			if x > curr {
				status.mu.Lock()
				status.code = x
				status.mu.Unlock()
				curr = x
			}
		case <-ctx.Done():
			done = true
		default:
			continue
		}
	}
}

func setErrorStatus(description string) {
	channels.Status <- 2
	status.mu.Lock()
	status.description = description
	status.mu.Unlock()
}

func getDirSize(path string) (uint64, error) {
	var size uint64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			size += uint64(info.Size())
		}
		return err
	})
	return size, err
}
