import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import TestCrop from '../components/TestCrop.vue'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            name: 'home',
            component: Home
        },
        {
            path: '/test',
            name: 'test',
            component: TestCrop
        }
    ]
})

export default router
