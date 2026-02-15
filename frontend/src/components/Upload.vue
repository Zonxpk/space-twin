<script setup>
import { ref } from "vue";
import * as pdfjsLib from "pdfjs-dist";
import { useMutation } from "@tanstack/vue-query";
import { uploadFloorplanMutation } from "../client/@tanstack/vue-query.gen";

// We will set the worker path dynamically before rendering.
// This is a more robust way to handle it with bundlers like Vite.

const emit = defineEmits(["analysis-complete"]);

const uploading = ref(false);
const error = ref(null);
const isDragOver = ref(false);

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
    processFile(file);
  }
};

const handleFileSelect = (event) => {
  const file = event.target.files[0];
  if (file) {
    processFile(file);
  }
};

const { mutateAsync: uploadFloorplan } = useMutation({
  ...uploadFloorplanMutation(),
});

const processFile = async (file) => {
  uploading.value = true;
  error.value = null;

  try {
    let fileToUpload = file;

    // If PDF, render to Image first
    if (file.type === "application/pdf") {
      fileToUpload = await convertPdfToImage(file);
    }

    const data = await uploadFloorplan({
      body: {
        file: fileToUpload,
      },
    });

    emit("analysis-complete", { rooms: data.rooms, image: data.image });
  } catch (err) {
    console.error(err);
    error.value = err.message || "Failed to analyze floorplan";
  } finally {
    uploading.value = false;
  }
};

const convertPdfToImage = async (file) => {
  // Dynamically set the worker source path. This is a robust method for Vite.
  pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
    "pdfjs-dist/build/pdf.worker.min.js",
    import.meta.url,
  ).toString();

  const arrayBuffer = await file.arrayBuffer();
  const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise;
  const page = await pdf.getPage(1);

  const scale = 2.0; // Use a higher scale for better analysis quality
  const viewport = page.getViewport({ scale });

  const canvas = document.createElement("canvas");
  const context = canvas.getContext("2d");
  canvas.height = viewport.height;
  canvas.width = viewport.width;

  const renderContext = {
    canvasContext: context,
    viewport: viewport,
    background: "rgba(255, 255, 255, 1)", // Ensure a white background
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
</script>

<template>
  <div
    class="upload-container"
    :class="{ 'is-dragover': isDragOver }"
    @dragover="handleDragOver"
    @dragleave="handleDragLeave"
    @drop.prevent="handleDrop"
    @click="$refs.fileInput.click()"
  >
    <div v-if="uploading" class="upload-content">
      <div class="spinner"></div>
      <p>Analyzing floorplan...</p>
      <small>Converting PDF & Detect Rooms</small>
    </div>

    <div v-else class="upload-content">
      <div class="icon"></div>
      <h3>Drop PDF floorplan here</h3>
      <p class="subtext">or click to browse</p>
      <input
        type="file"
        ref="fileInput"
        @change="handleFileSelect"
        accept=".pdf,.png,.jpg,.jpeg"
        style="display: none"
      />
    </div>

    <div v-if="error" class="error">
      {{ error }}
    </div>
  </div>
</template>

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
