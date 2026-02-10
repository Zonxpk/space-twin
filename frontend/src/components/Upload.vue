<template>
  <div
    class="upload-container"
    :class="{ 'is-dragover': isDragOver }"
    @dragover.prevent="isDragOver = true"
    @dragleave.prevent="isDragOver = false"
    @drop.prevent="handleDrop"
    @click="triggerFileInput"
  >
    <div v-if="uploading" class="upload-content">
      <div class="spinner"></div>
      <p>Analyzing Floorplan...</p>
    </div>
    <div v-else class="upload-content">
      <p class="icon">📂</p>
      <h3>Click or Drag & Drop to Upload</h3>
      <p class="subtext">Supports Images (PNG, JPG) and PDF</p>
      <input
        type="file"
        ref="fileInput"
        @change="handleFileSelect"
        accept="image/*,application/pdf"
        style="display: none"
      />
    </div>
    <div v-if="error" class="error" @click.stop>{{ error }}</div>
  </div>
</template>

<script setup>
import { ref } from "vue";

const emit = defineEmits(["analysis-complete"]);

const fileInput = ref(null);
const isDragOver = ref(false);
const uploading = ref(false);
const error = ref(null);

const triggerFileInput = () => {
  if (!uploading.value) {
    fileInput.value.click();
  }
};

const handleFileSelect = (event) => {
  const file = event.target.files[0];
  if (file) {
    processFile(file);
  }
};

const handleDrop = (event) => {
  isDragOver.value = false;
  const file = event.dataTransfer.files[0];
  if (file) {
    processFile(file);
  }
};

const processFile = async (file) => {
  uploading.value = true;
  error.value = null;

  const formData = new FormData();
  formData.append("file", file);

  try {
    const response = await fetch("http://localhost:8080/api/v1/upload", {
      method: "POST",
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`Upload failed: ${response.statusText}`);
    }

    const data = await response.json();
    emit("analysis-complete", { rooms: data.rooms, image: data.image });
  } catch (err) {
    console.error(err);
    error.value = err.message || "Failed to analyze floorplan";
  } finally {
    uploading.value = false;
  }
};
</script>

<style scoped>
.upload-container {
  padding: 40px;
  border: 3px dashed #ccc;
  border-radius: 12px;
  text-align: center;
  margin-bottom: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  background-color: #f9f9f9;
  min-height: 200px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.upload-container:hover {
  border-color: #3498db;
  background-color: #eef7fc;
}

.upload-container.is-dragover {
  border-color: #2ecc71;
  background-color: #e8f8f5;
  transform: scale(1.02);
}

.upload-content {
  pointer-events: none; /* Let clicks pass through to container */
}

.icon {
  font-size: 3rem;
  margin-bottom: 10px;
}

.subtext {
  color: #777;
  font-size: 0.9rem;
  margin-top: 5px;
}

.error {
  color: #e74c3c;
  margin-top: 15px;
  font-weight: bold;
  cursor: default;
}

/* Simple CSS Spinner */
.spinner {
  border: 4px solid #f3f3f3;
  border-top: 4px solid #3498db;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  animation: spin 1s linear infinite;
  margin: 0 auto 15px;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
</style>
