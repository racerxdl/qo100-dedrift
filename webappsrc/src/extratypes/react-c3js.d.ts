type C3ChartProps = {
  data: any,
  title?: any,
  size?: any,
  padding?: any,
  color?: any,
  interaction?: any,
  transition?: any,
  oninit?: (d: any, i: any) => boolean | void,
  onrendered?: (d: any, i: any) => boolean | void,
  onmouseover?: (d: any, i: any) => boolean | void,
  onmouseout?: (d: any, i: any) => boolean | void,
  onresize?: (d: any, i: any) => boolean | void,
  onresized?: (d: any, i: any) => boolean | void,
  onclick?: (d: any, i: any) => boolean | void,
  axis?: any,
  grid?: any,
  regions?: any | void | null[],
  legend?: any,
  tooltip?: any,
  subchart?: any,
  zoom?: any,
  point?: any,
  line?: any,
  area?: any,
  bar?: any,
  pie?: any,
  donut?: any,
  gauge?: any,
  className?: string,
  style?: any,
  unloadBeforeLoad?: boolean,
}

declare module 'react-c3js' {
  import {Component} from "react";

  class C3Chart extends Component<C3ChartProps> {}

  export = C3Chart;
}
// export default Component;
