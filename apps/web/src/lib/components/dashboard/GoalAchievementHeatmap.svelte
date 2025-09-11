<script lang="ts">
  import { Chart } from "svelte-echarts"
  import { init, use } from "echarts/core"
  import { HeatmapChart } from "echarts/charts"
  import {
    TitleComponent,
    TooltipComponent,
    GridComponent,
    CalendarComponent
  } from "echarts/components"
  import { CanvasRenderer } from "echarts/renderers"

  // Initialize ECharts
  use([
    HeatmapChart,
    TitleComponent,
    TooltipComponent,
    GridComponent,
    CalendarComponent,
    CanvasRenderer
  ])

  export let consumptions: any[] = []
  export let goalsData: any = null
  export let timeWindow: "7d" | "30d" | "90d" = "30d"

  // Key nutrients to track for heatmap
  const keyNutrients = [
    { key: 'calories', name: 'Calories' },
    { key: 'protein_g', name: 'Protein' },
    { key: 'total_carbs_g', name: 'Carbs' },
    { key: 'total_fat_g', name: 'Fat' },
    { key: 'dietary_fiber_g', name: 'Fiber' },
    { key: 'vitamin_c_mg', name: 'Vitamin C' },
    { key: 'calcium_mg', name: 'Calcium' },
    { key: 'iron_mg', name: 'Iron' },
    { key: 'potassium_mg', name: 'Potassium' },
    { key: 'sodium_mg', name: 'Sodium' }
  ]

  $: heatmapData = calculateGoalAchievementHeatmap(consumptions, goalsData, timeWindow)

  function calculateGoalAchievementHeatmap(data: any[], goals: any, window: string) {
    if (!data || data.length === 0 || !goals) {
      return { dates: [], nutrients: [], matrix: [], overallScores: [] }
    }

    const days = parseInt(window.replace('d', ''))
    
    // Generate date labels for the past N days
    const dateLabels = []
    const today = new Date()
    for (let i = days - 1; i >= 0; i--) {
      const date = new Date(today)
      date.setDate(date.getDate() - i)
      dateLabels.push({
        date: date.toISOString().split('T')[0],
        label: date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
      })
    }

    // Group consumptions by date
    const consumptionsByDate = data.reduce((acc: any, consumption: any) => {
      const date = new Date(consumption.consumed_at || consumption.created_at).toISOString().split('T')[0]
      if (!acc[date]) acc[date] = []
      acc[date].push(consumption)
      return acc
    }, {})

    // Calculate achievement matrix
    const matrix: number[][] = []
    const overallScores: number[] = []

    dateLabels.forEach((dayInfo, dayIndex) => {
      const dayConsumptions = consumptionsByDate[dayInfo.date] || []
      
      // Sum up nutrients for this day
      const dayTotals = dayConsumptions.reduce((acc: any, consumption: any) => {
        keyNutrients.forEach(nutrient => {
          const value = consumption.summary?.totals?.[nutrient.key] || 0
          acc[nutrient.key] = (acc[nutrient.key] || 0) + value
        })
        return acc
      }, {})

      const dayRow: number[] = []
      let dayScore = 0
      let validNutrients = 0

      keyNutrients.forEach((nutrient, nutrientIndex) => {
        const actual = dayTotals[nutrient.key] || 0
        const target = goals.targets?.[nutrient.key]
        const upperLimit = goals.upper_limits?.[nutrient.key]
        
        let achievementScore = 0

        if (target !== undefined && target > 0) {
          // For nutrients with targets (minimum goals)
          if (actual >= target) {
            if (upperLimit !== undefined && actual > upperLimit) {
              // Over the upper limit - partial score
              achievementScore = 0.6
            } else {
              // Met target within bounds - full score
              achievementScore = 1.0
            }
          } else {
            // Below target - proportional score
            achievementScore = Math.min(actual / target, 0.8)
          }
          validNutrients++
        } else if (upperLimit !== undefined && upperLimit > 0) {
          // For nutrients with only upper limits (minimize these)
          if (actual === 0) {
            achievementScore = 1.0
          } else if (actual <= upperLimit) {
            achievementScore = Math.max(0.2, 1 - (actual / upperLimit))
          } else {
            achievementScore = 0
          }
          validNutrients++
        }

        dayRow.push(achievementScore)
        dayScore += achievementScore
      })

      matrix.push(dayRow)
      overallScores.push(validNutrients > 0 ? dayScore / validNutrients : 0)
    })

    return {
      dates: dateLabels,
      nutrients: keyNutrients,
      matrix,
      overallScores
    }
  }

  $: options = {
    backgroundColor: 'transparent',
    title: {
      text: 'Goal Achievement Heatmap',
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
        const dayIndex = params.data[0]
        const nutrientIndex = params.data[1]
        const score = params.data[2]
        
        if (heatmapData.dates[dayIndex] && heatmapData.nutrients[nutrientIndex]) {
          const date = heatmapData.dates[dayIndex].label
          const nutrient = heatmapData.nutrients[nutrientIndex].name
          const percentage = (score * 100).toFixed(0)
          
          let status = 'No Goal'
          if (score === 0) status = 'Not Met'
          else if (score < 0.5) status = 'Poor'
          else if (score < 0.8) status = 'Fair'
          else if (score < 1.0) status = 'Good'
          else status = 'Excellent'
          
          return `${date}<br/>${nutrient}: ${percentage}%<br/>Status: ${status}`
        }
        return ''
      }
    },
    grid: {
      left: '10%',
      right: '10%',
      top: '15%',
      bottom: '15%',
      containLabel: true
    },
    xAxis: {
      type: 'category' as const,
      data: heatmapData.dates.map(d => d.label),
      axisLine: {
        lineStyle: {
          color: '#dbd9d5'
        }
      },
      axisLabel: {
        color: '#3c3d42',
        rotate: 45,
        fontSize: 10
      },
      splitLine: {
        show: false
      }
    },
    yAxis: {
      type: 'category' as const,
      data: heatmapData.nutrients.map(n => n.name),
      axisLine: {
        lineStyle: {
          color: '#dbd9d5'
        }
      },
      axisLabel: {
        color: '#3c3d42',
        fontSize: 10
      },
      splitLine: {
        show: false
      }
    },
    visualMap: {
      min: 0,
      max: 1,
      calculable: true,
      realtime: false,
      inRange: {
        color: ['#fee2e2', '#fef3c7', '#d9f99d', '#86efac', '#34d399']
      },
      text: ['High', 'Low'],
      textStyle: {
        color: '#3c3d42'
      },
      left: 'right',
      top: 'center'
    },
    series: [{
      name: 'Goal Achievement',
      type: 'heatmap' as const,
      data: heatmapData.matrix.flatMap((row, dayIndex) => 
        row.map((score, nutrientIndex) => [dayIndex, nutrientIndex, score])
      ),
      emphasis: {
        itemStyle: {
          borderColor: '#333',
          borderWidth: 2
        }
      }
    }]
  }

  // Calculate overall statistics
  $: stats = calculateOverallStats(heatmapData)

  function calculateOverallStats(data: any) {
    if (!data.overallScores || data.overallScores.length === 0) {
      return { avgScore: 0, bestDay: null, worstDay: null, streak: 0 }
    }

    const avgScore = data.overallScores.reduce((sum: number, score: number) => sum + score, 0) / data.overallScores.length
    
    let bestDayIndex = 0
    let worstDayIndex = 0
    let bestScore = data.overallScores[0]
    let worstScore = data.overallScores[0]
    
    data.overallScores.forEach((score: number, index: number) => {
      if (score > bestScore) {
        bestScore = score
        bestDayIndex = index
      }
      if (score < worstScore) {
        worstScore = score
        worstDayIndex = index
      }
    })

    // Calculate current streak of good days (score >= 0.7)
    let streak = 0
    for (let i = data.overallScores.length - 1; i >= 0; i--) {
      if (data.overallScores[i] >= 0.7) {
        streak++
      } else {
        break
      }
    }

    return {
      avgScore,
      bestDay: data.dates[bestDayIndex]?.label || null,
      worstDay: data.dates[worstDayIndex]?.label || null,
      streak
    }
  }
</script>

<div class="space-y-4">
  <!-- Header with stats -->
  <div class="flex justify-between items-start">
    <div>
      <h3 class="text-lg font-semibold text-base-content">Goal Achievement Overview</h3>
      <p class="text-sm text-base-content/70">
        Track daily progress across {heatmapData.nutrients.length} key nutrients
      </p>
    </div>
    
    <!-- Time window selector -->
    <div class="flex gap-2">
      <button 
        class="btn btn-sm {timeWindow === '7d' ? 'btn-primary' : 'btn-outline'}"
        on:click={() => timeWindow = '7d'}
      >
        7 Days
      </button>
      <button 
        class="btn btn-sm {timeWindow === '30d' ? 'btn-primary' : 'btn-outline'}"
        on:click={() => timeWindow = '30d'}
      >
        30 Days
      </button>
      <button 
        class="btn btn-sm {timeWindow === '90d' ? 'btn-primary' : 'btn-outline'}"
        on:click={() => timeWindow = '90d'}
      >
        90 Days
      </button>
    </div>
  </div>

  <!-- Stats row -->
  {#if stats.avgScore > 0}
    <div class="stats stats-horizontal bg-base-100 shadow-sm">
      <div class="stat">
        <div class="stat-title">Average Score</div>
        <div class="stat-value text-lg {stats.avgScore >= 0.8 ? 'text-success' : stats.avgScore >= 0.6 ? 'text-warning' : 'text-error'}">
          {(stats.avgScore * 100).toFixed(0)}%
        </div>
        <div class="stat-desc">Daily goal achievement</div>
      </div>
      
      <div class="stat">
        <div class="stat-title">Best Day</div>
        <div class="stat-value text-lg text-success">{stats.bestDay || 'N/A'}</div>
        <div class="stat-desc">Highest achievement</div>
      </div>
      
      <div class="stat">
        <div class="stat-title">Current Streak</div>
        <div class="stat-value text-lg text-info">{stats.streak}</div>
        <div class="stat-desc">Good days (≥70%)</div>
      </div>
    </div>
  {/if}

  <!-- Heatmap chart -->
  <div class="w-full" style="height: 400px;">
    {#if heatmapData.matrix.length > 0}
      <Chart {init} {options} />
    {:else}
      <div class="h-full flex items-center justify-center text-base-content-lighter bg-base-200 rounded-lg">
        <div class="text-center">
          <svg class="w-12 h-12 mx-auto mb-2 opacity-50" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h6a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h6a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h4a1 1 0 110 2H4a1 1 0 01-1-1z" clip-rule="evenodd" />
          </svg>
          <p class="text-sm">No goal data available</p>
          <p class="text-xs text-base-content/50 mt-1">Set up nutrition goals to see achievement tracking</p>
        </div>
      </div>
    {/if}
  </div>

  <!-- Legend -->
  <div class="text-xs text-base-content/70 bg-base-100 p-3 rounded-lg">
    <div class="flex items-center justify-center gap-6">
      <div class="flex items-center gap-2">
        <div class="w-3 h-3 bg-red-200 rounded"></div>
        <span>Not Met (0-49%)</span>
      </div>
      <div class="flex items-center gap-2">
        <div class="w-3 h-3 bg-yellow-200 rounded"></div>
        <span>Fair (50-79%)</span>
      </div>
      <div class="flex items-center gap-2">
        <div class="w-3 h-3 bg-green-200 rounded"></div>
        <span>Good (80-99%)</span>
      </div>
      <div class="flex items-center gap-2">
        <div class="w-3 h-3 bg-green-400 rounded"></div>
        <span>Excellent (100%)</span>
      </div>
    </div>
  </div>
</div>