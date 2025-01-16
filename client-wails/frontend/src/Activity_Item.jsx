import React from "react";
import "./Activity_Item.css";
import success from "../activity_images/success.png";
import failure from "../activity_images/failed.png";
import partial from "../activity_images/partial.webp";
function Activity_Item({ Status, Type, Description, Time }) {
  const date = new Date(Time);
  const year = date.getFullYear();
  const month = date.getMonth() + 1;
  const day = date.getDate();

  const formattedDate = `${year}-${month < 10 ? "0" + month : month}-${
    day < 10 ? "0" + day : day
  }`;
  const hours = date.getHours();
  const minutes = date.getMinutes();
  const seconds = date.getSeconds();

  const formattedTime = `${hours < 10 ? "0" + hours : hours}:${
    minutes < 10 ? "0" + minutes : minutes
  }:${seconds < 10 ? "0" + seconds : seconds}`;

  var image = "";
  if (Status === "Success") {
    image = success;
  } else if (Status === "Failure") {
    image = failure;
  } else if (Status === "Partial") {
    image = partial;
  }
  return (
    <div className="activity_item_container">
      <img className="activity__status_image" src={image}></img>
      <div className="backUp__Status">
        <p className="status__msg">
          {Type}: {Description}
        </p>
        <p className="backUp__Name">
          On {formattedDate} at {formattedTime}
        </p>
      </div>
    </div>
  );
}

export default Activity_Item;
