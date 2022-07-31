import {LogManager} from "aurelia-framework";
export var log = LogManager.getLogger('qpep');

export class StatusGraphCustomElement {
    constructor() {
        this.counter = 0;

        this.dataUpload = {
            x: [ 0 ],
            y: [ 0 ],
            mode: 'lines+markers',
            name: 'upload (Kb/s)',
            line: {
                color: 'rgb(219, 64, 82)',
                width: 1,
                shape: 'spline',
            }
        };
        this.dataDownload = {
            x: [ 0 ],
            y: [ 0 ],
            mode: 'lines+markers',
            name: 'download (Kb/s)',
            line: {
                color: 'rgb(55, 128, 191)',
                width: 1,
                shape: 'spline',
            }
        };
        this.layout = {
            title: 'Speed',
            width: window.innerWidth, 
            height: window.innerHeight * 0.75,
        };

        this.updateTimer = setInterval( () => this.periodicUpdate(), 1000 );
    }

    attached() {
        this.gd = document.getElementById('gd');
        Plotly.newPlot( this.gd, [this.dataUpload, this.dataDownload], this.layout, {responsive: true});
    }

    detached() {
        clearInterval(this.updateTimer);
    }

    periodicUpdate() {
        log.info( "constructor" );
        var up = Math.random() * 1000;
        var dw = Math.random() * 1000;

        this.counter++;
        if ( this.counter > 60 ) {
            this.dataUpload.y.shift();
            this.dataDownload.y.shift();
        } else {
            this.dataUpload.x.push( this.counter );
            this.dataDownload.x.push( this.counter );
        }

        this.dataUpload.y.push( up );
        this.dataDownload.y.push( up );

        log.info( up, dw );

        Plotly.extendTraces(this.gd, {
            y: [[up], [dw]]
        }, [0, 1]);
    }
}
