<script setup>
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <main class="app-container">
    <header>
      <h1>SmartFloor Digital Twin</h1>
      <nav>
        <RouterLink to="/" class="nav-link">Dashboard</RouterLink>
        <RouterLink to="/test" class="nav-link">Debug Crop</RouterLink>
      </nav>
    </header>
    
    <RouterView />
  </main>
</template>

<style scoped>
.app-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
  font-family: 'Inter', sans-serif;
}

header {
  text-align: center;
  margin-bottom: 2rem;
}

h1 {
  color: #2c3e50;
  font-weight: 700;
  margin-bottom: 1rem;
}

nav {
  display: flex;
  justify-content: center;
  gap: 20px;
}

.nav-link {
  text-decoration: none;
  color: #666;
  font-weight: 500;
  padding: 5px 10px;
  border-radius: 4px;
  transition: all 0.2s;
}

.nav-link:hover {
  background-color: #f0f0f0;
  color: #333;
}

.router-link-active {
  color: #3498db;
  background-color: #e8f7fc;
}
</style>
