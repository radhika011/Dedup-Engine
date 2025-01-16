package retrieve

import (
	"client-background/common"
	"client-background/types"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// return 0 when completed -> 250 from server, 1 when successfully file written, -1 when server still processing -> server sends 102
func GetFile(ctx context.Context, rootPath string) (int, error) {
	funcName := "GetFile"
	sessionDetails := ctx.Value(common.KEY).(types.SessionDetails)

	requestURL := fmt.Sprintf("%s://%s:%s/%s/%s", common.SERVER_TYPE, common.SERVER_HOST, common.SERVER_PORT, sessionDetails.Type, "file")
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)

	if err != nil { // log failure
		err = errors.New(funcName + " :: " + "Error occurred while creating HTTP request: " + err.Error())
		return -1, err
	}

	sessionIDCookie := &http.Cookie{
		Name:  "sessionID",
		Value: sessionDetails.SessionID,
	}
	req.AddCookie(sessionIDCookie)

	timeoutIntervals := []int64{10, 20, 20, 40, 60}

	for i := 0; i < 5; i++ {
		res, err := common.BoundedRequest(ctx, req)
		if err != nil { // log failure

			d := time.Duration(timeoutIntervals[i]) * time.Second
			time.Sleep(d)
			continue
		}

		defer res.Body.Close()

		if res.StatusCode == http.StatusInternalServerError {
			err = errors.New(funcName + " :: " + "Internal server error occurred")
			return -1, err
		}
		if res.StatusCode == http.StatusOK {

			dirPath := ""
			for _, cookie := range res.Cookies() {
				if cookie.Name == "dirPath" {
					dirPath = cookie.Value
				}
			}

			dirPath = filepath.FromSlash(dirPath)
			path := filepath.Join(rootPath, dirPath)
			parentPath := filepath.Dir(path)

			err = os.MkdirAll(parentPath, 0777)
			if err != nil {
				err = errors.New(funcName + " :: " + "Error occurred while creating restore directories: " + err.Error())
				return -1, err
			}

			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				err = errors.New(funcName + " :: " + "Error occurred while opening file: " + err.Error())
				return -1, err
			}

			// defer file.Close()

			_, err = io.Copy(file, res.Body)
			if err != nil {
				err = errors.New(funcName + " :: " + "Error occurred while writing response body to file: " + err.Error())

				return -1, err
			}
			file.Close()

			info, err := os.Lstat(path)
			if err == nil && info.Mode().IsRegular() {
				dataStatus.mu.Lock()
				dataStatus.size += uint64(info.Size())
				dataStatus.mu.Unlock()
			}

			return 1, nil

		}

		if res.StatusCode == http.StatusAccepted { // make into processing code
			return -1, nil
		}

		if res.StatusCode == common.HTTP_STATUS_RETRIEVAL_COMPLETE {
			return 0, nil
		}

		return -1, errors.New(funcName + " :: " + "Unknown error occurred")

	}

	return -1, errors.New(funcName + " :: " + "Failed to send HTTP request" + err.Error())

}
