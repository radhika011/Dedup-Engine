import React, { useState } from "react";
import "./Login.css";
import { TfiEmail } from "../node_modules/react-icons/tfi";
import { VscKey } from "../node_modules/react-icons/vsc";
import CryptoJS from "crypto-js";
import { LoginUser } from "../wailsjs/go/main/App";
import { ToastContainer, toast } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import { Link, useNavigate } from "react-router-dom";
function Login() {
  const [emailID, setEmailId] = useState("");
  const [password, setPassword] = useState("");
  const navigate = useNavigate();
  const navPass = () => {
    navigate('/forgotPassword')
  }
  const handleSubmit = () => {
    const emailWarning = document.getElementById("emailWarning");
    const passwordWarning = document.getElementById("passwordWarning");
    if (emailID === "") {
      emailWarning.style.display = "block";
      passwordWarning.style.display = "none";
    } else if (password === "") {
      emailWarning.style.display = "none";
      passwordWarning.style.display = "block";
    } else {
      emailWarning.style.display = "none";
      passwordWarning.style.display = "none";
      // let hashedPassword = CryptoJS.SHA256(password).toString();
      // console.log(hashedPassword);
      LoginUser(emailID, password).then((result) => {
        const jsonResponseData = JSON.parse(result);
      console.log(jsonResponseData.Status);
      console.log(jsonResponseData.Description);
        if (jsonResponseData.Status === true) {
          navigate("/Dashboard");
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
  };

  return (
    <div className="login">
      <ToastContainer />
      <div className="login__container">
        <h3 className="login_title">LOGIN</h3>
        <p className="email__id__title">Email</p>
        <div className="email__id">
          <TfiEmail className="login_icons"></TfiEmail>
          <input
            className="login__emailId"
            type={"email"}
            placeholder="Enter Email-Id"
            onChange={(e) => setEmailId(e.target.value)}
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
        <p className="password__title">Password</p>
        <div className="password">
          <VscKey className="login_icons"></VscKey>
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
        <a href="#" className="forgotPassword" onClick={navPass}>
          Forgot Password?
        </a>
        <br />
        <button className="login_button" onClick={handleSubmit}>
          LOGIN
        </button>
        <p className="not_register__msg">Not registered ?</p>
        <Link className="register_msg" to="/Registration">
          Create an account
        </Link>
      </div>
    </div>
  );
}

export default Login;
