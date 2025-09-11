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
    ToolboxComponent,
    BrushComponent
  } from "echarts/components"
  import { CanvasRenderer } from "echarts/renderers"
  
  // Register ECharts
  use([
    LineChart,
    TitleComponent,
    TooltipComponent,
    LegendComponent,
    GridComponent,
    DataZoomComponent,
    ToolboxComponent,
    BrushComponent,
    CanvasRenderer
  ])

  export let consumptions: any[] = []

  let chartInstance: any = null

  // Process data for chart
  $: chartData = processNutritionData(consumptions)

    // Auto-enable dataZoom tool when chart instance becomes available
  $: if (chartInstance) {
    setTimeout(() => {
      try {
        console.debug('Chart instance available, setting up brush events and auto-enabling dataZoom tool')
        
        // Register brush event listeners
        chartInstance.on('brushSelected', handleBrushSelected)
        chartInstance.on('brushEnd', handleBrushEnd)
        
        // Auto-enable the dataZoom brush tool
        chartInstance.dispatchAction({
          type: 'toolboxDataZoom'
        })
        
        console.debug('DataZoom tool auto-activation attempted')
      } catch (error) {
        console.debug('Error setting up chart or auto-activating dataZoom tool:', error)
      }
    }, 200)
  }

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
        { name: 'Calories', data: calories, color: '#BF8711' }, // honey color
        { name: 'Protein (g)', data: protein, color: '#657280' }, // secondary gray (dumbbell icon color)
        { name: 'Carbs (g)', data: carbs, color: '#be7454' }, // primary brownish-orange
        { name: 'Fat (g)', data: fat, color: '#74b986' }, // accent green
        { name: 'Fiber (g)', data: fiber, color: '#87888a' } // neutral gray
      ]
    }
  }

  $: options = {
    backgroundColor: 'transparent',
    title: {
      text: 'Macro Nutrition Trends',
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
          yAxisIndex: 'none' as const,
          title: {
            zoom: 'Zoom In',
            back: 'Zoom Out'
          }
        },
        brush: {
          type: ['lineX', 'clear'] as ('lineX' | 'clear')[],
          title: {
            lineX: 'Select to Zoom',
            clear: 'Clear Selection'
          }
        },
        restore: {
          title: 'Reset Zoom'
        },
        saveAsImage: {
          backgroundColor: '#faf9f5',
          title: 'Save as Image',
          pixelRatio: 2,
          excludeComponents: ['toolbox']
        }
      },
      iconStyle: {
        borderColor: '#657280'
      }
    },
    brush: {
      toolbox: ['lineX', 'clear'] as ('lineX' | 'clear')[],
      xAxisIndex: 0,
      brushType: 'lineX' as const,
      brushMode: 'single' as const,
      transformable: true,
      removeOnClick: false,
      inBrush: {
        opacity: 1
      },
      outOfBrush: {
        colorAlpha: 0.1
      },
      brushStyle: {
        borderColor: 'rgba(116, 185, 134, 0.8)',
        color: 'rgba(116, 185, 134, 0.2)',
        borderWidth: 2
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
        end: 100,
        zoomOnMouseWheel: true,
        moveOnMouseMove: true,
        moveOnMouseWheel: false
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

  // Handle brush selection to zoom
  function handleBrushSelected(params: any) {
    console.debug('Brush selected:', params)
    
    if (params.batch && params.batch[0]) {
      const batch = params.batch[0]
      if (batch.selected && batch.selected[0]) {
        const selection = batch.selected[0]
        if (selection.brushType === 'lineX' && selection.coordRange) {
          const [startCoord, endCoord] = selection.coordRange
          
          // Convert pixel coordinates to data indices
          const totalDates = chartData.dates.length
          const startPercent = Math.max(0, (startCoord / totalDates) * 100)
          const endPercent = Math.min(100, (endCoord / totalDates) * 100)
          
          console.debug('Zooming to range:', startPercent, endPercent)
          
          // Update both dataZoom components
          options = {
            ...options,
            dataZoom: [
              {
                ...options.dataZoom[0],
                start: startPercent,
                end: endPercent
              },
              {
                ...options.dataZoom[1], 
                start: startPercent,
                end: endPercent
              }
            ]
          }
        }
      }
    }
  }

  // Handle brush end event (when user finishes selection)
  function handleBrushEnd(params: any) {
    console.debug('Brush end:', params)
    
    if (params.areas && params.areas.length > 0) {
      const area = params.areas[0]
      if (area.brushType === 'lineX' && area.coordRange) {
        const [startCoord, endCoord] = area.coordRange
        
        // Convert coordinates to percentages
        const totalDates = chartData.dates.length
        const startPercent = Math.max(0, (startCoord / totalDates) * 100)  
        const endPercent = Math.min(100, (endCoord / totalDates) * 100)
        
        console.debug('Brush end - zooming to:', startPercent, endPercent)
        
        // Update dataZoom
        options = {
          ...options,
          dataZoom: [
            {
              ...options.dataZoom[0],
              start: startPercent,
              end: endPercent
            },
            {
              ...options.dataZoom[1],
              start: startPercent, 
              end: endPercent
            }
          ]
        }
      }
    }
  }
</script>

<div class="w-full h-64 sm:h-80 lg:h-96">
  {#if chartData.dates.length > 0}
    <Chart 
      {init} 
      {options} 
      bind:chart={chartInstance}
    />
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
