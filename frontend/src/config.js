// src/config.js
const env = window.__ENV__ || {}

export const API_HOST = env.API_HOST || 'http://localhost:8080'
export const WEBSOCKET_URL = env.WEBSOCKET_URL || 'ws://localhost:8080'
