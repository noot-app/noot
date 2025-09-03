<script lang="ts">
  import Label from "./Label.svelte"

  export let selectedColor: string = "#FFD700"
  export let previewName: string = "Preview"
  export let className: string = ""

  // Color palette (16 common colors)
  const colorPalette = [
    "FF0000", // Red
    "FF8C00", // Dark Orange  
    "FFD700", // Gold
    "74B986", // Green
    "1E90FF", // Blue
    "9B59B6", // Purple
    "9CA3AF", // Gray
    "2DD4BF", // Teal
    "A8E6CF", // Mint
    "F59E0B", // Orange
    "FF69B4", // Hot Pink
    "8A2BE2", // Blue Violet
    "00CED1", // Dark Turquoise
    "32CD32", // Lime Green
    "DC143C", // Crimson
    "4B0082", // Indigo
  ]

  function selectColor(color: string) {
    selectedColor = `#${color}`
  }

  function randomColor() {
    // Generate a random hex color (avoiding very light colors for readability)
    function randByte() {
      return Math.floor(Math.random() * 256)
    }
    function toHex(n: number) {
      return n.toString(16).padStart(2, "0")
    }
    function isTooLight(r: number, g: number, b: number) {
      // Perceived luminance (ITU-R BT.709)
      const luminance =
        0.2126 * (r / 255) + 0.7152 * (g / 255) + 0.0722 * (b / 255)
      return luminance > 0.85 // very light
    }

    let r = randByte(),
      g = randByte(),
      b = randByte()
    // Re-roll a couple times if too light to keep text-white readable
    for (let i = 0; i < 3 && isTooLight(r, g, b); i++) {
      r = randByte()
      g = randByte()
      b = randByte()
    }
    selectedColor = `#${toHex(r)}${toHex(g)}${toHex(b)}`
  }
</script>

<div class="form-control {className}">
  <!-- Color palette -->
  <div class="grid grid-cols-8 gap-2 mb-4">
    {#each colorPalette as color}
      <button
        type="button"
        class="w-8 h-8 rounded border-2 transition-all hover:scale-110"
        class:border-primary={selectedColor === `#${color}`}
        class:border-base-300={selectedColor !== `#${color}`}
        style="background-color: #{color}"
        aria-label={`Select color #${color}`}
        title={`Select color #${color}`}
        on:click={() => selectColor(color)}
      ></button>
    {/each}
  </div>

  <!-- Random color and manual input -->
  <div class="flex gap-2 mb-2">
    <button
      type="button"
      class="btn btn-sm btn-outline flex-1"
      on:click={randomColor}
    >
      🎲 Random Color
    </button>
    <input
      type="text"
      class="input input-bordered input-sm flex-1"
      bind:value={selectedColor}
      placeholder="#FFFFFF"
      maxlength={7}
    />
  </div>

  <!-- Color preview -->
  <div class="flex items-center gap-2 mt-2">
    <Label
      name={previewName}
      color={selectedColor}
      ariaLabel="Color preview"
    />
  </div>
</div>
