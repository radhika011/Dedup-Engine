import React from "react";
import "./FileDisplayComponent.css";
function FileDisplayComponent({ objArray, marginVal }) {
  return (
    <div>
      <p>
        {objArray?.map((obj) => (
          <div>
            <div>
              <p style={{ marginLeft: marginVal }}>{obj.Name}</p>
            </div>

            <FileDisplayComponent
              objArray={obj?.Children}
              marginVal={marginVal + 100}
            ></FileDisplayComponent>
          </div>
        ))}
      </p>
    </div>
  );
}
export default FileDisplayComponent;
