import { createRouter, createWebHistory } from "vue-router";
import Home from "../views/Home.vue";
import SpaceTwin from "../views/SpaceTwin.vue";
import TestCrop from "../components/TestCrop.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: Home,
    },
    {
      path: "/test",
      name: "test",
      component: TestCrop,
    },
    {
      path: "/space-twin",
      name: "space-twin",
      component: SpaceTwin,
    },
  ],
});

export default router;
