package retrieve

import (
	"dedup-server/mongoutil"
	"dedup-server/types"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// const filesPath = "/home/shruti/Project/BTechProj-dedup/server"
const MaxFileWorkers = 10

func RetrieveHandler(sess *RetrieveSession, backupStruct types.BackUpDirStruct, backupSizeChan chan uint64) {
	funcName := "RetrieveHandler"

	fullDirStruct, err := DirStructDBHandler(backupStruct)

	backupSizeChan <- fullDirStruct.Size

	if err != nil {
		sess.Core.SetError(err)
		sess.Core.Logger.Info("Error during DirStruct retrieval",
			zap.Error(err),
			zap.String("handler", funcName))
		return
	}

	fileGet := make(chan types.Dirfile)

	spawnFileWorkers(sess, fileGet, sess.Files)

	for i, dir := range fullDirStruct.DirectoryArray {
		dirName := filepath.Base(dir.Name)
		dir.Name = fmt.Sprintf("%d-%s", i, dirName)

		processFiles("", dir, fileGet)
	}

	close(fileGet)
}

func processFiles(curpath string, dirfile types.Dirfile, fileGet chan types.Dirfile) {
	if !dirfile.Valid {
		return
	}
	curpath = filepath.Join(curpath, dirfile.Name)
	if dirfile.Type == "FILE" {
		dirfile.Name = filepath.ToSlash(curpath)
		fileGet <- dirfile
	} else {
		for _, child := range dirfile.Children {
			processFiles(curpath, child, fileGet)
		}
	}
}

func spawnFileWorkers(sess *RetrieveSession, fileGet chan types.Dirfile, fileSend chan FileStruct) {
	funcName := "spawnFileWorkers"

	for i := 0; i < MaxFileWorkers; i++ { // replace hardcoded value later
		go fileWorker(sess, fileGet, fileSend)
	}
	sess.Core.Logger.Info("Spawned file workers",
		zap.Int("number", MaxFileWorkers),
		zap.String("handler", funcName)) // hardcoded
}

func fileWorker(sess *RetrieveSession, fileGet chan types.Dirfile, fileSend chan FileStruct) {
	funcName := "fileWorker"

	for dirfile := range fileGet {

		fileHash := dirfile.Hash
		collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("FileMetadata")
		fileMetadata, err := mongoutil.GetFileMD(collection, fileHash)
		if err != nil {
			sess.Core.SetError(err)

			sess.Core.Logger.Error("Could not obtain fileMetadata",
				zap.Error(err),
				zap.String("handler", funcName))
			continue
		}
		fileName := hex.EncodeToString(fileHash[:])
		filePath := filepath.Join(sess.Core.Path, fileName)

		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_RDWR|os.O_EXCL, 0644)

		if errors.Is(err, os.ErrExist) {
			file.Close()
			fileStruct := FileStruct{
				ServerPath: filePath,
				DirPath:    dirfile.Name,
			}
			fileSend <- fileStruct
			continue
		} else if err != nil {
			sess.Core.SetError(err)
			sess.Core.Logger.Error("Could not reconstruct file",
				zap.Error(err),
				zap.String("handler", funcName))
			continue
		}

		// _, err = os.Stat(filePath)
		// if err == nil {
		// 	continue
		// } else if !errors.Is(err, os.ErrNotExist) {
		// 	sess.Core.Logger.Error("Error while checking if file exists",
		// 		zap.Error(err),
		// 		zap.String("handler", funcName))
		// }

		collection = mongoutil.Client.Database(mongoutil.DB_NAME).Collection("ChunkStore")
		results, err := mongoutil.FindManyChunks(collection, fileMetadata.ChunkHashesArray)
		if err != nil {
			sess.Core.SetError(err)
			sess.Core.Logger.Error("Could not retrieve chunks",
				zap.Error(err),
				zap.String("handler", funcName))
			continue
		}

		// tempfile, ferr := os.OpenFile(filePath+".txt", os.O_CREATE|os.O_APPEND, 0644)

		// if ferr != nil {
		// 	fmt.Println(ferr)
		// }

		// fmt.Println("len(fileMetadata.ChunkHashesArray): ", len(fileMetadata.ChunkHashesArray))
		for _, hashBin := range fileMetadata.ChunkHashesArray {
			hash := (*[32]byte)(hashBin.Data)
			_, err := file.Write(results[*hash])
			// if ferr == nil {
			// 	chunkHash := hex.EncodeToString(hash[:])
			// 	tempfile.WriteString(chunkHash)
			// 	tempfile.WriteString("\n")
			// }
			if err != nil {
				//HELPPPPP
				sess.Core.SetError(err)
				sess.Core.Logger.Error("Could not write chunk to file",
					zap.Error(err),
					zap.String("handler", funcName))
				continue
			}
		}

		// tempfile.Close()
		file.Close()
		fileStruct := FileStruct{
			ServerPath: filePath,
			DirPath:    dirfile.Name,
		}
		// fmt.Println("My dirpath", dirfile.Name)
		fileSend <- fileStruct
	}
	sess.CountMu.Lock()
	sess.FileWorkerCompletionCount++
	sess.CountMu.Unlock()
}

func DirStructDBHandler(bkUpDir types.BackUpDirStruct) (types.BackUpDirStruct, error) {
	// funcName := "DirStructDBHandler"
	collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("BackUpDirStruct")

	fullDirStruct, err := mongoutil.FindDirStruct(collection, bkUpDir)

	if err != nil {
		return fullDirStruct, err
	}

	return fullDirStruct, nil
}
