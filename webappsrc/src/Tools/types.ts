type MetricType = {
  name?: string,
  value?: string | number | void,
  labels?: { [id: string]: string } | void
  count?: number;
  sum?: number;
  [id: string]: any
  help?: string
}

export type Metric = MetricType;
