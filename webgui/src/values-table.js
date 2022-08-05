import { LogManager, bindable } from "aurelia-framework";
export var log = LogManager.getLogger("qpep");

var $ = require("jquery");
var DataTable = require("datatables.net-dt");

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

    let table = new DataTable("#"+this.tableId, {
      paging: false,
      search: true,
      ordering: false,
      info: false,
      ajax: function (d, cb) {
        fetch(source)
          .then((response) => response.json())
          .then((data) => cb(data));
      },
      columns: [{ data: "attribute" }, { data: "value" }],
    });
  }
}
