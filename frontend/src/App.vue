<script setup>
import { RouterLink, RouterView } from "vue-router";
import { Analytics } from '@vercel/analytics/vue';
</script>

<template>
  <Analytics />
  <div class="space-twin-container">
    <!-- Grid Background Overlay -->
    <div class="grid-overlay"></div>

    <!-- Header / Navigation -->
    <header class="header">
      <div class="logo" data-text="SPACE_TWIN [BETA]">SPACE_TWIN [BETA]</div>
      <nav class="nav">
        <RouterLink to="/" class="nav-link">HOME</RouterLink>
        <RouterLink to="/upload" class="nav-link">UPLOAD</RouterLink>
        <RouterLink to="/test" class="nav-link">DEBUG</RouterLink>
        <a href="#login" class="nav-link">LOGIN_</a>
      </nav>
    </header>

    <!-- Main Content -->
    <main class="main-content">
      <RouterView />
    </main>
  </div>
</template>

<style>
/* Global Reset / Base Styles */
:root {
  --color-primary: #0f172a;
  --color-secondary: #334155;
  --color-cta: #0369a1;
  --color-text: #f8fafc;
  --color-accent: #00ff9d; /* Neon Green for status */
  --font-mono: "Space Mono", monospace;
}

body {
  margin: 0;
  padding: 0;
  background-color: var(--color-primary);
  color: var(--color-text);
  font-family: var(--font-mono);
}

a {
  text-decoration: none;
  color: inherit;
}
</style>

<style scoped>
.space-twin-container {
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
  background-color: var(--color-primary);
  color: var(--color-text);
  font-family: var(--font-mono);
}

/* Architectural Grid Background */
.grid-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-image:
    linear-gradient(rgba(51, 65, 85, 0.3) 1px, transparent 1px),
    linear-gradient(90deg, rgba(51, 65, 85, 0.3) 1px, transparent 1px);
  background-size: 40px 40px;
  pointer-events: none;
  z-index: 0;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 2rem 4rem;
  border-bottom: 1px solid #334155;
  position: relative;
  z-index: 10;
  background: rgba(15, 23, 42, 0.9);
  backdrop-filter: blur(5px);
}

.logo {
  font-weight: 700;
  font-size: 1.5rem;
  letter-spacing: -1px;
  position: relative;
  display: inline-block;
}

.logo::before,
.logo::after {
  content: attr(data-text);
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: var(--color-primary);
}

.logo::before {
  left: 2px;
  text-shadow: -1px 0 #ff00c1;
  clip-path: inset(44% 0 61% 0);
  animation: glitch-anim 2s infinite linear alternate-reverse;
}

.logo::after {
  left: -2px;
  text-shadow: -1px 0 #00fff9;
  clip-path: inset(58% 0 43% 0);
  animation: glitch-anim 2s infinite linear alternate-reverse;
}

@keyframes glitch-anim {
  0% {
    clip-path: inset(40% 0 61% 0);
  }
  20% {
    clip-path: inset(92% 0 1% 0);
  }
  40% {
    clip-path: inset(43% 0 1% 0);
  }
  60% {
    clip-path: inset(25% 0 58% 0);
  }
  80% {
    clip-path: inset(54% 0 7% 0);
  }
  100% {
    clip-path: inset(58% 0 43% 0);
  }
}

.nav-link {
  margin-left: 2rem;
  color: #94a3b8;
  text-decoration: none;
  font-size: 0.9rem;
  transition: color 0.2s;
}

.nav-link:hover,
.router-link-active {
  color: #f8fafc;
  text-decoration: underline;
  text-underline-offset: 4px;
}

.main-content {
  position: relative;
  z-index: 1;
  padding: 4rem;
  max-width: 1440px;
  margin: 0 auto;
}

/* Common Components */
:deep(.btn-primary) {
  display: inline-block;
  text-decoration: none;
  text-align: center;
  background-color: #0369a1;
  color: white;
  border: none;
  padding: 1rem 2rem;
  font-family: "Space Mono", monospace;
  font-weight: 700;
  font-size: 1rem;
  cursor: pointer;
  border: 1px solid #0369a1;
  transition: all 0.2s;
  text-transform: uppercase;
}

:deep(.btn-primary:hover) {
  background-color: #0284c7;
  color: white;
  transform: translateY(-2px);
  box-shadow: 4px 4px 0px rgba(0, 0, 0, 0.5);
}

:deep(.btn-secondary) {
  display: inline-block;
  text-decoration: none;
  text-align: center;
  background-color: transparent;
  color: #f8fafc;
  border: 1px solid #334155;
  padding: 1rem 2rem;
  font-family: "Space Mono", monospace;
  font-weight: 700;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

:deep(.btn-secondary:hover) {
  border-color: #f8fafc;
  background: rgba(255, 255, 255, 0.05);
}

@media (max-width: 768px) {
  .header {
    padding: 1rem 2rem;
    flex-direction: column;
    gap: 1rem;
  }

  .nav {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 1rem;
  }

  .nav-link {
    margin-left: 0;
  }

  .main-content {
    padding: 2rem;
  }
}
</style>
