package delete

import (
	"dedup-server/mongoutil"
	"dedup-server/types"

	"go.uber.org/zap"
)

const MaxFileWorkers = 10

func deleteHandler(sess *DeleteSession, backupStruct types.BackUpDirStruct) {
	funcName := "deleteHandler"

	fullDirStruct, err := dirStructDBHandler(backupStruct)
	if err != nil {
		sess.Core.SetError(err)
		sess.Core.Logger.Info("Error during DirStruct retrieval",
			zap.Error(err),
			zap.String("handler", funcName))
		return
	}

	sess.IsCompleted = true
	sess.Core.Logger.Info("DirStruct deleted",
		zap.String("handler", funcName))

	fileSend := make(chan types.Dirfile)
	spawnFileWorkers(sess, fileSend)

	for _, dir := range fullDirStruct.DirectoryArray {
		processFiles(dir, fileSend)
	}

	close(fileSend)
	// sess.Logger.Info("Processing Done",
	// 	zap.String("handler", funcName))
}

func dirStructDBHandler(bkUpDir types.BackUpDirStruct) (types.BackUpDirStruct, error) {
	// funcName := "DirStructDBHandler"
	collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("BackUpDirStruct")

	fullDirStruct, err := mongoutil.DeleteDirStruct(collection, bkUpDir)

	if err != nil {
		// fmt.Println("BkUpDirDBHandler :: error while inserting BackUpDirStruct: ", err)
		return fullDirStruct, err
	}

	return fullDirStruct, nil
}

func processFiles(dirfile types.Dirfile, fileSend chan types.Dirfile) {
	if !dirfile.Valid {
		return
	}
	if dirfile.Type == "FILE" {
		fileSend <- dirfile
	} else {
		for _, child := range dirfile.Children {
			processFiles(child, fileSend)
		}
	}
}

func spawnFileWorkers(sess *DeleteSession, fileSend chan types.Dirfile) {
	funcName := "spawnFileWorkers"

	for i := 0; i < MaxFileWorkers; i++ { // replace hardcoded value later
		go fileWorker(sess, fileSend)
	}
	sess.Core.Logger.Info("Spawned file workers",
		zap.Int("number", MaxFileWorkers),
		zap.String("handler", funcName)) // hardcoded
}

// move to dbhahndlers?
func fileWorker(sess *DeleteSession, fileSend chan types.Dirfile) {
	funcName := "fileWorker"

	for dirfile := range fileSend {
		fileHash := dirfile.Hash
		collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("FileMetadata")
		fileMetadata, err := mongoutil.GetFileMD(collection, fileHash)
		if err != nil {
			sess.FileMeta.Append(fileHash[:])
			sess.Core.SetError(err)

			sess.Core.Logger.Error("Could not obtain fileMetadata",
				zap.Error(err),
				zap.String("handler", funcName))
			continue
		}
		err = mongoutil.UpdFileMDRefCount(collection, [][32]byte{fileHash}, -1)
		if err != nil {
			sess.FileMeta.Append(fileHash[:])
			sess.Core.SetError(err)

			sess.Core.Logger.Error("Could not delete fileMetadata",
				zap.Error(err),
				zap.String("handler", funcName))
		}

		collection = mongoutil.Client.Database(mongoutil.DB_NAME).Collection("ChunkStore")
		err = mongoutil.UpdChunkRefCount(collection, fileMetadata.ChunkHashesArray, -1) // make this work
		if err != nil {
			// marshal chunkhash and write to session file
			sess.Core.SetError(err)
			sess.Core.Logger.Error("Could not delete chunk",
				zap.Error(err),
				zap.String("handler", funcName))
		}
	}
	sess.CountMu.Lock()
	sess.FileWorkerCompletionCount++
	sess.CountMu.Unlock()
}
