<template>
  <div class="test-crop-container">
    <h2>Edge Detection & Crop Test</h2>
    <p>Upload an image to test edge detection and automatic cropping.</p>

    <div class="controls">
      <input
        type="file"
        @change="handleFileSelect"
        accept="image/*,application/pdf"
      />
      <button @click="submitFile" :disabled="!selectedFile">
        Detect Edges
      </button>
      <button @click="submitCrop" :disabled="!selectedFile">
        Crop Floorplan
      </button>
      <button @click="$emit('close')">Close</button>
    </div>

    <!-- Client-side Preview (before detection) -->
    <div
      v-if="selectedFile && !result && !cropResult"
      class="image-box"
      style="margin-bottom: 20px"
    >
      <h3>Original</h3>
      <img
        :src="getPreviewUrl()"
        alt="Original"
        style="max-width: 100%; border: 2px solid #666"
      />
    </div>

    <div v-if="loading">Processing...</div>
    <div v-if="error" style="color: red; font-weight: bold">{{ error }}</div>

    <!-- Edge Detection Result -->
    <div v-if="result" class="comparison">
      <div class="image-box">
        <h3>Original</h3>
        <img
          :src="getPreviewUrl()"
          alt="Original"
          style="max-width: 100%; border: 2px solid #666"
        />
      </div>
      <div class="image-box">
        <h3>Edge Detection Result</h3>
        <img
          v-if="result.processed_image"
          :src="result.processed_image"
          style="max-width: 100%; border: 2px solid #666"
        />
        <p v-else>No result</p>
      </div>
    </div>

    <!-- Crop Result -->
    <div v-if="cropResult" class="comparison">
      <div class="image-box">
        <h3>Original</h3>
        <img
          :src="getPreviewUrl()"
          alt="Original"
          style="max-width: 100%; border: 2px solid #666"
        />
      </div>
      <div class="image-box">
        <h3>Cropped Floorplan</h3>
        <img
          v-if="cropResult.cropped_image"
          :src="cropResult.cropped_image"
          style="max-width: 100%; border: 2px solid #666"
        />
        <p v-else>No result</p>
        <p
          v-if="cropResult.message"
          style="margin-top: 10px; font-size: 0.9em; color: #666"
        >
          {{ cropResult.message }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import * as pdfjsLib from "pdfjs-dist";
import { useMutation } from "@tanstack/vue-query";
import {
  detectEdgesMutation,
  cropFloorplanMutation,
} from "../client/@tanstack/vue-query.gen";
import type { DetectEdgesResponse } from "../client/types.gen";

// Set worker source for pdfjs-dist
pdfjsLib.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjsLib.version}/build/pdf.worker.min.mjs`;

const selectedFile = ref<File | null>(null);
const result = ref<DetectEdgesResponse | null>(null);
const cropResult = ref<{ cropped_image: string; message: string } | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);

const { mutateAsync: detectEdges } = useMutation({
  ...detectEdgesMutation(),
});

const { mutateAsync: cropFloorplan } = useMutation({
  ...cropFloorplanMutation(),
});

const getPreviewUrl = (): string => {
  if (!selectedFile.value) return "";
  return URL.createObjectURL(selectedFile.value);
};

const handleFileSelect = async (event: Event) => {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  error.value = null;
  result.value = null;
  cropResult.value = null;

  if (!file) {
    selectedFile.value = null;
    return;
  }

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

      canvas.toBlob((blob) => {
        if (blob) {
          selectedFile.value = new File([blob], "converted.png", {
            type: "image/png",
          });
        }
      }, "image/png");
    } else if (file.type.startsWith("image/")) {
      selectedFile.value = file;
    }
  } catch (err: any) {
    console.error("Error previewing file:", err);
    error.value = "Failed to create preview: " + err.message;
  }
};

const submitFile = async () => {
  if (!selectedFile.value) return;

  loading.value = true;
  error.value = null;
  result.value = null;
  cropResult.value = null;

  try {
    const reader = new FileReader();
    reader.onload = async (e) => {
      if (!e.target?.result) return;

      console.log("Sending request to /api/v1/process/edges...");
      const data = await detectEdges({
        body: {
          file: selectedFile.value,
        },
      });

      console.log("Received data:", data);
      result.value = data as DetectEdgesResponse;
      loading.value = false;
    };
    reader.readAsDataURL(selectedFile.value);
  } catch (e: any) {
    console.error(e);
    error.value = e.message;
    loading.value = false;
  }
};

const submitCrop = async () => {
  if (!selectedFile.value) return;

  loading.value = true;
  error.value = null;
  result.value = null;
  cropResult.value = null;

  try {
    const reader = new FileReader();
    reader.onload = async (e) => {
      if (!e.target?.result) return;

      const base64Data = e.target.result as string;
      console.log("Sending crop request to /api/v1/process/crop...");

      const data = await cropFloorplan({
        body: {
          image: base64Data,
          options: {
            blur_radius: 1.2,
            canny_low: 50,
            canny_high: 150,
            resize_max_width: 800,
          },
        },
      });

      console.log("Crop result:", data);
      cropResult.value = data.cropped_image
        ? {
            cropped_image: data.cropped_image,
            message: data.message || "",
          }
        : null;
      loading.value = false;
    };
    reader.readAsDataURL(selectedFile.value);
  } catch (e: any) {
    console.error(e);
    error.value = e.message;
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
  border: 5px solid #333;
  display: block;
}
</style>
