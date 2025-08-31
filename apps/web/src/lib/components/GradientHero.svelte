<script lang="ts">
	import { onMount } from 'svelte';

	let mounted = false;
	onMount(() => {
		mounted = true;
	});
</script>

<div class="gradient-hero-wrapper">
	<div class="gradient-hero" class:mounted>
		<div class="rising-sun"></div>
	</div>
	<div class="gradient-hero-content">
		<slot />
	</div>
</div>

<style>
	.gradient-hero-wrapper {
		position: relative;
		overflow: hidden;
		width: 100%;
		min-height: 100vh;
	}

	.gradient-hero {
		position: absolute;
		inset: 0;
		background: linear-gradient(
				to top,
				#d35400 0%,
				#e67e22 15%,
				#f39c12 35%,
				#f5c28b 50%,
				hsl(var(--b1)) 70%
			),
			linear-gradient(135deg, hsl(var(--b1)), hsl(var(--b2)));
		background-size: 100% 120%;
		z-index: 1;
		transition: opacity 1s ease-in-out;
		opacity: 0;
	}

	.gradient-hero.mounted {
		opacity: 1;
	}

	.gradient-hero::after {
		content: '';
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		height: 20vh;
		background: linear-gradient(
			to bottom,
			transparent 0%,
			rgba(255, 255, 255, 0.3) 50%,
			rgba(255, 255, 255, 0.8) 90%,
			hsl(var(--b1)) 100%
		);
		z-index: 3;
		pointer-events: none;
	}

	.rising-sun {
		position: absolute;
		bottom: -10vh;
		left: 0;
		right: 0;
		height: 150vh;
		background: radial-gradient(
				ellipse 60vw 50vh at center bottom,
				#ff3000 0%,
				#ff4500 10%,
				#ff6b35 20%,
				#ff8c42 35%,
				transparent 70%
			),
			radial-gradient(
				ellipse 50vw 45vh at center bottom,
				#d2691e 0%,
				#e67e22 5%,
				#f39c12 10%,
				transparent 60%
			),
			radial-gradient(
				ellipse 40vw 40vh at center bottom,
				#b22222 0%,
				#dc143c 3%,
				#e74c3c 5%,
				transparent 50%
			),
			radial-gradient(
				ellipse 80vw 65vh at center bottom,
				#e67e22 40%,
				#d35400 60%,
				transparent 90%
			);
		background-blend-mode: screen, multiply, overlay, normal;
		opacity: 0.6;
		z-index: -10;
		mask: linear-gradient(to bottom, black 0%, black 70%, transparent 95%);
		-webkit-mask: linear-gradient(to bottom, black 0%, black 70%, transparent 95%);
		pointer-events: none;
	}

	.rising-sun::after {
		content: '';
		position: absolute;
		inset: 0;
		background-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='80' height='80'><filter id='coarseGrain'><feTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='2' stitchTiles='stitch'/></filter><rect width='100%' height='100%' filter='url(%23coarseGrain)' opacity='1'/></svg>"),
			url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='120' height='120'><filter id='medGrain'><feTurbulence type='turbulence' baseFrequency='1.2' numOctaves='3' seed='7'/></filter><rect width='100%' height='100%' filter='url(%23medGrain)' opacity='0.9'/></svg>"),
			url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='60' height='60'><filter id='heavyGrain'><feTurbulence type='fractalNoise' baseFrequency='0.5' numOctaves='1' seed='3'/></filter><rect width='100%' height='100%' filter='url(%23heavyGrain)' opacity='1'/></svg>"),
			url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='40' height='40'><filter id='fineGrain'><feTurbulence type='turbulence' baseFrequency='1.8' numOctaves='2' seed='11'/></filter><rect width='100%' height='100%' filter='url(%23fineGrain)' opacity='0.7'/></svg>");
		background-size: 80px 80px, 120px 120px, 60px 60px, 40px 40px;
		opacity: 0.98;
		mix-blend-mode: hard-light;
		filter: contrast(1.15);
		pointer-events: none;
	}

	.gradient-hero-content {
		position: relative;
		z-index: 2;
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		justify-content: center;
	}

	@media (max-width: 768px) {
		.rising-sun {
			/* Rich sunrise on mobile, slightly toned down compared to <=420px */
			background:
				radial-gradient(ellipse 120vw 80vh at center bottom, #ff6b35 0%, #ff8c42 20%, transparent 62%),
				radial-gradient(ellipse 90vw 65vh at center bottom, #ff4500 0%, #ff6b35 20%, transparent 68%),
				radial-gradient(ellipse 55vw 45vh at center bottom, #ff3000 0%, #ff4500 10%, transparent 48%);
			background-blend-mode: screen, overlay, multiply;
			opacity: 0.9;
			/* Pull the sun up so warm hues are visible above the fold */
			bottom: 0vh;
			height: 200vh;
			/* Make the mask less aggressive on mobile so top colors show through */
			mask: linear-gradient(to bottom, black 0%, black 90%, transparent 100%);
			-webkit-mask: linear-gradient(to bottom, black 0%, black 90%, transparent 100%);
			will-change: transform, opacity;
			transform: translateZ(0);
		}

		/* Let the hero overlay be smaller on mobile so it doesn't wash out the sun */
		.gradient-hero::after {
			height: 8vh;
		}
	}

	/* Extra strong adjustments for very small phones */
	@media (max-width: 420px) {
		.rising-sun {
			/* Stronger, more orange/red gradients for sunrise-y look */
			background:
				radial-gradient(ellipse 130vw 90vh at center bottom, #ff6b35 0%, #ff8c42 18%, transparent 60%),
				radial-gradient(ellipse 100vw 70vh at center bottom, #ff4500 0%, #ff6b35 22%, #ff8c42 38%, transparent 70%),
				radial-gradient(ellipse 60vw 50vh at center bottom, #ff3000 0%, #ff4500 12%, transparent 45%);
			background-blend-mode: screen, overlay, multiply;
			opacity: 0.98; /* nearly full intensity */
			bottom: 60vh; /* pull even higher */
			height: 240vh; /* larger so the gradients span the viewport */
			/* make mask less harsh so top orange hues are visible */
			mask: linear-gradient(to bottom, black 0%, black 92%, transparent 100%);
			-webkit-mask: linear-gradient(to bottom, black 0%, black 92%, transparent 100%);
			will-change: transform, opacity;
			transform: translateZ(0);
		}
	}

</style>
