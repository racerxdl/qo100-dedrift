import Grid from "@material-ui/core/Grid";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import * as React from "react";
import Typography from "@material-ui/core/Typography";
import C3Chart from "react-c3js";

import {Metric} from "../../Tools/types";
import {toHzNotation, toNotationUnit} from "../../Tools";
import {cardStyle} from "./styles";
import {gaugePresets} from "./presets";
import {toBytesNotation} from "../../Tools/format";


const MakeGauge = (metric: Metric) => {
  const m0 = metric.metrics[0];
  let fType = 'number';
  const help = (metric.help || '').toLocaleLowerCase();

  const name = metric.name || '_';
  let preset = gaugePresets['_'];
  if (gaugePresets[name]) {
    preset = gaugePresets[name];
  }

  if (help.indexOf('byte') > -1) {
    fType = 'byte';
  } else if (help.indexOf('hertz') > -1) {
    fType = 'hertz';
  }

  const formatFunc = (value: any) => {
    switch (fType) {
      case 'byte':
        return toBytesNotation(value);
      case 'hertz':
        return toHzNotation(value);
    }
    const o = toNotationUnit(value);
    return `${o[0]} ${o[1]}`
  };

  preset.label = preset.label || {};
  preset.format = preset.format || formatFunc;

  let val = m0.value;

  if (preset.preCompute) {
    const {units, value} = preset.preCompute(m0.value);
    val = value;
    preset.gauge.units = units;
  }

  const gaugeData = {
    data: {
      type: 'gauge',
      columns: [
        [metric.help, val],
      ],
    },
    ...preset,
  };

  return <Grid item xs key={metric.name}>
    <Card style={cardStyle}>
      <CardContent>
        <C3Chart {...gaugeData} />
      </CardContent>
    </Card>
  </Grid>
};

const MakeCounter = (metric: Metric) => {
  const m0 = metric.metrics[0];
  const help = (metric.help || '').toLocaleLowerCase();
  let val = m0.value;

  if (help.indexOf('byte') > -1) {
    val = toBytesNotation(val);
  } else if (help.indexOf('hertz') > -1) {
    val = toHzNotation(val);
  } else {
    const o = toNotationUnit(val);
    val = `${o[0]} ${o[1]}`;
  }

  return (
    <Grid item xs key={metric.name}>
      <Card style={cardStyle}>
        <CardContent>
          <Typography variant="h4" component="h2">
            {val}
          </Typography>
          <Typography>
            {metric.help}
          </Typography>
        </CardContent>
      </Card>
    </Grid>
  )
};

export {
  MakeGauge,
  MakeCounter
}
