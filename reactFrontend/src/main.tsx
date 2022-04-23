import React from "react";
import ReactDOM from 'react-dom/client';

function App(){
    return (
        <h1>test</h1>
    );
};

export function deleteRow(ID: string){
    alert(ID + typeof(ID))
    var row = document.getElementById(ID)
    if (!(row === null)) {
        row.remove()
    }
}

const container:any = document.getElementById('root')
const root = ReactDOM.createRoot(container)
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);