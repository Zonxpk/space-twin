<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import Upload from "../components/Upload.vue";
import FloorplanMap from "../components/FloorplanMap.vue";
import { WebSocketService } from "../services/websocket";

const rooms = ref([]);
const bgImage = ref(null);
const ws = new WebSocketService("ws://localhost:8080/ws");

onMounted(() => {
  ws.connect();
  ws.onMessage((msg) => {
    if (msg.type === "update" && msg.data) {
      handleRealtimeUpdate(msg.data);
    }
  });
});

onUnmounted(() => {
  ws.close();
});

const handleRealtimeUpdate = (updates) => {
  // updates is array of {name, temperature, occupancy}
  // We match by name.

  // Create map for O(1) lookup
  const updateMap = new Map(updates.map((u) => [u.name, u]));

  rooms.value = rooms.value.map((room) => {
    const update = updateMap.get(room.name);
    if (update) {
      return { ...room, ...update }; // Merge update
    }
    return room;
  });
};

const handleAnalysis = (data) => {
  console.log("Received AI Data:", data);

  if (data.rooms) {
    rooms.value = data.rooms;
  }
  if (data.image) {
    bgImage.value = data.image;
  }
};

const resetView = () => {
  rooms.value = [];
  bgImage.value = null;
};
</script>

<template>
  <div class="content-wrapper">
    <!-- Show Upload centered if no floorplan loaded -->
    <transition name="fade" mode="out-in">
      <div v-if="rooms.length === 0" class="upload-centered" key="upload">
        <Upload @analysis-complete="handleAnalysis" />
      </div>

      <div v-else class="dashboard" key="dashboard">
        <div class="dashboard-header">
          <button @click="resetView" class="reset-btn">
            ← Upload New Floorplan
          </button>
        </div>
        <FloorplanMap
          :rooms="rooms"
          :image="bgImage"
          :width="1000"
          :height="700"
        />

        <!-- T012: Debug JSON -->
        <div class="debug-json">
          <h3>Debug Data</h3>
          <pre>{{ JSON.stringify(rooms, null, 2) }}</pre>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
/* 
  Neo-Brutalist Upload Styles 
  Inherits variables from App.vue
*/
.content-wrapper {
  min-height: 600px;
  display: flex;
  flex-direction: column;
  padding: 2rem;
  color: var(--color-text);
  font-family: var(--font-mono);
}

/* Centered Upload State */
.upload-centered {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  border: 1px dashed var(--color-secondary);
  background: rgba(15, 23, 42, 0.5);
  margin: 2rem;
  min-height: 400px;
}

/* Dashboard State */
.dashboard {
  animation: fadeIn 0.5s ease;
  width: 100%;
}

.dashboard-header {
  margin-bottom: 2rem;
  display: flex;
  justify-content: flex-start;
}

.reset-btn {
  background: transparent;
  color: var(--color-cta);
  border: 1px solid var(--color-cta);
  padding: 1rem 2rem;
  cursor: pointer;
  font-family: var(--font-mono);
  font-weight: 700;
  text-transform: uppercase;
  transition: all 0.2s;
}

.reset-btn:hover {
  background: var(--color-cta);
  color: white;
  transform: translateY(-2px);
  box-shadow: 4px 4px 0px rgba(0, 0, 0, 0.5);
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.debug-json {
  margin-top: 2rem;
  background: rgba(0, 0, 0, 0.3);
  padding: 1rem;
  border: 1px solid var(--color-secondary);
  font-size: 0.8rem;
  color: #94a3b8;
  max-height: 300px;
  overflow-y: auto;
  font-family: var(--font-mono);
}

.debug-json h3 {
  color: var(--color-accent);
  margin-top: 0;
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 1px;
}
</style>
