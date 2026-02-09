<template>
  <div class="map-container">
    <v-stage :config="configStage">
      <!-- Layer 1: Background Image -->
      <v-layer>
        <v-image :config="{ image: imageObj }" />
      </v-layer>

      <!-- Layer 2: Spotlight Overlay (Visual Noise Removal) -->
      <v-layer>
        <v-group>
            <!-- A. Dark Overlay covering entire map -->
            <v-rect :config="{
                x: 0, 
                y: 0, 
                width: configStage.width, 
                height: configStage.height, 
                fill: 'rgba(255, 255, 255, 0.85)'
            }" />
            
            <!-- B. Cut out bubbles for rooms -->
            <!-- globalCompositeOperation 'destination-out' erases from THIS LAYER only -->
            <v-group :config="{ globalCompositeOperation: 'destination-out' }">
                 <v-rect 
                    v-for="(room, index) in rooms" 
                    :key="'cutout-'+index" 
                    :config="{
                        x: room.rect[0],
                        y: room.rect[1],
                        width: room.rect[2],
                        height: room.rect[3],
                        fill: 'black' 
                    }"
                />
            </v-group>
        </v-group>
      </v-layer>

      <!-- Layer 3: Interactive UI (Highlights & Labels) -->
      <v-layer>
        <!-- Room Highlights -->
        <v-rect 
            v-for="(room, index) in rooms" 
            :key="index" 
            :config="getRoomConfig(room)"
        />
        <!-- Labels -->
        <v-text 
            v-for="(room, index) in rooms" 
            :key="'label-'+index" 
            :config="getTextConfig(room)"
        />
      </v-layer>
    </v-stage>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';

const props = defineProps({
  rooms: {
    type: Array,
    default: () => []
  },
  image: {
    type: String,
    default: null
  },
  width: {
    type: Number,
    default: 800
  },
  height: {
    type: Number,
    default: 600
  }
});

const imageObj = ref(null);

watch(() => props.image, (newVal) => {
    if (newVal) {
        const img = new Image();
        img.src = newVal;
        img.onload = () => {
            imageObj.value = img;
            // Optionally adjust stage size to match image aspect ratio if needed
        };
    }
}, { immediate: true });

const configStage = computed(() => ({
  width: props.width,
  height: props.height
}));

const getRoomConfig = (room) => {
    // Expected format: rect [x, y, w, h] or [ymin, xmin, ymax, xmax] depending on prompt.
    // The prompt used was: [x, y, w, h].
    // room.rect is array.
    
    // Safety check
    if (!room.rect || room.rect.length < 4) return {};

    const [x, y, w, h] = room.rect;
    
    // Determine color based on occupancy
    let fillColor = 'rgba(0, 255, 0, 0.2)'; // Green (Free)
    if (room.occupancy > 0) {
        fillColor = 'rgba(255, 0, 0, 0.2)'; // Red (Busy)
    } else if (room.occupancy === undefined) {
        fillColor = 'rgba(0, 0, 255, 0.1)'; // Blue (Unknown/Default)
    }
    
    return {
        x: x,
        y: y,
        width: w,
        height: h,
        stroke: 'blue',
        strokeWidth: 2,
        fill: fillColor,
        draggable: true
    };
};

const getTextConfig = (room) => {
    if (!room.rect || room.rect.length < 4) return {};
    const [x, y, w, h] = room.rect;

    let label = room.name || 'Unnamed';
    if (room.temperature) {
        label += `\n${room.temperature}°C`;
    }
    if (room.occupancy !== undefined) {
        label += `\n${room.occupancy} ppl`;
    }

    return {
        x: x + 5,
        y: y + 5,
        text: label,
        fontSize: 14,
        fill: 'black'
    };
};

</script>

<style scoped>
.map-container {
  border: 1px solid #ddd;
  margin-top: 20px;
}
</style>
