import { LogManager } from "aurelia-framework";
export var log = LogManager.getLogger("qpep");

var $ = require("jquery");

export class App {
  constructor() {
    this.title = 'QPep high-latency network accelerator';
    this.modules = [ 'status-graph', 'statistics' ];

    $('.lds-grid').fadeIn(1000, "linear");

    setTimeout(() => {
      $('.lds-grid').fadeOut(1000, "linear", () => {
        $('.lds-grid').remove();
        log.info($('#fixed-tab-status-graph'));
        $('#fixed-tab-status-graph').addClass('is-active');
        $('#fixed-tab-status-graph > span').trigger('click');
      });
    }, 3000);
  }
}
