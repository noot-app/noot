<script lang="ts">
  import { Chart } from "svelte-echarts"
  import { init, use } from "echarts/core"
  import { ScatterChart } from "echarts/charts"
  import {
    TitleComponent,
    TooltipComponent,
    LegendComponent,
    GridComponent
  } from "echarts/components"
  import { CanvasRenderer } from "echarts/renderers"

  // Initialize ECharts
  use([
    ScatterChart,
    TitleComponent,
    TooltipComponent,
    LegendComponent,
    GridComponent,
    CanvasRenderer
  ])

  export let consumptions: any[] = []
  export let events: any[] = []

  // Process correlation data
  $: correlationData = processCorrelationData(consumptions, events)

  function processCorrelationData(consumptionData: any[], eventData: any[]) {
    if (!consumptionData || !eventData || consumptionData.length === 0 || eventData.length === 0) {
      return { labels: [], correlations: [] }
    }

    // Extract all unique labels from consumptions
    const labelMap = new Map()
    
    consumptionData.forEach(consumption => {
      const date = new Date(consumption.created_at).toDateString()
      consumption.labels?.forEach((label: any) => {
        if (!labelMap.has(label.name)) {
          labelMap.set(label.name, { dates: new Set(), count: 0 })
        }
        labelMap.get(label.name).dates.add(date)
        labelMap.get(label.name).count++
      })
    })

    // Count events by date
    const eventsByDate = new Map()
    eventData.forEach(event => {
      const date = new Date(event.started_at).toDateString()
      eventsByDate.set(date, (eventsByDate.get(date) || 0) + 1)
    })

    // Calculate correlations
    const correlations = []
    for (const [labelName, labelData] of labelMap.entries()) {
      let correlatedEvents = 0
      let totalDaysWithLabel = labelData.dates.size

      // Count events on days with this label
      for (const date of labelData.dates) {
        if (eventsByDate.has(date)) {
          correlatedEvents += eventsByDate.get(date)
        }
      }

      const correlationRate = totalDaysWithLabel > 0 ? (correlatedEvents / totalDaysWithLabel) : 0
      
      if (totalDaysWithLabel >= 2) { // Only include labels with sufficient data
        correlations.push({
          label: labelName,
          frequency: labelData.count,
          eventRate: correlationRate,
          daysWithLabel: totalDaysWithLabel
        })
      }
    }

    // Sort by correlation strength
    correlations.sort((a, b) => b.eventRate - a.eventRate)

    return {
      labels: correlations.map(c => c.label),
      correlations: correlations.map(c => [c.frequency, c.eventRate, c.daysWithLabel])
    }
  }

  $: options = {
    backgroundColor: 'transparent',
    title: {
      text: 'Label-Event Correlations',
      textStyle: {
        color: '#3c3d42',
        fontSize: 16,
        fontWeight: 'normal' as const
      },
      left: 'left'
    },
    tooltip: {
      trigger: 'item' as const,
      backgroundColor: '#faf9f5',
      borderColor: '#dbd9d5',
      textStyle: {
        color: '#3c3d42'
      },
      formatter: function(params: any) {
        const label = correlationData.labels[params.dataIndex]
        const data = params.data
        return `${label}<br/>
                Frequency: ${data[0]} times<br/>
                Event Rate: ${data[1].toFixed(2)} events/day<br/>
                Days with Label: ${data[2]}`
      }
    },
    grid: {
      left: '10%',
      right: '10%',
      bottom: '15%',
      top: '20%'
    },
    xAxis: {
      type: 'value' as const,
      name: 'Label Frequency',
      nameLocation: 'middle' as const,
      nameGap: 25,
      axisLine: {
        lineStyle: {
          color: '#dbd9d5'
        }
      },
      axisLabel: {
        color: '#3c3d42'
      },
      splitLine: {
        lineStyle: {
          color: '#f0efeb'
        }
      }
    },
    yAxis: {
      type: 'value' as const,
      name: 'Events per Day',
      nameLocation: 'middle' as const,
      nameGap: 40,
      axisLine: {
        lineStyle: {
          color: '#dbd9d5'
        }
      },
      axisLabel: {
        color: '#3c3d42'
      },
      splitLine: {
        lineStyle: {
          color: '#f0efeb'
        }
      }
    },
    series: [
      {
        type: 'scatter' as const,
        data: correlationData.correlations,
        symbolSize: function(data: any) {
          return Math.max(8, Math.min(20, data[2] * 2)) // Size by days with label
        },
        itemStyle: {
          color: '#be7454',
          opacity: 0.7
        },
        emphasis: {
          itemStyle: {
            color: '#be7454',
            opacity: 1,
            borderColor: '#3c3d42',
            borderWidth: 1
          }
        }
      }
    ]
  }
</script>

<div class="w-full h-48 sm:h-56 lg:h-64">
  {#if correlationData.labels.length > 0}
    <Chart {init} {options} />
  {:else}
    <div class="h-full flex items-center justify-center text-base-content-lighter">
      <div class="text-center">
        <svg class="w-12 h-12 mx-auto mb-2 opacity-50" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M6 2a2 2 0 00-2 2v12a2 2 0 002 2h8a2 2 0 002-2V4a2 2 0 00-2-2H6zm1 2a1 1 0 000 2h6a1 1 0 100-2H7zm6 7a1 1 0 011 1v3a1 1 0 11-2 0v-3a1 1 0 011-1zm-3 3a1 1 0 100 2h.01a1 1 0 100-2H10zm-4 1a1 1 0 011-1h.01a1 1 0 110 2H7a1 1 0 01-1-1zm1-4a1 1 0 100 2h.01a1 1 0 100-2H7zm2 1a1 1 0 011-1h.01a1 1 0 110 2H10a1 1 0 01-1-1zm4-4a1 1 0 100 2h.01a1 1 0 100-2H13zM9 9a1 1 0 011-1h.01a1 1 0 110 2H10a1 1 0 01-1-1zM7 8a1 1 0 000 2h.01a1 1 0 000-2H7z" clip-rule="evenodd" />
        </svg>
        <p class="text-sm">Not enough data for correlations</p>
        <p class="text-xs opacity-75">Add more labeled meals and events</p>
      </div>
    </div>
  {/if}
</div>