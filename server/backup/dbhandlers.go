package backup

import (
	"dedup-server/clienttypes"
	"dedup-server/mongoutil"
	"dedup-server/types"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// backupdirstruct
// status
// filemetadata
// chunk
// unmodifiedchunks

func ChunkDBHandler(chunkData clienttypes.Chunk) error {
	// funcName := "ChunkDBHandler"
	collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("ChunkStore")

	chunkExists, err := mongoutil.IfChunkExists(collection, chunkData.Hash[:])

	if err != nil {
		fmt.Println("ChunkDBHandler :: error while checking if new chunk: ", err)
		return err
	}

	if chunkExists {
		hashBinary := primitive.Binary{Data: chunkData.Hash[:]}
		err := mongoutil.UpdChunkRefCount(collection, []primitive.Binary{hashBinary}, 1)
		if err != nil {
			fmt.Println("ChunkDBHandler :: error while updating chunk reference count: ", err)
			return err
		}

	} else {
		chunk := types.Chunk{
			Hash:     primitive.Binary{Data: chunkData.Hash[:]},
			Data:     primitive.Binary{Data: chunkData.Data},
			RefCount: 1,
		}

		alreadyExists, err := mongoutil.InsertChunk(collection, chunk)

		if err != nil {
			fmt.Println("ChunkDBHandler :: error while inserting chunk: ", err)
			return err
		}

		if alreadyExists {
			// hashBinary := primitive.Binary{Data: chunkData.Hash[:]}
			err := mongoutil.UpdChunkRefCount(collection, []primitive.Binary{chunk.Hash}, 1)
			if err != nil {
				fmt.Println("ChunkDBHandler :: error while updating chunk reference count: ", err)
				return err
			}
		}

	}

	return nil
}

func UnmodChunksDBHandler(unmodChunks clienttypes.ChunkHashes) error {
	// funcName := "UnmodChunksDBHandler"
	collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("ChunkStore")
	var hashesBinary []primitive.Binary
	for _, hash := range unmodChunks.Hashes {
		chunkHash := []byte{}
		chunkHash = append(chunkHash, hash[:]...)

		hashesBinary = append(hashesBinary, primitive.Binary{Data: chunkHash})
	}
	err := mongoutil.UpdChunkRefCount(collection, hashesBinary, 1)

	if err != nil {
		// fmt.Println("UnmodChunksDBHandler :: error while updating chunk reference counts: ", err)
		return err
	}

	return nil
}

func FileMetaDBHandler(fileData clienttypes.FileMetadata) error {
	// funcName := "FileMetaDBHandler"
	collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("FileMetadata")

	fileExists, err := mongoutil.IfFileMDExists(collection, fileData.FileHash)
	if err != nil {
		// fmt.Println("FileDBHandler :: error while checking if new file: ", err)
		return err
	}

	if fileExists {
		err := mongoutil.UpdFileMDRefCount(collection, [][32]byte{fileData.FileHash}, 1)
		if err != nil {
			// fmt.Println("FileDBHandler :: error while updating file metadata reference count: ", err)
			return err
		}

	} else {

		arrLen := len(fileData.ChunkHashesArray)
		chunkHashes := make([]primitive.Binary, arrLen)

		// tempname := filepath.Join(common.DataPath, hex.EncodeToString(fileData.FileHash[:])) + ".txt"
		// tempfile, ferr := os.OpenFile(tempname, os.O_CREATE|os.O_APPEND, 0644)

		// if ferr != nil {
		// 	fmt.Println(ferr)
		// }

		for i, hash := range fileData.ChunkHashesArray {
			chunkHash := []byte{}
			chunkHash = append(chunkHash, hash[:]...)

			chunkHashes[i] = primitive.Binary{Data: chunkHash}
			// if ferr == nil {
			// 	chunkHash1 := hex.EncodeToString(hash[:])
			// 	tempfile.WriteString(chunkHash1)
			// 	tempfile.WriteString("\n")
			// 	chunkHash2 := hex.EncodeToString(chunkHashes[i].Data)
			// 	tempfile.WriteString(chunkHash2)
			// 	tempfile.WriteString("\n")
			// }
		}

		// if ferr == nil {
		// 	tempfile.WriteString("\n\n\n")

		// 	for i, h := range chunkHashes {
		// 		tempfile.WriteString(fmt.Sprint(i) + "  " + hex.EncodeToString(h.Data))
		// 		tempfile.WriteString("\n")
		// 	}
		// 	tempfile.Close()
		// }

		fileMD := types.FileMetadata{
			FileHash:         primitive.Binary{Data: fileData.FileHash[:]},
			ChunkHashesArray: chunkHashes,
			RefCount:         1,
		}

		alreadyExists, err := mongoutil.InsertFileMD(collection, fileMD)
		if err != nil {
			// fmt.Println("FileDBHandler :: error while inserting file metadata: ", err)
			return err
		}

		if alreadyExists {
			err := mongoutil.UpdChunkRefCount(collection, []primitive.Binary{fileMD.FileHash}, 1)
			if err != nil {
				fmt.Println("ChunkDBHandler :: error while updating chunk reference count: ", err)
				return err
			}
		}
	}

	return nil
}

func DirStructDBHandler(bkUpDir types.BackUpDirStruct) error {
	// funcName := "DirStructDBHandler"
	collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("BackUpDirStruct")

	err := mongoutil.InsertDirStruct(collection, bkUpDir)

	if err != nil {
		// fmt.Println("BkUpDirDBHandler :: error while inserting BackUpDirStruct: ", err)
		return err
	}

	return nil
}

func InvalidChunksDBHandler(chunkHashes [][32]byte) error {
	// funcName := "InvalidChunksDBHandler"
	collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("ChunkStore")
	var hashesBinary []primitive.Binary
	for _, hash := range chunkHashes {
		chunkHash := []byte{}
		chunkHash = append(chunkHash, hash[:]...)

		hashesBinary = append(hashesBinary, primitive.Binary{Data: chunkHash})
	}
	err := mongoutil.UpdChunkRefCount(collection, hashesBinary, -1)

	if err != nil {
		// fmt.Println("InvalidChunksDBHandler :: error while updating chunk reference counts: ", err)
		return err
	}

	return nil
}

func RollBackChunksDBHandler(chunkHashes [][32]byte) error {
	// funcName := "RollBackChunksDBHandler"
	collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("ChunkStore")
	var hashesBinary []primitive.Binary
	for _, hash := range chunkHashes {
		chunkHash := []byte{}
		chunkHash = append(chunkHash, hash[:]...)
		hashesBinary = append(hashesBinary, primitive.Binary{Data: chunkHash})
	}
	err := mongoutil.UpdChunkRefCount(collection, hashesBinary, -1)

	if err != nil {
		// fmt.Println("RollBackChunksDBHandler :: error while updating chunk reference counts: ", err)
		return err
	}

	return nil
}

func RollBackFilesDBHandler(fileHashes [][32]byte) error {
	// funcName := "RollBackFilesDBHandler"
	collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("FileMeta")
	err := mongoutil.UpdFileMDRefCount(collection, fileHashes, -1)

	if err != nil {
		// fmt.Println("RollBackFilesDBHandler :: error while updating chunk reference counts: ", err)
		return err
	}

	return nil
}
