<script>
  import { onMount } from "svelte";
  import { Heading, Card, Spinner } from "flowbite-svelte";
  import ApexChart from "$components/ApexChart.svelte";
  import PrometheusMetricsService from "$lib/prometheus/prom.js";

  let loading = $state(true);
  let error = $state(null);
  let dnsRequestTime = $state(null);
  let queryCounters = $state([]);
  let cacheHitCounters = $state([]);
  let blockedDomainCounters = $state([]);
  let refreshInterval = null;

  const PROMETHEUS_URL = "http://localhost:8085"; // Update with your Prometheus URL
  const REFRESH_RATE = 30000; // Refresh every 30 seconds

  const metricsService = new PrometheusMetricsService(PROMETHEUS_URL);

  const fetchDNSMetrics = async () => {
    try {
      loading = true;
      error = null;

      // Fetch histogram for DNS request time
      dnsRequestTime = await metricsService.getHistogram("dns_req_time");

      // Fetch counter vectors
      queryCounters = await metricsService.getCounterVec("dns_query_counter");
      cacheHitCounters = await metricsService.getCounterVec(
        "dns_cache_hit_counter",
      );
      blockedDomainCounters = await metricsService.getCounterVec(
        "dns_blocked_domain_counter",
      );
    } catch (err) {
      error = err.message;
      console.error("Failed to fetch DNS metrics:", err);
    } finally {
      loading = false;
    }
  };

  onMount(() => {
    fetchDNSMetrics();

    // Set up automatic refresh
    refreshInterval = setInterval(fetchDNSMetrics, REFRESH_RATE);

    return () => {
      if (refreshInterval) {
        clearInterval(refreshInterval);
      }
    };
  });

  // Prepare histogram chart data
  const histogramChartData = $derived.by(() => {
    if (!dnsRequestTime?.buckets) {
      return { series: [], categories: [] };
    }

    const buckets = dnsRequestTime.buckets;

    // Calculate incremental counts for each bucket
    const incrementalCounts = [];
    for (let i = 0; i < buckets.length; i++) {
      const current = buckets[i].count;
      const previous = i > 0 ? buckets[i - 1].count : 0;
      incrementalCounts.push(current - previous);
    }

    return {
      series: [
        {
          name: "Request Count",
          data: incrementalCounts,
        },
      ],
      categories: buckets.map((b) => `≤${b.upperBound}ms`),
    };
  });

  // Aggregate query counters by domain (summing across upstreams and results)
  const queryByDomain = $derived.by(() => {
    if (!queryCounters || queryCounters.length === 0) return [];

    const domainMap = new Map();

    for (const counter of queryCounters) {
      const domain = counter.labels.domain || "unknown";
      const current = domainMap.get(domain) || 0;
      domainMap.set(domain, current + counter.value);
    }

    return Array.from(domainMap.entries())
      .map(([domain, count]) => ({ domain, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10); // Top 10 domains
  });

  const countByUpstream = $derived.by(() => {
    if (!queryCounters || queryCounters.length === 0) return [];

    const upstreamMap = new Map();
    for (const counter of queryCounters) {
      const upstream = counter.labels.upstream || "unknown";
      const current = upstreamMap.get(upstream) || 0;
      upstreamMap.set(upstream, current + counter.value);
    }

    return Array.from(upstreamMap.entries()).map(([upstream, count]) => ({
      upstream,
      count,
    }));
  });

  // Aggregate blocked domains
  const blockedDomains = $derived.by(() => {
    if (!blockedDomainCounters || blockedDomainCounters.length === 0) return [];

    return blockedDomainCounters
      .map((counter) => ({
        domain: counter.labels.domain || "unknown",
        count: counter.value,
      }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10); // Top 10 blocked domains
  });

  // Prepare top domains chart
  const topDomainsChartData = $derived.by(() => {
    if (!queryByDomain || queryByDomain.length === 0) {
      return { series: [], categories: [] };
    }

    return {
      series: [
        {
          name: "Queries",
          data: queryByDomain.map((d) => d.count),
        },
      ],
      categories: queryByDomain.map((d) => d.domain),
    };
  });

  // Prepare blocked domains chart
  const blockedDomainsChartData = $derived.by(() => {
    if (!blockedDomains || blockedDomains.length === 0) {
      return { series: [], categories: [] };
    }

    return {
      series: [
        {
          name: "Blocked Requests",
          data: blockedDomains.map((d) => d.count),
        },
      ],
      categories: blockedDomains.map((d) => d.domain),
    };
  });

  const totalCacheHits = $derived(
    cacheHitCounters.reduce((sum, counter) => sum + counter.value, 0),
  );

  const totalBlockedRequests = $derived(
    blockedDomainCounters.reduce((sum, counter) => sum + counter.value, 0),
  );

  const cacheHitRate = $derived(
    dnsRequestTime?.count > 0
      ? ((totalCacheHits / dnsRequestTime.count) * 100).toFixed(1)
      : "0.0",
  );

  const upstreamRate = $derived.by(() => {
    let data = {
      series: [],
      labels: [],
    };
    if (!countByUpstream || countByUpstream.length === 0) return data;

    for (const counter of countByUpstream) {
      data.series.push(counter.count);
      data.labels.push(counter.upstream);
    }
    return data;
  });
</script>

<div class="flex flex-col gap-5">
  <Heading tag="h3">DNS Statistics</Heading>

  {#if loading && !dnsRequestTime}
    <div class="flex items-center justify-center p-12">
      <Spinner size="12" />
    </div>
  {:else if error}
    <Card
      class="bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800"
    >
      <p class="text-red-800 dark:text-red-200">
        Error loading DNS metrics: {error}
      </p>
    </Card>
  {:else if dnsRequestTime?.count > 0}
    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <!-- Total Requests -->
      <Card class="p-4">
        <div class="space-y-2">
          <p class="text-sm text-gray-500 dark:text-gray-400">Total Requests</p>
          <p class="text-3xl font-bold text-gray-900 dark:text-white">
            {dnsRequestTime.count.toLocaleString()}
          </p>
        </div>
      </Card>

      <!-- Average Response Time -->
      <Card class="p-4">
        <div class="space-y-2">
          <p class="text-sm text-gray-500 dark:text-gray-400">
            Avg Response Time
          </p>
          <p class="text-3xl font-bold text-gray-900 dark:text-white">
            {dnsRequestTime.average.toFixed(2)} ms
          </p>
        </div>
      </Card>

      <!-- Cache Hit Rate -->
      <Card class="p-4">
        <div class="space-y-2">
          <p class="text-sm text-gray-500 dark:text-gray-400">Cache Hit Rate</p>
          <p class="text-3xl font-bold text-gray-900 dark:text-white">
            {cacheHitRate}%
          </p>
        </div>
      </Card>

      <!-- Blocked Requests -->
      <Card class="p-4">
        <div class="space-y-2">
          <p class="text-sm text-gray-500 dark:text-gray-400">
            Blocked Requests
          </p>
          <p class="text-3xl font-bold text-red-600 dark:text-red-400">
            {totalBlockedRequests.toLocaleString()}
          </p>
        </div>
      </Card>
    </div>

    <!-- Request Time Distribution Chart -->
    <div
      class="flex flex-col gap-2 p-4 border shadow-lg rounded-lg border-gray-200 dark:border-gray-700 dark:bg-gray-800"
    >
      <div class="space-y-4">
        <ApexChart
          type="bar"
          series={histogramChartData.series}
          options={{
            plotOptions: {
              bar: {
                borderRadius: 4,
              },
            },
            dataLabels: {
              enabled: false,
            },
            xaxis: {
              categories: histogramChartData.categories,
              labels: {
                rotate: -45,
              },
            },
            title: "DNS Request Time Distribution",
            yaxis: {
              title: {
                text: "Number of Requests",
              },
            },
            colors: ["#3b82f6"],
            tooltip: {
              y: {
                formatter: (value) => `${value} requests`,
              },
            },
          }}
          height={350}
        />
      </div>
    </div>

    <!-- Top Queried Domains Chart -->
    <div
      class="flex flex-col gap-2 p-4 border shadow-lg rounded-lg border-gray-200 dark:border-gray-700 dark:bg-gray-800"
    >
      <div class="space-y-4">
        <ApexChart
          type="bar"
          series={topDomainsChartData.series}
          labels={topDomainsChartData.labels}
          options={{
            plotOptions: {
              bar: {
                borderRadius: 4,
                horizontal: true,
              },
            },
            dataLabels: {
              enabled: false,
            },
            xaxis: {
              categories: topDomainsChartData.categories,
            },
            yaxis: {
              labels: {
                maxWidth: 200,
              },
            },
            title: "Top Queried Domains",
            colors: ["#10b981"],
            tooltip: {
              y: {
                formatter: (value) => `${value} queries`,
              },
            },
          }}
          height={400}
        />
      </div>
    </div>

    <!-- Two column grid for pie charts -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <!-- Queries by Upstream Chart -->
      <div
        class="flex flex-col gap-2 p-4 border shadow-lg rounded-lg border-gray-200 dark:border-gray-700 dark:bg-gray-800"
      >
        <div class="space-y-4">
          <ApexChart
            type="donut"
            series={upstreamRate.series}
            options={{
              title: "Queries by Upstream",
              labels: upstreamRate.labels,
              legend: {
                position: "bottom",
              },
            }}
            height={400}
          />
        </div>
      </div>

      <!-- Blocked vs Allowed Chart -->
      <div
        class="flex flex-col gap-2 p-4 border shadow-lg rounded-lg border-gray-200 dark:border-gray-700 dark:bg-gray-800"
      >
        <div class="space-y-4">
          <ApexChart
            type="donut"
            series={[
              dnsRequestTime.count - totalBlockedRequests,
              totalBlockedRequests,
            ]}
            options={{
              title: "Allowed vs Blocked Requests",
              labels: ["Allowed", "Blocked"],
              colors: ["#10b981", "#ef4444"],
              legend: {
                position: "bottom",
              },
              tooltip: {
                y: {
                  formatter: (value) => `${value.toLocaleString()} requests`,
                },
              },
            }}
            height={400}
          />
        </div>
      </div>
    </div>

    <!-- Blocked Domains Chart -->
    {#if blockedDomains.length > 0}
      <div
        class="flex flex-col gap-2 p-4 border shadow-lg rounded-lg border-gray-200 dark:border-gray-700 dark:bg-gray-800"
      >
        <div class="space-y-4">
          <ApexChart
            type="bar"
            series={blockedDomainsChartData.series}
            options={{
              plotOptions: {
                bar: {
                  borderRadius: 4,
                  horizontal: true,
                },
              },
              dataLabels: {
                enabled: false,
              },
              xaxis: {
                categories: blockedDomainsChartData.categories,
              },
              title: "Top Blocked Domains",
              colors: ["#ef4444"],
              tooltip: {
                y: {
                  formatter: (value) => `${value} blocked requests`,
                },
              },
            }}
            height={400}
          />
        </div>
      </div>
    {/if}
  {:else}
    <p class="text-gray-500 dark:text-gray-400 text-center py-8">
      No DNS request data available yet
    </p>
  {/if}
</div>
