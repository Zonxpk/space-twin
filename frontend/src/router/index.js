import { createRouter, createWebHistory } from "vue-router";
import Upload from "../views/Upload.vue";
import SpaceTwin from "../views/SpaceTwin.vue";
import TestCrop from "../components/TestCrop.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: SpaceTwin,
    },
    {
      path: "/test",
      name: "test",
      component: TestCrop,
    },
    {
      path: "/upload",
      name: "upload",
      component: Upload,
    },
  ],
});

export default router;
