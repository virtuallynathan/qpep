
import store from './store';
import { LogManager } from "aurelia-framework";
export var log = LogManager.getLogger("actions");

import * as Mutations from './mutations';

function setHostModeAndPort(mode, port) {
    if( typeof(mode) !== 'string' || typeof(port) !== 'number' ) {
        throw 'It\'s required to pass valid values for mode and port parameters';
    }
    mode = mode.toLowerCase();
    switch(mode) {
        case 'client':
        case 'server':
            break;
        default:
            throw 'The only admitted values for mode are \'client\' or \'server\'';
    }
    if( port <= 0 )
        throw 'The port parameter must be a positive integer';

    store.dispatch(Mutations.setHostModeAndPort, mode, port);
}

function setServerSelectedAddress(address) {
    if( typeof(address) !== 'string' || address.length == 0 )
        return;

    store.dispatch(Mutations.setServerSelectedAddress, address);
}

function showMessage(msg, type, timeout) {
    if( typeof(type) !== 'string') {
        throw 'It\'s required to pass valid value for type parameter';
    }
    type = type.toLowerCase();
    switch(type) {
        case 'info':
        case 'error':
            break;
        default:
            throw 'The only admitted values for type are \'info\' or \'error\'';
    }

    store.dispatch(Mutations.showMessage, msg, type, Math.max( 0, timeout ));
}

function clearMessage() {
    store.dispatch(Mutations.clearMessage);
}

function showLoader() {
    store.dispatch(Mutations.showLoader);
}
function hideLoader() {
    store.dispatch(Mutations.hideLoader);
}

export {
    setHostModeAndPort,
    setServerSelectedAddress,
    showMessage,
    clearMessage,
    showLoader,
    hideLoader,
}
