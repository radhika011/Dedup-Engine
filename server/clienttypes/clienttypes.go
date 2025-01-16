package clienttypes

type Chunk struct {
	Hash [32]byte
	Data []byte
}

type FileMetadata struct {
	FileHash         [32]byte
	ChunkHashesArray [][32]byte
}

type ChunkHashes struct {
	Hashes [][32]byte
}

type Status struct {
	Code         int
	StatusString string
	Count        int
}
