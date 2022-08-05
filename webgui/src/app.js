import { LogManager } from "aurelia-framework";
export var log = LogManager.getLogger("qpep");
import { inject } from "aurelia-dependency-injection";
import { Store } from "aurelia-store";

import { EventAggregator } from 'aurelia-event-aggregator';

import { setHostTypeAndPort, showMessage, showLoader, hideLoader } from "./actions";

var $ = require("jquery");
$.fn.exists = function () {
  return this.length > 0 ? this : false;
};

@inject(EventAggregator, Store)
export class App {
  configureRouter(config, router) {
    config.title = "QPep Status";
    config.options.pushState = true;
    config.options.root = "/";

    config.map([
      {
        route: ["", "home"],
        name: "home",
        moduleId: PLATFORM.moduleName("home"),
      },
    ]);
  }

  constructor(eventAggregator, router, store) {
    this.store = store;
    this.router = router;

    this.title = "QPep high-latency network accelerator";
    this.modules = ["status-graph", "statistics"];

    eventAggregator.subscribe('router:navigation:complete', this.routeNavigationCompleted); 
    showLoader();
  }

  stateChanged(newState, oldState) {}

  routeNavigationCompleted = (eventArgs, eventName) => {
    try {
      // var type = "client";
      // var port = 444;
      var type = eventArgs.instruction.queryParams.type;
      var port = ~~(eventArgs.instruction.queryParams.port);

      setHostTypeAndPort( type, port );

      hideLoader();
    } catch (err) {
      log.error(
        "Cannot work without the required 'type' and 'port' parameters, please restart"
      );
      showMessage(
        "Failed to start for an internal error, check the console log and restart",
        "error"
      );

      $(function () {
        $(".mdl-layout__tab").removeAttr("href");
        $(".is-active").removeClass("is-active");
      });
    }
  }
}
