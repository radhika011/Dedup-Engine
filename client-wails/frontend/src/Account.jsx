import React from "react";
import { Tab, Tabs, TabList, TabPanel } from "react-tabs";
import Header from "./Header";
import "./Account.css";
import { GetUserName,Update } from "../wailsjs/go/main/App";
import  { useState, useEffect } from "react";
function Account() {
  const [fName, setFname] = useState("");
  const [lName, setLname] = useState("");
  const [phoneNo, setPhoneNo] = useState("");
  var [username,setUsername] = useState("");
  const updateUsername = (result) => {
    setUsername(result)
  }
  function updateOperation() {
    Update(fName,lName,phoneNo).then((result) => {
      const jsonResponseData = JSON.parse(result);
      console.log(jsonResponseData.Status);
      console.log(jsonResponseData.Description);
      if(jsonResponseData.Status==true){
        alert("Update successful!")
      }
      else{
        alert(jsonResponseData.Description)
      }
      
    }
      
    )
  }
  function getUsername() {
    GetUserName().then(updateUsername)
  }
  useEffect(()=>{
    getUsername()
  },[])
  return (
    <div className="account__container">
      <Header pageTitle="ACCOUNT"></Header>
      <div className="account__info">
        <div className="account__background"></div>
        <div className="account__info__content">
          
            <div className="account_image_and_tabs">
              <img src="https://qai.org.au/wp-content/uploads/2021/03/grey-person-icon-300x298.png"></img>
              <p
                style={{
                  width: "100%",
                  padding: "1vh 0 3vh",
                  fontSize: "1.2vw",
                  fontWeight: 800,
                  color: "#2dc464",
                }}
              >
                {username}
              </p>
            </div>
            <div className="account__tab__info">
             
                <div className="panel-content account__settings">
                  <h3>Account Settings</h3>
                  <form
                    style={{
                      height: "100%",
                      width: "75%",
                      display: "flex",
                      flexDirection: "column",
                      alignItems: "center",
                      justifyContent: "center",
                      textAlign: "left",
                    }}
                  >
                    <div className="row">
                      <div className="col">
                        <label for="firstName" style={{ width: "100%" }}>
                          First Name
                        </label>
                        <input
                          style={{ width: "95%", marginRight: "5px" }}
                          id="firstName"
                          type={"text"}
                          placeholder="Enter First Name"
                          onChange={(e) => setFname(e.target.value)}
                        ></input>
                      </div>
                      <div className="col">
                        <label for="lastName" style={{ width: "100%" }}>
                          Last Name{" "}
                        </label>
                        <input
                          style={{ width: "100%" }}
                          id="lastName"
                          type={"text"}
                          placeholder="Enter Last Name"
                          onChange={(e) => setLname(e.target.value)}
                        ></input>
                      </div>
                    </div>
                    <div className="col" >
                      <label for="email">Email Id</label>
                      <input
                        id="email"
                        type={"email"}
                        placeholder={username}
                        disabled="true"
                        style = {{cursor:"not-allowed"}}
                      ></input>
                    </div>
                    <div className="col">
                      <label for="phn__number">Phone No.</label>
                      <input
                        id="phn__number"
                        type={"tel"}
                        placeholder="Enter Phone Number"
                        onChange={(e) => setPhoneNo(e.target.value)}
                      ></input>
                    </div>

                    <div className="row buttons">
                     
                      <button type={"button"} className="Save" onClick={updateOperation}>
                        Save
                      </button>
                    </div>
                  </form>
                </div>
              
             
              
            </div>
          
        </div>
      </div>
    </div>
  );
}

export default Account;
