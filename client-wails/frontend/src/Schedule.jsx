import React, { useState, useEffect } from "react";
import "./Schedule.css";
import { TextField } from "@mui/material";
import "./Header_Style.css";
import { AiFillEdit } from "react-icons/ai";
import { Backup } from "../wailsjs/go/main/App";
import { SaveSchedule, GetScheduleDetails } from "../wailsjs/go/main/App";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import {
  LocalizationProvider,
  TimePicker,
  StaticDatePicker,
} from "@mui/x-date-pickers";
import { ToastContainer, toast } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import { fontGrid } from "@mui/material/styles/cssUtils";
import Header from "./Header";
function Schedule() {
  const [disabled, setDisabled] = useState(true);
  const [freq, setFreq] = useState();

  const [isScheduleSet, updateIsScheduleSet] = useState(false);
  const [value, setValue] = useState(new Date());
  const [nextBackUpDetails, setNextBackUpDetails] = useState([]);
  const [timeValue, setTimeValue] = useState("");
  const [dateValue, setDateValue] = useState("");
  const [isInputEnabled, setIsInputEnabled] = useState(false);
  const [selectedOption, setSelectedOption] = useState("");
  const [inputValue, setInputValue] = useState("");
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
  const updateResult = (result) => {
    if (result == "") {
    } else {
      updateIsScheduleSet(true);
      setNextBackUpDetails([
        JSON.parse(result).Frequency,
        JSON.parse(result).NextBackUpDate,
        JSON.parse(result).Time,
      ]);
      setValue(
        new Date(
          0,
          0,
          0,
          JSON.parse(result).Time.substring(0, 2),
          JSON.parse(result).Time.substring(3, 5)
        )
      );
      console.log(result);
      let frequency = JSON.parse(result).Frequency;
      setSelectedOption(frequency);
      setTimeValue(JSON.parse(result).Time);
      if (
        frequency != 0 &&
        frequency != 1 &&
        frequency != 7 &&
        frequency != 30
      ) {
        setInputValue(frequency);
      }

      setDateValue(JSON.parse(result).NextBackUpDate);
    }
  };

  useEffect(() => {
    displayScheduleDetails();
  }, []);

  function displayScheduleDetails() {
    GetScheduleDetails().then(updateResult);
  }

  const handleDateChange = (newValue) => {
    setDateValue(
      newValue.toLocaleDateString("en-US", {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
      })
    );
    console.log(dateValue);
  };
  const handleTimeChange = (newValue) => {
    const hrs = newValue.getHours().toLocaleString("en-US", {
      minimumIntegerDigits: 2,
      useGrouping: false,
    });
    const mins = newValue.getMinutes().toLocaleString("en-US", {
      minimumIntegerDigits: 2,
      useGrouping: false,
    });
    const time = hrs + ":" + mins;

    setValue(newValue);
    console.log(newValue);
    setTimeValue(time);
  };
  const clickButton = () => {
    console.log("button clicked");
    setDisabled(false);
  };
  const saveSchedule = () => {
    const time_array = timeValue.split(":")
    const hrs = time_array[0]
    const mins = time_array[1]
    if (!selectedOption.toString()) {
      alert("Please select a backup frequency!");
    } else if (isNaN(selectedOption) || selectedOption < 0) {
      alert("Please enter appropriate number of days!");
    } else if (!timeValue) {
      alert("Please select backup time!");
    } else if (!dateValue) {
      alert("Please select next backup date!");
    } else if (
      parseInt(dateValue.substring(0, 2)) == new Date().getMonth() + 1 &&
      parseInt(dateValue.substring(3, 5)) == new Date().getDate() &&
      parseInt(timeValue.substring(0, 2)) < new Date().getHours()
    ) {
      alert("Please select appropriate time!");
    } else if (isNaN(time_array[0])||isNaN(time_array[1])){
        alert("Please select backup time!");
    } else {
      setDisabled(true);
      updateIsScheduleSet(true);
      SaveSchedule(selectedOption.toString(), timeValue, dateValue);
      setNextBackUpDetails([freq, dateValue, timeValue]);
      setIsInputEnabled(false);
    }
    console.log(timeValue);
    console.log(parseInt(dateValue.substring(0, 2)));
    console.log(new Date().getMonth());
    console.log(dateValue.substring(3, 5));
    console.log(new Date().getDate());
  };

  function handleInputChange(event) {
    setInputValue(event.target.value);
    setSelectedOption(event.target.value);
  }
  const handleOptionChange = (event) => {
    setSelectedOption(event.target.value);
    setInputValue("");
    setIsInputEnabled(false);
  };
  const handleCustomOptionChange = () => {
    setSelectedOption(inputValue);
    setIsInputEnabled(true);
  };

  return (
    <div className="page__container">
      <ToastContainer></ToastContainer>
      <Header pageTitle="Schedule"></Header>
      <div className="page__content">
        <p className="schedule_top_card_title"> NEXT BACKUP INFO</p>
        {isScheduleSet ? (
          <div className="schedule__card_top">
            <p>
              Next backup is scheduled at :
              <span
                style={{
                  fontSize: "1.15vw",
                  color: "rgb(69, 69, 69)",
                  fontWeight: "800",
                  paddingLeft: "10px",
                }}
              >
                {nextBackUpDetails[1].split("/")[1]}
                    {"/"}
                    {nextBackUpDetails[1].split("/")[0]}
                    {"/"}
                    {nextBackUpDetails[1].split("/")[2]} {nextBackUpDetails[2]}
              </span>
            </p>
            <button className="button__style_top" onClick={backupNow}>
              Backup Now
            </button>
          </div>
        ) : (
          <p>Schedule is not set</p>
        )}

        <p className="schedule_bottom_card_title">SCHEDULE SETTINGS</p>
        <div className="schedule__card_bottom">
          <button className="edit_button" onClick={clickButton}>
            <AiFillEdit
              style={{ fontSize: "30px", paddingTop: "5px" }}
            ></AiFillEdit>
          </button>
          <div className="schedule__edit">
            <div className="schedule__container">
              <div className="schedule_edit_left">
                <p className="backup_setting_subtitle">BACKUP FREQUENCY</p>
                <div className="radiogroup__frequency">
                  <input
                    type="radio"
                    value="1"
                    name="freq"
                    onChange={handleOptionChange}
                    disabled={disabled}
                    checked={selectedOption === 1 ? "checked" : null}
                  />
                  Daily
                  <input
                    type="radio"
                    value="7"
                    name="freq"
                    disabled={disabled}
                    onChange={handleOptionChange}
                    checked={selectedOption === 7 ? "checked" : null}
                  />
                  Weekly
                  <input
                    type="radio"
                    value="30"
                    name="freq"
                    disabled={disabled}
                    onChange={handleOptionChange}
                    checked={selectedOption === 30 ? "checked" : null}
                  />
                  Monthly (30 days)
                  <input
                    type="radio"
                    value="Custom"
                    name="freq"
                    disabled={disabled}
                    onChange={handleCustomOptionChange}
                    checked={
                      selectedOption != 0 &&
                      selectedOption != 1 &&
                      selectedOption != 7 &&
                      selectedOption != 30
                        ? "checked"
                        : null
                    }
                  />
                  Custom
                  <input
                    type="radio"
                    value="0"
                    name="freq"
                    disabled={disabled}
                    onChange={handleOptionChange}
                    checked={selectedOption === 0 ? "checked" : null}
                  />
                  Never
                </div>

                <div className="schedule__input">
                  <p disabled={!isInputEnabled}>Backup after every </p>
                  <input
                    type={Text}
                    value={inputValue}
                    onChange={handleInputChange}
                    disabled={!isInputEnabled}
                  ></input>
                  <p>days</p>
                </div>
                <p className="backup_setting_subtitle">BACKUP TIME</p>
                <div className="TimeModule">
                  <LocalizationProvider dateAdapter={AdapterDateFns}>
                    <TimePicker
                      label="Select Backup Time"
                      value={isScheduleSet || timeValue != "" ? value : null}
                      onChange={handleTimeChange}
                      renderInput={(params) => <TextField {...params} />}
                      disabled={disabled}
                    />
                  </LocalizationProvider>
                </div>
              </div>
              <div className="schedule_edit_right">
                <p className="backup_setting_subtitle_start">
                  BACKUP START DATE
                </p>
                <LocalizationProvider dateAdapter={AdapterDateFns}>
                  <StaticDatePicker
                    label="Select start date"
                    value={dateValue}
                    onChange={handleDateChange}
                    renderInput={(params) => <TextField {...params} />}
                    disabled={disabled}
                    disablePast
                  />
                </LocalizationProvider>
              </div>
            </div>
            <button
              className="button__style_bottom"
              onClick={saveSchedule}
              disabled={disabled}
            >
              Save Changes
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Schedule;
