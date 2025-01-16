package types

import (
	"time"
)

type Chunk struct {
	Hash Hash
	Data []byte
}

type FileMetadata struct {
	FileHash         Hash
	ChunkHashesArray []Hash
}

type BackUpDirStruct struct {
	Username       string
	TimeStamp      time.Time
	ClientUtilID   int
	DirectoryArray []Dirfile
	Size           uint64
}

type Dirfile struct {
	Name     string
	Type     string
	Hash     Hash
	Children []Dirfile
	Valid    bool
}

type ChunkHashes struct {
	Hashes []Hash
}

type Channels struct {
	Err      chan error
	SendData chan SendPacket
	Status   chan int
}

type Status struct {
	Code         int
	StatusString string
	Count        int
}

type SendPacket struct {
	JsonBody []byte
	Endpoint string
}

type SessionDetails struct {
	SessionID string
	Type      string
}

type InterfaceRequest struct {
	Type       string
	Parameters []byte
}

type SysHistoryEntry struct {
	Type        string
	Timestamp   time.Time
	Status      string
	Description string
}

type SessionKey string

type Hash [32]byte

type UserData struct {
	FirstName   string
	LastName    string
	PhoneNumber string
	EmailID     string
	Password    Hash
}

// type LoginCredentials struct {
// 	EmailID  string
// 	Password Hash
// }

type InterfaceResponse struct {
	Code       int // success 0 , failure -1 , processing 1
	Parameters []byte
}

type ResponseParam struct {
	ProcessedData uint64
	TotalData     uint64
}

type UserResponseParams struct {
	Description string
	Code        int
	UserInfo    UserData
}

type CurrentUser struct {
	UserName string `json:"UserName"`
}

type Persist struct {
	LastBackUpTime time.Time
	CacheIsValid   bool
}
