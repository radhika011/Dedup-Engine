import React, { useState, useEffect } from "react";
import ListGroup from "react-bootstrap/ListGroup";
import "./Activity.css";
import Activity_Item from "./Activity_Item";
import Pagination from "./Pagination";
import { ReadSystemHistory } from "../wailsjs/go/main/App";
import Header from "./Header";
import "./Header_Style.css";

function Activity() {
  const [statusPerPage] = useState(5);
  const [currentPage, setCurrentPage] = useState(1);
  const [historyList, setHistoryList] = useState([]);
  const updateResult = (result) => {
    if(result!=null)
    {
      setHistoryList(result)
    }
  };
  const pages = [];
  useEffect(() => {
    displaySystemHistory();
  }, []);
  function displaySystemHistory() {
    ReadSystemHistory().then(updateResult);
  }

  for (let page = 1; page < Math.ceil(historyList.length); page++) {
    pages.push(page);
  }
  const indexOfLastStatus = currentPage * statusPerPage;
  const indexOfFirstStatus = indexOfLastStatus - statusPerPage;
  const currentStatus = historyList.slice(
    indexOfFirstStatus,
    indexOfLastStatus
  );

  const paginate = (pageNumber) => setCurrentPage(pageNumber);
  return (
    <div className="activity__container">
      <Header pageTitle="Activity"></Header>
      <p
        style={{
          margin: "0",
          backgroundColor: "#f3f7f7",
          paddingLeft: "2.5%",
          width: "100%",
          textAlign: "left",
          height: "12%",
          fontWeight: 700,
          display: "flex",
          alignItems: "center",
          fontSize: "12px",
        }}
      >
        ACTIVITY STATUS
      </p>
      <div className="activites__list">
        {historyList.length == 0 ? (
          <p
            style={{
              display: "flex",
              justifyContent: "center",
              fontWeight: "700",
              paddingTop: "5%",
            }}
          >
            No Backup History
          </p>
        ) : (
          <ListGroup style={{ height: "100%" }}>
            {currentStatus.map((perStatus) => (
              <Activity_Item
                Status={JSON.parse(perStatus).status}
                Type={JSON.parse(perStatus).type}
                Description={JSON.parse(perStatus).description}
                Time={JSON.parse(perStatus).timestamp}
              ></Activity_Item>
            ))}
          </ListGroup>
        )}
      </div>
      <Pagination
        statusPerPage={statusPerPage}
        totalStatus={historyList.length}
        paginate={paginate}
        currentPage={currentPage}
        setCurrentPage={setCurrentPage}
      ></Pagination>
    </div>
  );
}

export default Activity;
