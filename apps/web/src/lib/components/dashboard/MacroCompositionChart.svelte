<script lang="ts">
  import { Chart } from "svelte-echarts"
  import { init, use } from "echarts/core"
  import { PieChart } from "echarts/charts"
  import {
    TitleComponent,
    TooltipComponent,
    LegendComponent
  } from "echarts/components"
  import { CanvasRenderer } from "echarts/renderers"

  // Initialize ECharts
  use([
    PieChart,
    TitleComponent,
    TooltipComponent,
    LegendComponent,
    CanvasRenderer
  ])

  export let consumptions: any[] = []

  // Calculate macro totals
  $: macroData = calculateMacros(consumptions)

  function calculateMacros(data: any[]) {
    if (!data || data.length === 0) {
      return { protein: 0, carbs: 0, fat: 0, total: 0 }
    }

    const totals = data.reduce((acc, consumption) => {
      acc.protein += consumption.summary?.total_protein_g || 0
      acc.carbs += consumption.summary?.total_carbs_g || 0
      acc.fat += consumption.summary?.total_fat_g || 0
      return acc
    }, { protein: 0, carbs: 0, fat: 0 })

    // Convert to calories (protein: 4 cal/g, carbs: 4 cal/g, fat: 9 cal/g)
    const proteinCals = totals.protein * 4
    const carbsCals = totals.carbs * 4
    const fatCals = totals.fat * 9
    const total = proteinCals + carbsCals + fatCals

    return {
      protein: Math.round(proteinCals),
      carbs: Math.round(carbsCals),
      fat: Math.round(fatCals),
      total: Math.round(total)
    }
  }

  $: options = {
    backgroundColor: 'transparent',
    title: {
      text: 'Macro Composition',
      textStyle: {
        color: '#3c3d42',
        fontSize: 16,
        fontWeight: 'normal' as const
      },
      left: 'center'
    },
    tooltip: {
      trigger: 'item' as const,
      backgroundColor: '#faf9f5',
      borderColor: '#dbd9d5',
      textStyle: {
        color: '#3c3d42'
      },
      formatter: '{a} <br/>{b}: {c} cal ({d}%)'
    },
    legend: {
      orient: 'vertical' as const,
      left: 'left',
      top: 'middle',
      textStyle: {
        color: '#3c3d42'
      }
    },
    series: [
      {
        name: 'Macros',
        type: 'pie' as const,
        radius: ['40%', '70%'],
        center: ['60%', '50%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 8,
          borderColor: '#faf9f5',
          borderWidth: 2
        },
        label: {
          show: false,
          position: 'center' as const
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 20,
            fontWeight: 'bold' as const,
            color: '#3c3d42'
          }
        },
        labelLine: {
          show: false
        },
        data: [
          {
            value: macroData.protein,
            name: 'Protein',
            itemStyle: { color: '#74b986' }
          },
          {
            value: macroData.carbs,
            name: 'Carbs',
            itemStyle: { color: '#657280' }
          },
          {
            value: macroData.fat,
            name: 'Fat',
            itemStyle: { color: '#be7454' }
          }
        ]
      }
    ]
  }
</script>

<div class="w-full h-48 sm:h-56 lg:h-64">
  {#if macroData.total > 0}
    <Chart {init} {options} />
  {:else}
    <div class="h-full flex items-center justify-center text-base-content-lighter">
      <div class="text-center">
        <svg class="w-12 h-12 mx-auto mb-2 opacity-50" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM4.332 8.027a6.012 6.012 0 011.912-2.706C6.512 5.73 6.974 6 7.5 6A1.5 1.5 0 019 7.5V8a2 2 0 004 0 2 2 0 011.523-1.943A5.977 5.977 0 0116 10c0 .34-.028.675-.083 1H15a2 2 0 00-2 2v2.197A5.973 5.973 0 0110 16v-2a2 2 0 00-2-2 2 2 0 01-2-2 2 2 0 00-1.668-1.973z" clip-rule="evenodd" />
        </svg>
        <p class="text-sm">No macro data available</p>
      </div>
    </div>
  {/if}
</div>