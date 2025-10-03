<script>
  import { onMount } from "svelte";
  import ApexCharts from "apexcharts";

  let chartElement = $state(null);
  let chart = $state(null);
  let isDarkMode = $state(false);

  let { options = {}, series = [], type = "line", height = 350 } = $props();

  // Detect dark mode
  function checkDarkMode() {
    isDarkMode = document.documentElement.classList.contains("dark");
  }

  // Get theme-aware colors
  const getThemedOptions = $derived.by(() => {
    const textColor = isDarkMode ? "#e5e7eb" : "#374151";
    const gridColor = isDarkMode ? "#4b5563" : "#e5e7eb";
    const themeMode = isDarkMode ? "dark" : "light";

    return {
      ...options,
      chart: {
        ...options.chart,
        type,
        height,
        animations: {
          enabled: true,
          easing: "easeinout",
          speed: 400,
          ...options.chart?.animations,
        },
        foreColor: textColor,
        background: "transparent",
      },
      title: {
        text: options.title,
        style: {
          fontSize: "22px",
        },
      },
      theme: {
        mode: themeMode,
        ...options.theme,
      },
      grid: {
        borderColor: gridColor,
        ...options.grid,
      },
      xaxis: {
        ...options.xaxis,
        labels: {
          ...options.xaxis?.labels,
          style: {
            colors: textColor,
            ...options.xaxis?.labels?.style,
          },
        },
      },
      yaxis: {
        ...options.yaxis,
        labels: {
          ...options.yaxis?.labels,
          style: {
            colors: textColor,
            ...options.yaxis?.labels?.style,
          },
        },
        title: {
          ...options.yaxis?.title,
          style: {
            color: textColor,
            ...options.yaxis?.title?.style,
          },
        },
      },
      legend: {
        ...options.legend,
        labels: {
          colors: textColor,
          ...options.legend?.labels,
        },
      },
      tooltip: {
        theme: themeMode,
        ...options.tooltip,
      },
      labels: options.labels || [],
      series,
    };
  });

  onMount(() => {
    checkDarkMode();

    // Watch for theme changes
    const observer = new MutationObserver(() => {
      checkDarkMode();
    });

    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["class"],
    });

    if (chartElement) {
      setTimeout(() => {
        chart = new ApexCharts(chartElement, getThemedOptions);
        chart.render();
      }, 10);
    }

    return () => {
      observer.disconnect();
      if (chart) {
        chart.destroy();
      }
    };
  });

  // Update chart when theme or data changes
  $effect(() => {
    if (chart) {
      chart.updateOptions(getThemedOptions);
    }
  });
</script>

<div bind:this={chartElement}></div>
