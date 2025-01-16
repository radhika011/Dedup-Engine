import React from "react";
import "./SideBar.css";
import { NavLink } from "react-router-dom";
import {
  CDBSidebar,
  CDBSidebarContent,
  CDBSidebarHeader,
  CDBSidebarMenu,
  CDBSidebarMenuItem,
  CDBSidebarFooter,
} from "cdbreact";
import { LogoutUser } from "../wailsjs/go/main/App";
import { ToastContainer, toast } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import {useNavigate } from "react-router-dom";


function SideBar() {
  const navigate = useNavigate();
  var allowLogout = false;
  function logoutIfAllowed(){
    LogoutUser().then((result)=>{
      allowLogout = result
      if(allowLogout==true){
        console.log("Result:",result)
        allowLogout = result
        navigate('/');
      }
      else{
     alert("Cannot logout!");
    }})
  }
  return (
    <div className="sidebarDiv">
      <CDBSidebar id="sidebar">
        <CDBSidebarHeader prefix={<i className="fa fa-bars" />}>
          Dedup Engine
        </CDBSidebarHeader>
        <CDBSidebarContent id="sidebarContent">
          <CDBSidebarMenu>
            <NavLink to="/Dashboard">
              <CDBSidebarMenuItem className="sidebar__item " icon="home">
                Dashboard
              </CDBSidebarMenuItem>
            </NavLink>
            <NavLink to="/account">
              <CDBSidebarMenuItem className="sidebar__item" icon="user-edit">
                Account
              </CDBSidebarMenuItem>
            </NavLink>
            <NavLink to="/schedule">
              <CDBSidebarMenuItem
                className="sidebar__item"
                icon="clock"
                iconType="solid"
              >
                Schedule
              </CDBSidebarMenuItem>
            </NavLink>
            <NavLink to="/manageDirectories">
              <CDBSidebarMenuItem className="sidebar__item" icon="folder-open">
                Manage Directories
              </CDBSidebarMenuItem>
            </NavLink>
            <NavLink to="/recoverBackup">
              <CDBSidebarMenuItem
                className="sidebar__item"
                icon="cloud-download-alt"
              >
                Retrieve Backup
              </CDBSidebarMenuItem>
            </NavLink>
            <NavLink to="/activity">
              <CDBSidebarMenuItem className="sidebar__item" icon="undo-alt">
                Activity
              </CDBSidebarMenuItem>
            </NavLink>
          </CDBSidebarMenu>
        </CDBSidebarContent>

        <CDBSidebarFooter style={{ paddingBottom: "10%" }}>
          
            <CDBSidebarMenuItem
              className="sidebar__item"
              icon="power-off"
              iconType="solid"
              onClick={logoutIfAllowed}
            >
              Sign Out
            </CDBSidebarMenuItem>
         
        </CDBSidebarFooter>
      </CDBSidebar>
    </div>
  );
}

export default SideBar;
