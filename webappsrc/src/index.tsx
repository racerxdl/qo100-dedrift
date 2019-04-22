import React from 'react';
import ReactDOM from 'react-dom';
import {createStore} from "redux";
import {Provider} from "react-redux";

import './index.css';
import App from './App';
import {Client} from "./Client";
import * as serviceWorker from './serviceWorker';
import appReducers from "./actions/reducers";
import {AddFullFFT, AddMetrics, AddSegmentFFT, SetSettings, SetStatus} from "./actions/actions";
import {Metric} from "./Tools/types";
import {SettingsState} from "./actions/types";

const store = createStore(appReducers);

// const client = new Client('localhost:8000');
const client = new Client();

client.setOnFullFFT((samples: number[]) => {
  store.dispatch(AddFullFFT(client.getServerCenterFrequency(), client.getServerSampleRate(), samples));
});

client.setOnSegFFT((samples: number[]) => {
  store.dispatch(AddSegmentFFT(client.getSegmentCenterFrequency(), client.getSegmentSampleRate(), samples));
});

client.setOnMetrics((metrics: Metric[]) => {
  store.dispatch(AddMetrics(metrics));
});

client.setOnSettings((settings: SettingsState) => {
  store.dispatch(SetSettings(settings));
});

client.setOnClose(() => {
  store.dispatch(SetStatus({
    wsConnected: false,
  }))
});

client.setOnOpen(() => {
  store.dispatch(SetStatus({
    wsConnected: true,
  }))
});

if (!client.start()) {
  alert('No websockets!');
}

ReactDOM.render((
  <Provider store={store}>
    <App/>
  </Provider>), document.getElementById('root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
