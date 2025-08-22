# MuseoModerno Font Usage Guide

## Available Fonts

- `MuseoModerno-VariableFont_wght.ttf` - Regular (100-900 weight range)
- `MuseoModerno-Italic-VariableFont_wght.ttf` - Italic (100-900 weight range)

## CSS Classes

- `.font-museo` - Applies MuseoModerno font family
- `.noot-logo` - Pre-styled NOOT logo class with primary color and bold weight

## Usage Examples

### HTML/Svelte

```svelte
<!-- Logo usage -->
<span class="text-2xl noot-logo">NOOT</span>

<!-- Custom weights with MuseoModerno -->
<h1 class="font-museo" style="font-weight: 100;">Thin</h1>
<h1 class="font-museo" style="font-weight: 300;">Light</h1>
<h1 class="font-museo" style="font-weight: 400;">Regular</h1>
<h1 class="font-museo" style="font-weight: 600;">Semi-Bold</h1>
<h1 class="font-museo" style="font-weight: 700;">Bold</h1>
<h1 class="font-museo" style="font-weight: 900;">Black</h1>

<!-- Italic usage -->
<span class="font-museo italic" style="font-weight: 500;">Italic Medium</span>
```

### CSS

```css
.custom-heading {
  font-family: 'MuseoModerno', sans-serif;
  font-weight: 650; /* Any value between 100-900 works! */
  font-style: normal;
}

.custom-italic {
  font-family: 'MuseoModerno', sans-serif;
  font-weight: 400;
  font-style: italic;
}
```

## Benefits of Variable Fonts

- **Smaller bundle**: 2 files (392KB) vs 18 static files (~1.8MB)
- **Infinite weights**: Use any weight between 100-900, not just predefined steps
- **Better performance**: Fewer HTTP requests, faster loading
- **Future-proof**: Excellent browser support for variable fonts
