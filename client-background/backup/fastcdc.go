package backup

import (
	"bufio"
	"io"
	"time"
)

func getNextChunk(reader *bufio.Reader) ([]byte, error) {

	startTime := time.Now()

	defer ChunkingSpeedMeasure(startTime)

	var MinSize = 16384     // 16 KB
	var MaxSize = 262144    // 256 KB
	var NormalSize = 131072 // 128 KB

	const MaskS uint64 = 0x0000d9ff03530000 // 19
	const MaskL uint64 = 0x0000d90f03530000 // 15
	const MaskS_ls uint64 = (MaskS << 1)
	const MaskL_ls uint64 = (MaskL << 1)

	var fp uint64 = 0
	var chunk = make([]byte, 0, 262144)
	size := 0

	for ; size <= MinSize; size++ {
		b, err := reader.ReadByte()
		if err == io.EOF {
			return chunk, nil
		} else if err != nil {
			return nil, err
		}
		chunk = append(chunk, b)
	}

	for size <= NormalSize-1 {

		b, err := reader.ReadByte()
		if err == io.EOF {
			return chunk, nil
		} else if err != nil {
			return nil, err
		}
		chunk = append(chunk, b)
		size++
		fp = (fp << 2) + Gear_ls[b]
		if (fp & MaskS_ls) == 0 {
			return chunk, nil
		}

		b2, err2 := reader.ReadByte()
		if err2 == io.EOF {
			return chunk, nil
		} else if err2 != nil {
			return nil, err2
		}
		chunk = append(chunk, b2)
		size++

		fp += Gear[b2]
		if (fp & MaskS) == 0 {
			return chunk, nil
		}
	}

	for size <= MaxSize-1 {
		b, err := reader.ReadByte()
		if err == io.EOF {
			return chunk, nil
		} else if err != nil {
			return nil, err
		}
		chunk = append(chunk, b)
		size++
		if size == MaxSize {
			return chunk, nil
		}

		fp = (fp << 2) + Gear_ls[b]
		if (fp & MaskL_ls) == 0 {
			return chunk, nil
		}
		b2, err2 := reader.ReadByte()
		if err2 == io.EOF {
			break
		}
		chunk = append(chunk, b2)
		size++
		fp += Gear[b2]
		if (fp & MaskL) == 0 {
			return chunk, nil
		}
	}

	return chunk, nil
}

func ChunkingSpeedMeasure(startTime time.Time) {
	elapsedTime := time.Since(startTime)
	ChunkingStats.mu.Lock()
	ChunkingStats.Duration += elapsedTime
	ChunkingStats.mu.Unlock()
}
