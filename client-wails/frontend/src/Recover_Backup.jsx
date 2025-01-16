import React from "react";
import "./Recover_Backup.css";
import "./Header_Style.css";
import ListGroup from "react-bootstrap/ListGroup";
import { useState, useEffect } from "react";
import BackupList_Item from "./BackupList_Item";
import Pagination from "./Pagination";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import { LocalizationProvider, DatePicker } from "@mui/x-date-pickers";
import { GetBackupList, GetBackupListRange } from "../wailsjs/go/main/App";
import Header from "./Header";
import TextField from "@mui/material/TextField";

function Recover_Backup() {
  const [currentList, setCurrentList] = useState(1);
  const [backupPerPage] = useState(5);
  const pages = [];
  const [backupList, setBackups] = useState([]);
  const [count, setCount] = useState(currentList);
  const updateResult = (result) => setBackups(result);
  const [startDate, setStartDate] = useState(new Date("2000-01-01"));
  const [endDate, setEndDate] = useState(new Date());

  const handleChangeStart = (newValue) => {
    setStartDate(newValue);
    console.log("Start date selected:", { newValue });
  };
  const handleChangeEnd = (newValue) => {
    setEndDate(newValue);
    console.log("End date selected:", { newValue });
  };

  useEffect(() => {
    getBackups();
  }, []);
  useEffect(() => {
    getBackups();
    console.log("count updated");
  }, [count]);

  function getBackups() {
    GetBackupList()?.then(updateResult);
  }

  for (let page = 1; page < Math.ceil(backupList.length); page++) {
    pages.push(page);
  }
  const indexOfLastStatus = currentList * backupPerPage;
  const indexOfFirstStatus = indexOfLastStatus - backupPerPage;
  const currentStatus = backupList.slice(indexOfFirstStatus, indexOfLastStatus);

  const handleCountUpdate = (newCount) => {
    setCount(newCount);
  };
  const handleFilter = () => {
    if (startDate > endDate) {
      alert("End date cannot be greater than start date");
    } else {
      GetBackupListRange(startDate.toJSON(), endDate.toJSON())?.then(
        updateResult
      );
      console.log("Filter applied!");
    }
  };
  const paginate = (pageNumber) => setCurrentList(pageNumber);
  return (
    <div className="page__container">
      <Header pageTitle="Retrieve Backup"></Header>
      <div className="page__content">
        <div className="recover__filter_title">
          <div className="recover__range">
            <div className="from_date">
              <p>Start Date</p>
              <div className="DateModule">
                <LocalizationProvider dateAdapter={AdapterDateFns}>
                  <DatePicker
                    label="Select Date"
                    value={startDate}
                    onChange={handleChangeStart}
                    renderInput={(params) => <TextField {...params} />}
                  />
                </LocalizationProvider>
              </div>
            </div>
            <div className="to_date">
              <p>End Date</p>
              <div className="DateModule">
                <LocalizationProvider dateAdapter={AdapterDateFns}>
                  <DatePicker
                    label="Select Date"
                    value={endDate}
                    onChange={handleChangeEnd}
                    renderInput={(params) => <TextField {...params} />}
                  />
                </LocalizationProvider>
              </div>
            </div>
          </div>
          <button
            className="recover__search_button"
            onClick={() => handleFilter()}
          >
            Apply Filter
          </button>
        </div>
        {backupList.length == 0 ? (
          <p
            style={{
              display: "flex",

              justifyContent: "center",
              fontWeight: "700",
              paddingTop: "5%",
            }}
          >
            No Backups
          </p>
        ) : (
          <ListGroup className="recover__backup_list">
            {currentStatus.map((perStatus) => (
              <BackupList_Item
                backup_name={perStatus}
                count={count}
                onCountUpdate={handleCountUpdate}
              ></BackupList_Item>
            ))}
          </ListGroup>
        )}
        <Pagination
          statusPerPage={backupPerPage}
          totalStatus={backupList.length}
          paginate={paginate}
          currentPage={currentList}
          setCurrentPage={setCurrentList}
        ></Pagination>
      </div>
    </div>
  );
}

export default Recover_Backup;
