import {BufferToFloatArray, ParseMetrics} from "../Tools";
import {Metric} from "../Tools/types";
import {SettingsState} from "../actions/types";

type OnClose = () => void;
type OnOpen = () => void;
type OnFFT = (samples: number[]) => void;
type OnMetrics = (metrics: Metric[]) => void;
type OnSettings = (settings: SettingsState) => void;

class Client {
  conn?: WebSocket;
  onSegFFT?: OnFFT;
  onFullFFT?: OnFFT;
  onClose?: OnClose;
  onOpen?: OnOpen;
  onMetrics?: OnMetrics;
  onSettings?: OnSettings;

  host: string;
  tmp: boolean;
  isSSL: boolean;
  websocketUrl: string;
  metricsUrl: string;
  metrics: Metric[];
  running: boolean;
  settings?: SettingsState;

  serverSampleRate: number;
  serverCenterFrequency: number;

  segmentSampleRate: number;
  segmentCenterFrequency: number;

  constructor(host?: string) {
    this.host = host || document.location.host;
    this.isSSL = document.location.protocol !== 'http:';
    this.websocketUrl = `${this.isSSL ? 'wss://' : 'ws://'}${this.host}/ws`;
    this.metricsUrl = `${this.isSSL ? 'https://' : 'http://'}${this.host}/metrics`;
    this.tmp = false;
    this.metrics = [];
    this.serverSampleRate = 0;
    this.serverCenterFrequency = 0;
    this.segmentSampleRate = 0;
    this.segmentCenterFrequency = 0;
    this.running = false;
    this.updateMetrics();
    this.updateSettings();
  }

  updateSettings = async () => {
    const d = await fetch('/settings.json');
    this.settings = await d.json();
    if (this.onSettings && this.settings) {
      this.onSettings(this.settings);
    }
  };

  updateMetrics = async () => {
    const data = await fetch(this.metricsUrl);
    const metricsText = await data.text();
    this.metrics = ParseMetrics(metricsText);
    this.refreshCache();
    await this.sendKeepAlive();
    await this.updateSettings();
    if (this.onMetrics) {
      this.onMetrics(this.metrics);
    }
    setTimeout(this.updateMetrics, 1000);
  };

  sendKeepAlive = () => {
    if (this.conn) {
      this.conn.send("KEEP");
    }
  };

  refreshCache = () => {
    for (let i = 0; i < this.metrics.length; i++) {
      const m = this.metrics[i];
      switch (m.name) {
        case 'server_samplerate':
          this.serverSampleRate = m.metrics[0].value;
          break;
        case 'server_center_frequency':
          this.serverCenterFrequency = m.metrics[0].value;
          break;
        case 'segment_samplerate':
          this.segmentSampleRate = m.metrics[0].value;
          break;
        case 'segment_center_frequency':
          this.segmentCenterFrequency = m.metrics[0].value;
          break;
      }
    }
  };

  getServerCenterFrequency = () => {
    return this.serverCenterFrequency;
  };

  getServerSampleRate = () => {
    return this.serverSampleRate;
  };

  getSegmentCenterFrequency = () => {
    return this.segmentCenterFrequency;
  };

  getSegmentSampleRate = () => {
    return this.segmentSampleRate;
  };

  start = () => {
    if (WebSocket) {
      this.running = true;
      this.conn = new WebSocket("ws://" + this.host + "/ws");
      this.conn.binaryType = 'arraybuffer';
      this.conn.onerror = (err) => {
        console.log(`WS Error: `, err);
        if (this.conn) {
          this.conn.close();
          this.conn = undefined;
        }
      };
      this.conn.onclose = () => {
        if (this.onClose) {
          this.onClose();
        }
        this.conn = undefined;
        if (this.running) {
          setTimeout(this.start, 1000);
        }
      };
      this.conn.onopen = () => {
        if (this.onOpen) {
          this.onOpen();
        }
      };
      this.conn.onmessage = (evt) => {
        const {data} = evt;
        const dv = new DataView(data);
        const d = BufferToFloatArray(data.slice(1));
        switch (dv.getUint8(0)) {
          case 0: // FullFFT
            if (this.onFullFFT) {
              this.onFullFFT(d);
            }
            break;
          case 1: // SegFFT
            if (this.onSegFFT) {
              this.onSegFFT(d);
            }
            break;
        }
      };
    } else {
      return false
    }

    return true;
  };

  stop = () => {
    if (this.conn) {
      this.running = false;
      this.conn.close();
    }
  };

  setOnFullFFT(cb: OnFFT) {
    this.onFullFFT = cb;
  }

  setOnSegFFT(cb: OnFFT) {
    this.onSegFFT = cb;
  }

  setOnClose(cb: OnClose) {
    this.onClose = cb;
  }

  setOnOpen(cb: OnOpen) {
    this.onOpen = cb;
  }

  setOnMetrics(cb: OnMetrics) {
    this.onMetrics = cb;
  }

  setOnSettings(cb: OnSettings) {
    this.onSettings = cb;
    if (this.settings) {
      cb(this.settings);
    }
  }
}

export {
  Client,
}
