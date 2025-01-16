import React from "react";
import "./App.css";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import SideBar from "./SideBar";
import Dashboard from "./Dashboard";
import Account from "./Account";
import Schedule from "./Schedule";
import ManageDirectories from "./Manage_Directories";
import RecoverBackup from "./Recover_Backup";
import Activity from "./Activity";
import Registration from "./Registration";
import ForgotPassword from "./ForgotPassword";
import Login from "./Login";
import RetrievePage from "./RetrievePage";
import 'bootstrap/dist/css/bootstrap.min.css';
function App() {
  return (
    <BrowserRouter>
      <div id="App">
        <Routes>
          <Route
            exact
            path="/Dashboard"
            element={
              <>
                <SideBar></SideBar>
                <Dashboard></Dashboard>
              </>
            }
          ></Route>
          <Route
            exact
            path="/account"
            element={
              <>
                <SideBar></SideBar>
                <Account></Account>
              </>
            }
          ></Route>
          <Route
            exact
            path="/schedule"
            element={
              <>
                <SideBar></SideBar>
                <Schedule></Schedule>
              </>
            }
          ></Route>
          <Route
            exact
            path="/manageDirectories"
            element={
              <>
                <SideBar></SideBar>
                <ManageDirectories></ManageDirectories>
              </>
            }
          ></Route>
          <Route
            exact
            path="/recoverBackup"
            element={
              <>
                <SideBar></SideBar>
                <RecoverBackup></RecoverBackup>
              </>
            }
          ></Route>
          <Route
            exact
            path="/activity"
            element={
              <>
                <SideBar></SideBar>
                <Activity></Activity>
              </>
            }
          ></Route>
          <Route
            exact
            path="/Registration"
            element={<Registration></Registration>}
          ></Route>
          <Route
            exact
            path="/"
            element={
              <>
                <Login></Login>
              </>
            }
          ></Route>
          <Route
            exact
            path="/forgotPassword"
            element={
              <>
                <ForgotPassword></ForgotPassword>
              </>
            }
          ></Route>
          <Route
            exact
            path="/retrieve"
            element={
              <>
                <SideBar></SideBar>
                <RetrievePage></RetrievePage>
              </>
            }
          ></Route>
        </Routes>
      </div>
    </BrowserRouter>
  );
}

export default App;
