// Based on https://github.com/yunyu/parse-prometheus-text-format
// Ported to Typescript

import {ShallowObjectEquals, UnescapeHelp} from "./parsers";
import {Metric} from "./types";

const ERR_MSG = 'Invalid line: ';
const SUMMARY_TYPE = 'SUMMARY';
const HISTOGRAM_TYPE = 'HISTOGRAM';

function ParseMetrics(metrics: string) {
  const lines = metrics.split('\n');
  const converted = [];

  let metric, help, type, samples = [];

  for (let i = 0; i < lines.length; ++i) {
    const line = lines[i].trim();
    let lineMetric = null, lineHelp = null, lineType = null, lineSample = null;
    if (line.length === 0) {
      // ignore blank lines
    } else if (line.startsWith('# ')) { // process metadata lines
      let lineData = line.substring(2);
      let instr = null;
      if (lineData.startsWith('HELP ')) {
        instr = 1;
      } else if (lineData.startsWith('TYPE ')) {
        instr = 2;
      }
      if (instr) {
        lineData = lineData.substring(5);
        const spaceIndex = lineData.indexOf(' ');
        if (spaceIndex !== -1) { // expect another token
          lineMetric = lineData.substring(0, spaceIndex);
          const remain = lineData.substring(spaceIndex + 1);
          if (instr === 1) { // HELP
            lineHelp = UnescapeHelp(remain); // remain could be empty
          } else { // TYPE
            if (remain.indexOf(' ') !== -1) {
              throw ERR_MSG + line;
            }
            lineType = remain.toUpperCase();
          }
        } else {
          throw ERR_MSG + line;
        }
      }
      // 100% pure comment line, ignore
    } else { // process sample lines
      lineSample = parseSampleLine(line);
      lineMetric = lineSample.name;
    }

    if (lineMetric === metric) { // metadata always has same name
      if (!help && lineHelp) {
        help = lineHelp;
      } else if (!type && lineType) {
        type = lineType;
      }
    }

    // different types allow different suffixes
    const suffixedCount = metric + '_count';
    const suffixedSum = metric + '_sum';
    const suffixedBucket = metric + '_bucket';
    const allowedNames = [metric];
    if (type === SUMMARY_TYPE || type === HISTOGRAM_TYPE) {
      allowedNames.push(suffixedCount);
      allowedNames.push(suffixedSum);
    }
    if (type === HISTOGRAM_TYPE) {
      allowedNames.push(suffixedBucket);
    }

    // encountered new metric family or end of input
    if (i + 1 === lines.length || (lineMetric && allowedNames.indexOf(lineMetric) === -1)) {
      // write current
      if (metric) {
        if (type === SUMMARY_TYPE) {
          samples = flattenMetrics(samples, 'quantiles', 'quantile', 'value');
        } else if (type === HISTOGRAM_TYPE) {
          samples = flattenMetrics(samples, 'buckets', 'le', 'bucket');
        }
        converted.push({
          name: metric,
          help: help ? help : '',
          type: type ? type : 'UNTYPED',
          metrics: samples
        });
      }
      // reset for new metric family
      metric = lineMetric;
      help = lineHelp ? lineHelp : null;
      type = lineType ? lineType : null;
      samples = [];
    }
    if (lineSample) {
      // key is not called value in official implementation if suffixed count, sum, or bucket
      if (lineSample.name !== metric) {
        if (type === SUMMARY_TYPE || type === HISTOGRAM_TYPE) {
          if (lineSample.name === suffixedCount) {
            lineSample.count = lineSample.value ? parseFloat(`${lineSample.value}`) : 0;
          } else if (lineSample.name === suffixedSum) {
            lineSample.sum = lineSample.value ? parseFloat(`${lineSample.value}`) : 0;
          }
        }
        if (type === HISTOGRAM_TYPE && lineSample.name === suffixedBucket) {
          lineSample.bucket = lineSample.value;
        }
        delete lineSample.value;
      }
      delete lineSample.name;
      // merge into existing sample if labels are deep equal
      const samplesLen = samples.length;
      const lastSample = samplesLen === 0 ? null : samples[samplesLen - 1];
      if (lastSample && ShallowObjectEquals(lineSample.labels, lastSample.labels)) {
        delete lineSample.labels;
        for (const key in lineSample) {
          lastSample[key] = lineSample[key];
        }
      } else {
        samples.push(lineSample);
      }
    }
  }

  return converted;
}

function flattenMetrics(metrics: Metric[], groupName: string, keyName: string, valueName: string): Metric[] {
  let flattened: Metric | null = null;
  for (let i = 0; i < metrics.length; ++i) {
    const sample = metrics[i];
    if (sample.labels && sample.labels[keyName] && sample[valueName]) {
      if (!flattened) {
        flattened = {};
        flattened[groupName] = {};
      }
      flattened[groupName][sample.labels[keyName]] = sample[valueName];
    } else if (!sample.labels && flattened) {
      if (sample.count !== undefined) {
        flattened.count = sample.count;
      }
      if (sample.sum !== undefined) {
        flattened.sum = sample.sum;
      }
    }
  }
  if (flattened) {
    return [flattened];
  } else {
    return metrics;
  }
}

function parseSampleLine(line: string): Metric {
  const STATE_NAME = 0;
  const STATE_STARTOFLABELNAME = 1;
  const STATE_ENDOFNAME = 2;
  const STATE_VALUE = 3;
  const STATE_ENDOFLABELS = 4;
  const STATE_LABELNAME = 5;
  const STATE_LABELVALUEQUOTE = 6;
  const STATE_LABELVALUEEQUALS = 7;
  const STATE_LABELVALUE = 8;
  const STATE_LABELVALUESLASH = 9;
  const STATE_NEXTLABEL = 10;

  let name = '', labelname = '', labelvalue = '', value = '', labels: { [id: string]: string } | void = undefined;
  let state = STATE_NAME;

  for (let c = 0; c < line.length; ++c) {
    const char = line.charAt(c);
    if (state === STATE_NAME) {
      if (char === '{') {
        state = STATE_STARTOFLABELNAME;
      } else if (char === ' ' || char === '\t') {
        state = STATE_ENDOFNAME;
      } else {
        name += char;
      }
    } else if (state === STATE_ENDOFNAME) {
      if (char === ' ' || char === '\t') {
        // do nothing
      } else if (char === '{') {
        state = STATE_STARTOFLABELNAME;
      } else {
        value += char;
        state = STATE_VALUE;
      }
    } else if (state === STATE_STARTOFLABELNAME) {
      if (char === ' ' || char === '\t') {
        // do nothing
      } else if (char === '}') {
        state = STATE_ENDOFLABELS;
      } else {
        labelname += char;
        state = STATE_LABELNAME;
      }
    } else if (state === STATE_LABELNAME) {
      if (char === '=') {
        state = STATE_LABELVALUEQUOTE;
      } else if (char === '}') {
        state = STATE_ENDOFLABELS;
      } else if (char === ' ' || char === '\t') {
        state = STATE_LABELVALUEEQUALS;
      } else {
        labelname += char;
      }
    } else if (state === STATE_LABELVALUEEQUALS) {
      if (char === '=') {
        state = STATE_LABELVALUEQUOTE;
      } else if (char === ' ' || char === '\t') {
        // do nothing
      } else {
        throw ERR_MSG + line;
      }
    } else if (state === STATE_LABELVALUEQUOTE) {
      if (char === '"') {
        state = STATE_LABELVALUE;
      } else if (char === ' ' || char === '\t') {
        // do nothing
      } else {
        throw ERR_MSG + line;
      }
    } else if (state === STATE_LABELVALUE) {
      if (char === '\\') {
        state = STATE_LABELVALUESLASH;
      } else if (char === '"') {
        if (!labels) {
          labels = {};
        }
        labels[labelname] = labelvalue;
        labelname = '';
        labelvalue = '';
        state = STATE_NEXTLABEL;
      } else {
        labelvalue += char;
      }
    } else if (state === STATE_LABELVALUESLASH) {
      state = STATE_LABELVALUE;
      if (char === '\\') {
        labelvalue += '\\';
      } else if (char === 'n') {
        labelvalue += '\n';
      } else if (char === '"') {
        labelvalue += '"';
      } else {
        labelvalue += ('\\' + char);
      }
    } else if (state === STATE_NEXTLABEL) {
      if (char === ',') {
        state = STATE_LABELNAME;
      } else if (char === '}') {
        state = STATE_ENDOFLABELS;
      } else if (char === ' ' || char === '\t') {
        // do nothing
      } else {
        throw ERR_MSG + line;
      }
    } else if (state === STATE_ENDOFLABELS) {
      if (char === ' ' || char === '\t') {
        // do nothing
      } else {
        value += char;
        state = STATE_VALUE;
      }
    } else if (state === STATE_VALUE) {
      if (char === ' ' || char === '\t') {
        break; // timestamps are NOT supported - ignoring
      } else {
        value += char;
      }
    }
  }

  return {
    name: name,
    value: value,
    labels: labels
  };
}

export {
  ParseMetrics,
}
