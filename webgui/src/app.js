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
    this.mode = "";
    this.port = 0;

    this.title = "QPep high-latency network accelerator";
    this.modules = ["status-graph", "statistics"];
    this.clientVersion = "N/A";
    this.serverVersion = "N/A";
  
    this.updateDataTimer = setInterval(() => this.updateVersionsFooter(), 5000);

    eventAggregator.subscribe('router:navigation:complete', this.routeNavigationCompleted); 
    showLoader();
  }

  stateChanged(newState, oldState) {
    //log.info( 'app-state', oldState, '->', newState );
  }

  routeNavigationCompleted = (eventArgs, eventName) => {
    try {
      this.mode = eventArgs.instruction.queryParams.mode;
      this.port = ~~(eventArgs.instruction.queryParams.port);

      setHostModeAndPort( this.mode, this.port );

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

  updateVersionsFooter() {
    let source = `http://127.0.0.1:${this.port}/api/v1/${this.mode}/versions`;

    fetch(source)
      .then((response) => {
        return response.json();
      })
      .then((obj) => {
        this.clientVersion = obj.client;
        if( obj.server.length > 0 )
          this.serverVersion = obj.server;
        else
          this.serverVersion = "N/A";
      })
      .catch((error) => {
        showMessage(error, "error", 1000);
      });

    ;
  }
}
