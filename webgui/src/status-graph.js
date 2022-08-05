import { LogManager } from "aurelia-framework";
export var log = LogManager.getLogger("graph");

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

export class StatusGraphCustomElement {
  constructor() {
    var now = (new Date).timeNow();

    this.currentMax = 60;
    this.source = '';

    this.dataUpload = {
      x: [now],
      y: [0],
      mode: "lines+markers",
      name: "upload",
      line: {
        color: "rgb(219, 64, 82)",
        width: 1,
        shape: "spline",
      },
    };
    this.dataDownload = {
      x: [now],
      y: [0],
      mode: "lines+markers",
      name: "download",
      line: {
        color: "rgb(55, 128, 191)",
        width: 1,
        shape: "spline",
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
        rangemode: 'tozero',
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

    this.updateDataTimer = setInterval(() => this.periodicDataUpdate(), 1000);
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

    var up = this.dataUpload.y[0] + 200 * (0.5 + Math.random());
    var dw = this.dataDownload.y[0] + 200 * (0.5 + Math.random());

    fetch(this.source)
      //.then((response) => response.json())
      .then((data) => {
        this.periodicGraphUpdate({
          upload: up,
          download: dw,
        });
      });
  }

  periodicGraphUpdate(data) {
    var now = (new Date).timeNow();
    var up = data.upload;
    var dw = data.download;

    // ensure the data is valid
    if( up === undefined || up === null || dw === undefined || up === undefined )
      return;

    // discard old values
    var currMax = Math.max( 1, (this.dataUpload.x.length - this.currentMax) / 2 );
    if (this.dataUpload.x.length > this.currentMax) {
      for( var i=0; i<currMax; i++ ) {
        this.dataUpload.y.shift();
        this.dataDownload.y.shift();
        this.dataUpload.x.shift();
        this.dataDownload.x.shift();
      }
    }

    this.dataUpload.x.push( now );
    this.dataDownload.x.push( now );

    this.dataUpload.y.push( up );
    this.dataDownload.y.push( dw );

    Plotly.extendTraces(
      this.gd,
      {
        y: [ [up], [dw] ],
      },
      [0, 1]
    );
  }

}
