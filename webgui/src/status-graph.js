import { LogManager } from "aurelia-framework";
export var log = LogManager.getLogger("graph");

import { showMessage } from "./actions";

import { connectTo } from "aurelia-store";

var $ = require("jquery");

Date.prototype.timeNow = function () {
  return (
    (this.getHours() < 10 ? "0" : "") +
    this.getHours() +
    ":" +
    (this.getMinutes() < 10 ? "0" : "") +
    this.getMinutes() +
    ":" +
    (this.getSeconds() < 10 ? "0" : "") +
    this.getSeconds()
  );
};

@connectTo()
export class StatusGraphCustomElement {
  constructor(signaler) {
    var now = new Date().timeNow();

    this.currentMax = 60;
    this.source = "";
    this.apiPort = 444;
    this.failed = false;

    this.dataUpload = {
      x: [now],
      y: [0],
      mode: "lines+markers",
      name: "upload",
      line: {
        color: "rgb(219, 64, 82)",
        width: 2,
      },
    };
    this.dataDownload = {
      x: [now],
      y: [0],
      mode: "lines+markers",
      name: "download",
      line: {
        color: "rgb(55, 128, 191)",
        width: 2,
      },
    };
    this.layout = {
      width: window.innerWidth,
      height: window.innerHeight * 0.8,
      xaxis: {
        autotick: true,
        ticks: "outside",
        tick0: 0,
        ticklen: 8,
        tickwidth: 4,
        tickcolor: "#000",
        nticks: 20,
        title: {
          text: "Time (seconds)",
          standoff: 20,
        },
      },
      yaxis: {
        autotick: true,
        rangemode: "tozero",
        ticks: "outside",
        tick0: 0,
        ticklen: 8,
        tickwidth: 4,
        nticks: 20,
        tickcolor: "#000",
        title: {
          text: "Speed (Kb/s)",
          standoff: 20,
        },
      },
    };

    this.updateDataTimer = setInterval(() => this.periodicDataUpdate(), 1100);
  }

  attached() {
    this.gd = document.getElementById("gd");
    Plotly.newPlot(this.gd, [this.dataUpload, this.dataDownload], this.layout, {
      responsive: true,
    });
  }

  detached() {
    clearInterval(this.updateDataTimer);
  }

  periodicDataUpdate() {
    var $tab = $("status-graph");
    if ($tab.is(":visible") !== true) return; // skip update

    fetch(this.source)
      .then((response) => {
        if (!response.ok) {
          showMessage(`HTTP Error Status: ${response.status}`, "error", 1000);
          return cb({
            data: [],
          });
        }

        this.failed = false;

        return response.json();
      })
      .then((obj) => {
        var up = null;
        var dw = null;
        obj = obj.data;

        for (var i = 0; i < obj.length; i++) {
          if (obj[i].attribute == "Current Download Speed") {
            dw = ~~obj[i].value;
          }
          if (obj[i].attribute == "Current Upload Speed") {
            up = ~~obj[i].value;
          }
        }

        this.periodicGraphUpdate({
          upload: up !== null ? up : 0,
          download: dw !== null ? dw : 0,
        });
      })
      .catch((error) => {
        if (!this.failed) {
          showMessage(error, "error", 1000);
          this.failed = true;
        }
        this.periodicGraphUpdate({
          upload: 0,
          download: 0,
        });
      });
  }

  periodicGraphUpdate(data) {
    var now = new Date().timeNow();
    var up = data.upload;
    var dw = data.download;

    // ensure the data is valid
    if (up === undefined || up === null || dw === undefined || up === undefined)
      return;

    Plotly.extendTraces(
      this.gd,
      {
        x: [[now], [now]],
        y: [[up], [dw]],
      },
      [0, 1]
    );

    // scales plot x range to only last seconds (via the slider value)
    var update = {
      "xaxis.range": [this.dataUpload.x.length - this.currentMax, this.dataUpload.x.length], // updates the xaxis range
    };
    Plotly.relayout(this.gd, update);
  }

  stateChanged(newState, oldState) {
    this.apiPort = newState.port;
    this.source = `http://127.0.0.1:${newState.port}/api/v1/${newState.mode}/statistics/data`;
  }
}
