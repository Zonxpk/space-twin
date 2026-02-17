<script setup>
import { ref } from "vue";
import * as pdfjsLib from "pdfjs-dist";
import { useMutation } from "@tanstack/vue-query";
import {
  uploadFloorplanMutation,
  cropFloorplanMutation,
} from "../client/@tanstack/vue-query.gen";

const emit = defineEmits(["analysis-complete"]);

// --- Wizard state ---
const currentStep = ref(1);
const fileInput = ref(null);

// --- Step 1 state ---
const isDragOver = ref(false);
const uploadedFile = ref(null);
const originalPreviewUrl = ref(null);
const converting = ref(false);

// --- Step 2 state ---
const cropping = ref(false);
const croppedImageUrl = ref(null);
const cropMessage = ref(null);

// --- Step 3 state ---
const analyzing = ref(false);

// --- Shared ---
const error = ref(null);

// --- Mutations ---
const { mutateAsync: uploadFloorplan } = useMutation({
  ...uploadFloorplanMutation(),
});

const { mutateAsync: cropFloorplan } = useMutation({
  ...cropFloorplanMutation(),
});

// --- Helpers ---

const readFileAsDataUrl = (file) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = (e) => resolve(e.target.result);
    reader.onerror = () => reject(new Error("Failed to read file"));
    reader.readAsDataURL(file);
  });
};

const base64ToFile = async (dataUrl, filename) => {
  const response = await fetch(dataUrl);
  const blob = await response.blob();
  return new File([blob], filename, { type: "image/png" });
};

// --- Step 1: Upload ---

const handleDragOver = (e) => {
  e.preventDefault();
  isDragOver.value = true;
};

const handleDragLeave = () => {
  isDragOver.value = false;
};

const handleDrop = (event) => {
  isDragOver.value = false;
  const file = event.dataTransfer.files[0];
  if (file) {
    processUpload(file);
  }
};

const handleFileSelect = (event) => {
  const file = event.target.files[0];
  if (file) {
    processUpload(file);
  }
};

const processUpload = async (file) => {
  converting.value = true;
  error.value = null;
  croppedImageUrl.value = null;
  cropMessage.value = null;

  try {
    let fileToProcess = file;

    if (file.type === "application/pdf") {
      fileToProcess = await convertPdfToImage(file);
    }

    uploadedFile.value = fileToProcess;
    originalPreviewUrl.value = await readFileAsDataUrl(fileToProcess);
    currentStep.value = 2;
    autoCrop();
  } catch (err) {
    console.error(err);
    error.value = err.message || "Failed to process file";
  } finally {
    converting.value = false;
  }
};

const convertPdfToImage = async (file) => {
  pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
    "pdfjs-dist/build/pdf.worker.min.js",
    import.meta.url,
  ).toString();

  const arrayBuffer = await file.arrayBuffer();
  const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise;
  const page = await pdf.getPage(1);

  const scale = 2.0;
  const viewport = page.getViewport({ scale });

  const canvas = document.createElement("canvas");
  const context = canvas.getContext("2d");
  canvas.height = viewport.height;
  canvas.width = viewport.width;

  const renderContext = {
    canvasContext: context,
    viewport: viewport,
    background: "rgba(255, 255, 255, 1)",
  };

  await page.render(renderContext).promise;

  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (blob) {
        const newFile = new File([blob], file.name.replace(".pdf", ".png"), {
          type: "image/png",
        });
        resolve(newFile);
      } else {
        reject(new Error("Canvas to Blob conversion failed"));
      }
    }, "image/png");
  });
};

// --- Step 2: Crop ---

const autoCrop = async () => {
  cropping.value = true;
  error.value = null;
  croppedImageUrl.value = null;
  cropMessage.value = null;

  try {
    const data = await cropFloorplan({
      body: {
        image: originalPreviewUrl.value,
        options: {
          blur_radius: 1.2,
          canny_low: 50,
          canny_high: 150,
          resize_max_width: 800,
        },
      },
    });

    if (data.cropped_image) {
      croppedImageUrl.value = data.cropped_image;
      cropMessage.value = data.message || "";
    } else {
      error.value =
        "Crop returned no image. You may proceed with the original.";
      croppedImageUrl.value = originalPreviewUrl.value;
    }
  } catch (err) {
    console.error(err);
    error.value = "Auto-crop failed: " + (err.message || "Unknown error");
    croppedImageUrl.value = originalPreviewUrl.value;
  } finally {
    cropping.value = false;
  }
};

// --- Step 3: Analyze ---

const proceedToAnalyze = () => {
  currentStep.value = 3;
  analyzeFloorplan();
};

const analyzeFloorplan = async () => {
  analyzing.value = true;
  error.value = null;

  try {
    const imageToSend = croppedImageUrl.value || originalPreviewUrl.value;
    const fileToSend = await base64ToFile(imageToSend, "floorplan.png");

    const data = await uploadFloorplan({
      body: {
        file: fileToSend,
      },
    });

    emit("analysis-complete", { rooms: data.rooms, image: data.image });
  } catch (err) {
    console.error(err);
    error.value = err.message || "Failed to analyze floorplan";
  } finally {
    analyzing.value = false;
  }
};

// --- Reset ---

const reUpload = () => {
  currentStep.value = 1;
  uploadedFile.value = null;
  originalPreviewUrl.value = null;
  croppedImageUrl.value = null;
  cropMessage.value = null;
  error.value = null;
  converting.value = false;
  cropping.value = false;
  analyzing.value = false;
  if (fileInput.value) {
    fileInput.value.value = "";
  }
};
</script>

<template>
  <div class="upload-wizard">
    <!-- Step indicator -->
    <div class="step-indicator">
      <div
        class="step"
        :class="{ active: currentStep === 1, completed: currentStep > 1 }"
      >
        <span class="step-number">1</span>
        <span class="step-label">Upload</span>
      </div>
      <div class="step-divider" :class="{ completed: currentStep > 1 }"></div>
      <div
        class="step"
        :class="{ active: currentStep === 2, completed: currentStep > 2 }"
      >
        <span class="step-number">2</span>
        <span class="step-label">Crop</span>
      </div>
      <div class="step-divider" :class="{ completed: currentStep > 2 }"></div>
      <div class="step" :class="{ active: currentStep === 3 }">
        <span class="step-number">3</span>
        <span class="step-label">Analyze</span>
      </div>
    </div>

    <!-- STEP 1: Upload -->
    <div v-if="currentStep === 1" class="wizard-step">
      <div
        class="drop-zone"
        :class="{ 'is-dragover': isDragOver }"
        @dragover.prevent="handleDragOver"
        @dragleave="handleDragLeave"
        @drop.prevent="handleDrop"
        @click="fileInput?.click()"
      >
        <div v-if="converting" class="upload-content">
          <div class="spinner"></div>
          <p>Converting PDF to image...</p>
        </div>
        <div v-else class="upload-content">
          <div class="icon"></div>
          <h3>Drop PDF floorplan here</h3>
          <p class="subtext">or click to browse</p>
        </div>
        <input
          type="file"
          ref="fileInput"
          @change="handleFileSelect"
          accept=".pdf,.png,.jpg,.jpeg"
          style="display: none"
        />
      </div>
    </div>

    <!-- STEP 2: Crop -->
    <div v-else-if="currentStep === 2" class="wizard-step">
      <div class="comparison">
        <div class="image-box">
          <h3>Original</h3>
          <img :src="originalPreviewUrl" alt="Original floorplan" />
        </div>
        <div class="image-box">
          <h3>Cropped</h3>
          <div v-if="cropping" class="loading-placeholder">
            <div class="spinner"></div>
            <p>Auto-cropping floorplan...</p>
          </div>
          <img
            v-else-if="croppedImageUrl"
            :src="croppedImageUrl"
            alt="Cropped floorplan"
          />
          <p v-else>No cropped image available.</p>
          <p v-if="cropMessage" class="crop-message">{{ cropMessage }}</p>
        </div>
      </div>
      <div class="step-actions">
        <button class="btn btn-secondary" @click="reUpload">Re-upload</button>
        <button
          class="btn btn-primary"
          @click="proceedToAnalyze"
          :disabled="cropping"
        >
          Analyze Floorplan
        </button>
      </div>
    </div>

    <!-- STEP 3: Analyze -->
    <div v-else-if="currentStep === 3" class="wizard-step">
      <div v-if="analyzing" class="analyze-loading">
        <div class="spinner"></div>
        <p>Analyzing floorplan with AI...</p>
        <small>Detecting rooms and features</small>
      </div>
      <div v-else class="analyze-loading">
        <p>Analysis complete.</p>
      </div>
      <div class="step-actions">
        <button
          class="btn btn-secondary"
          @click="reUpload"
          :disabled="analyzing"
        >
          Start Over
        </button>
      </div>
    </div>

    <!-- Error display (visible in any step) -->
    <div v-if="error" class="error">
      {{ error }}
    </div>
  </div>
</template>

<style scoped>
.upload-wizard {
  width: 100%;
}

/* Step Indicator */
.step-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
  gap: 0;
}

.step {
  display: flex;
  align-items: center;
  gap: 6px;
  opacity: 0.4;
  transition: opacity 0.3s ease;
}

.step.active,
.step.completed {
  opacity: 1;
}

.step-number {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #ccc;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.85rem;
  font-weight: 600;
  transition: background 0.3s ease;
}

.step.active .step-number {
  background: #3498db;
}

.step.completed .step-number {
  background: #2ecc71;
}

.step-label {
  font-size: 0.85rem;
  color: #666;
  font-weight: 500;
}

.step.active .step-label {
  color: #333;
}

.step-divider {
  width: 40px;
  height: 2px;
  background: #ddd;
  margin: 0 8px;
  transition: background 0.3s ease;
}

.step-divider.completed {
  background: #2ecc71;
}

/* Wizard step container */
.wizard-step {
  animation: fadeIn 0.3s ease;
}

/* Drop zone (Step 1) */
.drop-zone {
  padding: 40px;
  border: 3px dashed #ccc;
  border-radius: 12px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  background-color: #f9f9f9;
  min-height: 200px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.drop-zone:hover {
  border-color: #3498db;
  background-color: #eef7fc;
}

.drop-zone.is-dragover {
  border-color: #2ecc71;
  background-color: #e8f8f5;
  transform: scale(1.02);
}

.upload-content {
  pointer-events: none;
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

/* Comparison layout (Step 2) */
.comparison {
  display: flex;
  flex-direction: row;
  gap: 20px;
  margin-bottom: 20px;
}

.image-box {
  flex: 1;
  min-width: 0;
  box-sizing: border-box;
  border: 1px solid #ddd;
  padding: 10px;
  background: white;
  border-radius: 8px;
}

.image-box h3 {
  margin: 0 0 10px 0;
  font-size: 0.95rem;
  color: #555;
}

.image-box img {
  max-width: 100%;
  height: auto;
  border: 2px solid #eee;
  border-radius: 4px;
  display: block;
}

.loading-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  color: #888;
}

.crop-message {
  margin-top: 10px;
  font-size: 0.85rem;
  color: #888;
}

/* Analyze loading (Step 3) */
.analyze-loading {
  text-align: center;
  padding: 40px;
}

.analyze-loading small {
  color: #888;
}

/* Action buttons */
.step-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 16px;
}

.btn {
  padding: 10px 24px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: #3498db;
  color: #fff;
}

.btn-primary:hover:not(:disabled) {
  background: #2980b9;
}

.btn-secondary {
  background: #f0f0f0;
  color: #666;
  border: 1px solid #ddd;
}

.btn-secondary:hover:not(:disabled) {
  background: #e0e0e0;
  color: #333;
}

/* Error */
.error {
  color: #e74c3c;
  margin-top: 15px;
  font-weight: bold;
  text-align: center;
}

/* Spinner */
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
</style>
