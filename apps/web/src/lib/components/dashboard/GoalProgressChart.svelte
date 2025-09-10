<script lang="ts">
  import { Chart } from "svelte-echarts"
  import { init, use } from "echarts/core"
  import { BarChart } from "echarts/charts"
  import {
    TitleComponent,
    TooltipComponent,
    LegendComponent,
    GridComponent
  } from "echarts/components"
  import { CanvasRenderer } from "echarts/renderers"

  // Initialize ECharts
  use([
    BarChart,
    TitleComponent,
    TooltipComponent,
    LegendComponent,
    GridComponent,
    CanvasRenderer
  ])

  export let consumptions: any[] = []
  export let goals: any = null

  // Calculate nutrient totals and goal progress
  $: progressData = calculateGoalProgress(consumptions, goals)

  function calculateGoalProgress(data: any[], goalsData: any) {
    if (!data || data.length === 0 || !goalsData) {
      return { nutrients: [], actual: [], targets: [], percentages: [] }
    }

    // Sum up all nutrients from consumptions
    const totals = data.reduce((acc, consumption) => {
      acc.calories += consumption.summary?.total_calories || 0
      acc.protein += consumption.summary?.total_protein_g || 0
      acc.carbs += consumption.summary?.total_carbs_g || 0
      acc.fat += consumption.summary?.total_fat_g || 0
      acc.fiber += consumption.summary?.total_fiber_g || 0
      acc.sodium += consumption.summary?.total_sodium_mg || 0
      acc.sugar += consumption.summary?.total_sugar_g || 0
      return acc
    }, { 
      calories: 0, protein: 0, carbs: 0, fat: 0, 
      fiber: 0, sodium: 0, sugar: 0 
    })

    // Map to goal targets
    const targets = goalsData.goals?.targets || {}
    
    const nutrients = [
      { name: 'Calories', actual: Math.round(totals.calories), target: targets.calories || 2000, unit: 'cal' },
      { name: 'Protein', actual: Math.round(totals.protein), target: targets.protein_g || 50, unit: 'g' },
      { name: 'Carbs', actual: Math.round(totals.carbs), target: targets.carbs_g || 300, unit: 'g' },
      { name: 'Fat', actual: Math.round(totals.fat), target: targets.fat_g || 65, unit: 'g' },
      { name: 'Fiber', actual: Math.round(totals.fiber), target: targets.fiber_g || 25, unit: 'g' },
      { name: 'Sodium', actual: Math.round(totals.sodium), target: targets.sodium_mg || 2300, unit: 'mg' }
    ]

    return {
      nutrients: nutrients.map(n => n.name),
      actual: nutrients.map(n => n.actual),
      targets: nutrients.map(n => n.target),
      percentages: nutrients.map(n => Math.round((n.actual / n.target) * 100)),
      details: nutrients
    }
  }

  $: options = {
    backgroundColor: 'transparent',
    title: {
      text: 'Goal Progress',
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
        const nutrient = params[0].axisValue
        const detail = progressData.details?.find(d => d.name === nutrient)
        if (!detail) return ''
        
        const actualData = params.find((p: any) => p.seriesName === 'Actual')
        const targetData = params.find((p: any) => p.seriesName === 'Target')
        
        return `${nutrient}<br/>
                Actual: ${actualData.value} ${detail.unit}<br/>
                Target: ${targetData.value} ${detail.unit}<br/>
                Progress: ${Math.round((actualData.value / targetData.value) * 100)}%`
      }
    },
    legend: {
      data: ['Actual', 'Target'],
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
      data: progressData.nutrients,
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
        name: 'Actual',
        type: 'bar' as const,
        data: progressData.actual,
        itemStyle: {
          color: '#74b986'
        }
      },
      {
        name: 'Target',
        type: 'bar' as const,
        data: progressData.targets,
        itemStyle: {
          color: '#be7454',
          opacity: 0.6
        }
      }
    ]
  }
</script>

<div class="w-full h-48 sm:h-56 lg:h-64">
  {#if progressData.nutrients.length > 0}
    <Chart {init} {options} />
  {:else}
    <div class="h-full flex items-center justify-center text-base-content-lighter">
      <div class="text-center">
        <svg class="w-12 h-12 mx-auto mb-2 opacity-50" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M12.395 2.553a1 1 0 00-1.45-.385c-.345.23-.614.558-.822.88-.214.33-.403.713-.57 1.116-.334.804-.614 1.768-.84 2.734a31.365 31.365 0 00-.613 3.58 2.64 2.64 0 01-.945-1.067c-.328-.68-.398-1.534-.398-2.654A1 1 0 005.05 6.05 6.981 6.981 0 003 11a7 7 0 1011.95-4.95c-.592-.591-.98-.985-1.348-1.467-.363-.476-.724-1.063-1.207-2.03zM12.12 15.12A3 3 0 017 13s.879.5 2.5.5c0-1 .5-4 1.25-4.5.5 1 .786 1.293 1.371 1.879A2.99 2.99 0 0113 13a2.99 2.99 0 01-.879 2.121z" clip-rule="evenodd" />
        </svg>
        <p class="text-sm">No goal data available</p>
      </div>
    </div>
  {/if}
</div>