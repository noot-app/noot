<script lang="ts">
  import { Chart } from "svelte-echarts"
  import { init, use } from "echarts/core"
  import { LineChart } from "echarts/charts"
  import {
    TitleComponent,
    TooltipComponent,
    LegendComponent,
    GridComponent,
    MarkLineComponent
  } from "echarts/components"
  import { CanvasRenderer } from "echarts/renderers"

  // Initialize ECharts
  use([
    LineChart,
    TitleComponent,
    TooltipComponent,
    LegendComponent,
    GridComponent,
    MarkLineComponent,
    CanvasRenderer
  ])

  export let nutrientKey: string = ""
  export let nutrientName: string = ""
  export let unit: string = ""
  export let consumptions: any[] = []
  export let goalsData: any = null
  export let timeWindow: "1d" | "7d" | "30d" = "7d"

  // Calculate time-series data for the selected nutrient
  $: chartData = calculateNutrientTimeSeries(consumptions, nutrientKey, timeWindow, goalsData)

  function calculateNutrientTimeSeries(data: any[], nutrient: string, window: string, goals: any) {
    if (!data || data.length === 0) {
      return { dates: [], values: [], target: null, upperLimit: null }
    }

    const days = parseInt(window.replace('d', ''))
    
    // Generate date labels for the past N days
    const dateLabels = []
    const today = new Date()
    for (let i = days - 1; i >= 0; i--) {
      const date = new Date(today)
      date.setDate(date.getDate() - i)
      dateLabels.push(date.toISOString().split('T')[0])
    }

    // Group consumptions by date
    const consumptionsByDate = data.reduce((acc: any, consumption: any) => {
      const date = new Date(consumption.consumed_at || consumption.created_at).toISOString().split('T')[0]
      if (!acc[date]) acc[date] = []
      acc[date].push(consumption)
      return acc
    }, {})

    // Calculate daily values for the nutrient
    const values = dateLabels.map(date => {
      const dayConsumptions = consumptionsByDate[date] || []
      return dayConsumptions.reduce((total: number, consumption: any) => {
        const value = consumption.summary?.totals?.[nutrient] || 0
        return total + value
      }, 0)
    })

    // Get goal values
    const target = goals?.targets?.[nutrient] || null
    const upperLimit = goals?.upper_limits?.[nutrient] || null

    return {
      dates: dateLabels,
      values,
      target,
      upperLimit
    }
  }

  $: options = {
    backgroundColor: 'transparent',
    title: {
      text: `${nutrientName} Intake`,
      textStyle: {
        color: '#3c3d42',
        fontSize: 16,
        fontWeight: 'normal' as const
      },
      left: 'left'
    },
    tooltip: {
      trigger: 'axis' as const,
      backgroundColor: '#faf9f5',
      borderColor: '#dbd9d5',
      textStyle: {
        color: '#3c3d42'
      },
      formatter: function(params: any) {
        const date = params[0].axisValue
        const value = params[0].value
        let content = `${date}<br/>${nutrientName}: ${value.toFixed(2)} ${unit}`
        
        if (chartData.target !== null) {
          const percentage = (value / chartData.target * 100).toFixed(0)
          content += `<br/>Target: ${chartData.target} ${unit} (${percentage}%)`
        }
        
        if (chartData.upperLimit !== null) {
          const percentage = (value / chartData.upperLimit * 100).toFixed(0)
          content += `<br/>Upper Limit: ${chartData.upperLimit} ${unit} (${percentage}%)`
        }
        
        return content
      }
    },
    legend: {
      data: ['Intake', 'Target', 'Upper Limit'],
      top: 30,
      textStyle: {
        color: '#3c3d42'
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: '20%',
      containLabel: true
    },
    xAxis: {
      type: 'category' as const,
      data: chartData.dates.map(date => {
        const d = new Date(date)
        return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
      }),
      axisLine: {
        lineStyle: {
          color: '#dbd9d5'
        }
      },
      axisLabel: {
        color: '#3c3d42',
        rotate: 45
      }
    },
    yAxis: {
      type: 'value' as const,
      axisLine: {
        lineStyle: {
          color: '#dbd9d5'
        }
      },
      axisLabel: {
        color: '#3c3d42',
        formatter: function(value: number) {
          return `${value.toFixed(1)} ${unit}`
        }
      },
      splitLine: {
        lineStyle: {
          color: '#f0efeb'
        }
      }
    },
    series: [
      {
        name: 'Intake',
        type: 'line' as const,
        data: chartData.values,
        smooth: true,
        lineStyle: {
          color: '#74b986',
          width: 3
        },
        itemStyle: {
          color: '#74b986'
        },
        areaStyle: {
          color: {
            type: 'linear' as const,
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [{
              offset: 0, color: 'rgba(116, 185, 134, 0.3)'
            }, {
              offset: 1, color: 'rgba(116, 185, 134, 0.1)'
            }]
          }
        },
        markLine: {
          animation: false,
          data: [
            ...(chartData.target !== null ? [{
              name: 'Target',
              yAxis: chartData.target,
              lineStyle: {
                color: '#be7454',
                type: 'dashed' as const,
                width: 2
              },
              label: {
                formatter: `Target: ${chartData.target} ${unit}`,
                position: 'end' as const
              }
            }] : []),
            ...(chartData.upperLimit !== null ? [{
              name: 'Upper Limit',
              yAxis: chartData.upperLimit,
              lineStyle: {
                color: '#dc2626',
                type: 'dashed' as const,
                width: 2
              },
              label: {
                formatter: `Limit: ${chartData.upperLimit} ${unit}`,
                position: 'end' as const
              }
            }] : [])
          ]
        }
      }
    ]
  }

  // Helper to determine status based on target and upper limit
  $: currentStatus = getCurrentStatus(chartData.values[chartData.values.length - 1] || 0, chartData.target, chartData.upperLimit)

  function getCurrentStatus(currentValue: number, target: number | null, upperLimit: number | null) {
    if (upperLimit !== null && currentValue > upperLimit) {
      return { status: 'over-limit', color: 'text-error', message: 'Above upper limit' }
    }
    if (target !== null) {
      if (currentValue >= target) {
        if (upperLimit !== null && currentValue <= upperLimit) {
          return { status: 'optimal', color: 'text-success', message: 'In optimal range' }
        }
        return { status: 'met', color: 'text-success', message: 'Target met' }
      } else {
        const percentage = (currentValue / target * 100).toFixed(0)
        return { status: 'below', color: 'text-warning', message: `${percentage}% of target` }
      }
    }
    return { status: 'no-goal', color: 'text-base-content', message: 'No goal set' }
  }
</script>

<div class="w-full h-64 lg:h-72">
  <div class="flex justify-between items-center mb-2">
    <h3 class="text-sm font-semibold text-base-content">{nutrientName}</h3>
    <div class="text-xs {currentStatus.color}">
      {currentStatus.message}
    </div>
  </div>
  
  {#if chartData.values.length > 0}
    <Chart {init} {options} />
  {:else}
    <div class="h-full flex items-center justify-center text-base-content-lighter">
      <div class="text-center">
        <svg class="w-8 h-8 mx-auto mb-2 opacity-50" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h6a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h6a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h4a1 1 0 110 2H4a1 1 0 01-1-1z" clip-rule="evenodd" />
        </svg>
        <p class="text-xs">No data available</p>
      </div>
    </div>
  {/if}
</div>