import { LogManager, bindable } from "aurelia-framework";
export var log = LogManager.getLogger("values-tables");

var $ = require("jquery");
var DataTable = require("datatables.net-dt");

import { showMessage } from "./actions";

export class ValuesTableCustomElement {
  @bindable paragraph;
  @bindable tableId;
  @bindable source;

  constructor() {
    this.paragraph = "Title";
    this.tableId = "table";
    this.source = "testdata_server.json";
  }

  attached() {
    var $tab = $("statistics");
    if ($tab.is(":visible") !== true) return; // skip update

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

    setInterval(function () {
      table.ajax.reload();
    }, 3000);
  }
}
