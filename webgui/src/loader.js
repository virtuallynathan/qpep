import { LogManager } from "aurelia-framework";
export var log = LogManager.getLogger("qpep");

import { connectTo } from "aurelia-store";

var $ = require("jquery");

@connectTo()
export class LoaderCustomElement {
  constructor() {}

  stateChanged(newState, oldState) {
    log.info(oldState, newState);
    if (newState.showLoader) {
      this.show();
    } else if (!newState.showLoader) {
      this.hide();
    }
  }

  show() {
    $(".lds").fadeIn(1000, "linear");
  }

  hide() {
    setTimeout(() => {
      $(".lds").fadeOut(1000, "linear", () => {
        var $activeTab = $(".is-active");
        if (!$activeTab.exists()) {
          $("#fixed-tab-status-graph").addClass("is-active");
          $("#fixed-tab-status-graph > span").trigger("click");
        }
      });
    }, 2000);
  }
}
