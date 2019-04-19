import React, {Component} from 'react';
import 'c3/c3.css';

import './App.css';
import {connect} from "react-redux";
import {FFTBoard, MetricsBoard} from './Components';
import AppBar from "@material-ui/core/AppBar";
import Typography from "@material-ui/core/Typography";
import Toolbar from "@material-ui/core/Toolbar";

type AppProps = {
  serverName: string
  wsConnected: boolean
}
type AppState = {}

class App extends Component<AppProps, AppState> {
  render() {
    return (
      <div className="App">
        <AppBar position="static" color={this.props.wsConnected ? "primary" : "secondary"}>
          <Toolbar>
            <Typography variant="h6" color="inherit">
              {this.props.serverName}
            </Typography>
          </Toolbar>
        </AppBar>
        <br/>
        <FFTBoard/>
        <MetricsBoard/>
      </div>
    );
  }
}

const mapStateToProps = (state: any) => {
  return ({
    serverName: state.settings.name,
    wsConnected: state.status.wsConnected,
  });
};

export default connect(mapStateToProps)(App);
