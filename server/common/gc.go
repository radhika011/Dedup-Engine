package common

import (
	"dedup-server/mongoutil"
	"time"

	"go.uber.org/zap"
)

func GarbageCollector(logger *zap.Logger) {
	chunkCollection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("ChunkStore")
	fileCollection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("FileMetadata")
	for {
		count := 0
		ids, err := mongoutil.FindDeadEntries(chunkCollection)
		if err != nil {
			logger.Error("Failed to find dead chunk entries",
				zap.String("Operation", "Garbage Collection"),
				zap.Error(err))
		} else {
			err = mongoutil.DeleteDeadEntries(chunkCollection, ids)
			if err != nil {
				logger.Error("Failed to delete dead chunk entries",
					zap.String("Operation", "Garbage Collection"),
					zap.Error(err))
			}
		}
		count = int(len(ids))
		ids, err = mongoutil.FindDeadEntries(fileCollection)
		if err != nil {
			logger.Error("Failed to find dead file metadata entries",
				zap.String("Operation", "Garbage Collection"),
				zap.Error(err))
		} else {
			err = mongoutil.DeleteDeadEntries(fileCollection, ids)
			if err != nil {
				logger.Error("Failed to delete dead file metadata entries",
					zap.String("Operation", "Garbage Collection"),
					zap.Error(err))
			}
		}

		count += int(len(ids))
		logger.Info("Running routine garbage collection", zap.Int("Total documents removed", count))

		time.Sleep(10 * time.Minute)
	}
}
