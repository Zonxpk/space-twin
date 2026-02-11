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

    <!-- Client-side Preview -->
    <div v-if="previewUrl" class="image-box" style="margin-bottom: 20px">
      <h3>Client Preview</h3>
      <img
        :src="previewUrl"
        alt="Client Preview"
        style="max-width: 100%; border: 2px solid #666"
      />
    </div>

    <div v-if="loading">Processing...</div>
    <div v-if="error" style="color: red; font-weight: bold">{{ error }}</div>

    <div v-if="result" class="comparison">
      <div class="image-box">
        <h3>Original (from Server)</h3>
        <canvas
          ref="pdfCanvas"
          class="pdf-canvas"
          v-show="isPdfResult"
        ></canvas>
        <img
          v-if="result && (result.original || result.file) && !isPdfResult"
          :src="result.original || result.file"
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
import { ref, watch, computed } from "vue";
import * as pdfjsLib from "pdfjs-dist";
import { useMutation } from "@tanstack/vue-query";
import { debugCropMutation } from "../client/@tanstack/vue-query.gen";

// Set worker source for pdfjs-dist
pdfjsLib.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjsLib.version}/build/pdf.worker.min.mjs`;

const selectedFile = ref(null);
const previewUrl = ref(null);
const result = ref(null);
const loading = ref(false);
const error = ref(null);
const pdfCanvas = ref(null);

const isPdfResult = computed(() => {
  return (
    result.value && (result.value.pdf || result.value.file_type === ".pdf")
  );
});

const { mutateAsync: debugCrop } = useMutation({
  ...debugCropMutation(),
});

watch(result, (newResult) => {
  if (newResult && (newResult.pdf || newResult.file_type === ".pdf")) {
    renderPdfResult(newResult.pdf || newResult.file, newResult.content_box);
  }
});

const handleFileSelect = async (event) => {
  const file = event.target.files[0];
  selectedFile.value = file;
  previewUrl.value = null;
  error.value = null;

  if (!file) return;

  try {
    if (file.type === "application/pdf") {
      const arrayBuffer = await file.arrayBuffer();
      const loadingTask = pdfjsLib.getDocument(arrayBuffer);
      const pdf = await loadingTask.promise;
      const page = await pdf.getPage(1);

      const viewport = page.getViewport({ scale: 1.5 });
      const canvas = document.createElement("canvas");
      const context = canvas.getContext("2d");
      canvas.height = viewport.height;
      canvas.width = viewport.width;

      const renderContext = {
        canvasContext: context,
        viewport: viewport,
      };
      await page.render(renderContext).promise;

      previewUrl.value = canvas.toDataURL("image/png");
    } else if (file.type.startsWith("image/")) {
      const reader = new FileReader();
      reader.onload = (e) => {
        previewUrl.value = e.target.result;
      };
      reader.readAsDataURL(file);
    }
  } catch (err) {
    console.error("Error previewing file:", err);
    error.value = "Failed to create preview: " + err.message;
  }
};

const renderPdfResult = async (pdfData, contentBox) => {
  try {
    // pdfData is base64 string from server
    const loadingTask = pdfjsLib.getDocument({
      data: atob(pdfData.split(",")[1]),
    });
    const pdf = await loadingTask.promise;
    const page = await pdf.getPage(1);
    const viewport = page.getViewport({ scale: 1.5 });

    const canvas = pdfCanvas.value;
    if (!canvas) return;

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
  } catch (err) {
    console.error("Error rendering result PDF:", err);
    error.value = "Failed to render result PDF: " + err.message;
  }
};

const submitFile = async () => {
  if (!selectedFile.value) return;

  loading.value = true;
  error.value = null;
  result.value = null;

  try {
    console.log("Sending request to /api/v1/debug/crop...");
    const data = await debugCrop({
      body: {
        file: selectedFile.value,
      },
    });

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
