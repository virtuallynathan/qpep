import { LogManager, bindable } from "aurelia-framework";
export var log = LogManager.getLogger("values-tables");

var $ = require("jquery");
var DataTable = require("datatables.net-dt");

import { showMessage } from "./actions";

import { inject } from "aurelia-dependency-injection";
import { TaskQueue } from "aurelia-task-queue";

@inject(TaskQueue)
export class ValuesTableCustomElement {
  @bindable paragraph;
  @bindable tableId;
  @bindable source;
  @bindable shown;

  constructor(queue) {
    this.paragraph = "Title";
    this.tableId = "table";
    this.prevSource = "testdata_server.json";
    this.source = "testdata_server.json";
    this.queue = queue;
    this.shown = false;
  }

  attached() {
    this.queue.queueMicroTask(() => {
      if( !this.shown || this.prevSource == this.source ) {
        return
      }
      this.prevSource = this.source;

      clearInterval( this.update );
      this.table.destroy();

      this.initGraph();
    });

    this.initGraph();
  }

  initGraph() {
    if( !this.shown )
      return;

    var source = this.source;

    let table = new DataTable("#" + this.tableId, {
      paging: false,
      search: true,
      ordering: false,
      info: false,
      ajax: function (d, cb) {
        fetch(source)
          .then((response) => response.json())
          .then((data) => cb(data))
          .catch((error) => {
            showMessage(error, "error", 3000);
          });
      },

      columns: [{ data: "attribute" }, { data: "value" }],
    });
    this.table = table;

    this.update = setInterval(function () {
      var $tab = $("statistics");
      if ($tab.is(":visible") !== true) return; // skip update

      table.ajax.reload();
    }, 3000);
  }

}
