import React, { useState } from "react";
import CryptoJS from "crypto-js";
import "./Registration.css";
import { TfiEmail, TfiMobile } from "../node_modules/react-icons/tfi";
import { VscKey } from "../node_modules/react-icons/vsc";
import { BsPersonCircle } from "../node_modules/react-icons/bs";
import { Register } from "../wailsjs/go/main/App";
import { Link } from "react-router-dom";
import { toast, ToastContainer } from "react-toastify";
import {useNavigate } from "react-router-dom";
function Registration() {
  const navigate = useNavigate();
  const [fName, setFname] = useState("");
  const [lName, setLname] = useState("");
  const [phoneNo, setPhoneNo] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  function register() {
    const fNameWarning = document.getElementById("fNameWarning");
    const lNameWarning = document.getElementById("lNameWarning");
    const emailWarning = document.getElementById("emailWarning");
    const phoneNoWarning = document.getElementById("phoneNoWarning");
    const passwordWarning = document.getElementById("passwordWarning");

    if (fName === "") {
      fNameWarning.style.display = "block";
      lNameWarning.style.display = "none";
      emailWarning.style.display = "none";
      passwordWarning.style.display = "none";
      phoneNoWarning.style.display = "none";
    } else if (lName === "") {
      lNameWarning.style.display = "block";
      fNameWarning.style.display = "none";
      emailWarning.style.display = "none";
      passwordWarning.style.display = "none";
      phoneNoWarning.style.display = "none";
    } else if (email === "") {
      emailWarning.style.display = "block";
      lNameWarning.style.display = "none";
      fNameWarning.style.display = "none";
      passwordWarning.style.display = "none";
      phoneNoWarning.style.display = "none";
    } else if (phoneNo === "") {
      phoneNoWarning.style.display = "block";
      passwordWarning.style.display = "none";
      lNameWarning.style.display = "none";
      emailWarning.style.display = "none";
      fNameWarning.style.display = "none";
    } else if (password === "") {
      passwordWarning.style.display = "block";
      lNameWarning.style.display = "none";
      emailWarning.style.display = "none";
      fNameWarning.style.display = "none";
      phoneNoWarning.style.display = "none";
    } else {
      lNameWarning.style.display = "none";
      fNameWarning.style.display = "none";
      emailWarning.style.display = "none";
      passwordWarning.style.display = "none";
      phoneNoWarning.style.display = "none";
      // let hashedPassword = CryptoJS.SHA256(password).toString();
      Register(fName, lName, email, phoneNo, password).then((result) => {
        const jsonResponseData = JSON.parse(result);
      console.log(jsonResponseData.Status);
      console.log(jsonResponseData.Description);
        if (jsonResponseData.Status === true) {
          toast.success("Registration Successful", {
            position: "top-right",
            autoClose: 2000,
            hideProgressBar: false,
            pauseOnHover: false,
            closeOnClick: true,
            draggable: true,
          });

          alert("Registration successful!")
          navigate('/')
        } else {
          toast.error(jsonResponseData.Description, {
            position: "top-right",
            autoClose: 2000,
            hideProgressBar: false,
            pauseOnHover: false,
            closeOnClick: true,
            draggable: true,
          });
        }
      });
    }
  }
  return (
    <div className="registration">
      <ToastContainer />
      <div className="registration__container">
        <h3 className="login_title">SIGN UP</h3>
        <p className="first_name__title">First Name</p>
        <div className="first_name">
          <BsPersonCircle className="register_icons"></BsPersonCircle>
          <input
            className="register_first_name"
            type={"text"}
            placeholder="Enter First Name"
            onChange={(e) => setFname(e.target.value)}
          />
        </div>
        <p
          id="fNameWarning"
          style={{
            display: "none",
          }}
        >
          Please enter your first name
        </p>
        <p className="last_name__title">Last Name</p>
        <div className="last_name">
          <BsPersonCircle className="register_icons"></BsPersonCircle>
          <input
            className="register_last_name"
            type={"text"}
            placeholder="Enter Last Name"
            onChange={(e) => setLname(e.target.value)}
          />
        </div>
        <p
          id="lNameWarning"
          style={{
            display: "none",
          }}
        >
          Please enter your last name
        </p>
        <p className="email__id__title">Email</p>
        <div className="email__id">
          <TfiEmail className="register_icons"></TfiEmail>
          <input
            className="login__emailId"
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
        <p className="phoneNo__title">Phone Number</p>
        <div className="phone_No">
          <TfiMobile className="register_icons"></TfiMobile>
          <input
            className="login__phoneNo"
            type={"tel"}
            placeholder="Enter Phone No."
            onChange={(e) => setPhoneNo(e.target.value)}
          />
        </div>
        <p
          id="phoneNoWarning"
          style={{
            display: "none",
          }}
        >
          Please enter the Phone Number
        </p>
        <p className="password__title">Password</p>
        <div className="password">
          <VscKey className="register_icons"></VscKey>
          <input
            className="login__password"
            type={"password"}
            placeholder="Enter Password"
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        <p
          id="passwordWarning"
          style={{
            display: "none",
          }}
        >
          Please enter the password
        </p>
        <button className="registered_button" onClick={register}>
          REGISTER
        </button>
        <p className="registered__msg">Already a member ?</p>
        <Link className="login__Existing" to="/">
          Login
        </Link>
      </div>
    </div>
  );
}

export default Registration;
