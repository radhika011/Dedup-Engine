import "./Header_Style.css";
import "./Manage_Directories.css";
import Header from "./Header";
import ListGroup from "react-bootstrap/ListGroup";
import Pagination from "./Pagination";
import React, { useState, useEffect } from "react";
import DirList_Item from "./DirList_Item";
import { AddNewDirectory } from "../wailsjs/go/main/App";
import { GetDirectories } from "../wailsjs/go/main/App";
import { ToastContainer, toast } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
function Manage_Directories() {
  const [newDirectory, setNewDirectory] = useState();
  const [addDirResult, setAddDirResult] = useState("");

  const [count, setCount] = useState(0);
  const [currentList, setCurrentList] = useState(1);
  const [dirPerPage] = useState(5);
  const [directoryList, setDirectories] = useState([]);
  const updateResult = (result) => setDirectories(result);
  useEffect(() => {
    getDirs();
    console.log("useEffect called", { count });
  }, [count]);

  function getDirs() {
    GetDirectories().then(updateResult);
  }

  function updateDirName(event) {
    setNewDirectory(event.target.value);
  }
  const handleCountUpdate = (newCount) => {
    setCount(newCount);
    console.log("manage dir count :", count);
  };
  function addNewDirectory(event) {
    event.preventDefault();
    if (newDirectory != undefined){
      AddNewDirectory(newDirectory).then((addDirResult) => {
        setNewDirectory("");
        document.getElementById("dirToAddName").value = "";
        setCount(count + 1);
        console.log(addDirResult);
        toast.success(newDirectory + addDirResult, {
          position: "top-right",
          autoClose: 2000,
          hideProgressBar: false,
          pauseOnHover: false,
          closeOnClick: true,
          draggable: true,
        });
        console.log(addDirResult);
      });
      console.log(addDirResult);
    } else {
      toast.warning("Please enter directory/file name", {
        position: "top-right",
        autoClose: 2000,
        hideProgressBar: false,
        pauseOnHover: false,
        closeOnClick: true,
        draggable: true,
      });
    }
    
  }

  const pages = [];

  for (let page = 1; page < Math.ceil(directoryList.length); page++) {
    pages.push(page);
  }
  const indexOfLastStatus = currentList * dirPerPage;
  const indexOfFirstStatus = indexOfLastStatus - dirPerPage;
  const currentStatus = directoryList.slice(
    indexOfFirstStatus,
    indexOfLastStatus
  );

  const paginate = (pageNumber) => setCurrentList(pageNumber);

  return (
    <div className="page__container">
      <ToastContainer />
      <Header pageTitle="Manage Directories"></Header>
      <div className="manage_dir_page__content">
        <div className="dir_list">
          {directoryList.length == 0 ? (
            <p>No directories added for backup</p>
          ) : (
            <ListGroup style={{ height: "100%" }}>
              {currentStatus.map((perStatus) => (
                <DirList_Item
                  directory_name={perStatus}
                  count={count}
                  onCountUpdate={handleCountUpdate}
                ></DirList_Item>
              ))}
            </ListGroup>
          )}
        </div>
        <Pagination
          statusPerPage={dirPerPage}
          totalStatus={directoryList.length}
          paginate={paginate}
          currentPage={currentList}
          setCurrentPage={setCurrentList}
        ></Pagination>
        <div className="add_dir">
          <p className="enter_dir">Enter Directory Path:</p>
          <input
            id="dirToAddName"
            style={{
              width: "50%",
              backgroundColor: "white",
              width: "40%",
              height: "60%",
            }}
            onChange={updateDirName}
          ></input>
          <button className="directory__button" onClick={addNewDirectory}>
            Add Directory
          </button>
        </div>
      </div>
    </div>
  );
}

export default Manage_Directories;
