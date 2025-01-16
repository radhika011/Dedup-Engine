import React from "react";
import "./DirList_Item.css";
import { BsFillTrashFill } from "react-icons/bs";
import ListGroup from "react-bootstrap/ListGroup";
import { DeleteDirectory } from "../wailsjs/go/main/App";
import { ToastContainer, toast } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
function DirList_Item(props) {
  function deleteDir(event) {
    DeleteDirectory(props.directory_name);
    const newCount = props.count + 1;
    props.onCountUpdate(newCount);
    toast.success(props.directory_name + " deleted!", {
      position: "top-right",
      autoClose: 3000,
      hideProgressBar: false,
      closeOnClick: true,
      pauseOnHover: true,
      draggable: true,
    });
    event.preventDefault();
  }
  return (
    <div className="dirlist_item_container">
      <ToastContainer />
      <ListGroup.Item className="listitem">
        <p className="name">{props.directory_name}</p>
        <button className="trash_button" onClick={deleteDir}>
          <BsFillTrashFill></BsFillTrashFill>
        </button>
      </ListGroup.Item>
    </div>
  );
}

export default DirList_Item;
