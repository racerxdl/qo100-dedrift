import {toNotationUnit} from "../../Tools";

const gaugePresets: { [id: string]: any } = {
  '_': {
    gauge: {},
    color: {
      pattern: [
        '#FF0000',
        '#F97600',
        '#F6C600',
        '#60B044'
      ],
      threshold: {
        values: [0, 25, 50, 100],
      }
    },
  },
  'connections': {
    gauge: {
      units: '',
      label: {
        format: (value: number) => toNotationUnit(value)[0],
      },
    },
    color: {
      pattern: [
        '#60B044',
        '#F6C600',
        '#F97600',
        '#FF0000',
      ],
      threshold: {
        values: [15, 30, 60, 100],
      }
    },
  },
  'segment_center_frequency': {
    preCompute: (value: number) => {
      const o = toNotationUnit(value);
      return {
        units: ` ${o[1]}Hz`,
        value: o[0],
      };
    },
    gauge: {
      label: {
        format: (value: number) => toNotationUnit(value)[0],
      },
      units: '',
      min: 24,
      max: 1800,
    },
    color: {
      pattern: [
        '#FF0000',
        '#F97600',
        '#F6C600',
        '#60B044',
      ],
      threshold: {
        values: [0, 25, 50, 100],
      },
    },
  },
  'server_center_frequency': {
    preCompute: (value: number) => {
      const o = toNotationUnit(value);
      return {
        units: ` ${o[1]}Hz`,
        value: o[0],
      };
    },
    gauge: {
      label: {
        format: (value: number) => toNotationUnit(value)[0],
      },
      units: '',
      min: 24,
      max: 1800,
    },
    color: {
      pattern: [
        '#FF0000',
        '#F97600',
        '#F6C600',
        '#60B044',
      ],
      threshold: {
        unit: 'percentage',
        values: [0, 25, 50, 100],
      },
    },
  },
  'lock_offset': {
    preCompute: (value: number) => {
      return {
        value: Math.round(value * 100) / 100,
        units: 'Hz'
      };
    },
    gauge: {
      units: 'Hz',
      min: -10e6,
      max: +10e6,
      label: {
        format: (value: number) => `${value}`,
      },
    },
    color: {
      pattern: [
        '#60B044',
        '#F6C600',
        '#F97600',
        '#FF0000'
      ],
      threshold: {
        values: [0, 10, 50, 70, 100],
      }
    },
  },
};

export {
  gaugePresets,
}
