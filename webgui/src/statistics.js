import { LogManager, computedFrom } from "aurelia-framework";
export var log = LogManager.getLogger("stats");

import { inject } from "aurelia-dependency-injection";
import {BindingSignaler} from 'aurelia-templating-resources';
import { connectTo } from "aurelia-store";

@inject(BindingSignaler)
@connectTo()
export class StatisticsCustomElement {
  isServer = false;
  isClient = false;
  apiPort = 0;
  signaler = null;
  selectedHostAddress = 'localhost';

  constructor(signaler) {
    this.signaler = signaler;
  }
  
  stateChanged(newState, oldState) {
    this.isServer = (newState.mode == 'server');
    this.isClient = (newState.mode == 'client');
    this.apiPort = newState.port;
    
    this.signaler.signal("update");
  }

  @computedFrom('apiPort')
  get serverSourceSelect(){
    return `http://127.0.0.1:${this.apiPort}/api/v1/server/statistics/hosts`;
  }
  @computedFrom('apiPort', 'selectedHostAddress')
  get serverGeneralInfo(){
    return `http://127.0.0.1:${this.apiPort}/api/v1/server/statistics/${this.selectedHostAddress}/info`;
  }
  @computedFrom('apiPort', 'selectedHostAddress')
  get serverStatisticsData(){
    return `http://127.0.0.1:${this.apiPort}/api/v1/server/statistics/${this.selectedHostAddress}/data`;
  }

  @computedFrom('apiPort')
  get clientGeneralInfo(){
    return `http://127.0.0.1:${this.apiPort}/api/v1/client/statistics/info`;
  }
  @computedFrom('apiPort')
  get clientStatisticsData(){
    return `http://127.0.0.1:${this.apiPort}/api/v1/client/statistics/data`;
  }

  @computedFrom('isServer')
  get serverIsShown() {
    return this.isServer;
  }
  @computedFrom('isClient')
  get clientIsShown() {
    return this.isClient;
  }
  
}
