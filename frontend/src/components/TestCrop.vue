<template>
  <div class="test-crop-container">
    <h2>Debug: Image Cropping Test</h2>
    <p>Upload an image to test backend cropping logic (No AI involved).</p>

    <div class="controls">
      <input
        type="file"
        @change="handleFileSelect"
        accept="image/*,application/pdf"
      />
      <button @click="submitFile" :disabled="!selectedFile">Test Crop</button>
      <button @click="$emit('close')">Close Debug</button>
    </div>

    <div v-if="loading">Processing...</div>
    <div v-if="error" style="color: red; font-weight: bold">{{ error }}</div>

    <div v-if="result" class="comparison">
      <div class="image-box">
        <h3>Original</h3>
        <canvas ref="pdfCanvas" class="pdf-canvas"></canvas>
        <img
          v-if="result && result.original"
          :src="result.original"
          alt="Original Image"
        />
      </div>
      <div class="image-box">
        <h3>Cropped</h3>
        <img
          v-if="result && result.cropped"
          :src="result.cropped"
          alt="Image with Bounding Box"
        />
      </div>
    </div>

    <div v-if="result" class="info">{{ result.info }}</div>
    <div v-if="result && result.content_box" style="margin-top: 12px">
      <strong>Detected content_box:</strong>
      <pre style="margin: 6px 0; background: #222; color: #fff; padding: 8px">{{
        JSON.stringify(result.content_box)
      }}</pre>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from "vue";

const selectedFile = ref(null);
const result = ref(null);
const loading = ref(false);
const error = ref(null);
const pdfCanvas = ref(null);

watch(result, (newResult) => {
  if (newResult && newResult.pdf) {
    renderPdf(newResult.pdf, newResult.content_box);
  }
});

const renderPdf = async (pdfData, contentBox) => {
  const pdfjsLib = window["pdfjs-dist/build/pdf"];
  pdfjsLib.GlobalWorkerOptions.workerSrc =
    "https://cdnjs.cloudflare.com/ajax/libs/pdf.js/2.10.377/pdf.worker.min.js";

  const loadingTask = pdfjsLib.getDocument({
    data: atob(pdfData.split(",")[1]),
  });
  const pdf = await loadingTask.promise;
  const page = await pdf.getPage(1);
  const viewport = page.getViewport({ scale: 1.5 });

  const canvas = pdfCanvas.value;
  const context = canvas.getContext("2d");
  canvas.height = viewport.height;
  canvas.width = viewport.width;

  const renderContext = {
    canvasContext: context,
    viewport: viewport,
  };
  await page.render(renderContext).promise;

  if (contentBox) {
    const [yMin, xMin, yMax, xMax] = contentBox;
    context.strokeStyle = "red";
    context.lineWidth = 2;
    context.setLineDash([10, 5]);
    context.strokeRect(
      (xMin / 1000) * canvas.width,
      (yMin / 1000) * canvas.height,
      ((xMax - xMin) / 1000) * canvas.width,
      ((yMax - yMin) / 1000) * canvas.height,
    );
  }
};

const handleFileSelect = (event) => {
  selectedFile.value = event.target.files[0];
};

const submitFile = async () => {
  if (!selectedFile.value) return;

  loading.value = true;
  error.value = null;
  result.value = null;

  const formData = new FormData();
  formData.append("file", selectedFile.value);

  try {
    console.log("Sending request to /debug/crop...");
    const response = await fetch("http://localhost:8080/debug/crop", {
      method: "POST",
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
.pdf-canvas {
  max-width: 100%;
  height: auto;
  border: 5px solid #333;
  display: block;
}
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
