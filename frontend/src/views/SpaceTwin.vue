<template>
  <div>
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content">
        <div class="status-indicator">
          <span class="blinking-dot"></span> SYSTEM_ONLINE
        </div>
        <h1 class="hero-title">
          ARCHITECTURAL<br />
          INTELLIGENCE<br />
          <span class="highlight">UNLEASHED</span>
        </h1>
        <p class="hero-subtitle">
          // CONVERT 2D BLUEPRINTS TO 3D ENVIRONMENTS.<br />
          // REAL-TIME AI LABELING.<br />
          // STRUCTURAL TRANSPARENCY.
        </p>

        <div class="cta-group">
          <button class="btn-primary">START_TWIN_ENGINE_</button>
          <button class="btn-secondary">VIEW_DEMO_LOGS</button>
        </div>
      </div>

      <!-- Animation / Visual -->
      <div class="hero-visual">
        <div class="blueprint-card">
          <div class="scan-line"></div>
          <div class="blueprint-content">
            <div class="room-label" style="top: 20%; left: 30%">
              KITCHEN [DETECTED]
            </div>
            <div class="room-label" style="top: 60%; left: 70%">
              LIVING [DETECTED]
            </div>
            <svg viewBox="0 0 100 100" class="wireframe">
              <rect
                x="10"
                y="10"
                width="80"
                height="80"
                fill="none"
                stroke="currentColor"
                stroke-width="0.5"
              />
              <line
                x1="10"
                y1="50"
                x2="90"
                y2="50"
                stroke="currentColor"
                stroke-width="0.5"
              />
              <line
                x1="50"
                y1="10"
                x2="50"
                y2="90"
                stroke="currentColor"
                stroke-width="0.5"
              />
            </svg>
          </div>
        </div>
      </div>
    </section>

    <!-- Data Dashboard / Feature Strip -->
    <section class="data-strip">
      <div class="data-item">
        <span class="data-label">LATENCY</span>
        <span class="data-value">12ms</span>
      </div>
      <div class="data-item">
        <span class="data-label">ACCURACY</span>
        <span class="data-value">99.8%</span>
      </div>
      <div class="data-item">
        <span class="data-label">MODELS_GENERATED</span>
        <span class="data-value">8,421</span>
      </div>
      <div class="data-item">
        <span class="data-label">SERVER_STATUS</span>
        <span class="data-value status-ok">OPERATIONAL</span>
      </div>
    </section>
  </div>
</template>

<script setup>
// No script logic needed for this static landing page yet
</script>

<style scoped>
.hero-section {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4rem;
  align-items: center;
  min-height: 60vh;
}

.status-indicator {
  display: flex;
  align-items: center;
  font-size: 0.8rem;
  color: #00ff9d;
  margin-bottom: 1rem;
}

.blinking-dot {
  width: 8px;
  height: 8px;
  background-color: #00ff9d;
  border-radius: 50%;
  margin-right: 8px;
  animation: blink 1.5s infinite;
}

@keyframes blink {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.3;
  }
  100% {
    opacity: 1;
  }
}

.hero-title {
  font-size: 4rem;
  line-height: 1.1;
  margin-bottom: 1.5rem;
  font-weight: 700;
  letter-spacing: -2px;
}

.hero-title .highlight {
  color: #0369a1;
  background: rgba(3, 105, 161, 0.1);
  padding: 0 10px;
}

.hero-subtitle {
  font-size: 1.1rem;
  color: #94a3b8;
  margin-bottom: 2.5rem;
  line-height: 1.6;
}

.cta-group {
  display: flex;
  gap: 1rem;
}

/* Hero Visual Animation */
.hero-visual {
  position: relative;
  height: 400px;
  border: 1px solid #334155;
  background: rgba(15, 23, 42, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.blueprint-card {
  width: 80%;
  height: 80%;
  border: 1px dashed #334155;
  position: relative;
}

.scan-line {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 2px;
  background: #00ff9d;
  box-shadow: 0 0 10px #00ff9d;
  animation: scan 3s linear infinite;
  z-index: 5;
}

@keyframes scan {
  0% {
    top: 0;
  }
  100% {
    top: 100%;
  }
}

.room-label {
  position: absolute;
  background: #0369a1;
  color: white;
  padding: 2px 6px;
  font-size: 0.7rem;
  transform: translate(-50%, -50%);
}

.wireframe {
  width: 100%;
  height: 100%;
  color: #334155;
}

/* Data Strip */
.data-strip {
  margin-top: 4rem;
  border-top: 1px solid #334155;
  padding-top: 2rem;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 2rem;
}

.data-item {
  display: flex;
  flex-direction: column;
}

.data-label {
  font-size: 0.8rem;
  color: #64748b;
  margin-bottom: 0.5rem;
}

.data-value {
  font-size: 1.5rem;
  font-weight: 700;
}

.status-ok {
  color: #00ff9d;
}

/* Responsive */
@media (max-width: 1024px) {
  .hero-section {
    grid-template-columns: 1fr;
    text-align: center;
  }

  .hero-content {
    order: 2;
  }

  .hero-visual {
    order: 1;
    margin-bottom: 2rem;
  }

  .status-indicator {
    justify-content: center;
  }
}
</style>
