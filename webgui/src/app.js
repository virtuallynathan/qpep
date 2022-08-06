import { LogManager } from "aurelia-framework";
export var log = LogManager.getLogger("app");
import { inject } from "aurelia-dependency-injection";
import { Store } from "aurelia-store";

import { EventAggregator } from 'aurelia-event-aggregator';

import { setHostModeAndPort, showMessage, showLoader, hideLoader } from "./actions";

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
        route: ["", "index", "home"],
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

  stateChanged(newState, oldState) {
    log.info( 'app-state', oldState, '->', newState );
  }

  routeNavigationCompleted = (eventArgs, eventName) => {
    try {
      var mode = eventArgs.instruction.queryParams.mode;
      var port = ~~(eventArgs.instruction.queryParams.port);

      setHostModeAndPort( mode, port );

      hideLoader();
    } catch (err) {
      log.error(err);
      showMessage( err, "error", 0 );

      // disable tabs
      $(function () {
        $(".mdl-layout__tab").removeAttr("href");
        $(".is-active").removeClass("is-active");
      });
    }
  }
}
