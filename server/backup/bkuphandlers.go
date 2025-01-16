package backup

import (
	"dedup-server/clienttypes"
	"dedup-server/types"

	"go.uber.org/zap"
)

func ChunkHandler(sess *BackUpSession, chunk clienttypes.Chunk) {
	startedReqProcessing(sess)

	funcName := "ChunkHandler"

	err := ChunkDBHandler(chunk)
	if err != nil {
		sess.Core.Logger.Error("Database operation failed",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName),
			zap.Error(err))
		sess.Core.SetError(err)
		return
	}

	err = sess.Chunks.Append(chunk.Hash[:])
	if err != nil {
		sess.Core.Logger.Error("File append operation failed",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName),
			zap.String("file", "chunks.dat"),
			zap.Binary("chunkHash", chunk.Hash[:]),
			zap.Error(err))
		sess.Core.SetError(err)
		return
	}

	// sess.Core.Logger.Info("Processed chunk successfully",
	// 	zap.String("sessionID", sess.Core.SessionID),
	// 	zap.String("handler", funcName))
	finishedReqProcessing(sess)
}

func FileMetaHandler(sess *BackUpSession, fileMD clienttypes.FileMetadata) {

	startedReqProcessing(sess)
	funcName := "FileMetaHandler"
	err := FileMetaDBHandler(fileMD)
	if err != nil {
		sess.Core.Logger.Error("Database operation failed",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName),
			zap.Error(err))
		sess.Core.SetError(err)
		return
	}

	// tempname := filepath.Join(common.SessionsPath, sess.Core.SessionID, hex.EncodeToString(fileMD.FileHash[:])) + ".txt"
	// tempfile, ferr := os.OpenFile(tempname, os.O_CREATE|os.O_APPEND, 0644)

	// if ferr != nil {
	// 	fmt.Println(ferr)
	// } else {
	// 	for _, hash := range fileMD.ChunkHashesArray {
	// 		chunkHash := hex.EncodeToString(hash[:])
	// 		tempfile.WriteString(chunkHash)
	// 		tempfile.WriteString("\n")
	// 	}
	// 	tempfile.Close()
	// }

	err = sess.FileMeta.Append(fileMD.FileHash[:])
	if err != nil {
		sess.Core.Logger.Error("File append operation failed",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName),
			zap.String("file", "filemeta.dat"),
			zap.Binary("fileHash", fileMD.FileHash[:]),
			zap.Error(err))
		sess.Core.SetError(err)
		return
	}
	finishedReqProcessing(sess)
	sess.Core.Logger.Debug("Processed filemetadata successfully",
		zap.String("handler", funcName))
}

func UnmodChunksHandler(sess *BackUpSession, chunkHashes clienttypes.ChunkHashes) {

	startedReqProcessing(sess)
	funcName := "UnmodChunksHandler"
	err := UnmodChunksDBHandler(chunkHashes)
	if err != nil {
		sess.Core.Logger.Error("Database operation failed",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName),
			zap.Error(err))
		sess.Core.SetError(err)
		return
	}

	for _, hash := range chunkHashes.Hashes {
		err := sess.Chunks.Append(hash[:])
		if err != nil {
			sess.Core.Logger.Error("File append operation failed",
				zap.String("sessionID", sess.Core.SessionID),
				zap.String("handler", funcName),
				zap.String("file", "chunks.dat"),
				zap.Binary("chunkHash", hash[:]),
				zap.Error(err))
			sess.Core.SetError(err)
			return
		}
	}
	finishedReqProcessing(sess)

}

func DirStructHandler(sess *BackUpSession, dirStruct types.BackUpDirStruct) {

	startedReqProcessing(sess)
	funcName := "DirStructHandler"
	err := DirStructDBHandler(dirStruct)
	if err != nil {
		sess.Core.Logger.Error("Database operation failed",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName),
			zap.Error(err))
		sess.Core.SetError(err)
		return
	}

	//SetSessionStatusCompleted(sess, true)

	invalidChunks, err := ReadHashesFromFile(sess.InvalidChunks)
	if err != nil {
		sess.Core.Logger.Error("Failed to read invalid chunk hashes from file",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName),
			zap.String("file", "invalidchunks.dat"),
			zap.Error(err))
		sess.Core.SetError(err) //-- this is a server side issue, client does not lose data
		return
	}

	err = InvalidChunksDBHandler(invalidChunks)
	if err != nil {
		sess.Core.Logger.Error("Database operations for roll back of invalid chunks failed",
			zap.String("sessionID", sess.Core.SessionID),
			zap.String("handler", funcName),
			zap.Error(err))
		sess.Core.SetError(err) //-- this is a server side issue, client does not lose data
		return
	}
	finishedReqProcessing(sess)
	sess.Core.Logger.Debug("Processed backupdirstruct successfully",
		zap.String("handler", funcName))
}

func InvalidChunksHandler(sess *BackUpSession, chunkHashes clienttypes.ChunkHashes) {

	startedReqProcessing(sess)
	funcName := "InvalidChunksHandler"
	for _, hash := range chunkHashes.Hashes {
		err := sess.InvalidChunks.Append(hash[:])
		if err != nil {
			sess.Core.Logger.Error("File append operation failed",
				zap.String("sessionID", sess.Core.SessionID),
				zap.String("handler", funcName),
				zap.Binary("chunkHash", hash[:]),
				zap.Error(err))
			sess.Core.SetError(err)
			return
		}
	}
	finishedReqProcessing(sess)

}
