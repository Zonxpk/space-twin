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
.content-wrapper {
  min-height: 600px;
  display: flex;
  flex-direction: column;
}

/* Centered Upload State */
.upload-centered {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  flex-direction: column;
}

/* Dashboard State */
.dashboard {
  animation: fadeIn 0.5s ease;
}

.dashboard-header {
  margin-bottom: 1rem;
}

.reset-btn {
  background: none;
  border: 1px solid #ccc;
  padding: 8px 16px;
  cursor: pointer;
  border-radius: 4px;
  color: #666;
  transition: all 0.2s;
}

.reset-btn:hover {
  background: #f0f0f0;
  color: #333;
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
  margin-top: 20px;
  padding: 10px;
  background: #f4f4f4;
  border: 1px solid #ddd;
  border-radius: 4px;
  max-height: 200px;
  overflow-y: auto;
  font-size: 12px;
}
</style>
