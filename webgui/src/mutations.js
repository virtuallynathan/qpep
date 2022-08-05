import store from "./store";
import { LogManager } from "aurelia-framework";
export var log = LogManager.getLogger("qpep");

function setHostTypeAndPort(state, type, port) {
  const newState = Object.assign({}, state, { type: type, port: port });
  return newState;
}

function showMessage(state, text, type) {
  const newState = Object.assign({}, state, { msg: text, msgType: type });
  return newState;
}

function showLoader(state) {
  const newState = Object.assign({}, state, { showLoader: true });
  return newState;
}

function hideLoader(state) {
  const newState = Object.assign({}, state, { showLoader: false });
  return newState;
}

store.registerAction('setHostTypeAndPort', setHostTypeAndPort);
store.registerAction('showMessage', showMessage);
store.registerAction('showLoader', showLoader);
store.registerAction('hideLoader', hideLoader);

export {
    setHostTypeAndPort,
    showMessage,
    showLoader, 
    hideLoader,
};
