import { createRouter, createWebHistory } from 'vue-router'
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // {
    //   path: '/',
    //   name: 'home',
    //   component: HomeView,
    // },
    // 猜數字遊戲相關路由
    {
      path: '/',
      name: '',
      component: () => import('../views/LoginView.vue'),
    },
    {
      path: '/room-selection',
      name: 'roomSelection',
      component: () => import('../views/RoomSelectionView.vue'),
    },
    {
      path: '/game',
      name: 'game',
      component: () => import('../views/GameView.vue'),
    },
  ],
})

export default router