import { LogManager } from "aurelia-framework";
export var log = LogManager.getLogger("toast");
import { BindingSignaler } from "aurelia-templating-resources";
import { inject } from "aurelia-framework";

import { connectTo } from "aurelia-store";

import { clearMessage } from "./actions";

@inject(BindingSignaler)
@connectTo()
export class Toast {
  msg = "<null>";
  type = "info";
  shown = false;
  shownClass = "toast-hidden";
  timeout = 0;

  signaler = null;

  constructor(signaler) {
    this.signaler = signaler;
  }

  stateChanged(newState, oldState) {
    if (
      oldState !== undefined &&
      newState.toast_msg !== null &&
      newState.toast_msg !== this.msg
    ) {
      if (this.shown) {
        log.error(
          "Error was ignored because another error notification is already shown: ",
          newState.toast_msg
        );
        return;
      }
      log.info("newstate: ", newState);
      this.show(
        newState.toast_msg,
        newState.toast_type,
        newState.toast_timeout
      );
    }
  }

  show(msg, type, timeout) {
    this.msg = msg;
    this.type = type;
    this.timeout = timeout;

    this.shown = true;
    this.shownClass = "toast-shown";

    setTimeout(() => {
      this.signaler.signal("update");

      if (this.timeout > 0) {
        // other errors / info, with normal timeout no reload
        setTimeout(() => {
          this.shown = false;
          this.shownClass = 'toast-hidden';
          
          this.signaler.signal("update");
        }, this.timeout);

        // this is to workaround the bad hide behavior
        setTimeout(() => {
          this.msg = "";
          this.type = 'info';
          
          this.signaler.signal("update");
        }, this.timeout + 1000 );

        return;
      }

      // on loading screen, reload in case of error
      setTimeout(() => {
        this.shown = false;
        this.msg = "";
      }, 8000);

      setTimeout(() => {
        location.reload();
      }, 9000);
    }, 1000);
  }
}
