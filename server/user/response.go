package user

import "dedup-server/types"

type Response struct {
	// 0 - success
	// 1 - wrong password during verification
	// 2 - account to be verified or updated doesn't exist
	// 3 - account attempting to get registered already exists
	Status   int
	UserData types.UserData
}
