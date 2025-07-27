<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { API_HOST } from '../config'

// 響應式資料
const username = ref('')
const roomId = ref('')
const isJoining = ref(false)
const isCreating = ref(false)
const showLeaderboard = ref(false)
const isLoadingLeaderboard = ref(false)
const leaderboardData = ref([])
const router = useRouter()

/**
 * 頁面載入時檢查使用者是否已登入
 */
onMounted(() => {
  const storedUsername = localStorage.getItem('username')
  const token = localStorage.getItem('token')
  
  if (!storedUsername || !token) {
    router.push('/')
  } else {
    username.value = storedUsername
  }
})

/**
 * 創建新房間
 */
const createRoom = async () => {
  isCreating.value = true
  
  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`${API_HOST}/api/v1/auth/createGame`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      }
      // 不包含 body
    })

    const data = await response.json()

    if (response.ok) {
      // 將房間ID儲存到localStorage
      localStorage.setItem('currentRoomId', data.roomId || data.game_id || data.gameId)
      // 跳轉到遊戲頁面
      router.push('/game')
    } else {
      alert(data.error || '創建房間失敗')
    }
  } catch (error) {
    console.error('創建房間錯誤:', error)
    alert('網路錯誤，請稍後再試')
  } finally {
    isCreating.value = false
  }
}

/**
 * 進入特定房間
 */
const joinSpecificRoom = async () => {
  if (!roomId.value.trim()) {
    alert('請輸入房間ID')
    return
  }

  isJoining.value = true

  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`${API_HOST}/api/v1/auth/joinGame`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        game_id: roomId.value.trim()
      })
    })

    const data = await response.json()

    if (response.ok) {
      // 將房間ID儲存到localStorage
      localStorage.setItem('currentRoomId', data.roomId || data.game_id || data.gameId || roomId.value.trim())
      // 跳轉到遊戲頁面
      router.push('/game')
    } else {
      alert(data.error || '加入房間失敗')
    }
  } catch (error) {
    console.error('加入特定房間錯誤:', error)
    alert('網路錯誤，請稍後再試')
  } finally {
    isJoining.value = false
  }
}

/**
 * 獲取排行榜資料
 */
const fetchLeaderboard = async () => {
  isLoadingLeaderboard.value = true
  
  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`${API_HOST}/api/v1/auth/leaderboard`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    const data = await response.json()

    if (response.ok) {
      // 直接使用回傳的陣列
      leaderboardData.value = Array.isArray(data) ? data : []
    } else {
      console.error('獲取排行榜失敗:', data.error)
      leaderboardData.value = []
    }
  } catch (error) {
    console.error('獲取排行榜錯誤:', error)
    leaderboardData.value = []
  } finally {
    isLoadingLeaderboard.value = false
  }
}

/**
 * 切換排行榜顯示狀態
 */
const toggleLeaderboard = async () => {
  if (!showLeaderboard.value) {
    await fetchLeaderboard()
  }
  showLeaderboard.value = !showLeaderboard.value
}

/**
 * 關閉排行榜
 */
const closeLeaderboard = () => {
  showLeaderboard.value = false
}

/**
 * 處理房間ID輸入的 Enter 鍵事件
 */
const handleRoomIdKeyPress = (event) => {
  if (event.key === 'Enter') {
    joinSpecificRoom()
  }
}

/**
 * 登出功能
 */
const logout = () => {
  localStorage.removeItem('username')
  localStorage.removeItem('token')
  localStorage.removeItem('currentRoomId')
  router.push('/')
}
</script>

<template>
  <div class="room-selection-container">
    <div class="room-selection-content">
      <!--  左上角排行榜按鈕 -->
      <div class="top-left-actions">
        <button @click="toggleLeaderboard" class="leaderboard-button">
          🏆 排行榜
        </button>
      </div>
      
      <div class="user-info">
        <h2>歡迎, {{ username }}!</h2>
        <div class="user-actions">
          <button @click="logout" class="logout-button">登出</button>
        </div>
      </div>
      
      <div class="room-options">
        <h1>選擇加入方式</h1>
        
        <div class="option-card">
          <h3>🏠 創建房間</h3>
          <p>創建一個新的遊戲房間，您將成為房主</p>
          <button 
            @click="createRoom" 
            :disabled="isCreating"
            class="join-button create"
          >
            <span v-if="isCreating">創建中...</span>
            <span v-else>創建房間</span>
          </button>
        </div>
        
        <div class="option-card">
          <h3>🎯 加入特定房間</h3>
          <p>輸入房間ID加入指定的遊戲房間</p>
          <input
            v-model="roomId"
            type="text"
            placeholder="請輸入房間ID"
            @keydown="handleRoomIdKeyPress"
            :disabled="isJoining"
            class="room-input"
          />
          <button 
            @click="joinSpecificRoom" 
            :disabled="isJoining || !roomId.trim()"
            class="join-button specific"
          >
            <span v-if="isJoining">加入中...</span>
            <span v-else>加入房間</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 排行榜彈出視窗 -->
    <div v-if="showLeaderboard" class="leaderboard-overlay" @click="closeLeaderboard">
      <div class="leaderboard-modal" @click.stop>
        <div class="leaderboard-header">
          <h3>🏆 分數排行榜</h3>
          <button @click="closeLeaderboard" class="close-button">✕</button>
        </div>
        
        <div class="leaderboard-content">
          <div v-if="isLoadingLeaderboard" class="loading-state">
            <div class="loading-spinner"></div>
            <p>載入排行榜中...</p>
          </div>
          
          <div v-else-if="leaderboardData.length === 0" class="empty-state">
            <div class="empty-icon">📊</div>
            <p>暫無排行榜資料</p>
          </div>
          
          <div v-else class="leaderboard-list">
            <div class="leaderboard-header-row">
              <span class="rank-col">排名</span>
              <span class="player-col">玩家</span>
              <span class="score-col">勝場</span>
              <!-- 移除用戶ID欄位 -->
            </div>
            
            <div 
              v-for="(player, index) in leaderboardData" 
              :key="player.Username + '_' + index"
              :class="['leaderboard-row', { 
                'current-user': player.Username === username,
                'top-three': index < 3
              }]"
            >
              <span class="rank-col">
                <span v-if="index === 0" class="medal gold">🥇</span>
                <span v-else-if="index === 1" class="medal silver">🥈</span>
                <span v-else-if="index === 2" class="medal bronze">🥉</span>
                <span v-else class="rank-text">{{ index + 1 }}</span>
              </span>
              <span class="player-col">{{ player.Username || '未知玩家' }}</span>
              <span class="score-col">{{ player.WinCount || 0 }}</span>
              <!-- 移除用戶ID顯示 -->
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.room-selection-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 2rem;
  background: #1a1a1a;
  box-sizing: border-box;
}

.room-selection-content {
  background: #2d2d2d;
  border-radius: 12px;
  padding: 3rem;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  width: 100%;
  max-width: 800px;
  border: 1px solid #404040;
  position: relative;
}

/*  左上角排行榜按鈕 */
.top-left-actions {
  position: absolute;
  top: 1.5rem;
  left: 1.5rem;
  z-index: 10;
}

.user-info {
  text-align: center;
  margin-bottom: 3rem;
  position: relative;
}

.user-info h2 {
  color: #e8eaed;
  margin: 0;
  font-size: 2rem;
  padding-right: 100px; /* 為登出按鈕留空間 */
}

.user-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
}

.leaderboard-button {
  background: linear-gradient(135deg, #ffd700, #ffa500);
  color: #1a1a1a;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 600;
  white-space: nowrap;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(255, 215, 0, 0.2);
}

.leaderboard-button:hover {
  background: linear-gradient(135deg, #ffa500, #ff8c00);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(255, 215, 0, 0.3);
}

.logout-button {
  background: #e74c3c;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  white-space: nowrap;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(231, 76, 60, 0.2);
}

.logout-button:hover {
  background: #c0392b;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(231, 76, 60, 0.3);
}

.room-options h1 {
  color: #e8eaed;
  text-align: center;
  margin-bottom: 2rem;
  font-size: 2rem;
}

.room-options {
  display: grid;
  grid-template-columns: 1fr;
  gap: 2rem;
}

.room-options > h1 {
  grid-column: 1 / -1;
}

.room-options > .option-card {
  grid-column: 1;
}

.room-options > .option-card:first-of-type {
  grid-column: 1;
}

.room-options > .option-card:last-of-type {
  grid-column: 1;
}

@media (min-width: 769px) {
  .room-options {
    grid-template-columns: 1fr 1fr;
  }
  
  .room-options > .option-card:first-of-type {
    grid-column: 1;
  }
  
  .room-options > .option-card:last-of-type {
    grid-column: 2;
  }
}

.option-card {
  background: #242424;
  padding: 2rem;
  border-radius: 8px;
  text-align: center;
  border: 2px solid #404040;
  transition: border-color 0.3s ease;
}

.option-card:hover {
  border-color: #4db6e6;
}

.option-card h3 {
  color: #e8eaed;
  margin-bottom: 1rem;
  font-size: 1.5rem;
  margin-top: 0;
}

.option-card p {
  color: #a8a8a8;
  margin-bottom: 2rem;
  line-height: 1.5;
}

.room-input {
  width: 100%;
  padding: 10px 12px;
  border: 2px solid #404040;
  border-radius: 6px;
  font-size: 1rem;
  background: #1a1a1a;
  color: #e8eaed;
  margin-bottom: 1.5rem;
  box-sizing: border-box;
}

.room-input:focus {
  outline: none;
  border-color: #4db6e6;
}

.room-input::placeholder {
  color: #a8a8a8;
}

.join-button {
  background: #4db6e6;
  color: #1a1a1a;
  border: none;
  padding: 12px 24px;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.3s ease;
  min-width: 140px;
}

.join-button:hover {
  background: #3a9bc1;
}

.join-button:disabled {
  background: #666;
  cursor: not-allowed;
  opacity: 0.6;
}

/* 創建房間按鈕特殊樣式 */
.join-button.create {
  background: #28a745;
  color: white;
}

.join-button.create:hover {
  background: #218838;
}

.join-button.create:disabled {
  background: #666;
  cursor: not-allowed;
  opacity: 0.6;
}

/* 排行榜彈出視窗樣式 */
.leaderboard-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.75);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  padding: 2rem;
  backdrop-filter: blur(4px);
}

.leaderboard-modal {
  background: #2d2d2d;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
  border: 1px solid #404040;
  max-width: 650px;
  width: 100%;
  max-height: 85vh;
  overflow: hidden;
  animation: modalSlideIn 0.3s ease-out;
}

@keyframes modalSlideIn {
  from {
    opacity: 0;
    transform: scale(0.9) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.leaderboard-header {
  background: linear-gradient(135deg, #1a1a1a, #242424);
  padding: 1.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid #404040;
}

.leaderboard-header h3 {
  color: #e8eaed;
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
}

.close-button {
  background: none;
  border: none;
  color: #a8a8a8;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 50%;
  transition: all 0.3s ease;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-button:hover {
  background: #404040;
  color: #e8eaed;
  transform: scale(1.1);
}

.leaderboard-content {
  padding: 1.5rem;
  max-height: 60vh;
  overflow-y: auto;
}

.loading-state {
  text-align: center;
  padding: 3rem 1rem;
  color: #a8a8a8;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #404040;
  border-top: 3px solid #4db6e6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.empty-state {
  text-align: center;
  padding: 3rem 1rem;
  color: #a8a8a8;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.6;
}

.leaderboard-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.leaderboard-header-row {
  display: grid;
  grid-template-columns: 80px 1fr 100px;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: linear-gradient(135deg, #404040, #333);
  border-radius: 8px;
  font-weight: 700;
  color: #e8eaed;
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.leaderboard-row {
  display: grid;
  grid-template-columns: 80px 1fr 100px;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: #242424;
  border: 1px solid #404040;
  border-radius: 8px;
  align-items: center;
  transition: all 0.3s ease;
  font-size: 0.95rem;
}

.leaderboard-row:hover {
  background: #333;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.leaderboard-row.current-user {
  background: linear-gradient(135deg, rgba(77, 182, 230, 0.15), rgba(77, 182, 230, 0.05));
  border-color: #4db6e6;
  box-shadow: 0 0 15px rgba(77, 182, 230, 0.2);
}

.leaderboard-row.top-three {
  background: linear-gradient(135deg, rgba(255, 215, 0, 0.1), rgba(255, 215, 0, 0.05));
}

.rank-col {
  text-align: center;
  font-weight: 700;
}

.medal {
  font-size: 1.5rem;
  filter: drop-shadow(0 0 5px rgba(255, 215, 0, 0.5));
}

.rank-text {
  color: #a8a8a8;
  font-size: 1.2rem;
  font-weight: 600;
}

.player-col {
  color: #e8eaed;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.current-user .player-col {
  color: #4db6e6;
}

.score-col {
  color: #ffd700;
  font-weight: 700;
  text-align: center;
  font-size: 1.1rem;
}

.games-col {
  color: #a8a8a8;
  text-align: center;
  font-weight: 500;
}

/* 自定義滾動條樣式 */
.leaderboard-content::-webkit-scrollbar {
  width: 8px;
}

.leaderboard-content::-webkit-scrollbar-track {
  background: #404040;
  border-radius: 4px;
}

.leaderboard-content::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, #4db6e6, #3a9bc1);
  border-radius: 4px;
}

.leaderboard-content::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, #3a9bc1, #2980b9);
}

@media (max-width: 768px) {
  .room-selection-container {
    padding: 1rem;
  }
  
  .room-selection-content {
    padding: 2rem;
  }
  
  .top-left-actions {
    top: 1rem;
    left: 1rem;
  }
  
  .user-info h2 {
    font-size: 1.5rem;
    padding-right: 80px;
  }
  
  .user-actions {
    position: static;
    transform: none;
    justify-content: flex-end;
    margin-top: 1rem;
  }
  
  .leaderboard-button,
  .logout-button {
    font-size: 0.8rem;
    padding: 6px 12px;
  }
  
  .room-options h1 {
    font-size: 1.5rem;
  }
  
  .option-card {
    padding: 1.5rem;
  }
  
  .option-card h3 {
    font-size: 1.2rem;
  }
  
  .leaderboard-overlay {
    padding: 1rem;
  }
  
  .leaderboard-modal {
    max-height: 90vh;
  }
  
  .leaderboard-header-row,
  .leaderboard-row {
    grid-template-columns: 60px 1fr 70px 70px;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    font-size: 0.85rem;
  }
  
  .leaderboard-header h3 {
    font-size: 1.2rem;
  }
  
  .medal {
    font-size: 1.2rem;
  }
}

@media (max-width: 480px) {
  .leaderboard-header-row,
  .leaderboard-row {
    grid-template-columns: 50px 1fr 60px;
    gap: 0.25rem;
  }
  
  .games-col {
    display: none;
  }
}
</style>
