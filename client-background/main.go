package main

import (
	"client-background/common"
	"client-background/listener"
	"client-background/schedule"
	"fmt"
)

//TODO

// custom errors
// env variables

func main() {

	err := common.LoadEnvFile()
	if err != nil {
		fmt.Println(err)
		// common.GlobalLogger.Error("makeDataDirs failed", zap.Error(err))
	}

	err = common.MakeGlobalDirs()
	if err != nil {
		fmt.Println(err)
		// common.GlobalLogger.Error("makeDataDirs failed", zap.Error(err))
	}

	err = common.InitGlobalLogger()
	if err != nil {
		fmt.Println(err)
		// common.GlobalLogger.Error("makeDataDirs failed", zap.Error(err))
	}

	common.InitStateVars()
	common.LoadLoginState()
	common.LoadPersistedState()

	listener.InitTracker()

	tempChan := make(chan struct{})
	go schedule.Scheduler(tempChan)
	listener.Listen(tempChan)

	// ---------------------------------------------------------------
	// go listener.Listen(tempChan)
	// _, err = backup.BackUp(time.Now().Round(time.Second))

	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println("Done")

	// common.GlobalLogger.Info("Success",
	// 	zap.Uint64("average chunk size", (backup.ChunkingStats.Size/backup.ChunkingStats.Num)))

	// // time.Sleep(5 * time.Second)

	// speed := float64(backup.ChunkingStats.Size) / float64(backup.Duration.Seconds())
	// fmt.Println("Length : ", backup.ChunkingStats.Size)
	// fmt.Println("ChunkNum : ", backup.ChunkingStats.Num)
	// fmt.Println("Duration nano : ", float64(backup.Duration.Seconds()))
	// fmt.Println(speed)

	// err = delete.Delete(buds.TimeStamp)

	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(time.Now())

}
