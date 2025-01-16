import React from 'react'
import "./BackupList_Item.css"
import {BsFillTrashFill} from "react-icons/bs"
import ListGroup from "react-bootstrap/ListGroup";
import { FiArrowUpRight } from "react-icons/fi";
import {Delete} from "../wailsjs/go/main/App";
import {useNavigate } from "react-router-dom";

function BackupList_Item({backup_name,count,onCountUpdate}) {
  const navigate = useNavigate();
  const navigateBackup = () =>{
    navigate('/retrieve',{
      state:{
        name:backup_name
      }
    });
  
  }
  const name_array = backup_name.split('_');
  function deleteBackup(){
    alert("Backup Deletion Process Started")
    Delete(backup_name.slice(0,-5)).then((result)=>{
      if(result==true){
        alert("Backup Deleted Successfully")
        console.log("Deleted")
        const newcount = count + 1
        onCountUpdate(newcount)
      }
      else{
         alert("Backup Deletion Unsuccessful")
      }
    })
    
  }
  return (
   
    <div className="backuplist_item_container">
        <ListGroup.Item className="listitem-backup">
              <p className="listitem__text">Backup on {name_array[2]}/{name_array[1]}/{name_array[0]} at {name_array[3]}:{name_array[4]}:{name_array[5]}</p>
              <div className='recover__operations'>
              <button className='recover__button' onClick={()=>navigateBackup()}><FiArrowUpRight></FiArrowUpRight></button>
              <button className='recover__button' onClick={()=>deleteBackup()}><BsFillTrashFill></BsFillTrashFill></button>
              </div>
      </ListGroup.Item>
    </div>
  );
}

export default BackupList_Item
