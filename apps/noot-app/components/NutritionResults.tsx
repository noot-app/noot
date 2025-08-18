/**
 * Nutrition results display matching the original noot web frontend design
 * with cards, glass morphism, and proper nutrition layout
 */
import React from 'react';
import { View, ScrollView, StyleSheet } from 'react-native';
import { GlassCard, ThemedText } from './ThemedComponents';
import { Colors } from '@/constants/Colors';
import { useColorScheme } from '@/hooks/useColorScheme';

interface NutritionResultsProps {
  data: any;
}

interface PieChartProps {
  percentage: number;
  color: string;
}

// Simple circular progress indicator to mimic the original pie charts
function PieChart({ percentage, color }: PieChartProps) {
  // Simplified implementation - just a filled circle with opacity based on percentage
  return (
    <View style={styles.pieChartContainer}>
      <View 
        style={[
          styles.pieChart,
          { 
            borderColor: color,
            borderWidth: 3,
          }
        ]}
      >
        <View 
          style={[
            styles.pieChartFill,
            { 
              backgroundColor: color,
              opacity: percentage / 100,
            }
          ]}
        />
      </View>
    </View>
  );
}

export function NutritionResults({ data }: NutritionResultsProps) {
  const colorScheme = useColorScheme();
  const colors = Colors[colorScheme];
  const commonColors = Colors.common;

  if (!data) return null;

  const { transcript, summary, items } = data;

  const formatNumber = (num: number | undefined): string => {
    if (typeof num !== 'number' || isNaN(num)) return '0';
    return num % 1 === 0 ? num.toString() : num.toFixed(1);
  };

  const getPercentageColor = (percentage: number): string => {
    if (percentage < 25) return commonColors.percentageLow;
    if (percentage < 50) return commonColors.percentageMedium;
    if (percentage < 75) return commonColors.percentageGood;
    if (percentage < 100) return commonColors.percentageHigh;
    return commonColors.percentageOver;
  };

  const renderMacronutrients = () => {
    if (!summary?.totals) return null;
    
    const totals = summary.totals;
    const percentDaily = summary.percent_of_daily || {};
    
    const macros = [
      { 
        label: 'Calories', 
        value: formatNumber(totals.calories), 
        unit: 'kcal', 
        percent: Math.min(percentDaily.calories || 0, 100)
      },
      { 
        label: 'Protein', 
        value: formatNumber(totals.protein_g), 
        unit: 'g', 
        percent: Math.min(percentDaily.protein || 0, 100)
      },
      { 
        label: 'Carbs', 
        value: formatNumber(totals.total_carbs_g), 
        unit: 'g', 
        percent: Math.min(percentDaily.total_carbs || 0, 100)
      },
      { 
        label: 'Fat', 
        value: formatNumber(totals.total_fat_g), 
        unit: 'g', 
        percent: Math.min(percentDaily.total_fat || 0, 100)
      }
    ];

    return (
      <View style={styles.nutritionSection}>
        <ThemedText style={styles.sectionTitle}>⚡ Macronutrients</ThemedText>
        <View style={styles.nutritionGrid}>
          {macros.map((macro, index) => (
            <View key={index} style={[styles.nutritionItem, { borderColor: colors.borderColor }]}>
              <PieChart 
                percentage={macro.percent} 
                color={getPercentageColor(macro.percent)} 
              />
              <ThemedText type="tertiary" style={styles.nutritionLabel}>
                {macro.label}
              </ThemedText>
              <ThemedText style={styles.nutritionValue}>
                {macro.value}{macro.unit}
              </ThemedText>
              <ThemedText type="tertiary" style={styles.nutritionPercent}>
                {Math.round(macro.percent)}% daily
              </ThemedText>
            </View>
          ))}
        </View>
      </View>
    );
  };

  const renderSugarBreakdown = () => {
    if (!summary?.totals) return null;
    
    const totals = summary.totals;
    const totalSugarsG = totals.total_sugars_g || 0;
    const addedSugarsG = totals.added_sugars_g || 0;
    const naturalSugarsG = Math.max(0, totalSugarsG - addedSugarsG);
    
    if (totalSugarsG === 0) return null;
    
    const addedSugarLimit = 50;
    const addedSugarPercent = Math.min((addedSugarsG / addedSugarLimit) * 100, 100);
    
    const sugarItems = [];
    
    if (addedSugarsG > 0) {
      sugarItems.push({
        label: 'Added Sugar',
        value: formatNumber(addedSugarsG),
        unit: 'g',
        percent: addedSugarPercent,
        note: `of ${addedSugarLimit}g limit`,
        warning: addedSugarPercent > 80
      });
    }
    
    if (naturalSugarsG > 0) {
      sugarItems.push({
        label: 'Natural Sugar',
        value: formatNumber(naturalSugarsG),
        unit: 'g',
        percent: 0,
        note: 'from fruits & dairy',
        warning: false
      });
    }

    if (sugarItems.length === 0) return null;

    return (
      <View style={styles.nutritionSection}>
        <ThemedText style={styles.sectionTitle}>🍯 Sugar Breakdown</ThemedText>
        <View style={styles.nutritionGrid}>
          {sugarItems.map((item, index) => (
            <View key={index} style={[
              styles.nutritionItem, 
              { borderColor: item.warning ? commonColors.red : colors.borderColor }
            ]}>
              {item.percent > 0 && (
                <PieChart 
                  percentage={item.percent} 
                  color={item.warning ? commonColors.red : getPercentageColor(item.percent)} 
                />
              )}
              <ThemedText type="tertiary" style={styles.nutritionLabel}>
                {item.label}
              </ThemedText>
              <ThemedText style={styles.nutritionValue}>
                {item.value}{item.unit}
              </ThemedText>
              <ThemedText type="tertiary" style={styles.nutritionPercent}>
                {item.note}
              </ThemedText>
            </View>
          ))}
        </View>
      </View>
    );
  };

  const renderVitaminsAndMinerals = () => {
    if (!summary?.totals) return null;
    
    const totals = summary.totals;
    const percentDaily = summary.percent_of_daily || {};
    
    const nutrients = [
      { 
        label: 'Fiber', 
        value: formatNumber(totals.dietary_fiber_g), 
        unit: 'g', 
        percent: Math.min(percentDaily.dietary_fiber || 0, 100)
      },
      { 
        label: 'Calcium', 
        value: formatNumber(totals.calcium_mg), 
        unit: 'mg', 
        percent: Math.min(percentDaily.calcium || 0, 100)
      },
      { 
        label: 'Iron', 
        value: formatNumber(totals.iron_mg), 
        unit: 'mg', 
        percent: Math.min(percentDaily.iron || 0, 100)
      },
      { 
        label: 'Sodium', 
        value: formatNumber(totals.sodium_mg), 
        unit: 'mg', 
        percent: Math.min(percentDaily.sodium || 0, 100)
      },
      { 
        label: 'Potassium', 
        value: formatNumber(totals.potassium_mg), 
        unit: 'mg', 
        percent: Math.min(percentDaily.potassium || 0, 100)
      }
    ].filter(nutrient => parseFloat(nutrient.value) > 0);

    if (nutrients.length === 0) return null;

    return (
      <View style={styles.nutritionSection}>
        <ThemedText style={styles.sectionTitle}>💊 Vitamins & Minerals</ThemedText>
        <View style={styles.nutritionGrid}>
          {nutrients.map((nutrient, index) => (
            <View key={index} style={[styles.nutritionItem, { borderColor: colors.borderColor }]}>
              <PieChart 
                percentage={nutrient.percent} 
                color={getPercentageColor(nutrient.percent)} 
              />
              <ThemedText type="tertiary" style={styles.nutritionLabel}>
                {nutrient.label}
              </ThemedText>
              <ThemedText style={styles.nutritionValue}>
                {nutrient.value}{nutrient.unit}
              </ThemedText>
              <ThemedText type="tertiary" style={styles.nutritionPercent}>
                {Math.round(nutrient.percent)}% daily
              </ThemedText>
            </View>
          ))}
        </View>
      </View>
    );
  };

  return (
    <ScrollView style={styles.container} showsVerticalScrollIndicator={false}>
      {/* Nutrition Summary Card */}
      <GlassCard style={styles.summaryCard}>
        <ThemedText style={styles.cardTitle}>📊 Nutrition Summary</ThemedText>
        {renderMacronutrients()}
        {renderSugarBreakdown()}
        {renderVitaminsAndMinerals()}
      </GlassCard>

      {/* Individual Items Card */}
      {items && items.length > 0 && (
        <GlassCard style={styles.itemsCard}>
          <ThemedText style={styles.cardTitle}>🍽️ Individual Items</ThemedText>
          {items.map((itemData: any, index: number) => {
            const item = itemData.item;
            const nutrients = item?.nutrients;
            
            return (
              <View key={index} style={[styles.foodItem, { borderColor: colors.borderColor }]}>
                <View style={styles.itemHeader}>
                  <ThemedText style={styles.itemName}>{item?.name || 'Unknown item'}</ThemedText>
                  {item?.quantity && item?.unit && (
                    <View style={[styles.quantityBadge, { backgroundColor: colors.glassBg }]}>
                      <ThemedText type="tertiary" style={styles.quantityText}>
                        {formatNumber(item.quantity)} {item.unit}
                      </ThemedText>
                    </View>
                  )}
                </View>
                
                {nutrients && (
                  <View style={styles.itemNutrition}>
                    <View style={styles.keyNutrients}>
                      {[
                        { label: 'Calories', value: nutrients.calories, unit: 'kcal' },
                        { label: 'Protein', value: nutrients.protein_g, unit: 'g' },
                        { label: 'Carbs', value: nutrients.total_carbs_g, unit: 'g' },
                        { label: 'Fat', value: nutrients.total_fat_g, unit: 'g' }
                      ].map((n, i) => (
                        <View key={i} style={[styles.keyNutrientItem, { backgroundColor: colors.glassBg }]}>
                          <ThemedText type="tertiary" style={styles.keyNutrientLabel}>{n.label}</ThemedText>
                          <ThemedText style={styles.keyNutrientValue}>{formatNumber(n.value)}{n.unit}</ThemedText>
                        </View>
                      ))}
                    </View>
                  </View>
                )}
              </View>
            );
          })}
        </GlassCard>
      )}

      {/* Transcript Card */}
      <GlassCard style={styles.transcriptCard}>
        <ThemedText style={styles.transcriptTitle}>What you said</ThemedText>
        <View style={[styles.transcriptContainer, { 
          backgroundColor: colors.glassBg,
          borderLeftColor: colors.borderColor 
        }]}>
          <ThemedText type="secondary" style={styles.transcriptText}>
            &ldquo;{transcript || 'No transcript available'}&rdquo;
          </ThemedText>
        </View>
      </GlassCard>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    width: '100%',
    maxWidth: 896, // 56rem equivalent
  },
  summaryCard: {
    // GlassCard handles the styling
  },
  itemsCard: {
    // GlassCard handles the styling
  },
  transcriptCard: {
    // GlassCard handles the styling
  },
  cardTitle: {
    fontSize: 30,
    fontWeight: '700',
    marginBottom: 24,
    letterSpacing: -0.02,
  },
  nutritionSection: {
    marginBottom: 0,
    paddingVertical: 24,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: '600',
    marginBottom: 16,
    letterSpacing: -0.01,
  },
  nutritionGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 16,
  },
  nutritionItem: {
    flex: 1,
    minWidth: 100,
    padding: 16,
    borderRadius: 12,
    borderWidth: 1,
    alignItems: 'center',
    minHeight: 120,
  },
  pieChartContainer: {
    marginBottom: 8,
  },
  pieChart: {
    width: 32,
    height: 32,
    borderRadius: 16,
    borderWidth: 3,
    justifyContent: 'center',
    alignItems: 'center',
  },
  pieChartFill: {
    width: 20,
    height: 20,
    borderRadius: 10,
  },
  nutritionLabel: {
    fontSize: 14,
    marginBottom: 4,
    textAlign: 'center',
  },
  nutritionValue: {
    fontSize: 18,
    fontWeight: '700',
    marginBottom: 2,
    textAlign: 'center',
  },
  nutritionPercent: {
    fontSize: 12,
    textAlign: 'center',
  },
  foodItem: {
    padding: 24,
    borderRadius: 14,
    borderWidth: 1,
    marginBottom: 16,
  },
  itemHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 16,
    flexWrap: 'wrap',
    gap: 8,
  },
  itemName: {
    fontSize: 20,
    fontWeight: '600',
    letterSpacing: -0.01,
    flex: 1,
    minWidth: 200,
  },
  quantityBadge: {
    paddingHorizontal: 12,
    paddingVertical: 4,
    borderRadius: 16,
    borderWidth: 1,
  },
  quantityText: {
    fontSize: 14,
  },
  itemNutrition: {
    gap: 16,
  },
  keyNutrients: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
  },
  keyNutrientItem: {
    flex: 1,
    minWidth: 100,
    padding: 12,
    borderRadius: 8,
    borderWidth: 1,
    alignItems: 'center',
  },
  keyNutrientLabel: {
    fontSize: 12,
    marginBottom: 4,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
    textAlign: 'center',
  },
  keyNutrientValue: {
    fontSize: 16,
    fontWeight: '600',
    textAlign: 'center',
  },
  transcriptTitle: {
    fontSize: 24,
    fontWeight: '600',
    marginBottom: 16,
    letterSpacing: -0.01,
  },
  transcriptContainer: {
    padding: 24,
    borderRadius: 8,
    borderLeftWidth: 4,
    position: 'relative',
  },
  transcriptText: {
    fontSize: 18,
    lineHeight: 28,
    fontStyle: 'italic',
    letterSpacing: -0.005,
    paddingHorizontal: 16,
  },
});