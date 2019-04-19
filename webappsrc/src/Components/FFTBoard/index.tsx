import {Component, default as React} from "react";
import PropTypes from "prop-types";
import {connect} from "react-redux";
import Typography from "@material-ui/core/Typography";
import Grid from "@material-ui/core/Grid";
import FFT from '../FFT';
import {FFTState, SettingsState} from "../../actions/types";
import {FFTInitialState} from "../../actions/initialStates";

type FFTBoardProps = {
  segFFT: FFTState,
  fullFFT: FFTState,
  settings: SettingsState,
}

type FFTBoardState = {
  segFFT: FFTState,
  fullFFT: FFTState,
  delta: number,
  lastUpdate: number,
  fps: number,
}

class FFTBoard extends Component<FFTBoardProps, FFTBoardState> {
  rafId?: number;

  state = {
    segFFT: FFTInitialState,
    fullFFT: FFTInitialState,
    delta: 0,
    lastUpdate: Date.now(),
    fps: 0,
  };

  componentDidMount() {
    this.rafId = requestAnimationFrame(this.tick);
  }

  tick = () => {
    const delta = Date.now() - this.state.lastUpdate;
    const lastUpdate = Date.now();
    this.setState({
      segFFT: this.props.segFFT,
      fullFFT: this.props.fullFFT,
      lastUpdate,
      delta,
      fps: Math.round(1000 / delta),
    });

    this.rafId = requestAnimationFrame(this.tick);
  };

  render() {
    const {settings} = this.props;
    const {segFFT, fullFFT} = settings;
    return (
      <div>
        <Typography variant="h2" component="h1">
          Server Spectrum
        </Typography>
        <Grid container>
          <Grid item xs>
            <Typography variant="h5" component="h3">
              Beacon Segment
            </Typography>
            <FFT
              samples={this.state.segFFT.samples}
              maxVal={segFFT.maxVal}
              range={segFFT.range}
              width={segFFT.width}
              height={segFFT.height}
              fps={this.state.fps}
              centerFrequency={this.props.segFFT.centerFrequency}
              sampleRate={this.props.segFFT.sampleRate}
            />
          </Grid>
          <Grid item xs>
            <Typography variant="h5" component="h3">
              Full
            </Typography>
            <FFT
              samples={this.state.fullFFT.samples}
              maxVal={fullFFT.maxVal}
              range={fullFFT.range}
              width={fullFFT.width}
              height={fullFFT.height}
              fps={this.state.fps}
              centerFrequency={this.props.fullFFT.centerFrequency}
              sampleRate={this.props.fullFFT.sampleRate}
            />
          </Grid>
        </Grid>
      </div>
    )
  }

  static propTypes = {
    samples: PropTypes.arrayOf(PropTypes.number),
  }
}

const mapStateToProps = (state: any) => {
  return ({
    segFFT: state.segmentFFT,
    fullFFT: state.fullFFT,
    settings: state.settings,
  });
};

// @ts-ignore
export default connect(mapStateToProps)(FFTBoard);
