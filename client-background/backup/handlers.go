package backup

import (
	"bufio"
	"client-background/cache"
	"client-background/common"
	"client-background/types"
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

func directoryHandler(ctx context.Context, dirPath string, res chan map[string]types.Dirfile) {
	funcName := "directoryHandler"

	dirStruct := types.Dirfile{}
	m := make(map[string]types.Dirfile)

	dirInfo, err := os.Lstat(dirPath)
	if err != nil {
		logger.Error("Failed to read directory/file stats",
			zap.Error(err),
			zap.String("directory", dirPath),
			zap.String("handler", funcName))

		dirStruct.Valid = false
		m[dirPath] = dirStruct
		res <- m

		channels.Status <- 1

		return
	}

	if dirInfo.Mode().IsDir() {
		_, dirName := filepath.Split(dirPath)
		dirStruct.Name = dirName
		dirStruct.Type = "DIR"
		dirStruct.Hash = types.Hash{}
		logger.Debug("Processing directory",
			zap.String("name", dirPath))
		directory, err := os.ReadDir(dirPath)
		if err != nil {
			logger.Error("Failed to read directory entries",
				zap.Error(err),
				zap.String("directory", dirPath),
				zap.String("handler", funcName))

			dirStruct.Valid = false
			m[dirPath] = dirStruct
			res <- m

			channels.Status <- 1

			return

		}

		subDirStructs := []types.Dirfile{}

		results := make(chan map[string]types.Dirfile, len(directory))
		for _, entry := range directory {

			entryPath := filepath.Join(dirPath, entry.Name())
			go directoryHandler(ctx, entryPath, results)
		}

		for i := 0; i < len(directory); {
			select {
			case <-ctx.Done():
				return
			case subDirMap := <-results:
				for k := range subDirMap {
					subDirStructs = append(subDirStructs, subDirMap[k])
				}
				i++

			}
		}

		dirStruct.Children = subDirStructs

		dirStruct.Valid = true
		m[dirPath] = dirStruct
		res <- m
		return

	} else if dirInfo.Mode().IsRegular() {

		if ctx.Err() == nil {

			fileStruct := fileHandler(ctx, dirPath)

			if fileStruct.Valid {
				status.mu.Lock()
				status.size += uint64(dirInfo.Size())
				status.mu.Unlock()
			}

			m[dirPath] = fileStruct
			res <- m
		}
		return
	}

}

func fileHandler(ctx context.Context, fPath string) types.Dirfile {
	funcName := "fileHandler"
	_, fileName := filepath.Split(fPath)
	fileStruct := types.Dirfile{}

	file, err := os.Open(fPath)

	if err != nil {
		logger.Error("Failed to read file",
			zap.Error(err),
			zap.String("file", fPath),
			zap.String("handler", funcName))

		fileStruct.Valid = false
		channels.Status <- 1

		return fileStruct
	}

	if ctx.Err() != nil {
		return fileStruct
	}

	fileReader := bufio.NewReader(file)
	hashes, fileHash, err := getChunks(ctx, fileReader)

	fileHashArray := (*types.Hash)(fileHash)

	if err != nil { // file failure - reading/ sending?
		logger.Error("Failed to process file chunks, invalidating file",
			zap.Error(err),
			zap.String("file", fPath),
			zap.String("handler", funcName))

		fileStruct.Valid = false
		channels.Status <- 1

		cache.Remove(hashes)
		invalidChunksJSON, err := common.ToJSON(hashes)
		if err != nil {
			logger.Error("Could not marshal invalid chunk hashes into JSON",
				zap.Error(err),
				zap.String("file", fPath),
				zap.String("handler", funcName))

			fileStruct.Valid = false
			channels.Status <- 1

			return fileStruct
		}
		sendPacket := types.SendPacket{
			JsonBody: invalidChunksJSON,
			Endpoint: "invalidchunks",
		}
		channels.SendData <- sendPacket
		logger.Info("Sending invalid chunk hashes",
			zap.String("file", fPath),
			zap.String("handler", funcName))
		return fileStruct
	}

	fileMeta := types.FileMetadata{
		FileHash:         *fileHashArray,
		ChunkHashesArray: hashes,
	}

	fileMetaJSON, err := common.ToJSON(fileMeta)
	if err != nil {
		logger.Error("Could not marshal file metadata into JSON",
			zap.Error(err),
			zap.String("file", fPath),
			zap.String("handler", funcName))

		fileStruct.Valid = false
		channels.Status <- 1

		return fileStruct
	}
	sendPacket := types.SendPacket{
		JsonBody: fileMetaJSON,
		Endpoint: "filemetadata",
	}
	channels.SendData <- sendPacket

	fileStruct.Hash = *fileHashArray
	fileStruct.Type = "FILE"
	fileStruct.Children = nil
	fileStruct.Name = fileName
	fileStruct.Valid = true

	return fileStruct
}
