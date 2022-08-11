import { LogManager, bindable } from "aurelia-framework";
export var log = LogManager.getLogger("values-tables");

var $ = require("jquery");
$.fn.exists = function () {
  return this.length > 0 ? this : false;
};
var DataTable = require("datatables.net-dt");
require("datatables.net-select-dt");

import { showMessage, setServerSelectedAddress } from "./actions";

import { inject } from "aurelia-dependency-injection";
import { TaskQueue } from "aurelia-task-queue";

@inject(TaskQueue)
export class ValuesTableCustomElement {
  @bindable paragraph;
  @bindable tableId;
  @bindable source;
  @bindable shown;
  @bindable updateSelect;

  constructor(queue) {
    this.paragraph = "Title";
    this.tableId = "table";
    this.prevSource = "";
    this.source = "";
    this.updateSelect = false;
    this.queue = queue;
    this.shown = false;
    this.failed = false;
  }

  attached() {
    setInterval(() => {
      this.queueSignalObserver();
    }, 500);

    this.initDatatable();
  }

  initDatatable() {
    if (!this.shown) return;

    let self = this;

    $(document).ready(function () {
      let source = self.source;
      let id = "#" + self.tableId;
      let table = new DataTable(id, {
        paging: false,
        search: true,
        ordering: false,
        info: false,
        language: {
          emptyTable: "",
        },
        select: {
          style: "single",
          items: "row",
          className: "selected",
          blurable: false,
          toggleable: false,
          info: false,
        },
        deferRender: true,
        rowId: "id",
        ajax: function (d, cb) {
          fetch(source, {
            headers: new Headers({ Accept: "application/json" }),
          })
            .then((response) => {
              if (!response.ok) {
                if (
                  !self.failed &&
                  response.status < 400 &&
                  response.status >= 500
                ) {
                  showMessage(
                    `HTTP Error Status: ${response.status}`,
                    "error",
                    3000
                  );
                }
                self.failed = true;
                return cb({
                  data: [],
                });
              }

              self.failed = false;
              return response.json();
            })
            .then((data) => cb(data))
            .catch((error) => {
              if (!self.failed) {
                showMessage(error, "error", 3000);
              }
              self.failed = true;
            });
        },

        columns: [{ data: "attribute" }, { data: "value" }],
      });
      self.table = table;

      self.update = setInterval(function () {
        var $tab = $("statistics");
        if ($tab.is(":visible") !== true) return; // skip update

        table.ajax.reload(null, false);
      }, 3000);
    });
  }

  selectionListener() {
    if (!this.shown || !this.updateSelect) return;

    let table = $("#" + this.tableId);
    table.on("click", "tr", function () {
      var $tr = $(this);

      table.children("tr.selected").removeClass("selected");

      $tr.addClass("selected");

      setServerSelectedAddress($tr.children().last().html());
    });
  }

  queueSignalObserver() {
    this.queue.queueMicroTask(() => {
      if (!this.shown || this.prevSource == this.source) {
        return;
      }
      this.prevSource = this.source;

      if (this.updateSelect === "true") this.updateSelect = true;
      else this.updateSelect = false;

      clearInterval(this.update);
      this.table.destroy();

      this.initDatatable();
      this.selectionListener();
    });
  }
}
