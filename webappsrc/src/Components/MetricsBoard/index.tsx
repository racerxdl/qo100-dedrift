import {connect} from "react-redux";
import {Component, default as React} from "react";
import Grid from "@material-ui/core/Grid";
import {Metric} from "../../Tools/types";
import Typography from "@material-ui/core/Typography";

import {
  gridStyle,
  divStyle,
} from "./styles";

import {
  MakeGauge,
  MakeCounter
} from './generators';

type MetricsBoardProps = {
  metrics: Metric[]
}

const itemGenerator: { [id: string]: (metric: Metric) => any } = {
  'COUNTER': MakeCounter,
  'GAUGE': MakeGauge,
};

const generatorOverride: { [id: string]: ((metric: Metric) => any) | null } = {
  '_': null,
  'server_samplerate': MakeCounter,
  'segment_samplerate': MakeCounter,
};

class MetricsBoard extends Component<MetricsBoardProps> {
  render() {
    const baseItems: any[] = [];
    let items: any[] = [];
    const {metrics} = this.props;

    const numItems = metrics.length;
    const maxColumns = 4;

    let c = 0;

    for (const t in itemGenerator) {
      const fT = itemGenerator[t];
      for (let i = 0; i < numItems; i++) {
        if (c % maxColumns == 0 && items.length > 0) {
          baseItems.push(
            <Grid container style={gridStyle} spacing={16} key={`${t}_${c}`}>
              {items}
            </Grid>
          );
          items = [];
        }
        const m = metrics[i];
        const ov = generatorOverride[m.name || '_'];
        if (ov) {
          if (ov == fT) {
            items.push(ov(m));
            c++;
          }
        } else if (m.type === t) {
          items.push(fT(m));
          c++;
        }
      }

      if (items.length > 0) {
        baseItems.push(
          <Grid container style={gridStyle} spacing={16} key={`${t}_${c}`}>
            {items}
          </Grid>
        );
        items = [];
      }
      c = 0;
    }

    return (
      <div style={divStyle}>
        <Typography variant="h2" component="h1">
          Metrics
        </Typography>
        <br/>
        {baseItems}
      </div>
    )
  }
}

const mapStateToProps = (state: any) => {
  return ({
    metrics: state.metrics.metrics,
  });
};

export default connect(mapStateToProps)(MetricsBoard);
