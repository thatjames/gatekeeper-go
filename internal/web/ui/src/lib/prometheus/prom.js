import { env } from "$lib/api/api";

class PrometheusMetricsService {
  constructor() {
    this.url = env.metrics;
  }

  /**
   * Fetches raw Prometheus metrics from the endpoint
   */
  async fetchMetrics() {
    const response = await fetch(`${this.url}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch metrics: ${response.statusText}`);
    }
    return await response.text();
  }

  /**
   * Parses Prometheus text format into structured data
   */
  parseMetrics(metricsText) {
    const lines = metricsText.split("\n");
    const metrics = {};
    let currentMetric = null;

    for (const line of lines) {
      // Skip empty lines and comments that aren't HELP or TYPE
      if (
        !line.trim() ||
        (line.startsWith("#") &&
          !line.startsWith("# HELP") &&
          !line.startsWith("# TYPE"))
      ) {
        continue;
      }

      // Parse HELP lines
      if (line.startsWith("# HELP")) {
        const match = line.match(/# HELP (\S+) (.+)/);
        if (match) {
          currentMetric = match[1];
          if (!metrics[currentMetric]) {
            metrics[currentMetric] = { help: "", type: "", values: [] };
          }
          metrics[currentMetric].help = match[2];
        }
        continue;
      }

      // Parse TYPE lines
      if (line.startsWith("# TYPE")) {
        const match = line.match(/# TYPE (\S+) (\S+)/);
        if (match) {
          currentMetric = match[1];
          if (!metrics[currentMetric]) {
            metrics[currentMetric] = { help: "", type: "", values: [] };
          }
          metrics[currentMetric].type = match[2];
        }
        continue;
      }

      // Parse metric values
      const metricMatch = line.match(
        /^([a-zA-Z_:][a-zA-Z0-9_:]*?)(\{.*?\})?\s+(.+?)(\s+\d+)?$/
      );
      if (metricMatch) {
        const metricName = metricMatch[1];
        const labels = metricMatch[2] ? this.parseLabels(metricMatch[2]) : {};
        const value = metricMatch[3];

        if (!metrics[metricName]) {
          metrics[metricName] = { help: "", type: "", values: [] };
        }

        metrics[metricName].values.push({
          labels,
          value: value === "NaN" ? NaN : parseFloat(value),
        });
      }
    }

    return metrics;
  }

  /**
   * Parses label string like {le="100",type="test"} into object
   */
  parseLabels(labelString) {
    const labels = {};
    const labelRegex = /(\w+)="([^"]*)"/g;
    let match;

    while ((match = labelRegex.exec(labelString)) !== null) {
      labels[match[1]] = match[2];
    }

    return labels;
  }

  /**
   * Gets a histogram metric
   * @param {string} metricName - Base name of the histogram (without _bucket, _sum, _count suffixes)
   * @returns {Object} Histogram data with buckets, sum, count, and average
   */
  async getHistogram(metricName) {
    const metricsText = await this.fetchMetrics();
    const allMetrics = this.parseMetrics(metricsText);

    const bucketMetric = allMetrics[`${metricName}_bucket`] || { values: [] };
    const sumMetric = allMetrics[`${metricName}_sum`]?.values[0]?.value || 0;
    const countMetric =
      allMetrics[`${metricName}_count`]?.values[0]?.value || 0;

    const buckets = bucketMetric.values
      .filter((v) => v.labels.le !== "+Inf")
      .map((v) => ({
        upperBound: parseFloat(v.labels.le),
        count: v.value,
      }))
      .sort((a, b) => a.upperBound - b.upperBound);

    const average = countMetric > 0 ? sumMetric / countMetric : 0;

    return {
      buckets,
      sum: sumMetric,
      count: countMetric,
      average,
    };
  }

  /**
   * Gets a gauge metric
   * @param {string} metricName - Name of the gauge metric
   * @returns {number} Current gauge value
   */
  async getGauge(metricName) {
    const metricsText = await this.fetchMetrics();
    const allMetrics = this.parseMetrics(metricsText);

    const metric = allMetrics[metricName];
    if (!metric || metric.values.length === 0) {
      return 0;
    }

    return metric.values[0].value;
  }

  /**
   * Gets a counter metric
   * @param {string} metricName - Name of the counter metric
   * @returns {number} Current counter value
   */
  async getCounter(metricName) {
    const metricsText = await this.fetchMetrics();
    const allMetrics = this.parseMetrics(metricsText);

    const metric = allMetrics[metricName];
    if (!metric || metric.values.length === 0) {
      return 0;
    }

    return metric.values[0].value;
  }

  /**
   * Gets a counter vector metric (counter with labels)
   * @param {string} metricName - Name of the counter vector metric
   * @returns {Array} Array of {labels, value} objects
   */
  async getCounterVec(metricName) {
    const metricsText = await this.fetchMetrics();
    const allMetrics = this.parseMetrics(metricsText);

    const metric = allMetrics[metricName];
    if (!metric || metric.values.length === 0) {
      return [];
    }

    return metric.values.map((v) => ({
      labels: v.labels,
      value: v.value,
    }));
  }

  /**
   * Generic method to get any metric by name
   * @param {string} metricName - Name of the metric
   * @returns {Object} Raw metric data including type, help, and values
   */
  async getMetric(metricName) {
    const metricsText = await this.fetchMetrics();
    const allMetrics = this.parseMetrics(metricsText);

    return allMetrics[metricName] || null;
  }

  /**
   * Gets multiple metrics at once
   * @param {Array<string>} metricNames - Array of metric names
   * @returns {Object} Object with metric names as keys and their data as values
   */
  async getMetrics(metricNames) {
    const metricsText = await this.fetchMetrics();
    const allMetrics = this.parseMetrics(metricsText);

    const result = {};
    for (const name of metricNames) {
      result[name] = allMetrics[name] || null;
    }

    return result;
  }
}

export default PrometheusMetricsService;
