import { LogManager } from "aurelia-framework";
export var log = LogManager.getLogger("toast");
import {BindingSignaler} from 'aurelia-templating-resources';
import {inject} from 'aurelia-framework';

import { connectTo } from "aurelia-store";

@inject(BindingSignaler)
@connectTo()
export class Toast {
  msg = "<null>";
  type = "info";
  shown = false;
  shownClass = 'toast-hidden';

  signaler = null;

  constructor(signaler) {
    this.signaler = signaler;
  }

  stateChanged(newState, oldState) {
    if (oldState !== undefined && newState.msg != oldState.msg) {
      this.show(newState.msg, newState.msgType);
    }
  }

  show(msg, type) {
    this.msg = msg;
    this.type = type;
    this.shown = true;
    this.shownClass = 'toast-shown';

    setTimeout(() => {
      this.signaler.signal('update');
  
      setTimeout(() => {
        this.shown = false;
      }, 8000);
      setTimeout(() => {
        location.reload();
      }, 9000);

    }, 1000);
  }
}
