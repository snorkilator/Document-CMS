import React, { ReactElement, FC } from "react";
import ReactDOM from "react-dom/client";

interface page {
  Title: string;
  Body: string;
  ID: number;
}

//deleteRow tells server to delete row and only removes dom table row after okay from server
function deleteRow(ID: string) {
  fetch("/delete/" + ID.toString()).then((response) => {
    if (response.status == 200) {
      var row = document.getElementById(ID);
      if (!(row === null)) {
        row.remove();
      }
    }
  });
}

fetch("/titles.json")
  .then((response) => response.json())
  .then((unmarshalled) => display(unmarshalled));

function display(titles: page[]) {
  function App() {
    let rows = titles.map((page: page) => {
      return (
        <tr id={page.ID.toString()}>
          <td>
            <a href={"/view/" + page.ID}>{page.Title}</a>
          </td>
          <td>
            <button
              style={{ float: "right" }}
              className="button button1"
              onClick={() => deleteRow(page.ID.toString())}
            >
              X
            </button>
          </td>
        </tr>
      );
    });

    return (
      <div>
        <table style={{ width: "60%", maxWidth: "60rem" }}>{rows}</table>
      </div>
    );
  }

  const container: any = document.getElementById("root");
  const root = ReactDOM.createRoot(container);
  root.render(<App />);
}
