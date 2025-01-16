package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	BG_TCP_HOST string
	BG_TCP_PORT string
	BG_TCP_TYPE string

	ROOT_PATH    string
	DATA_PATH    string
	RESTORE_PATH string
	TIME_FORMAT  string
)

var USER_DATA_DIR string

const infinityInterval = 10000

var current_status BackupStatus

var currentUserFile string

// var (
// 	scheduleFile    = filepath.Join("data", "schedule.json")
// 	directoriesFile = filepath.Join("data", "directories.json")
// 	sysHistoryFile  = filepath.Join("data", "sysHistory.jsonl")
// 	backupsFolder   = filepath.Join("data", "backups")

// 	currentUserFile = filepath.Join("data", "currentUser.json")
// )

const timeInterval = 1

type SystemHistory struct {
	Type        string `json:"type"`
	Timestamp   string `json:"timestamp"`
	Status      string `json:"status"`
	Description string `json:"description"`
}
type DirectoryData struct {
	DirNames []string `json:"Dirs"`
}

type ScheduleData struct {
	Frequency    int    `json:"Frequency"`
	ScheduleDate string `json:"NextBackUpDate"`
	ScheduleTime string `json:"Time"`
}
type BackupData struct {
	Username       string               `json:"Username"`
	TimeStamp      string               `json:"TimeStamp"`
	ClientUtilID   int                  `json:"ClientUtilID"`
	DirectoryArray []DirectoryArrayItem `json:"DirectoryArray"`
	Size           uint64               `json:"Size"`
}
type DirectoryArrayItem struct {
	Name     string               `json:"Name"`
	Type     string               `json:"Type"`
	Hash     []int                `json:"Hash"`
	Children []DirectoryArrayItem `json:"Children"`
	Valid    bool                 `json:"Valid"`
}
type InterfaceRequest struct {
	Type       string
	Parameters []byte
}
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
type Hash [32]byte

type BackupStatus struct {
	ProcessedData  uint64
	TotalData      uint64
	BackupComplete bool
}
type UserData struct {
	FirstName   string `json:"FirstName"`
	LastName    string `json:"LastName"`
	PhoneNumber string `json:"PhoneNumber"`
	EmailID     string `json:"EmailID"`
	Password    Hash   `json:"Password"`
}

type ResponseData struct {
	Status      bool
	Description string
}

type LoginCredentials struct {
	EmailID  string
	Password Hash
}

type Backup_Item struct {
	Backup_Name string
}

type CurrentUser struct {
	UserName string `json:"UserName"`
}
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// //	TO RUN IN DEV MODE USE:

	// dir, err := os.Getwd()
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	// ROOT_PATH = filepath.Join(filepath.Dir(dir), "test")
	// fmt.Println(ROOT_PATH)

	// // -----------------------------------------------------------

	//	TO RUN IN PROD MODE USE:

	if runtime.GOOS == "windows" {
		ROOT_PATH = filepath.Join(os.Getenv("ProgramFiles"), "backup")
	} else if runtime.GOOS == "linux" {
		ROOT_PATH = filepath.Join(os.Getenv("HOME"), ".backup")
	}

	// // -----------------------------------------------------------

	DATA_PATH = filepath.Join(ROOT_PATH, "data")
	currentUserFile = filepath.Join(DATA_PATH, "currentUser.json")

	if !LoadEnvFile() {
		fmt.Println("quitting")
		os.Exit(1)
	}

	// currentUserFile = filepath.Join(ROOT_PATH, "currentUser.json")

}

func LoadEnvFile() bool {

	err := godotenv.Load(filepath.Join(ROOT_PATH, ".env.common"))
	if err != nil {
		fmt.Println(err)
		return false
	}
	BG_TCP_HOST = os.Getenv("BG_TCP_HOST")
	BG_TCP_PORT = os.Getenv("BG_TCP_PORT")
	BG_TCP_TYPE = os.Getenv("BG_TCP_TYPE")

	restorePath, pathSet := os.LookupEnv("RESTORE_PATH")

	if pathSet {
		RESTORE_PATH = restorePath
	} else {
		RESTORE_PATH = filepath.Join(ROOT_PATH, "restore")
	}

	TIME_FORMAT = os.Getenv("TIME_FORMAT")

	fmt.Println(BG_TCP_PORT)
	return true
}

func MakeUserDir(username string) error {

	USER_DATA_DIR = username
	fmt.Println("USER_DATA_DIR: ", USER_DATA_DIR)
	// scheduleFile = filepath.Join(ROOT_PATH, username, "schedule.json")
	// directoriesFile = filepath.Join(ROOT_PATH, username, "directories.json")
	// sysHistoryFile = filepath.Join(ROOT_PATH, username, "sysHistory.jsonl")
	// backupsFolder = filepath.Join(ROOT_PATH, username, "backups")

	err := os.MkdirAll(getBackupsFolder(), 0777)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = os.OpenFile(getScheduleFile(), os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}

	dirFile, err := os.OpenFile(getDirectoriesFile(), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = dirFile.WriteString("{\"Dirs\":[]}")
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = os.OpenFile(getSysHistoryFile(), os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (a *App) ManageBackupStatus(processed_data uint64, total_data uint64, complete bool, request bool) (uint64, uint64, bool) {
	if request {
		return current_status.ProcessedData, current_status.TotalData, current_status.BackupComplete
	}
	current_status.ProcessedData = processed_data
	current_status.TotalData = total_data
	current_status.BackupComplete = complete
	return current_status.ProcessedData, current_status.TotalData, current_status.BackupComplete

}

func (a *App) Backup() bool {
	fmt.Println("starting")
	conn, err := net.Dial(BG_TCP_TYPE, BG_TCP_HOST+":"+BG_TCP_PORT)
	if err != nil {
		fmt.Println(err)
		return false
	}
	parameters := map[string]interface{}{}
	paramData, err := json.Marshal(parameters)
	if err != nil {
		fmt.Println("Parameter Marshalling error!")
		return false
	}
	data := InterfaceRequest{
		Type:       "backup",
		Parameters: paramData,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Marshalling error!")
		return false
	}

	_, _ = conn.Write(jsonData)
	defer conn.Close()
	var done = make(chan bool)
	go a.waitForLongResponse(conn, done)
	select {
	case <-time.After(infinityInterval * time.Second):
		fmt.Println("timed out!")
		return false
	case successFlag := <-done:
		return successFlag
	}

}
func (a *App) waitForLongResponse(connection net.Conn, done chan bool) {
	for {
		fmt.Println("Trying to read response!")
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil { // fail safe
			fmt.Println("Error reading:", err.Error())
			if err.Error() == "EOF" {
				a.ManageBackupStatus(current_status.ProcessedData, current_status.TotalData, true, false)
				break
			}

		}

		var response InterfaceResponse
		err = json.Unmarshal(buffer[:mLen], &response)
		fmt.Println("Response: ", response)
		if err != nil {
			fmt.Println("Error unmarshalling:", err.Error())

		}

		if response.Code == 0 { // 0 and -1 are terminal, 1 indicates processing
			// fmt.Println("LastParams:", params)
			done <- true
			return
		}
		if response.Code == -1 { // 0 and -1 are terminal, 1 indicates processing
			done <- false
			return
		}

		var params ResponseParam
		err = json.Unmarshal(response.Parameters, &params)
		if err != nil {
			continue
		}
		fmt.Println("Params:", params)
		time.Sleep(timeInterval * time.Second)
		fmt.Println(params.ProcessedData, params.TotalData)
		a.ManageBackupStatus(params.ProcessedData, params.TotalData, false, false)
		//return params.ProcessedData, params.TotalData, true

	}
	done <- false

}

func (a *App) Schedule() {
	fmt.Println("starting")
	conn, err := net.Dial(BG_TCP_TYPE, BG_TCP_HOST+":"+BG_TCP_PORT)
	if err != nil {
		fmt.Println(err) // don't panic
		return
	}
	parameters := map[string]interface{}{}
	paramData, err := json.Marshal(parameters)
	if err != nil {
		fmt.Println("Parameter Marshalling error!")
		return
	}
	data := InterfaceRequest{
		Type:       "schedule",
		Parameters: paramData,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Marshalling error!")
		return
	}
	_, _ = conn.Write(jsonData)
	defer conn.Close()
}

func (a *App) Retrieve(timestamp string) bool {
	fmt.Println("starting")
	conn, err := net.Dial(BG_TCP_TYPE, BG_TCP_HOST+":"+BG_TCP_PORT)
	if err != nil {
		fmt.Println(err)
		return false
	}

	parameters := map[string]interface{}{
		"Timestamp": timestamp,
	}
	paramData, err := json.Marshal(parameters)
	if err != nil {
		fmt.Println("Parameter Marshalling error!")
		return false
	}
	data := InterfaceRequest{
		Type:       "retrieve",
		Parameters: paramData,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Marshalling error!")
		return false
	}
	_, _ = conn.Write(jsonData)
	defer conn.Close()

	done := make(chan bool)
	go a.waitForLongResponseRetrieve(conn, done)
	select {
	case <-time.After(infinityInterval * time.Minute):
		fmt.Println("timed out!")
		return false
	case flag := <-done:
		return flag
	}

}
func (a *App) waitForLongResponseRetrieve(connection net.Conn, done chan bool) {
	for {
		fmt.Println("Trying to read response!")
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil { // fail safe
			fmt.Println("Error reading:", err.Error())
			if err.Error() == "EOF" {
				break
			}

		}

		var response InterfaceResponse
		err = json.Unmarshal(buffer[:mLen], &response)
		fmt.Println("Response: ", response)
		if err != nil {
			fmt.Println("Error unmarshalling:", err.Error())

		}

		if response.Code == 0 { // 0 and -1 are terminal, 1 indicates processing
			done <- true
			return
		}
		if response.Code == -1 { // 0 and -1 are terminal, 1 indicates processing
			done <- false
			return
		}
		var params ResponseParam
		err = json.Unmarshal(response.Parameters, &params)
		if err != nil {
			continue
		}

		fmt.Println("Params:", params)
		time.Sleep(timeInterval * time.Second)
		fmt.Println(params.ProcessedData, params.TotalData)
		//return params.ProcessedData, params.TotalData, true
	}
	done <- false

}

func (a *App) Delete(timestamp string) bool {
	USER_DATA_DIR = a.GetUserName()
	fmt.Println("starting")
	// backupFileLocation := filepath.Join(getBackupsFolder(), timestamp) + ".bkup"
	// fmt.Println(backupFileLocation)
	// err := os.Remove(backupFileLocation)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return false
	// }

	conn, err := net.Dial(BG_TCP_TYPE, BG_TCP_HOST+":"+BG_TCP_PORT)
	if err != nil {
		fmt.Println(err)
		return false
	}

	parameters := map[string]interface{}{
		"Timestamp": timestamp,
	}
	paramData, err := json.Marshal(parameters)
	if err != nil {
		fmt.Println("Parameter Marshalling error!")
		return false
	}
	data := InterfaceRequest{
		Type:       "delete",
		Parameters: paramData,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Marshalling error!")
		return false
	}
	_, _ = conn.Write(jsonData)
	flag := waitForShortResponseDelete(conn)
	conn.Close()

	return flag

}

func waitForShortResponseDelete(connection net.Conn) bool {
	fmt.Println("Trying to read response!")
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return false
	}

	var response InterfaceResponse
	err = json.Unmarshal(buffer[:mLen], &response)
	if err != nil {
		fmt.Println("Error unmarshalling:", err.Error())
		return false
	}
	if response.Code == -1 {
		fmt.Println("Operation Failed!")
		return false
	}
	if response.Code == 0 {
		fmt.Println("Operation Success!")
		return true
	}

	return false
}

func createFileIfNotExist(filename string) error {
	// check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) { // change this
		// file does not exist, create it with {} as content
		err = os.WriteFile(filename, []byte("{}"), 0644)
		if err != nil {
			return fmt.Errorf("error creating file: %v", err)
		}
		fmt.Println("file created successfully")
	}
	return nil
}

func readDataFromDirectoryFile(filename string) (*DirectoryData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var dir DirectoryData
	err = json.Unmarshal(data, &dir)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON data: %v", err)
	}

	return &dir, nil
}
func marshalData(dir *DirectoryData) ([]byte, error) {

	newData, err := json.Marshal(dir)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON data: %v", err)
	}

	return newData, nil
}

func writeDataToFile(filename string, data []byte) error {
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

func (a *App) AddNewDirectory(dirName string) string {
	filename := getDirectoriesFile()

	err := createFileIfNotExist(filename)
	if err != nil {
		fmt.Println("Add dir :", err)
		return "Error occured while adding directory"
	}

	dir, err := readDataFromDirectoryFile(filename)
	if err != nil {
		fmt.Println("Add dir :", err)
		return "Error occured while adding directory"
	}

	if a.SearchForDirectory(dirName) {
		return " already exists"
	}
	dir.DirNames = append(dir.DirNames, dirName)

	newData, err := marshalData(dir)
	if err != nil {
		fmt.Println("Add dir :", err)
		return "Error occured while adding directory"
	}

	err = writeDataToFile(filename, newData)
	if err != nil {
		fmt.Println("Add dir :", err)
		return "Error occured while adding directory"
	}

	fmt.Println("Data added and saved successfully!")
	return " added successfully!"
}

func (a *App) GetDirectories() []string {
	filename := getDirectoriesFile()
	dir, err := readDataFromDirectoryFile(filename)
	if err != nil {
		fmt.Println("Get dir :", err)
		return []string{}
	}

	return dir.DirNames
}

func (a *App) SearchForDirectory(dirToSearch string) bool {
	filename := getDirectoriesFile()
	dir, err := readDataFromDirectoryFile(filename)
	if err != nil {
		fmt.Print("Search dir :", err)
	}
	for _, dirName := range dir.DirNames {
		if dirName == dirToSearch {
			return true
		}

	}
	return false
}

func (a *App) DeleteDirectory(dirToDelete string) {
	filename := getDirectoriesFile()

	dir, err := readDataFromDirectoryFile(filename)
	if err != nil {
		fmt.Println("Delete dir :", err)
		return
	}

	//to delete
	for i, dirName := range dir.DirNames {
		if dirName == dirToDelete {
			newDirList := append(dir.DirNames[:i], dir.DirNames[i+1:]...)
			dir.DirNames = newDirList
		}
	}

	newData, err := marshalData(dir)
	if err != nil {
		fmt.Println("Delete dir :", err)
		return
	}

	err = writeDataToFile(filename, newData)
	if err != nil {
		fmt.Println("Delete dir :", err)
		return
	}
	fmt.Println("Data deleted and saved successfully!")
}

func (a *App) ReadSystemHistory() []string {
	// if _, err := os.Stat(getSysHistoryFile()); os.IsNotExist(err) {
	// 	return []string{}
	// }

	file, err := os.OpenFile(getSysHistoryFile(), os.O_RDONLY, 0644)
	if err != nil {
		// handle error
		fmt.Println(err)
		return []string{}
	}
	defer file.Close()

	var systemHistoryList = []string{}

	decoder := json.NewDecoder(file)

	for {
		var sysHistory SystemHistory

		err := decoder.Decode(&sysHistory)
		if err == io.EOF {
			break // end of file
		} else if err != nil {
			fmt.Println(err)
			break
		}
		sysHistoryObject := SystemHistory{Type: sysHistory.Type, Timestamp: sysHistory.Timestamp, Status: sysHistory.Status, Description: sysHistory.Description}

		// Convert the Person object to a JSON string
		jsonBytes, err := json.Marshal(sysHistoryObject)
		if err != nil {
			fmt.Println(err)
			continue
		}

		jsonString := string(jsonBytes)
		systemHistoryList = append(systemHistoryList, jsonString)

	}

	// fmt.Println("systemHistoryList: ", systemHistoryList)
	// fmt.Printf("%T", systemHistoryList)
	return systemHistoryList
}

func (a *App) SaveSchedule(freq string, time string, date string) {

	schfrequency, err := strconv.Atoi(freq)
	if err != nil {
		fmt.Println(err)
	}

	schedule := ScheduleData{
		Frequency:    schfrequency,
		ScheduleTime: time,
		ScheduleDate: date,
	}
	// Open the file for writing, truncating it if it already exists
	// file, err := os.OpenFile("Schedule.json", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	file, err := os.OpenFile(getScheduleFile(), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Encode the struct as JSON and write it to the file
	encoder := json.NewEncoder(file)
	err = encoder.Encode(schedule)
	if err != nil {
		fmt.Println("here")
		log.Fatal(err)
	}

	a.Schedule()
}

func (a *App) GetScheduleDetails() string {
	file, err := os.Open(getScheduleFile())
	var scheduleDetails string
	if err != nil {
		fmt.Println(err)
		return ""
	}

	defer file.Close()
	var schedule ScheduleData
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&schedule)
	if err == io.EOF {
		return "" // end of file
	} else if err != nil {
		fmt.Println(err)
		return ""
	}
	schedule = ScheduleData{Frequency: schedule.Frequency, ScheduleDate: schedule.ScheduleDate, ScheduleTime: schedule.ScheduleTime}

	// Convert the Person object to a JSON string
	jsonBytes, err := json.Marshal(schedule)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	scheduleDetails = string(jsonBytes)

	return scheduleDetails
}

func (a *App) GetBackupList() []string {
	files, err := os.ReadDir(getBackupsFolder())
	if err != nil {
		return []string{}
	}
	file_array := []string{}
	if files != nil {
		for _, file := range files {
			file_array = append(file_array, file.Name())
		}

		return file_array
	}
	return []string{}
}

func (a *App) GetRecentBackupList() []string {
	file_array := a.GetBackupList()
	if len(file_array) < 5 {
		return file_array
	}
	recent_backups := file_array[len(file_array)-5:]
	return recent_backups
}

func (a *App) GetLastBackup() string {
	lastBackup := ""
	file_array := a.GetBackupList()
	if len(file_array) < 1 {
		return ""
	}
	recent_backups := file_array[len(file_array)-1:]
	lastBackup = recent_backups[0]
	defer fmt.Println("lb :", lastBackup)
	return lastBackup
}

func (a *App) GetLastBackupSize() string {
	timestamp := a.GetLastBackup()
	if timestamp == "" {
		return "0B"
	}
	file_name := filepath.Join(getBackupsFolder(), timestamp)
	backup_data, err := readBackupDetails(file_name)
	if err != nil {
		log.Fatal(err)
	}
	last_size := backup_data.Size
	description := ""
	if last_size < 1024 {
		description = fmt.Sprintf("%d", last_size) + " B"
	} else if last_size < 1024*1024 {
		description = fmt.Sprintf("%d", last_size/1024) + "KB"
	} else if last_size < 1024*1024*1024 {
		description = fmt.Sprintf("%d", last_size/(1024*1024)) + "MB"
	} else {
		description = fmt.Sprintf("%d", last_size/(1024*1024*1024)) + "GB"
	}

	//fmt.Println(last_size)
	return description
}

func readBackupDetails(filename string) (*BackupData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var dir BackupData
	err = json.Unmarshal(data, &dir)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON data: %v", err)
	}

	return &dir, nil
}

func (a *App) GetDirectoryStructure(timestamp string) string {

	file_name := filepath.Join(getBackupsFolder(), timestamp)
	//fmt.Println(file_name)
	backup_data, err := readBackupDetails(file_name)
	if err != nil {
		log.Fatal(err)
	}
	jsonBytes, err := json.Marshal(backup_data)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return string(jsonBytes)
}

func (a *App) GetBackupListRange(start_date string, end_date string) []string {
	files, err := os.ReadDir(getBackupsFolder())
	if err != nil {
		log.Fatal(err)
	}
	const format = "2006-01-02"
	start_date_array := strings.Split(start_date, "T")[0]
	end_date_array := strings.Split(end_date, "T")[0]

	startDate, err := time.Parse(format, start_date_array)
	endDate, err := time.Parse(format, end_date_array)

	file_array := []string{}
	if files != nil {
		for _, file := range files {
			file_name_array := strings.Split(file.Name(), "_")
			current_day := file_name_array[2]
			current_month := file_name_array[1]
			current_year := file_name_array[0]
			date_str := current_year + "-" + current_month + "-" + current_day

			currentDate, _ := time.Parse(format, date_str)
			fmt.Println(currentDate, startDate, endDate)
			if currentDate.After(startDate) && currentDate.Before(endDate) {
				file_array = append(file_array, file.Name())
			}
			if currentDate.Equal(startDate) || currentDate.Equal(endDate) {
				file_array = append(file_array, file.Name())
			}

		}

		return file_array
	}
	return nil
}

func readRegisterUserData(email_id string) (*UserData, error) {
	data, err := os.ReadFile("data/register_dummy.json")
	//fmt.Println(data)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var user UserData
	err = json.Unmarshal(data, &user)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON data: %v", err)
	}

	return &user, nil
}
func readLoginUserData(email_id string) (*UserData, error) {
	data, err := os.ReadFile("data/login_dummy.json")
	//fmt.Println(data)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var user UserData
	err = json.Unmarshal(data, &user)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON data: %v", err)
	}

	return &user, nil
}

func (a *App) LogoutUser() bool {
	fmt.Println("starting")
	conn, err := net.Dial(BG_TCP_TYPE, BG_TCP_HOST+":"+BG_TCP_PORT)
	if err != nil {
		fmt.Println(err)
		return false
	}
	parameters := map[string]interface{}{}
	paramData, err := json.Marshal(parameters)
	if err != nil {
		fmt.Println("Parameter Marshalling error!")
		return false
	}
	data := InterfaceRequest{
		Type:       "logout",
		Parameters: paramData,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Marshalling error!")
		return false
	}

	_, _ = conn.Write(jsonData)
	flag := waitForShortResponseDelete(conn)
	fmt.Println("logout flag: ", flag)
	conn.Close()

	return flag

}

func (a *App) LoginUser(email string, password string) string {
	// fmt.Println(email, password)
	var user LoginCredentials
	user.EmailID = email
	var pwd Hash = sha256.Sum256([]byte(password))
	user.Password = pwd

	conn, err := net.Dial(BG_TCP_TYPE, BG_TCP_HOST+":"+BG_TCP_PORT)
	if err != nil {
		fmt.Println(err) // don't panic
		fmt.Println("Parameter Marshalling error!")
		response := ResponseData{
			Status:      false,
			Description: "Something went wrong!",
		}
		jsonResponseData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		return string(jsonResponseData)
	}
	//jsonUserData, err := json.Marshal(user)
	parameters := map[string]interface{}{
		"userCredentials": user,
	}
	paramData, err := json.Marshal(parameters)
	if err != nil {
		fmt.Println("Parameter Marshalling error!")
		response := ResponseData{
			Status:      false,
			Description: "Something went wrong!",
		}
		jsonResponseData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		return string(jsonResponseData)
	}
	data := InterfaceRequest{
		Type:       "login",
		Parameters: paramData,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Marshalling error!")
		response := ResponseData{
			Status:      false,
			Description: "Something went wrong!",
		}
		jsonResponseData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		return string(jsonResponseData)
	}
	_, _ = conn.Write(jsonData)
	successFlag, descrip := waitForShortResponse(conn)

	conn.Close()

	if successFlag {
		person := CurrentUser{
			UserName: email,
		}
		userdata, err := json.MarshalIndent(person, "", "  ")
		if err != nil {
			fmt.Println(err)
			response := ResponseData{
				Status:      false,
				Description: "Something went wrong!",
			}
			jsonResponseData, err := json.Marshal(response)
			if err != nil {
				log.Fatalf("Error marshaling JSON: %v", err)
			}
			return string(jsonResponseData)
		}

		file, err := os.OpenFile(currentUserFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Println(err)
			response := ResponseData{
				Status:      false,
				Description: "Something went wrong!",
			}
			jsonResponseData, err := json.Marshal(response)
			if err != nil {
				log.Fatalf("Error marshaling JSON: %v", err)
			}
			return string(jsonResponseData)
		}

		_, err = file.Write(userdata)
		if err != nil {
			fmt.Println(err)
			response := ResponseData{
				Status:      false,
				Description: "Something went wrong!",
			}
			jsonResponseData, err := json.Marshal(response)
			if err != nil {
				log.Fatalf("Error marshaling JSON: %v", err)
			}
			return string(jsonResponseData)
		}

		file.Close()
		USER_DATA_DIR = email
		// scheduleFile = filepath.Join(ROOT_PATH, DATA_DIR, "schedule.json")
		// directoriesFile = filepath.Join(ROOT_PATH, DATA_DIR, "directories.json")
		// sysHistoryFile = filepath.Join(ROOT_PATH, DATA_DIR, "sysHistory.jsonl")
		// backupsFolder = filepath.Join(ROOT_PATH, DATA_DIR, "backups")
	}
	response := ResponseData{
		Status:      successFlag,
		Description: descrip,
	}
	jsonResponseData, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}
	return string(jsonResponseData)

}

func (a *App) Register(fname string, lname string, email string, phoneNo string, password string) string {
	//fmt.Println(fname, lname, email, password)
	//return true
	fmt.Println("starting")
	//user, err := readRegisterUserData("dummy")
	var user UserData
	user.FirstName = fname
	user.LastName = lname
	user.EmailID = email
	user.PhoneNumber = phoneNo //phone change
	var pwd Hash = sha256.Sum256([]byte(password))
	//fmt.Printf("%T", hashed_pwd)
	user.Password = pwd
	// Calculate and print the hash
	//var hashed_pwd []byte
	//hashed_pwd = h.Sum(nil)
	//fmt.Printf("%T", hashed_pwd)

	//jsonUserData, err := json.Marshal(user)
	conn, err := net.Dial(BG_TCP_TYPE, BG_TCP_HOST+":"+BG_TCP_PORT)
	if err != nil {
		fmt.Println(err) // don't panic
		response := ResponseData{
			Status:      false,
			Description: "Something went wrong!",
		}
		jsonResponseData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		return string(jsonResponseData)
	}
	parameters := map[string]interface{}{
		"UserData": user,
	}
	paramData, err := json.Marshal(parameters)
	if err != nil {
		fmt.Println("Parameter Marshalling error!")
		response := ResponseData{
			Status:      false,
			Description: "Something went wrong!",
		}
		jsonResponseData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		return string(jsonResponseData)
	}
	data := InterfaceRequest{
		Type:       "register",
		Parameters: paramData,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Marshalling error!")
		response := ResponseData{
			Status:      false,
			Description: "Something went wrong!",
		}
		jsonResponseData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		return string(jsonResponseData)
	}
	_, _ = conn.Write(jsonData)
	successFlag, descrip := waitForShortResponse(conn)

	conn.Close()

	if successFlag {
		fmt.Println("user: ", user)

		err = MakeUserDir(user.EmailID)
		if err != nil {
			response := ResponseData{
				Status:      false,
				Description: "Something went wrong!",
			}
			jsonResponseData, err := json.Marshal(response)
			if err != nil {
				log.Fatalf("Error marshaling JSON: %v", err)
			}
			return string(jsonResponseData)
		}
	}
	response := ResponseData{
		Status:      successFlag,
		Description: descrip,
	}
	jsonResponseData, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}
	return string(jsonResponseData)
}

func (a *App) Update(fname string, lname string, phoneNo string) string {
	//fmt.Println("starting")
	//user, err := readRegisterUserData("dummy")
	var user UserData
	user.FirstName = fname
	user.LastName = lname
	user.EmailID = a.GetUserName()
	user.PhoneNumber = phoneNo //phone change
	var pwd Hash = [32]byte{}
	//fmt.Printf("%T", hashed_pwd)
	user.Password = pwd
	//jsonUserData, err := json.Marshal(user)
	conn, err := net.Dial(BG_TCP_TYPE, BG_TCP_HOST+":"+BG_TCP_PORT)
	if err != nil {
		fmt.Println(err) // don't panic
		response := ResponseData{
			Status:      false,
			Description: "Something went wrong!",
		}
		jsonResponseData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		return string(jsonResponseData)
	}

	parameters := map[string]interface{}{
		"UserData": user,
	}

	paramData, err := json.Marshal(parameters)
	if err != nil {
		fmt.Println("Parameter Marshalling error!")
		response := ResponseData{
			Status:      false,
			Description: "something went wrong",
		}
		jsonResponseData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		return string(jsonResponseData)
	}
	data := InterfaceRequest{
		Type:       "update",
		Parameters: paramData,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Marshalling error!")
		response := ResponseData{
			Status:      false,
			Description: "something went wrong",
		}
		jsonResponseData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		return string(jsonResponseData)
	}
	_, _ = conn.Write(jsonData)
	successFlag, descrip := waitForShortResponse(conn)

	conn.Close()
	response := ResponseData{
		Status:      successFlag,
		Description: descrip,
	}
	jsonResponseData, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}
	return string(jsonResponseData)

}
func (a *App) UpdatePassword(emailId string, password string, confirmNewPassword string) string {
	//fmt.Println("starting")
	//user, err := readRegisterUserData("dummy")
	if password != confirmNewPassword {
		response := ResponseData{
			Status:      false,
			Description: "Re-entered password does not match!",
		}
		jsonResponseData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		return string(jsonResponseData)
	}
	var user UserData
	user.FirstName = ""
	user.LastName = ""
	user.EmailID = emailId
	user.PhoneNumber = "" //phone change
	var pwd Hash = sha256.Sum256([]byte(password))
	//fmt.Printf("%T", hashed_pwd)
	user.Password = pwd
	//jsonUserData, err := json.Marshal(user)
	conn, err := net.Dial(BG_TCP_TYPE, BG_TCP_HOST+":"+BG_TCP_PORT)
	if err != nil {
		fmt.Println(err) // don't panic
		response := ResponseData{
			Status:      false,
			Description: "Something went wrong!",
		}
		jsonResponseData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		return string(jsonResponseData)
	}

	parameters := map[string]interface{}{
		"UserData": user,
	}
	paramData, err := json.Marshal(parameters)
	if err != nil {
		fmt.Println("Parameter Marshalling error!")
		response := ResponseData{
			Status:      false,
			Description: "something went wrong",
		}
		jsonResponseData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		return string(jsonResponseData)

	}
	data := InterfaceRequest{
		Type:       "update",
		Parameters: paramData,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Marshalling error!")
		response := ResponseData{
			Status:      false,
			Description: "something went wrong",
		}
		jsonResponseData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		return string(jsonResponseData)
	}
	_, _ = conn.Write(jsonData)
	successFlag, descrip := waitForShortResponse(conn)

	conn.Close()

	response := ResponseData{
		Status:      successFlag,
		Description: descrip,
	}
	jsonResponseData, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}
	return string(jsonResponseData)
}

func waitForShortResponse(connection net.Conn) (bool, string) {
	fmt.Println("Trying to read response!")
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

	var response InterfaceResponse
	err = json.Unmarshal(buffer[:mLen], &response)
	if err != nil {
		fmt.Println("Error unmarshalling:", err.Error())
		return false, "something went wrong"
	}
	if response.Code == -1 {
		fmt.Println("Operation Failed!")
		return false, "something went wrong"
	}
	if response.Code == 0 {
		fmt.Println("Operation Success!")
		var respParams UserResponseParams
		err := json.Unmarshal(response.Parameters, &respParams)
		if err != nil {
			fmt.Println(err)
			return false, "something went wrong"
		}

		switch respParams.Code {
		case -1:
			return false, respParams.Description
		case 0:
			return true, respParams.Description
		case 1:
			return false, respParams.Description
		case 2:
			return false, respParams.Description
		case 3:
			return false, "something went wrong"
		}
	}

	return false, "something went wrong"

}

func (a *App) GetUserName() string {
	file, err := os.Open(currentUserFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal the JSON data into a struct
	var data CurrentUser
	err = json.Unmarshal(fileContent, &data)
	if err != nil {
		log.Fatal(err)
	}

	// Access the data
	return data.UserName
}

func getScheduleFile() string {
	return filepath.Join(DATA_PATH, USER_DATA_DIR, "schedule.json")
}

func getDirectoriesFile() string {
	return filepath.Join(DATA_PATH, USER_DATA_DIR, "directories.json")
}

func getSysHistoryFile() string {
	return filepath.Join(DATA_PATH, USER_DATA_DIR, "sysHistory.jsonl")
}

func getBackupsFolder() string {
	return filepath.Join(DATA_PATH, USER_DATA_DIR, "backups")
}
