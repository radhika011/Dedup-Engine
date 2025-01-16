import React from "react";
import { TfiKey, TfiEmail } from "../node_modules/react-icons/tfi";
import { VscKey } from "../node_modules/react-icons/vsc";
import { Link } from "react-router-dom";
import "./ForgotPassword.css";
import { GetUserName,UpdatePassword } from "../wailsjs/go/main/App";
import  { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";

function ForgotPassword() {
  const [newPassword, setNewPassword] = useState("");
  const [confirmNewPassword, setConfirmNewPassword] = useState("");
  const [emailId, setEmail] = useState("");
  const navigate = useNavigate()
  function updateOperation() {
    const emailWarning = document.getElementById("emailWarning");
    const newPasswordWarning = document.getElementById("newPasswordWarning");
    const confirmNewPasswordWarning = document.getElementById(
      "confirmNewPasswordWarning"
    );
    if (emailId === "") {
      emailWarning.style.display = "block";
      newPasswordWarning.style.display = "none";
      confirmNewPasswordWarning.style.display = "none";
    } else if (newPassword === "") {
      emailWarning.style.display = "none";
      newPasswordWarning.style.display = "block";
confirmNewPasswordWarning.style.display = "none";
    } else if (confirmNewPassword === "") {
      emailWarning.style.display = "none";
      newPasswordWarning.style.display = "none";
      confirmNewPasswordWarning.style.display = "block";
    } else {
      emailWarning.style.display = "none";
      newPasswordWarning.style.display = "none";
      confirmNewPasswordWarning.style.display = "none";
      UpdatePassword(emailId,newPassword,confirmNewPassword).then( (result) => {
        //console.log("Updated details:",fName,lName,phoneNo)
        const jsonResponseData = JSON.parse(result);
        console.log(jsonResponseData.Status);
        console.log(jsonResponseData.Description);
        if(jsonResponseData.Status==true){
          alert("Update successful!")
          navigate('/')
        }
        else{
          alert(jsonResponseData.Description)
        }
      } );
    }
   
  }
  const confirmNewPasswordWarning = document.getElementById(
    "confirmNewPasswordWarning"
  );
  const emailWarning = document.getElementById("emailWarning");
  const newPasswordWarning = document.getElementById("newPasswordWarning");
  
  return (
    <div className="ForgotPassword">
      <div className="forgotPassword_container">
        <h3 className="forgotPassword_title">Reset Password</h3>
        <p className="email__id__title">Email</p>
        <div className="forgotPass__email__id">
          <TfiEmail className="forgotPass_icons"></TfiEmail>
          <input
            className="forgotPassword__emailId"
            type={"email"}
            placeholder="Enter Email-Id"
            onChange={(e) => setEmail(e.target.value)}
          />
        </div>
        <p
          id="emailWarning"
          style={{
            display: "none",
          }}
        >
          Please enter the email-Id
        </p>
        <p className="newPassword__title"> New Password</p>
        <div className="newpassword">
          <VscKey className="forgotPass_icons"></VscKey>
          <input
            className="forgotPass__newPass"
            type={"password"}
            placeholder="Enter New Password"
            onChange={(e) => setNewPassword(e.target.value)}
          />
        </div>
        <p
          id="newPasswordWarning"
          style={{
            display: "none",
          }}
        >
          Please enter the new password
        </p>
        <p className="confirmNewPassword__title">Confirm New Password</p>
        <div className="confirmNewpassword">
          <VscKey className="forgotPass_icons"></VscKey>
          <input
            className="forgotPass__confirmNewPass"
            type={"password"}
            placeholder="Re-enter New Password"
            onChange={(e) => setConfirmNewPassword(e.target.value)}
          />
        </div>
        <p
          id="confirmNewPasswordWarning"
          style={{
            display: "none",
          }}
        >
          Please re-enter the new password
        </p>

        <br />
        <button className="forgotPass_button" onClick={updateOperation}>Reset Password</button>

        <Link className="backToLogin_msg" to="/">
          Back to Login
        </Link>
      </div>
    </div>
  );
}

export default ForgotPassword;
