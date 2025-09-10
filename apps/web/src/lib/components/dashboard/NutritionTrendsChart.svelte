<script lang="ts">
  import { Chart } from "svelte-echarts"
  import { init, use } from "echarts/core"
  import { LineChart } from "echarts/charts"
  import {
    TitleComponent,
    TooltipComponent,
    LegendComponent,
    GridComponent,
    DataZoomComponent,
    ToolboxComponent
  } from "echarts/components"
  import { CanvasRenderer } from "echarts/renderers"

  // Initialize ECharts
  use([
    LineChart,
    TitleComponent,
    TooltipComponent,
    LegendComponent,
    GridComponent,
    DataZoomComponent,
    ToolboxComponent,
    CanvasRenderer
  ])

  export let consumptions: any[] = []

  // Process data for chart
  $: chartData = processNutritionData(consumptions)

  function processNutritionData(data: any[]) {
    if (!data || data.length === 0) return { dates: [], series: [] }

    // Sort by date
    const sortedData = [...data].sort((a, b) => 
      new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    )

    // Group by date and sum nutrients
    const dailyData = new Map()
    
    sortedData.forEach(consumption => {
      const date = new Date(consumption.created_at).toLocaleDateString()
      
      if (!dailyData.has(date)) {
        dailyData.set(date, {
          calories: 0,
          protein: 0,
          carbs: 0,
          fat: 0,
          fiber: 0,
          count: 0
        })
      }
      
      const daily = dailyData.get(date)
      daily.calories += consumption.summary?.totals?.calories || 0
      daily.protein += consumption.summary?.totals?.protein_g || 0
      daily.carbs += consumption.summary?.totals?.total_carbs_g || 0
      daily.fat += consumption.summary?.totals?.total_fat_g || 0
      daily.fiber += consumption.summary?.totals?.dietary_fiber_g || 0
      daily.count += 1
    })

    // Convert to arrays for chart
    const dates = Array.from(dailyData.keys())
    const calories = Array.from(dailyData.values()).map(d => Math.round(d.calories))
    const protein = Array.from(dailyData.values()).map(d => Math.round(d.protein))
    const carbs = Array.from(dailyData.values()).map(d => Math.round(d.carbs))
    const fat = Array.from(dailyData.values()).map(d => Math.round(d.fat))
    const fiber = Array.from(dailyData.values()).map(d => Math.round(d.fiber))

    return {
      dates,
      series: [
        { name: 'Calories', data: calories, color: '#be7454' },
        { name: 'Protein (g)', data: protein, color: '#74b986' },
        { name: 'Carbs (g)', data: carbs, color: '#657280' },
        { name: 'Fat (g)', data: fat, color: '#667584' },
        { name: 'Fiber (g)', data: fiber, color: '#87888a' }
      ]
    }
  }

  $: options = {
    backgroundColor: 'transparent',
    title: {
      text: 'Nutrition Trends',
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
      axisPointer: {
        type: 'cross' as const,
        label: {
          backgroundColor: '#657280'
        }
      }
    },
    legend: {
      data: chartData.series.map(s => s.name),
      top: 30,
      textStyle: {
        color: '#3c3d42'
      }
    },
    toolbox: {
      feature: {
        dataZoom: {
          yAxisIndex: 'none' as const
        },
        restore: {},
        saveAsImage: {
          backgroundColor: '#faf9f5'
        }
      },
      iconStyle: {
        borderColor: '#657280'
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      top: '20%',
      containLabel: true
    },
    dataZoom: [
      {
        type: 'inside',
        start: 0,
        end: 100
      },
      {
        start: 0,
        end: 100,
        height: 30,
        bottom: 10,
        borderColor: '#dbd9d5',
        fillerColor: 'rgba(190, 116, 84, 0.2)',
        textStyle: {
          color: '#3c3d42'
        }
      }
    ],
    xAxis: {
      type: 'category' as const,
      boundaryGap: false,
      data: chartData.dates,
      axisLine: {
        lineStyle: {
          color: '#dbd9d5'
        }
      },
      axisLabel: {
        color: '#3c3d42'
      }
    },
    yAxis: [
      {
        type: 'value' as const,
        name: 'Amount',
        position: 'left' as const,
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
      }
    ],
    series: chartData.series.map(s => ({
      name: s.name,
      type: 'line' as const,
      data: s.data,
      smooth: true,
      lineStyle: {
        color: s.color,
        width: 2
      },
      itemStyle: {
        color: s.color
      },
      areaStyle: {
        color: {
          type: 'linear' as const,
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: s.color + '40' },
            { offset: 1, color: s.color + '10' }
          ]
        }
      }
    }))
  }
</script>

<div class="w-full h-64 sm:h-80 lg:h-96">
  {#if chartData.dates.length > 0}
    <Chart {init} {options} />
  {:else}
    <div class="h-full flex items-center justify-center text-base-content-lighter">
      <div class="text-center">
        <svg class="w-16 h-16 mx-auto mb-4 opacity-50" fill="currentColor" viewBox="0 0 20 20">
          <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z" />
        </svg>
        <p>No data available for the selected period</p>
      </div>
    </div>
  {/if}
</div>