import React, { useState, useEffect } from "react";
import "./Dashboard.css";
import { FiArrowUpRight } from "react-icons/fi";
import ListGroup from "react-bootstrap/ListGroup";
import { Backup, GetLastBackup } from "../wailsjs/go/main/App";
import { useNavigate } from "react-router-dom";
import {
  GetRecentBackupList,
  GetLastBackupSize,
  ManageBackupStatus,
} from "../wailsjs/go/main/App";
import { ToastContainer, toast } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import { GetScheduleDetails } from "../wailsjs/go/main/App";
import Header from "./Header";
import ProgressBar from "@ramonak/react-progress-bar";
function Dashboard() {
  const navigate = useNavigate();
  const [processedData, setProcessedData] = useState(0);
  const [totalData, setTotalData] = useState(0);
  const [ongoingBackup, setOngoingBackup] = useState(false);
  const [recentBackupList, setRecentBackupList] = useState([]);
  const [currentSelected, setCurrentSelected] = useState("");
  const [nextBackUpDetails, setNextBackUpDetails] = useState([]);
  const [isBackUpScheduled, updateIsBackupScheduled] = useState(false);
  const [lastBackup, setLastBackup] = useState("");
  const [backupExists, setBackupExists] = useState(false);
  const [lastBackupSize, setLastBackupSize] = useState("");
  const [count, setCount] = useState(0);
  const updateResult = (result) => {
    setRecentBackupList(result);
  };
  const updateBackupSize = (result) => {
    if (result != null){
      setLastBackupSize(result);
    }
  };
  const updateLastBackup = (result) => {
    console.log(result);
    setLastBackup(result);
    console.log("genius " + result);
    if (result == "" || result == null) {
      setBackupExists(false);
    } else {
      setBackupExists(true);
    }
    console.log(backupExists + " validate exist")
  };

  const updateBackupState = (result) => {
    if (JSON.parse(result) != null && JSON.parse(result).TotalData != 0) {
      setProcessedData(JSON.parse(result).ProcessedData);
      setTotalData(JSON.parse(result).TotalData);
      setOngoingBackup(!JSON.parse(result).BackupComplete);
    }

    console.log(processedData);
    console.log(totalData);
    console.log(ongoingBackup);
  };

  const updateNextBackUp = (result) => {
    if (result != "") {
      updateIsBackupScheduled(true);
      setNextBackUpDetails([
        JSON.parse(result).Frequency,
        JSON.parse(result).NextBackUpDate,
        JSON.parse(result).Time,
      ]);
    }
  };
  const navigateRecover = () => {
    navigate("/recoverBackup");
  };
  useEffect(() => {
    getRecentBackups();
    getLastBackup();
    getNextBackUp();
    getBackupSize();

  }, []);

  useEffect(() => {
    getRecentBackups();
    getLastBackup();
    getBackupSize();

  }, [count]);

  
  function getTimeString(input_time) {
    let name_array = input_time.split("_");
    return name_array;
  }
  function backupNow() {
    toast.success("Backup Started ", {
      position: "top-right",
      autoClose: 3000,
      hideProgressBar: false,
      closeOnClick: true,
      pauseOnHover: true,
      draggable: true,
    });
    Backup().then((result)=> {
        setCount(count + 1)
        console.log("idiot " + result)
        if (result){
          toast.success("Backup Completed ", {
            position: "top-right",
            autoClose: 3000,
            hideProgressBar: false,
            closeOnClick: true,
            pauseOnHover: true,
            draggable: true,
          });
        } else {
          toast.error("Backup Failed", {
            position: "top-right",
            autoClose: 3000,
            hideProgressBar: false,
            closeOnClick: true,
            pauseOnHover: true,
            draggable: true,
          });
        }
    });
  }

  function getRecentBackups() {
    GetRecentBackupList()?.then(updateResult);
  }
  function getLastBackup() {
    GetLastBackup()?.then(updateLastBackup);
  }
  function getNextBackUp() {
    GetScheduleDetails().then(updateNextBackUp);
  }
  function getBackupSize() {
    GetLastBackupSize().then(updateBackupSize);
  }
  function getBackupStatus() {
    ManageBackupStatus(0, 0, false, true)?.then(updateBackupState);
  }
  function getCompleted() {
    if (totalData == 0) return 0;
    else {
      return (float(processedData) / float(totalData)) * 100;
    }
  }

  const navigateBackup = () => {
    navigate("/retrieve", {
      state: {
        name: currentSelected,
      },
    });
  };
  useEffect(() => {
    console.log(currentSelected);
    if (currentSelected != "") {
      navigateBackup();
    }
  }, [currentSelected]);

  const redirectRetrieve = (backupName) => {
    setCurrentSelected(backupName);
  };

  return (
    <div className="dashboard__container">
      <ToastContainer></ToastContainer>
      <Header pageTitle="Dashboard"></Header>
      {ongoingBackup ? (
        <div className="progress_component">
          <p>Backup Progress</p>
          <div className="progress_bar">
            <ProgressBar
              completed={getCompleted()}
              bgColor="#2dc464"
            ></ProgressBar>
          </div>
        </div>
      ) : (
        <div></div>
      )}

      <div className="dashboard__content">
        <div className="dashboard__recentBackUp">
          <pre style={{ margin: "2% 1% 3%" }}>RECENT BACKUPS</pre>
          {!backupExists ? (
            <div
              style={{
                height: "70%",
                display: "flex",
                justifyContent: "center",
              }}
            >
              <h4
                style={{
                  display: "flex",
                  fontSize: "medium",
                  fontWeight: "700",
                  alignItems: "center",
                  justifyContent: "center",
                  height: "100%;",
                }}
              >
                No previous backups!
              </h4>
            </div>
          ) : (
            <ListGroup className="recent_backUp__list">
              {recentBackupList.map((perBackup) => (
                <ListGroup.Item className="recentBackUp__Item">
                  <p>
                    Backup on {getTimeString(perBackup)[2]}/
                    {getTimeString(perBackup)[1]}/{getTimeString(perBackup)[0]}{" "}
                    at {getTimeString(perBackup)[3]}:
                    {getTimeString(perBackup)[4]}:{getTimeString(perBackup)[5]}
                  </p>

                  <FiArrowUpRight
                    className="recentBackUp__ItemIcon"
                    onClick={() => redirectRetrieve(perBackup)}
                  ></FiArrowUpRight>
                </ListGroup.Item>
              ))}
            </ListGroup>
          )}
          <div className="recent_BackUp_ViewMore">
            <a href="#" onClick={navigateRecover}>
              View More
            </a>
          </div>
        </div>
        <div className="currentStatus">
          <pre
            style={{
              margin: "3% 1% 0%",
              width: "100%",
              textAlign: "left",
              height: "5%",
            }}
          >
            CURRENT STATUS
          </pre>
          <div className="currentStatus__cards">
            <div className="currentStatus_cards_1">
              {backupExists ? (
                <div className="dashboard__lastBackUp">
                  <p className="lastBackUp__title">Last Backup</p>
                  <p className="lastBackUp__time">
                    {getTimeString(lastBackup)[3]}:
                    {getTimeString(lastBackup)[4]}{" "}
                  </p>
                  <p className="lastBackUp__date">
                    {getTimeString(lastBackup)[2]}/
                    {getTimeString(lastBackup)[1]}/
                    {getTimeString(lastBackup)[0]}{" "}
                  </p>
                </div>
              ) : (
                <div className="dashboard__lastBackUp">No Previous Backups</div>
              )}

              {isBackUpScheduled ? (
                <div className="dashboard__nextBackUp">
                  <p className="nextBackUp__title">Next Backup</p>
                  <p className="nextBackUp__time">{nextBackUpDetails[2]}</p>
                  <p className="lastBackUp__date">{nextBackUpDetails[1].split("/")[1]}
                    {"/"}
                    {nextBackUpDetails[1].split("/")[0]}
                    {"/"}
                    {nextBackUpDetails[1].split("/")[2]}</p>
                </div>
              ) : (
                <div className="dashboard__nextBackUp">
                  Next Backup not Scheduled
                </div>
              )}
            </div>
            <div className="currentStatus_cards_2">
              <div className="dashboard__dataBackUp">
                <p className="dataBackUp__title">Last Backup Size</p>
                <p className="dataBackUp__time">{lastBackupSize}</p>
              </div>
              <div className="dashboard__backup__now">
                <button className="nextBackUp__backUpNow" onClick={backupNow}>
                  Backup Now
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Dashboard;
