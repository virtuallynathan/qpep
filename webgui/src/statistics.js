import { LogManager } from "aurelia-framework";
export var log = LogManager.getLogger("stats");

import { inject } from "aurelia-dependency-injection";
import {BindingSignaler} from 'aurelia-templating-resources';
import { connectTo } from "aurelia-store";

@inject(BindingSignaler)
@connectTo()
export class StatisticsCustomElement {
  isServer = true;
  apiPort = 0;
  signaler = null;

  constructor(signaler) {
    this.signaler = signaler;
  }
  
  stateChanged(newState, oldState) {
    this.isServer = (newState.mode == 'server');
    this.apiPort = newState.port;
  }

}
