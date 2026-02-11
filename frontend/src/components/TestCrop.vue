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

    <!-- Client-side Preview (before crop) -->
    <div
      v-if="previewUrl && !result"
      class="image-box"
      style="margin-bottom: 20px"
    >
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
        <h3>Original with Box</h3>
        <!-- Always use canvas for consistent drawing of box -->
        <canvas ref="originalCanvas" class="pdf-canvas"></canvas>
      </div>
      <div class="image-box">
        <h3>Cropped Client-Side</h3>
        <img
          v-if="croppedImageUrl"
          :src="croppedImageUrl"
          style="max-width: 100%; border: 2px solid #666"
        />
        <p v-else>No crop result</p>
      </div>
    </div>

    <div v-if="result && result.content_box" style="margin-top: 12px">
      <strong>Detected content_box:</strong>
      <pre style="margin: 6px 0; background: #222; color: #fff; padding: 8px">{{
        JSON.stringify(result.content_box)
      }}</pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from "vue";
import * as pdfjsLib from "pdfjs-dist";
import { useMutation } from "@tanstack/vue-query";
import { debugCropMutation } from "../client/@tanstack/vue-query.gen";
import type { DebugCropResponse } from "../client/types.gen";

// Set worker source for pdfjs-dist
pdfjsLib.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjsLib.version}/build/pdf.worker.min.mjs`;

const selectedFile = ref<File | null>(null);
const previewUrl = ref<string | null>(null);
const result = ref<DebugCropResponse | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);
const originalCanvas = ref<HTMLCanvasElement | null>(null);
const croppedImageUrl = ref<string | null>(null);

const { mutateAsync: debugCrop } = useMutation({
  ...debugCropMutation(),
});

watch(result, async (newResult) => {
  if (newResult && newResult.content_box && previewUrl.value) {
    await nextTick(); // Wait for canvas to be in DOM
    renderCropResult(previewUrl.value, newResult.content_box);
  }
});

const handleFileSelect = async (event: Event) => {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  selectedFile.value = file || null;
  previewUrl.value = null;
  error.value = null;
  result.value = null;
  croppedImageUrl.value = null;

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
      if (!context) throw new Error("Could not get 2d context");

      canvas.height = viewport.height;
      canvas.width = viewport.width;

      const renderContext = {
        canvasContext: context,
        viewport: viewport,
        transform: undefined as any,
        canvas: context.canvas,
      };
      await page.render(renderContext).promise;

      previewUrl.value = canvas.toDataURL("image/png");
    } else if (file.type.startsWith("image/")) {
      const reader = new FileReader();
      reader.onload = (e) => {
        if (e.target?.result) {
          previewUrl.value = e.target.result as string;
        }
      };
      reader.readAsDataURL(file);
    }
  } catch (err: any) {
    console.error("Error previewing file:", err);
    error.value = "Failed to create preview: " + err.message;
  }
};

const renderCropResult = async (
  imageUrl: string,
  contentBox: [number, number, number, number],
) => {
  try {
    const canvas = originalCanvas.value;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const img = new Image();
    img.src = imageUrl;
    await new Promise((resolve) => {
      img.onload = resolve;
    });

    canvas.width = img.width;
    canvas.height = img.height;
    ctx.drawImage(img, 0, 0);

    // Draw Box
    const [yMin, xMin, yMax, xMax] = contentBox;
    ctx.strokeStyle = "red";
    ctx.lineWidth = 4;
    ctx.setLineDash([10, 5]);
    ctx.strokeRect(
      (xMin / 1000) * canvas.width,
      (yMin / 1000) * canvas.height,
      ((xMax - xMin) / 1000) * canvas.width,
      ((yMax - yMin) / 1000) * canvas.height,
    );

    // Crop
    const cropW = ((xMax - xMin) / 1000) * img.width;
    const cropH = ((yMax - yMin) / 1000) * img.height;
    const cropX = (xMin / 1000) * img.width;
    const cropY = (yMin / 1000) * img.height;

    const cropCanvas = document.createElement("canvas");
    cropCanvas.width = cropW;
    cropCanvas.height = cropH;
    const cropCtx = cropCanvas.getContext("2d");
    if (!cropCtx) return;

    cropCtx.drawImage(img, cropX, cropY, cropW, cropH, 0, 0, cropW, cropH);

    croppedImageUrl.value = cropCanvas.toDataURL("image/png");
  } catch (err: any) {
    console.error("Error rendering crop result:", err);
    error.value = "Failed to render crop result: " + err.message;
  }
};

const submitFile = async () => {
  if (!selectedFile.value) return;

  loading.value = true;
  error.value = null;
  result.value = null;
  croppedImageUrl.value = null;

  try {
    console.log("Sending request to /api/v1/debug/crop...");
    const data = await debugCrop({
      body: {
        file: selectedFile.value,
      },
    });

    console.log("Received data:", data);
    result.value = data as DebugCropResponse;
  } catch (e: any) {
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
  flex-direction: row;
  gap: 20px;
  margin-top: 20px;
}
.image-box {
  flex: 1;
  min-width: 0;
  box-sizing: border-box;
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
