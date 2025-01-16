package backup

import (
	"bufio"
	"client-background/cache"
	"client-background/common"
	"client-background/types"
	"context"
	"crypto/sha256"
	"io"
	"sync"
	"time"

	"go.uber.org/zap"
)

var ChunkingStats struct {
	mu       sync.Mutex
	Size     uint64
	Num      uint64
	Duration time.Duration
}

func getChunks(ctx context.Context, reader *bufio.Reader) ([]types.Hash, []byte, error) { // return non-fail chunks to file-handler always
	var chunkHashes []types.Hash
	hasher := sha256.New()

	for !isEmpty(reader) {
		chunk, err := getNextChunk(reader)
		if err != nil {

			return chunkHashes, nil, err
		}

		size := len(chunk)

		ChunkingStats.mu.Lock()
		ChunkingStats.Size = ChunkingStats.Size + uint64(size)
		ChunkingStats.Num++
		ChunkingStats.mu.Unlock()

		hash, err := chunkHandler(ctx, chunk)
		if err != nil {
			return chunkHashes, nil, err
		}
		chunkHashes = append(chunkHashes, hash)
		hasher.Write(chunk)
	}
	filehash := hasher.Sum(nil)
	return chunkHashes, filehash, nil
}

func isEmpty(reader *bufio.Reader) bool {
	_, err := reader.Peek(1)
	return err == io.EOF
}

func chunkHandler(ctx context.Context, chunk []byte) (types.Hash, error) { // only sends have problems

	hash := sha256.Sum256(chunk)
	if ctx.Err() != nil {
		return hash, ctx.Err()
	}

	if cache.Check(hash) {
		UnmodChunkHandler(ctx, hash, false)
	} else {

		chunkStruct := types.Chunk{
			Data: chunk,
			Hash: hash,
		}

		chunkStructJSON, err := common.ToJSON(chunkStruct)
		if err != nil {

			return hash, err
		}

		sendPacket := types.SendPacket{
			JsonBody: chunkStructJSON,
			Endpoint: "chunk",
		}
		channels.SendData <- sendPacket
	}
	return hash, nil
}

func UnmodChunkHandler(ctx context.Context, hash types.Hash, flush bool) {
	funcName := "UnmodChunkHandler"
	unmodChunks.mu.Lock()
	defer unmodChunks.mu.Unlock()

	if (len(unmodChunks.chunkHashes.Hashes) >= 100) || flush {
		UnmodChunksJSON, err := common.ToJSON(unmodChunks.chunkHashes)
		if err != nil {
			logger.Error("Could not marshal unmodified chunk hashes into JSON",
				zap.Error(err),
				zap.String("handler", funcName))

			channels.Err <- err
			return
		}
		sendPacket := types.SendPacket{
			JsonBody: UnmodChunksJSON,
			Endpoint: "unmodifiedchunks",
		}
		channels.SendData <- sendPacket
		logger.Info("Sent unmodified chunk hashes to server",
			zap.String("handler", funcName))
		unmodChunks.chunkHashes = types.ChunkHashes{}

		if flush {
			return
		}
	}
	unmodChunks.chunkHashes.Hashes = append(unmodChunks.chunkHashes.Hashes, hash)
}
