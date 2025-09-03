export type Json =
  | string
  | number
  | boolean
  | null
  | { [key: string]: Json | undefined }
  | Json[]

export type Database = {
  public: {
    Tables: {
      api_keys: {
        Row: {
          created_at: string | null
          expires_at: string | null
          hash: string
          id: string
          last_used_at: string | null
          name: string
          prefix: string
          revoked_at: string | null
          scope: string
          user_id: string
        }
        Insert: {
          created_at?: string | null
          expires_at?: string | null
          hash: string
          id?: string
          last_used_at?: string | null
          name: string
          prefix: string
          revoked_at?: string | null
          scope: string
          user_id: string
        }
        Update: {
          created_at?: string | null
          expires_at?: string | null
          hash?: string
          id?: string
          last_used_at?: string | null
          name?: string
          prefix?: string
          revoked_at?: string | null
          scope?: string
          user_id?: string
        }
        Relationships: [
          {
            foreignKeyName: "api_keys_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "profiles"
            referencedColumns: ["id"]
          },
        ]
      }
      consumption_item_labels: {
        Row: {
          consumption_item_id: string
          created_at: string
          label_id: string
        }
        Insert: {
          consumption_item_id: string
          created_at?: string
          label_id: string
        }
        Update: {
          consumption_item_id?: string
          created_at?: string
          label_id?: string
        }
        Relationships: [
          {
            foreignKeyName: "consumption_item_labels_consumption_item_id_fkey"
            columns: ["consumption_item_id"]
            isOneToOne: false
            referencedRelation: "consumption_items"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "consumption_item_labels_label_id_fkey"
            columns: ["label_id"]
            isOneToOne: false
            referencedRelation: "labels"
            referencedColumns: ["id"]
          },
        ]
      }
      consumption_items: {
        Row: {
          added_sugars_g: number
          alcohol_g: number
          biotin_mcg: number
          brand: string
          caffeine_mg: number
          calcium_mg: number
          calories: number
          chloride_mg: number
          cholesterol_mg: number
          choline_mg: number
          chromium_mcg: number
          consumption_id: string
          copper_mg: number
          created_at: string
          creatine_mg: number
          dietary_fiber_g: number
          fluoride_mg: number
          folate_mcg: number
          grams: number
          id: string
          ingredients: Json | null
          iodine_mcg: number
          iron_mg: number
          item_id: string | null
          magnesium_mg: number
          manganese_mg: number
          molybdenum_mcg: number
          monounsaturated_fat_g: number
          name: string
          niacin_mg: number
          note: string | null
          omega3_ala_g: number
          omega3_dha_g: number
          omega3_epa_g: number
          omega6_g: number
          pantothenic_acid_mg: number
          phosphorus_mg: number
          polyunsaturated_fat_g: number
          potassium_mg: number
          protein_g: number
          riboflavin_mg: number
          saturated_fat_g: number
          selenium_mcg: number
          sodium_mg: number
          thiamine_mg: number
          total_carbs_g: number
          total_fat_g: number
          total_sugars_g: number
          trans_fat_g: number
          updated_at: string
          url: string | null
          user_quantity: number | null
          user_unit: string | null
          vitamin_a_mcg: number
          vitamin_b12_mcg: number
          vitamin_b6_mg: number
          vitamin_c_mg: number
          vitamin_d_mcg: number
          vitamin_e_mg: number
          vitamin_k_mcg: number
          zinc_mg: number
        }
        Insert: {
          added_sugars_g?: number
          alcohol_g?: number
          biotin_mcg?: number
          brand?: string
          caffeine_mg?: number
          calcium_mg?: number
          calories?: number
          chloride_mg?: number
          cholesterol_mg?: number
          choline_mg?: number
          chromium_mcg?: number
          consumption_id: string
          copper_mg?: number
          created_at?: string
          creatine_mg?: number
          dietary_fiber_g?: number
          fluoride_mg?: number
          folate_mcg?: number
          grams?: number
          id?: string
          ingredients?: Json | null
          iodine_mcg?: number
          iron_mg?: number
          item_id?: string | null
          magnesium_mg?: number
          manganese_mg?: number
          molybdenum_mcg?: number
          monounsaturated_fat_g?: number
          name: string
          niacin_mg?: number
          note?: string | null
          omega3_ala_g?: number
          omega3_dha_g?: number
          omega3_epa_g?: number
          omega6_g?: number
          pantothenic_acid_mg?: number
          phosphorus_mg?: number
          polyunsaturated_fat_g?: number
          potassium_mg?: number
          protein_g?: number
          riboflavin_mg?: number
          saturated_fat_g?: number
          selenium_mcg?: number
          sodium_mg?: number
          thiamine_mg?: number
          total_carbs_g?: number
          total_fat_g?: number
          total_sugars_g?: number
          trans_fat_g?: number
          updated_at?: string
          url?: string | null
          user_quantity?: number | null
          user_unit?: string | null
          vitamin_a_mcg?: number
          vitamin_b12_mcg?: number
          vitamin_b6_mg?: number
          vitamin_c_mg?: number
          vitamin_d_mcg?: number
          vitamin_e_mg?: number
          vitamin_k_mcg?: number
          zinc_mg?: number
        }
        Update: {
          added_sugars_g?: number
          alcohol_g?: number
          biotin_mcg?: number
          brand?: string
          caffeine_mg?: number
          calcium_mg?: number
          calories?: number
          chloride_mg?: number
          cholesterol_mg?: number
          choline_mg?: number
          chromium_mcg?: number
          consumption_id?: string
          copper_mg?: number
          created_at?: string
          creatine_mg?: number
          dietary_fiber_g?: number
          fluoride_mg?: number
          folate_mcg?: number
          grams?: number
          id?: string
          ingredients?: Json | null
          iodine_mcg?: number
          iron_mg?: number
          item_id?: string | null
          magnesium_mg?: number
          manganese_mg?: number
          molybdenum_mcg?: number
          monounsaturated_fat_g?: number
          name?: string
          niacin_mg?: number
          note?: string | null
          omega3_ala_g?: number
          omega3_dha_g?: number
          omega3_epa_g?: number
          omega6_g?: number
          pantothenic_acid_mg?: number
          phosphorus_mg?: number
          polyunsaturated_fat_g?: number
          potassium_mg?: number
          protein_g?: number
          riboflavin_mg?: number
          saturated_fat_g?: number
          selenium_mcg?: number
          sodium_mg?: number
          thiamine_mg?: number
          total_carbs_g?: number
          total_fat_g?: number
          total_sugars_g?: number
          trans_fat_g?: number
          updated_at?: string
          url?: string | null
          user_quantity?: number | null
          user_unit?: string | null
          vitamin_a_mcg?: number
          vitamin_b12_mcg?: number
          vitamin_b6_mg?: number
          vitamin_c_mg?: number
          vitamin_d_mcg?: number
          vitamin_e_mg?: number
          vitamin_k_mcg?: number
          zinc_mg?: number
        }
        Relationships: [
          {
            foreignKeyName: "consumption_items_consumption_id_fkey"
            columns: ["consumption_id"]
            isOneToOne: false
            referencedRelation: "consumptions"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "consumption_items_item_id_fkey"
            columns: ["item_id"]
            isOneToOne: false
            referencedRelation: "items"
            referencedColumns: ["id"]
          },
        ]
      }
      consumption_labels: {
        Row: {
          consumption_id: string
          created_at: string
          label_id: string
        }
        Insert: {
          consumption_id: string
          created_at?: string
          label_id: string
        }
        Update: {
          consumption_id?: string
          created_at?: string
          label_id?: string
        }
        Relationships: [
          {
            foreignKeyName: "consumption_labels_consumption_id_fkey"
            columns: ["consumption_id"]
            isOneToOne: false
            referencedRelation: "consumptions"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "consumption_labels_label_id_fkey"
            columns: ["label_id"]
            isOneToOne: false
            referencedRelation: "labels"
            referencedColumns: ["id"]
          },
        ]
      }
      consumptions: {
        Row: {
          added_sugars_g: number
          alcohol_g: number
          biotin_mcg: number
          caffeine_mg: number
          calcium_mg: number
          chloride_mg: number
          cholesterol_mg: number
          choline_mg: number
          chromium_mcg: number
          copper_mg: number
          created_at: string
          creatine_mg: number
          dietary_fiber_g: number
          fluoride_mg: number
          folate_mcg: number
          id: string
          iodine_mcg: number
          iron_mg: number
          is_public: boolean
          magnesium_mg: number
          manganese_mg: number
          molybdenum_mcg: number
          monounsaturated_fat_g: number
          niacin_mg: number
          note: string | null
          omega3_ala_g: number
          omega3_dha_g: number
          omega3_epa_g: number
          omega6_g: number
          pantothenic_acid_mg: number
          phosphorus_mg: number
          polyunsaturated_fat_g: number
          potassium_mg: number
          riboflavin_mg: number
          saturated_fat_g: number
          selenium_mcg: number
          thiamine_mg: number
          total_calories: number
          total_carbs_g: number
          total_fat_g: number
          total_protein_g: number
          total_sodium_mg: number
          total_sugars_g: number
          trans_fat_g: number
          transcript: string
          updated_at: string | null
          user_id: string
          vitamin_a_mcg: number
          vitamin_b12_mcg: number
          vitamin_b6_mg: number
          vitamin_c_mg: number
          vitamin_d_mcg: number
          vitamin_e_mg: number
          vitamin_k_mcg: number
          zinc_mg: number
        }
        Insert: {
          added_sugars_g?: number
          alcohol_g?: number
          biotin_mcg?: number
          caffeine_mg?: number
          calcium_mg?: number
          chloride_mg?: number
          cholesterol_mg?: number
          choline_mg?: number
          chromium_mcg?: number
          copper_mg?: number
          created_at?: string
          creatine_mg?: number
          dietary_fiber_g?: number
          fluoride_mg?: number
          folate_mcg?: number
          id?: string
          iodine_mcg?: number
          iron_mg?: number
          is_public?: boolean
          magnesium_mg?: number
          manganese_mg?: number
          molybdenum_mcg?: number
          monounsaturated_fat_g?: number
          niacin_mg?: number
          note?: string | null
          omega3_ala_g?: number
          omega3_dha_g?: number
          omega3_epa_g?: number
          omega6_g?: number
          pantothenic_acid_mg?: number
          phosphorus_mg?: number
          polyunsaturated_fat_g?: number
          potassium_mg?: number
          riboflavin_mg?: number
          saturated_fat_g?: number
          selenium_mcg?: number
          thiamine_mg?: number
          total_calories?: number
          total_carbs_g?: number
          total_fat_g?: number
          total_protein_g?: number
          total_sodium_mg?: number
          total_sugars_g?: number
          trans_fat_g?: number
          transcript: string
          updated_at?: string | null
          user_id: string
          vitamin_a_mcg?: number
          vitamin_b12_mcg?: number
          vitamin_b6_mg?: number
          vitamin_c_mg?: number
          vitamin_d_mcg?: number
          vitamin_e_mg?: number
          vitamin_k_mcg?: number
          zinc_mg?: number
        }
        Update: {
          added_sugars_g?: number
          alcohol_g?: number
          biotin_mcg?: number
          caffeine_mg?: number
          calcium_mg?: number
          chloride_mg?: number
          cholesterol_mg?: number
          choline_mg?: number
          chromium_mcg?: number
          copper_mg?: number
          created_at?: string
          creatine_mg?: number
          dietary_fiber_g?: number
          fluoride_mg?: number
          folate_mcg?: number
          id?: string
          iodine_mcg?: number
          iron_mg?: number
          is_public?: boolean
          magnesium_mg?: number
          manganese_mg?: number
          molybdenum_mcg?: number
          monounsaturated_fat_g?: number
          niacin_mg?: number
          note?: string | null
          omega3_ala_g?: number
          omega3_dha_g?: number
          omega3_epa_g?: number
          omega6_g?: number
          pantothenic_acid_mg?: number
          phosphorus_mg?: number
          polyunsaturated_fat_g?: number
          potassium_mg?: number
          riboflavin_mg?: number
          saturated_fat_g?: number
          selenium_mcg?: number
          thiamine_mg?: number
          total_calories?: number
          total_carbs_g?: number
          total_fat_g?: number
          total_protein_g?: number
          total_sodium_mg?: number
          total_sugars_g?: number
          trans_fat_g?: number
          transcript?: string
          updated_at?: string | null
          user_id?: string
          vitamin_a_mcg?: number
          vitamin_b12_mcg?: number
          vitamin_b6_mg?: number
          vitamin_c_mg?: number
          vitamin_d_mcg?: number
          vitamin_e_mg?: number
          vitamin_k_mcg?: number
          zinc_mg?: number
        }
        Relationships: [
          {
            foreignKeyName: "consumptions_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "profiles"
            referencedColumns: ["id"]
          },
        ]
      }
      event_labels: {
        Row: {
          created_at: string
          event_id: string
          label_id: string
        }
        Insert: {
          created_at?: string
          event_id: string
          label_id: string
        }
        Update: {
          created_at?: string
          event_id?: string
          label_id?: string
        }
        Relationships: [
          {
            foreignKeyName: "event_labels_event_id_fkey"
            columns: ["event_id"]
            isOneToOne: false
            referencedRelation: "events"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "event_labels_label_id_fkey"
            columns: ["label_id"]
            isOneToOne: false
            referencedRelation: "labels"
            referencedColumns: ["id"]
          },
        ]
      }
      event_links: {
        Row: {
          consumption_id: string | null
          consumption_item_id: string | null
          created_at: string
          event_id: string
          id: string
        }
        Insert: {
          consumption_id?: string | null
          consumption_item_id?: string | null
          created_at?: string
          event_id: string
          id?: string
        }
        Update: {
          consumption_id?: string | null
          consumption_item_id?: string | null
          created_at?: string
          event_id?: string
          id?: string
        }
        Relationships: [
          {
            foreignKeyName: "event_links_consumption_id_fkey"
            columns: ["consumption_id"]
            isOneToOne: false
            referencedRelation: "consumptions"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "event_links_consumption_item_id_fkey"
            columns: ["consumption_item_id"]
            isOneToOne: false
            referencedRelation: "consumption_items"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "event_links_event_id_fkey"
            columns: ["event_id"]
            isOneToOne: false
            referencedRelation: "events"
            referencedColumns: ["id"]
          },
        ]
      }
      event_types: {
        Row: {
          color: string
          created_at: string
          default_name: string | null
          description: string | null
          id: string
          name: string
          updated_at: string
          user_id: string
        }
        Insert: {
          color: string
          created_at?: string
          default_name?: string | null
          description?: string | null
          id?: string
          name: string
          updated_at?: string
          user_id: string
        }
        Update: {
          color?: string
          created_at?: string
          default_name?: string | null
          description?: string | null
          id?: string
          name?: string
          updated_at?: string
          user_id?: string
        }
        Relationships: [
          {
            foreignKeyName: "event_types_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "profiles"
            referencedColumns: ["id"]
          },
        ]
      }
      events: {
        Row: {
          color: string | null
          created_at: string
          ended_at: string | null
          event_type_id: string | null
          id: string
          level: number | null
          name: string
          note: string | null
          started_at: string
          updated_at: string
          user_id: string
        }
        Insert: {
          color?: string | null
          created_at?: string
          ended_at?: string | null
          event_type_id?: string | null
          id?: string
          level?: number | null
          name: string
          note?: string | null
          started_at: string
          updated_at?: string
          user_id: string
        }
        Update: {
          color?: string | null
          created_at?: string
          ended_at?: string | null
          event_type_id?: string | null
          id?: string
          level?: number | null
          name?: string
          note?: string | null
          started_at?: string
          updated_at?: string
          user_id?: string
        }
        Relationships: [
          {
            foreignKeyName: "events_event_type_id_fkey"
            columns: ["event_type_id"]
            isOneToOne: false
            referencedRelation: "event_types"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "events_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "profiles"
            referencedColumns: ["id"]
          },
        ]
      }
      item_aliases: {
        Row: {
          alias_brand: string
          alias_name: string
          canonical_brand: string
          canonical_name: string
          created_at: string
          id: string
        }
        Insert: {
          alias_brand?: string
          alias_name: string
          canonical_brand?: string
          canonical_name: string
          created_at?: string
          id?: string
        }
        Update: {
          alias_brand?: string
          alias_name?: string
          canonical_brand?: string
          canonical_name?: string
          created_at?: string
          id?: string
        }
        Relationships: []
      }
      items: {
        Row: {
          added_sugars_g_per_100g: number
          alcohol_g_per_100g: number
          biotin_mcg_per_100g: number
          caffeine_mg_per_100g: number
          calcium_mg_per_100g: number
          calories_per_100g: number
          chloride_mg_per_100g: number
          cholesterol_mg_per_100g: number
          choline_mg_per_100g: number
          chromium_mcg_per_100g: number
          copper_mg_per_100g: number
          created_at: string
          creatine_mg_per_100g: number
          dietary_fiber_g_per_100g: number
          display_brand: string | null
          display_name: string
          fluoride_mg_per_100g: number
          folate_mcg_per_100g: number
          id: string
          ingredients: Json | null
          iodine_mcg_per_100g: number
          iron_mg_per_100g: number
          label: string | null
          magnesium_mg_per_100g: number
          manganese_mg_per_100g: number
          molybdenum_mcg_per_100g: number
          monounsaturated_fat_g_per_100g: number
          niacin_mg_per_100g: number
          normalized_brand: string
          normalized_name: string
          note: string | null
          omega3_ala_g_per_100g: number
          omega3_dha_g_per_100g: number
          omega3_epa_g_per_100g: number
          omega6_g_per_100g: number
          original_added_sugars_g: number | null
          original_alcohol_g: number | null
          original_biotin_mcg: number | null
          original_caffeine_mg: number | null
          original_calcium_mg: number | null
          original_calories: number | null
          original_chloride_mg: number | null
          original_cholesterol_mg: number | null
          original_choline_mg: number | null
          original_chromium_mcg: number | null
          original_copper_mg: number | null
          original_creatine_mg: number | null
          original_dietary_fiber_g: number | null
          original_fluoride_mg: number | null
          original_folate_mcg: number | null
          original_iodine_mcg: number | null
          original_iron_mg: number | null
          original_magnesium_mg: number | null
          original_manganese_mg: number | null
          original_molybdenum_mcg: number | null
          original_monounsaturated_fat_g: number | null
          original_niacin_mg: number | null
          original_omega3_ala_g: number | null
          original_omega3_dha_g: number | null
          original_omega3_epa_g: number | null
          original_omega6_g: number | null
          original_pantothenic_acid_mg: number | null
          original_phosphorus_mg: number | null
          original_polyunsaturated_fat_g: number | null
          original_potassium_mg: number | null
          original_protein_g: number | null
          original_riboflavin_mg: number | null
          original_saturated_fat_g: number | null
          original_selenium_mcg: number | null
          original_serving_grams: number | null
          original_sodium_mg: number | null
          original_thiamine_mg: number | null
          original_total_carbs_g: number | null
          original_total_fat_g: number | null
          original_total_sugars_g: number | null
          original_trans_fat_g: number | null
          original_vitamin_a_mcg: number | null
          original_vitamin_b12_mcg: number | null
          original_vitamin_b6_mg: number | null
          original_vitamin_c_mg: number | null
          original_vitamin_d_mcg: number | null
          original_vitamin_e_mg: number | null
          original_vitamin_k_mcg: number | null
          original_zinc_mg: number | null
          pantothenic_acid_mg_per_100g: number
          phosphorus_mg_per_100g: number
          polyunsaturated_fat_g_per_100g: number
          potassium_mg_per_100g: number
          protein_g_per_100g: number
          riboflavin_mg_per_100g: number
          saturated_fat_g_per_100g: number
          selenium_mcg_per_100g: number
          sodium_mg_per_100g: number
          thiamine_mg_per_100g: number
          total_carbs_g_per_100g: number
          total_fat_g_per_100g: number
          total_sugars_g_per_100g: number
          trans_fat_g_per_100g: number
          updated_at: string
          url: string | null
          vitamin_a_mcg_per_100g: number
          vitamin_b12_mcg_per_100g: number
          vitamin_b6_mg_per_100g: number
          vitamin_c_mg_per_100g: number
          vitamin_d_mcg_per_100g: number
          vitamin_e_mg_per_100g: number
          vitamin_k_mcg_per_100g: number
          zinc_mg_per_100g: number
        }
        Insert: {
          added_sugars_g_per_100g?: number
          alcohol_g_per_100g?: number
          biotin_mcg_per_100g?: number
          caffeine_mg_per_100g?: number
          calcium_mg_per_100g?: number
          calories_per_100g?: number
          chloride_mg_per_100g?: number
          cholesterol_mg_per_100g?: number
          choline_mg_per_100g?: number
          chromium_mcg_per_100g?: number
          copper_mg_per_100g?: number
          created_at?: string
          creatine_mg_per_100g?: number
          dietary_fiber_g_per_100g?: number
          display_brand?: string | null
          display_name: string
          fluoride_mg_per_100g?: number
          folate_mcg_per_100g?: number
          id?: string
          ingredients?: Json | null
          iodine_mcg_per_100g?: number
          iron_mg_per_100g?: number
          label?: string | null
          magnesium_mg_per_100g?: number
          manganese_mg_per_100g?: number
          molybdenum_mcg_per_100g?: number
          monounsaturated_fat_g_per_100g?: number
          niacin_mg_per_100g?: number
          normalized_brand?: string
          normalized_name: string
          note?: string | null
          omega3_ala_g_per_100g?: number
          omega3_dha_g_per_100g?: number
          omega3_epa_g_per_100g?: number
          omega6_g_per_100g?: number
          original_added_sugars_g?: number | null
          original_alcohol_g?: number | null
          original_biotin_mcg?: number | null
          original_caffeine_mg?: number | null
          original_calcium_mg?: number | null
          original_calories?: number | null
          original_chloride_mg?: number | null
          original_cholesterol_mg?: number | null
          original_choline_mg?: number | null
          original_chromium_mcg?: number | null
          original_copper_mg?: number | null
          original_creatine_mg?: number | null
          original_dietary_fiber_g?: number | null
          original_fluoride_mg?: number | null
          original_folate_mcg?: number | null
          original_iodine_mcg?: number | null
          original_iron_mg?: number | null
          original_magnesium_mg?: number | null
          original_manganese_mg?: number | null
          original_molybdenum_mcg?: number | null
          original_monounsaturated_fat_g?: number | null
          original_niacin_mg?: number | null
          original_omega3_ala_g?: number | null
          original_omega3_dha_g?: number | null
          original_omega3_epa_g?: number | null
          original_omega6_g?: number | null
          original_pantothenic_acid_mg?: number | null
          original_phosphorus_mg?: number | null
          original_polyunsaturated_fat_g?: number | null
          original_potassium_mg?: number | null
          original_protein_g?: number | null
          original_riboflavin_mg?: number | null
          original_saturated_fat_g?: number | null
          original_selenium_mcg?: number | null
          original_serving_grams?: number | null
          original_sodium_mg?: number | null
          original_thiamine_mg?: number | null
          original_total_carbs_g?: number | null
          original_total_fat_g?: number | null
          original_total_sugars_g?: number | null
          original_trans_fat_g?: number | null
          original_vitamin_a_mcg?: number | null
          original_vitamin_b12_mcg?: number | null
          original_vitamin_b6_mg?: number | null
          original_vitamin_c_mg?: number | null
          original_vitamin_d_mcg?: number | null
          original_vitamin_e_mg?: number | null
          original_vitamin_k_mcg?: number | null
          original_zinc_mg?: number | null
          pantothenic_acid_mg_per_100g?: number
          phosphorus_mg_per_100g?: number
          polyunsaturated_fat_g_per_100g?: number
          potassium_mg_per_100g?: number
          protein_g_per_100g?: number
          riboflavin_mg_per_100g?: number
          saturated_fat_g_per_100g?: number
          selenium_mcg_per_100g?: number
          sodium_mg_per_100g?: number
          thiamine_mg_per_100g?: number
          total_carbs_g_per_100g?: number
          total_fat_g_per_100g?: number
          total_sugars_g_per_100g?: number
          trans_fat_g_per_100g?: number
          updated_at?: string
          url?: string | null
          vitamin_a_mcg_per_100g?: number
          vitamin_b12_mcg_per_100g?: number
          vitamin_b6_mg_per_100g?: number
          vitamin_c_mg_per_100g?: number
          vitamin_d_mcg_per_100g?: number
          vitamin_e_mg_per_100g?: number
          vitamin_k_mcg_per_100g?: number
          zinc_mg_per_100g?: number
        }
        Update: {
          added_sugars_g_per_100g?: number
          alcohol_g_per_100g?: number
          biotin_mcg_per_100g?: number
          caffeine_mg_per_100g?: number
          calcium_mg_per_100g?: number
          calories_per_100g?: number
          chloride_mg_per_100g?: number
          cholesterol_mg_per_100g?: number
          choline_mg_per_100g?: number
          chromium_mcg_per_100g?: number
          copper_mg_per_100g?: number
          created_at?: string
          creatine_mg_per_100g?: number
          dietary_fiber_g_per_100g?: number
          display_brand?: string | null
          display_name?: string
          fluoride_mg_per_100g?: number
          folate_mcg_per_100g?: number
          id?: string
          ingredients?: Json | null
          iodine_mcg_per_100g?: number
          iron_mg_per_100g?: number
          label?: string | null
          magnesium_mg_per_100g?: number
          manganese_mg_per_100g?: number
          molybdenum_mcg_per_100g?: number
          monounsaturated_fat_g_per_100g?: number
          niacin_mg_per_100g?: number
          normalized_brand?: string
          normalized_name?: string
          note?: string | null
          omega3_ala_g_per_100g?: number
          omega3_dha_g_per_100g?: number
          omega3_epa_g_per_100g?: number
          omega6_g_per_100g?: number
          original_added_sugars_g?: number | null
          original_alcohol_g?: number | null
          original_biotin_mcg?: number | null
          original_caffeine_mg?: number | null
          original_calcium_mg?: number | null
          original_calories?: number | null
          original_chloride_mg?: number | null
          original_cholesterol_mg?: number | null
          original_choline_mg?: number | null
          original_chromium_mcg?: number | null
          original_copper_mg?: number | null
          original_creatine_mg?: number | null
          original_dietary_fiber_g?: number | null
          original_fluoride_mg?: number | null
          original_folate_mcg?: number | null
          original_iodine_mcg?: number | null
          original_iron_mg?: number | null
          original_magnesium_mg?: number | null
          original_manganese_mg?: number | null
          original_molybdenum_mcg?: number | null
          original_monounsaturated_fat_g?: number | null
          original_niacin_mg?: number | null
          original_omega3_ala_g?: number | null
          original_omega3_dha_g?: number | null
          original_omega3_epa_g?: number | null
          original_omega6_g?: number | null
          original_pantothenic_acid_mg?: number | null
          original_phosphorus_mg?: number | null
          original_polyunsaturated_fat_g?: number | null
          original_potassium_mg?: number | null
          original_protein_g?: number | null
          original_riboflavin_mg?: number | null
          original_saturated_fat_g?: number | null
          original_selenium_mcg?: number | null
          original_serving_grams?: number | null
          original_sodium_mg?: number | null
          original_thiamine_mg?: number | null
          original_total_carbs_g?: number | null
          original_total_fat_g?: number | null
          original_total_sugars_g?: number | null
          original_trans_fat_g?: number | null
          original_vitamin_a_mcg?: number | null
          original_vitamin_b12_mcg?: number | null
          original_vitamin_b6_mg?: number | null
          original_vitamin_c_mg?: number | null
          original_vitamin_d_mcg?: number | null
          original_vitamin_e_mg?: number | null
          original_vitamin_k_mcg?: number | null
          original_zinc_mg?: number | null
          pantothenic_acid_mg_per_100g?: number
          phosphorus_mg_per_100g?: number
          polyunsaturated_fat_g_per_100g?: number
          potassium_mg_per_100g?: number
          protein_g_per_100g?: number
          riboflavin_mg_per_100g?: number
          saturated_fat_g_per_100g?: number
          selenium_mcg_per_100g?: number
          sodium_mg_per_100g?: number
          thiamine_mg_per_100g?: number
          total_carbs_g_per_100g?: number
          total_fat_g_per_100g?: number
          total_sugars_g_per_100g?: number
          trans_fat_g_per_100g?: number
          updated_at?: string
          url?: string | null
          vitamin_a_mcg_per_100g?: number
          vitamin_b12_mcg_per_100g?: number
          vitamin_b6_mg_per_100g?: number
          vitamin_c_mg_per_100g?: number
          vitamin_d_mcg_per_100g?: number
          vitamin_e_mg_per_100g?: number
          vitamin_k_mcg_per_100g?: number
          zinc_mg_per_100g?: number
        }
        Relationships: []
      }
      labels: {
        Row: {
          color: string
          created_at: string
          description: string | null
          id: string
          name: string
          updated_at: string
          user_id: string
        }
        Insert: {
          color: string
          created_at?: string
          description?: string | null
          id?: string
          name: string
          updated_at?: string
          user_id: string
        }
        Update: {
          color?: string
          created_at?: string
          description?: string | null
          id?: string
          name?: string
          updated_at?: string
          user_id?: string
        }
        Relationships: [
          {
            foreignKeyName: "labels_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "profiles"
            referencedColumns: ["id"]
          },
        ]
      }
      profiles: {
        Row: {
          active_goal_name: string | null
          avatar_url: string | null
          created_at: string | null
          email: string
          full_name: string | null
          handle: string | null
          id: string
          subscription_tier: string
          updated_at: string | null
        }
        Insert: {
          active_goal_name?: string | null
          avatar_url?: string | null
          created_at?: string | null
          email: string
          full_name?: string | null
          handle?: string | null
          id: string
          subscription_tier?: string
          updated_at?: string | null
        }
        Update: {
          active_goal_name?: string | null
          avatar_url?: string | null
          created_at?: string | null
          email?: string
          full_name?: string | null
          handle?: string | null
          id?: string
          subscription_tier?: string
          updated_at?: string | null
        }
        Relationships: []
      }
      user_biometrics: {
        Row: {
          activity_level: string | null
          birth_date: string | null
          created_at: string
          height_cm: number | null
          id: string
          sex: string | null
          updated_at: string
          user_id: string
          weight_kg: number | null
        }
        Insert: {
          activity_level?: string | null
          birth_date?: string | null
          created_at?: string
          height_cm?: number | null
          id?: string
          sex?: string | null
          updated_at?: string
          user_id: string
          weight_kg?: number | null
        }
        Update: {
          activity_level?: string | null
          birth_date?: string | null
          created_at?: string
          height_cm?: number | null
          id?: string
          sex?: string | null
          updated_at?: string
          user_id?: string
          weight_kg?: number | null
        }
        Relationships: [
          {
            foreignKeyName: "user_biometrics_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: true
            referencedRelation: "profiles"
            referencedColumns: ["id"]
          },
        ]
      }
      user_goals: {
        Row: {
          created_at: string
          id: string
          name: string
          overrides_json: string
          updated_at: string
          user_id: string
        }
        Insert: {
          created_at?: string
          id?: string
          name?: string
          overrides_json: string
          updated_at?: string
          user_id: string
        }
        Update: {
          created_at?: string
          id?: string
          name?: string
          overrides_json?: string
          updated_at?: string
          user_id?: string
        }
        Relationships: [
          {
            foreignKeyName: "user_goals_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "profiles"
            referencedColumns: ["id"]
          },
        ]
      }
    }
    Views: {
      [_ in never]: never
    }
    Functions: {
      [_ in never]: never
    }
    Enums: {
      [_ in never]: never
    }
    CompositeTypes: {
      [_ in never]: never
    }
  }
}

type DatabaseWithoutInternals = Omit<Database, "__InternalSupabase">

type DefaultSchema = DatabaseWithoutInternals[Extract<keyof Database, "public">]

export type Tables<
  DefaultSchemaTableNameOrOptions extends
    | keyof (DefaultSchema["Tables"] & DefaultSchema["Views"])
    | { schema: keyof DatabaseWithoutInternals },
  TableName extends DefaultSchemaTableNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof (DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"] &
        DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Views"])
    : never = never,
> = DefaultSchemaTableNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? (DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"] &
      DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Views"])[TableName] extends {
      Row: infer R
    }
    ? R
    : never
  : DefaultSchemaTableNameOrOptions extends keyof (DefaultSchema["Tables"] &
        DefaultSchema["Views"])
    ? (DefaultSchema["Tables"] &
        DefaultSchema["Views"])[DefaultSchemaTableNameOrOptions] extends {
        Row: infer R
      }
      ? R
      : never
    : never

export type TablesInsert<
  DefaultSchemaTableNameOrOptions extends
    | keyof DefaultSchema["Tables"]
    | { schema: keyof DatabaseWithoutInternals },
  TableName extends DefaultSchemaTableNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"]
    : never = never,
> = DefaultSchemaTableNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"][TableName] extends {
      Insert: infer I
    }
    ? I
    : never
  : DefaultSchemaTableNameOrOptions extends keyof DefaultSchema["Tables"]
    ? DefaultSchema["Tables"][DefaultSchemaTableNameOrOptions] extends {
        Insert: infer I
      }
      ? I
      : never
    : never

export type TablesUpdate<
  DefaultSchemaTableNameOrOptions extends
    | keyof DefaultSchema["Tables"]
    | { schema: keyof DatabaseWithoutInternals },
  TableName extends DefaultSchemaTableNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"]
    : never = never,
> = DefaultSchemaTableNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"][TableName] extends {
      Update: infer U
    }
    ? U
    : never
  : DefaultSchemaTableNameOrOptions extends keyof DefaultSchema["Tables"]
    ? DefaultSchema["Tables"][DefaultSchemaTableNameOrOptions] extends {
        Update: infer U
      }
      ? U
      : never
    : never

export type Enums<
  DefaultSchemaEnumNameOrOptions extends
    | keyof DefaultSchema["Enums"]
    | { schema: keyof DatabaseWithoutInternals },
  EnumName extends DefaultSchemaEnumNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[DefaultSchemaEnumNameOrOptions["schema"]]["Enums"]
    : never = never,
> = DefaultSchemaEnumNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[DefaultSchemaEnumNameOrOptions["schema"]]["Enums"][EnumName]
  : DefaultSchemaEnumNameOrOptions extends keyof DefaultSchema["Enums"]
    ? DefaultSchema["Enums"][DefaultSchemaEnumNameOrOptions]
    : never

export type CompositeTypes<
  PublicCompositeTypeNameOrOptions extends
    | keyof DefaultSchema["CompositeTypes"]
    | { schema: keyof DatabaseWithoutInternals },
  CompositeTypeName extends PublicCompositeTypeNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[PublicCompositeTypeNameOrOptions["schema"]]["CompositeTypes"]
    : never = never,
> = PublicCompositeTypeNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[PublicCompositeTypeNameOrOptions["schema"]]["CompositeTypes"][CompositeTypeName]
  : PublicCompositeTypeNameOrOptions extends keyof DefaultSchema["CompositeTypes"]
    ? DefaultSchema["CompositeTypes"][PublicCompositeTypeNameOrOptions]
    : never

export const Constants = {
  public: {
    Enums: {},
  },
} as const

