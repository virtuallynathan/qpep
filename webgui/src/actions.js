
import store from './store';
import { LogManager } from "aurelia-framework";
export var log = LogManager.getLogger("qpep");

import * as Mutations from './mutations';

function setHostTypeAndPort(type, port) {
    if( typeof(type) !== 'string' || typeof(port) !== 'number' ) {
        throw 'It\'s required to pass valid values for type and port parameters';
    }
    type = type.toLowerCase();
    switch(type) {
        case 'client':
        case 'server':
            break;
        default:
            throw 'The only admitted values for type are \'client\' or \'server\'';
    }
    if( port <= 0 )
        throw 'The port parameter must be a positive integer';

    store.dispatch(Mutations.setHostTypeAndPort, type, port);
}

function showMessage(msg, type) {
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

    store.dispatch(Mutations.showMessage, msg, type);
}

function showLoader() {
    store.dispatch(Mutations.showLoader);
}
function hideLoader() {
    store.dispatch(Mutations.hideLoader);
}

export {
    setHostTypeAndPort,
    showMessage,
    showLoader,
    hideLoader,
}
