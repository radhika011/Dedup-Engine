import * as React from 'react';
import TextField from '@mui/material/TextField';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import {  LocalizationProvider,DatePicker } from '@mui/x-date-pickers';
import { format } from 'date-fns';
function DateModule(){
    const [value, setValue] = React.useState(new Date());
 
    const handleChange = (newValue) => {
      setValue(newValue.toLocaleDateString("en-US", {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
      }));
     

    };
    console.log(value);
    return (
       <div className="DateModule">
        <LocalizationProvider dateAdapter={AdapterDateFns}>
        <DatePicker
      label="Select Date"
      value={value}
      
      onChange={handleChange}
      renderInput={(params) => <TextField {...params} 

      />}
    />
  </LocalizationProvider>
      </div>
    );
} 
export default DateModule;