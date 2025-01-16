package schedule

import (
	"client-background/backup"
	"client-background/common"
	"client-background/listener"
	"encoding/json"
	"os"
	"time"

	"go.uber.org/zap"
)

var Schedule struct {
	Frequency      int
	NextBackUpDate string // "MM/DD/YYYY"
	Time           string // "HH:MM"
}

var NextBackUpTimeStamp time.Time
var NextTimeSet = true

func Scheduler(scheduleChan chan struct{}) {
	getSchedule()

	backupAttempted := false
	for {
		select {
		case <-scheduleChan:
			getSchedule()
			// 			discuss
			backupAttempted = false

		default:
			if NextTimeSet {
				if backupAttempted {
					SetNextBackUpDate()
					backupAttempted = false
				} else if time.Now().After(NextBackUpTimeStamp) {
					backupAttempted = true
					for i := 0; i < 5; i++ {
						common.GlobalLogger.Info("Initiating scheduled backup")
						backupStart := time.Now().Round(time.Second)
						timestampStr := backupStart.Format(common.TIME_FORMAT)
						added := listener.AddToActiveProcesses("Backup", timestampStr)
						if added {
							_, err := backup.BackUp(backupStart)
							listener.RemoveFromActiveProcesses("Backup", timestampStr)
							if err != nil {
								common.GlobalLogger.Error("Error while performing scheduled backup",
									zap.Error(err))
								time.Sleep(10 * time.Second)
							} else {
								break
							}
						}
					}
				} else {
					time.Sleep(10 * time.Second)
				}
			}
		}
	}
}

func getSchedule() {
	op := "Getting schedule information"
	fi, err := os.Stat(common.GetScheduleFile())
	if err != nil {
		NextTimeSet = false
		common.GlobalLogger.Error("Could not find schedule file",
			zap.String("Operation", op),
			zap.Error(err))
		return
	}

	if fi.Size() == 0 {
		NextTimeSet = false
		common.GlobalLogger.Error("Schedule file is empty",
			zap.String("Operation", op),
			zap.Error(err))
	}

	file, err := os.OpenFile(common.GetScheduleFile(), os.O_RDONLY, 0644)
	if err != nil {
		NextTimeSet = false
		common.GlobalLogger.Error("Could not open schedule file",
			zap.String("Operation", op),
			zap.Error(err))
		return
	}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&Schedule)
	if err != nil {
		common.GlobalLogger.Error("Could not read schedule from file",
			zap.String("Operation", op),
			zap.Error(err))
		return
	}

	if Schedule.Frequency == 0 {
		NextTimeSet = false
		common.GlobalLogger.Warn("Schedule frequency is set to NEVER",
			zap.String("Operation", op))
	}

	valueStr := Schedule.NextBackUpDate + "_" + Schedule.Time
	formatStr := "01/02/2006_15:04"

	NextBackUpTimeStamp, err = time.ParseInLocation(formatStr, valueStr, time.Local)
	if err != nil {
		common.GlobalLogger.Error("Could not get next backup time from schedule",
			zap.String("Operation", op),
			zap.Error(err))
		return
	}

	NextTimeSet = true
}

func SetNextBackUpDate() error {
	op := "Setting next backup date"
	file, err := os.OpenFile(common.GetScheduleFile(), os.O_RDWR, 0644)
	if err != nil {
		common.GlobalLogger.Error("Could not open schedule file",
			zap.String("Operation", op),
			zap.Error(err))
		return err
	}

	nhours := time.Duration(Schedule.Frequency * 24)
	for NextBackUpTimeStamp.Before(time.Now()) {
		NextBackUpTimeStamp = NextBackUpTimeStamp.Add(nhours * time.Hour)
	}

	Schedule.NextBackUpDate = NextBackUpTimeStamp.Format("01/02/2006")
	ScheduleJSON, err := common.ToJSON(Schedule)
	if err != nil {
		common.GlobalLogger.Error("Could not marshal schedule to write to file",
			zap.String("Operation", op),
			zap.Error(err))
		return err
	}
	_, err = file.Write(ScheduleJSON)
	if err != nil {
		common.GlobalLogger.Error("Could not write to schedule file",
			zap.String("Operation", op),
			zap.Error(err))
		return err
	}
	return nil
}
