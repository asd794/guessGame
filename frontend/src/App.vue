<script setup>
import { RouterView } from 'vue-router'
</script>

<template>
  <RouterView />
</template>

<style scoped>
/* 移除所有原有的 header 和 nav 樣式 */
</style>

<style scoped>
header {
  line-height: 1.5;
  background: #2d2d2d;
  border-bottom: 1px solid #404040;
  padding: 1rem 0;
}

.wrapper {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
}

nav {
  width: 100%;
  font-size: 1rem;
  text-align: center;
}

nav a.router-link-exact-active {
  color: #4db6e6;
  font-weight: 600;
}

nav a.router-link-exact-active:hover {
  background-color: transparent;
}

nav a {
  display: inline-block;
  padding: 0.5rem 1rem;
  margin: 0 0.5rem;
  text-decoration: none;
  color: #e8eaed;
  border-radius: 4px;
  transition: background-color 0.3s ease;
}

nav a:hover {
  background-color: #404040;
}

@media (min-width: 1024px) {
  nav {
    text-align: left;
  }
  
  .wrapper {
    margin: 0 2rem 0 0;
  }

  header .wrapper {
    display: flex;
    place-items: flex-start;
    flex-wrap: wrap;
  }

  nav {
    text-align: left;
    margin-left: -1rem;
    font-size: 1rem;
    padding: 1rem 0;
    margin-top: 1rem;
  }
}
</style>
