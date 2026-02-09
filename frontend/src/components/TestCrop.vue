<template>
  <div class="test-crop-container">
    <h2>Debug: Image Cropping Test</h2>
    <p>Upload an image to test backend cropping logic (No AI involved).</p>
    
    <div class="controls">
      <input type="file" @change="handleFileSelect" accept="image/*,application/pdf" />
      <button @click="submitFile" :disabled="!selectedFile">Test Crop</button>
      <button @click="$emit('close')">Close Debug</button>
    </div>

    <div v-if="loading">Processing...</div>
    <div v-if="error" style="color: red; font-weight: bold;">{{ error }}</div>

    <div v-if="result" class="comparison">
      <div class="image-box">
        <h3>Original</h3>
        <img :src="result.original" alt="Original" />
      </div>
      <div class="image-box">
        <h3>Cropped (Simulated Center Cut)</h3>
        <img :src="result.cropped" alt="Cropped" />
      </div>
    </div>
    <div v-if="result" class="info">{{ result.info }}</div>
  </div>
</template>

<script setup>
import { ref } from 'vue';

const selectedFile = ref(null);
const result = ref(null);
const loading = ref(false);
const error = ref(null);

const handleFileSelect = (event) => {
  selectedFile.value = event.target.files[0];
};

const submitFile = async () => {
    if (!selectedFile.value) return;

    loading.value = true;
    error.value = null;
    result.value = null;

    const formData = new FormData();
    formData.append('file', selectedFile.value);

    try {
        console.log("Sending request to /debug/crop...");
        const response = await fetch('http://localhost:8080/debug/crop', {
            method: 'POST',
            body: formData,
        });
        
        if (!response.ok) {
            const errText = await response.text();
            throw new Error(`Server Error ${response.status}: ${errText}`);
        }

        const data = await response.json();
        console.log("Received data:", data);
        result.value = data;
    } catch (e) {
        console.error(e);
        error.value = e.message;
    } finally {
        loading.value = false;
    }
};
</script>

<style scoped>
.test-crop-container {
    padding: 20px;
    background: #f4f4f4;
    border: 1px solid #ccc;
    margin: 20px 0;
}
.controls {
    margin: 15px 0;
    display: flex;
    gap: 10px;
    align-items: center;
}
.comparison {
    display: flex;
    gap: 20px;
    margin-top: 20px;
}
.image-box {
    flex: 1;
    border: 1px solid #ddd;
    padding: 10px;
    background: white;
}
.image-box img {
    max-width: 100%;
    height: auto;
    border: 5px solid #333; /* Thick dark border to see white images */
    display: block;
}
.info {
    margin-top: 10px;
    font-family: monospace;
    background: #333;
    color: #fff;
    padding: 10px;
}
</style>
