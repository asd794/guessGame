<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import {  WEBSOCKET_URL } from '../config'

// 路由相關
const router = useRouter()

// 使用者資料
const username = ref('')
const roomId = ref('')

// WebSocket 相關
const ws = ref(null)
const isConnected = ref(false)

// 聊天室相關
const messages = ref([])
const newMessage = ref('')
const chatContainer = ref(null)

// 遊戲狀態
const gameState = ref('waiting') // waiting, ready, playing, finished
const isReady = ref(false)
const targetNumber = ref(null)
const guessNumber = ref('')
const guessHistory = ref([])
const gameResult = ref('')
const attempts = ref(0)
// const maxAttempts = ref(10)

// 在現有的 ref 變量後添加
const players = ref([])  // 玩家列表
const roomInfo = ref({   // 房間資訊
  maxPlayers: 4,
  currentPlayers: 0
})

// 遊戲結束相關變數
const gameWinner = ref('')           // 獲勝者名稱
const gameAnswer = ref(null)         // 遊戲答案

/**
 * 頁面初始化
 */
onMounted(() => {
  // 檢查使用者登入狀態
  const storedUsername = localStorage.getItem('username')
  const token = localStorage.getItem('token')
  const currentRoomId = localStorage.getItem('currentRoomId')
  
  if (!storedUsername || !token) {
    router.push('/')
    return
  }
  
  if (!currentRoomId) {
    router.push('/room-selection')
    return
  }
  
  username.value = storedUsername
  roomId.value = currentRoomId
  
  // 連接 WebSocket
  connectWebSocket(token, currentRoomId)
})

/**
 * 頁面卸載時清理
 */
onUnmounted(() => {
  // 離開前送出離開遊戲訊息
  if (ws.value && ws.value.readyState === WebSocket.OPEN) {
    ws.value.send(JSON.stringify({
      type: 'left_game',
      message: 'leave_game',
      gameId: roomId.value,
      from: username.value,
      timestamp: new Date().toISOString()
    }))
  }
  disconnectWebSocket()
})

/**
 * 連接 WebSocket
 */
const connectWebSocket = (token, gameId) => {
  try {
    //  WebSocket 使用 URL 參數傳遞 token
    const wsUrl = `${WEBSOCKET_URL}/api/v1/auth/wsGame?token=${encodeURIComponent(token)}&game_id=${encodeURIComponent(gameId)}`

    console.log('正在連接 WebSocket:', wsUrl)
    ws.value = new WebSocket(wsUrl)
    
    ws.value.onopen = (event) => {
      console.log('WebSocket 連接成功', event)
      isConnected.value = true
      addSystemMessage('已連線至遊戲')
      
    }
    
    ws.value.onmessage = (event) => {
      console.log('收到 WebSocket 訊息:', event.data)
      try {
        const data = JSON.parse(event.data)
        handleWebSocketMessage(data)
      } catch (error) {
        console.error('解析 WebSocket 訊息失敗:', error)
        addSystemMessage('❌ 收到無效的訊息格式')
      }
    }
    
    ws.value.onclose = (event) => {
      console.log('WebSocket 連接關閉', event)
      isConnected.value = false
      addSystemMessage('📡 與伺服器的連接已斷開')
      
      // 401 認證失敗，不要重連
      if (event.code === 1006) {
        addSystemMessage('❌ 認證失敗，請檢查登入狀態')
        return
      }
      
      // 其他錯誤才嘗試重連
      if (!event.wasClean && event.code !== 1000) {
        setTimeout(() => {
          addSystemMessage('🔄 嘗試重新連接...')
          connectWebSocket(token, gameId)
        }, 3000)
      }
    }
    
    ws.value.onerror = (error) => {
      console.error('WebSocket 錯誤:', error)
      addSystemMessage('❌ WebSocket 連接錯誤')
    }
    
  } catch (error) {
    console.error('建立 WebSocket 連接失敗:', error)
    addSystemMessage('❌ 無法連接到遊戲伺服器')
  }
}

/**
 * 斷開 WebSocket 連接
 */
const disconnectWebSocket = () => {
  if (ws.value) {
    console.log('主動斷開 WebSocket 連接')
    ws.value.close(1000, '使用者離開')
    ws.value = null
    isConnected.value = false
  }
}

/**
 * 處理 WebSocket 訊息
 */
const handleWebSocketMessage = (data) => {
  console.log('處理 WebSocket 訊息:', data)
  
  switch (data.type) {
    case 'chat':
      addChatMessage(data.from || data.playerName, data.message, data.timestamp)
      break
      
    case 'system':
      addSystemMessage(data.message)
      break

    case 'player_joined': 
      updatePlayerList(data)
      break

    case 'player_left':   
      addSystemMessage(data.message)
      updatePlayerList(data)
      break
      
    case 'player_ready':
      addSystemMessage(data.message)
      //  修復：添加安全檢查
      if (data.message && data.playerName) {
        updatePlayerReadyStatus(data.playerName, data.message.includes('已準備就緒'))
      }
      break
      
    case 'ready_confirm':
      console.log('收到準備確認:', data.message)
      break
      
    case 'ready_status':
      //  解析準備狀態並更新房間資訊
      if (data.message.includes('準備狀態:')) {
        const match = data.message.match(/準備狀態: (\d+)\/(\d+)/)
        if (match) {
          const [, ready, total] = match
          const readyCount = parseInt(ready)
          const totalCount = parseInt(total)
          
          roomInfo.value.currentPlayers = totalCount
          
          if (readyCount === totalCount && totalCount > 1) {
            addSystemMessage('🎯 所有玩家已準備就緒！')
          } else {
            addSystemMessage(`⏳ ${ready}/${total} 玩家已準備`)
          }
        } else {
          console.log('正則表達式沒有匹配到:', data.message)
          addSystemMessage(data.message)
        }
      } else {
        addSystemMessage(data.message)
      }
      console.log('準備狀態更新:', data.message)
      break

    case 'room_status_update':
      console.log('收到房間狀態更新:', data)
      updateRoomStatus(data)
      
      //  檢查是否為遊戲重置後的更新
      if (data.gameInfo && data.gameInfo.gameStatus === 'waiting') {
        handleGameReset(data)
      }
      break
      
    case 'game_status':
      updateGameStatus(data)
      break
      
    case 'all_players_ready':
    case 'all_ready':
      addSystemMessage(data.message)
      break
      
    //  修改：處理遊戲開始事件（提取 TurnOrder 資訊）
    case 'start_game':
    case 'game_started':
      addSystemMessage('🎯 遊戲開始！')
      gameState.value = 'playing'
      
      //  確保這行代碼存在並被執行
      handleGameStarted(data)
      
      // 重置遊戲相關狀態
      attempts.value = 0
      guessHistory.value = []
      targetNumber.value = null
      gameResult.value = ''
      break
      
    case 'player_guess':
      handleGuessResult(data)
      break
      
    //  處理玩家輪次
    case 'player_turn':
      addSystemMessage(`🎯 ${data.message}`)
      
      if (data.gameInfo && data.gameInfo.CurrentTurn !== undefined) {
        currentTurnIndex.value = data.gameInfo.CurrentTurn
        
        //  根據 CurrentTurn 找到對應的玩家並更新
        const currentPlayer = players.value.find(p => p.turnOrder === data.gameInfo.CurrentTurn)
        if (currentPlayer) {
          currentTurnPlayer.value = currentPlayer.name
          console.log('✅ 當前輪次玩家:', currentPlayer.name, 'TurnOrder:', currentPlayer.turnOrder)
          
          //  檢查是否是新回合的開始（該玩家之前已經猜過）
          if (currentPlayer.guessed) {
            console.log('🔄 檢測到新回合開始，重置玩家猜測狀態')
            resetAllPlayersGuessStatus()
          }
        } else {
          console.warn('⚠️ 找不到 TurnOrder 為', data.gameInfo.CurrentTurn, '的玩家')
        }
      } else {
        //  備用方案：使用事件中的 playerName
        highlightCurrentPlayer(data.playerName)
        
        //  檢查是否需要重置狀態
        const player = players.value.find(p => p.name === data.playerName)
        if (player && player.guessed) {
          console.log('🔄 檢測到新回合開始（備用方案），重置玩家猜測狀態')
          resetAllPlayersGuessStatus()
        }
      }
      break
      
    //  處理新回合開始事件
    case 'new_round':
    case 'round_start':
      handleNewRound(data)
      break
      
    //  處理輪次重置事件
    case 'reset_turns':
    case 'turns_reset':
      handleTurnsReset(data)
      break
      
    //  處理所有人都猜過但沒人猜中的情況
    case 'round_complete':
    case 'all_guessed':
      handleRoundComplete(data)
      break
      
    case 'game_end':
      handleGameEnd(data)
      break
      
    //  處理遊戲結束事件
    case 'game_over':
      handleGameOver(data)
      break
      
    case 'error':
      addSystemMessage(`❌ ${data.message}`)
      break
      
    default:
      console.log('未處理的訊息類型:', data.type, data)
      if (data.message) {
        addSystemMessage(`📋 ${data.message}`)
      }
  }
}

// 高亮當前輪次的玩家
const currentTurnPlayer = ref('')

const highlightCurrentPlayer = (playerName) => {
  currentTurnPlayer.value = playerName
  console.log('🔍 更新當前輪次玩家:', playerName)
  console.log('🔍 currentTurnPlayer.value:', currentTurnPlayer.value)
  console.log('🔍 當前玩家列表:', players.value.map(p => ({ name: p.name, turnOrder: p.turnOrder })))
}
/**
 * 更新玩家列表
 */
const updatePlayerList = (data) => {
  // 這裡可以根據後端提供的玩家資料更新
  if (data.playerCount !== undefined) {
    roomInfo.value.currentPlayers = data.playerCount
  }
  
  // 如果後端提供完整玩家列表，可以這樣更新：
  if (data.players) {
    players.value = data.players
  }
}

/**
 * 更新玩家準備狀態
 */
const updatePlayerReadyStatus = (playerName, isReady) => {
  //  添加參數檢查
  if (!playerName) {
    console.warn('updatePlayerReadyStatus: playerName 為空')
    return
  }
  
  const playerIndex = players.value.findIndex(p => p.name === playerName)
  if (playerIndex !== -1) {
    players.value[playerIndex].isReady = isReady
  } else {
    // 如果玩家不在列表中，添加他們
    players.value.push({
      name: playerName,
      isReady: isReady,
      uuid: '',  //  添加默認值
      turnOrder: 0,  //  添加默認值
      guessed: false,  //  添加默認值
      score: 0  //  添加默認值
    })
  }
}

/**
 * 更新遊戲狀態
 */
const updateGameStatus = (data) => {
  if (data.players) {
    players.value = data.players
  }
  if (data.gameInfo) {
    roomInfo.value = { ...roomInfo.value, ...data.gameInfo }
  }
}

/**
 * 更新房間狀態
 */
const updateRoomStatus = (data) => {
  console.log('開始更新房間狀態:', data)
  
  //  更新玩家列表
  if (data.players && Array.isArray(data.players)) {
    players.value = data.players.map(player => ({
      name: player.name,
      uuid: player.uuid,
      isReady: player.isReady || false,
      //  處理重置後的遊戲相關屬性
      turnOrder: player.turnOrder || 0,
      guessed: player.guessed || false,
      score: player.score || 0
    }))
    console.log('玩家列表已更新:', players.value)
  }
  
  //  更新房間資訊
  if (data.gameInfo) {
    roomInfo.value = {
      ...roomInfo.value,
      maxPlayers: data.gameInfo.maxPlayers || 4,
      currentPlayers: data.gameInfo.currentPlayers || 0,
      readyCount: data.gameInfo.readyCount || 0,
      gameStatus: data.gameInfo.gameStatus || 'waiting',
      minRange: data.gameInfo.minRange || 1,
      maxRange: data.gameInfo.maxRange || 100
    }
    console.log('房間資訊已更新:', roomInfo.value)
  }
  
  //  修改：確保在玩家列表更新後再更新當前玩家的準備狀態
  const currentPlayer = players.value.find(p => p.name === username.value)
  if (currentPlayer) {
    const previousReady = isReady.value
    isReady.value = currentPlayer.isReady
    console.log('當前玩家準備狀態已更新:', previousReady, '->', isReady.value)
  }
  
  //  更新遊戲狀態
  if (data.gameInfo && data.gameInfo.gameStatus) {
    const previousGameState = gameState.value
    gameState.value = data.gameInfo.gameStatus
    console.log('遊戲狀態已更新:', previousGameState, '->', gameState.value)
    
    //  如果從 finished 狀態變為 waiting 狀態，表示遊戲被重置
    if (previousGameState === 'finished' && gameState.value === 'waiting') {
      console.log('🔄 檢測到遊戲重置 (finished -> waiting)')
      // handleGameReset 會在上面的 room_status_update case 中被調用
    }
  }
}

//  輪次相關的響應式變數
const turnOrder = ref([])           // 輪次順序陣列
const currentTurnIndex = ref(0)     // 當前輪次索引


/**
 * 發送 WebSocket 訊息
 */
const sendWebSocketMessage = (messageData) => {
  if (ws.value && ws.value.readyState === WebSocket.OPEN) {
    const message = {
      ...messageData,
      gameId: roomId.value,
      from: username.value,
      timestamp: new Date().toISOString()
    }
    
    console.log('發送 WebSocket 訊息:', message)
    ws.value.send(JSON.stringify(message))
  } else {
    console.error('WebSocket 未連接，無法發送訊息')
    addSystemMessage('❌ 無法發送訊息：未連接到伺服器')
  }
}

/**
 * 添加系統訊息到聊天室
 */
const addSystemMessage = (content) => {
  messages.value.push({
    id: Date.now(),
    type: 'system',
    content,
    timestamp: new Date().toLocaleTimeString()
  })
  scrollToBottom()
}

/**
 * 添加聊天訊息到聊天室
 */
const addChatMessage = (from, content, timestamp = null) => {
  messages.value.push({
    id: Date.now(),
    type: 'user',
    username: from,
    content,
    timestamp: timestamp ? new Date(timestamp).toLocaleTimeString() : new Date().toLocaleTimeString()
  })
  scrollToBottom()
}

/**
 * 發送聊天訊息
 */
const sendMessage = () => {
  if (newMessage.value.trim()) {
    sendWebSocketMessage({
      type: 'chat',
      message: newMessage.value.trim()
    })
    newMessage.value = ''
  }
}

/**
 * 處理聊天輸入的 Enter 鍵事件
 */
const handleMessageKeyPress = (event) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    sendMessage()
  }
}

/**
 * 滾動聊天室到底部
 */
const scrollToBottom = () => {
  nextTick(() => {
    if (chatContainer.value) {
      chatContainer.value.scrollTop = chatContainer.value.scrollHeight
    }
  })
}

/**
 * 玩家準備
 */
const toggleReady = () => {
  console.log('準備切換前:', isReady.value)
  
  const newReadyState = !isReady.value
  
  sendWebSocketMessage({
    type: 'player_ready',
    message: newReadyState ? 'ready' : 'not_ready'
  })
  
  //  添加本地狀態更新以提供即時反饋
  isReady.value = newReadyState
  
  //  更新玩家列表中的狀態
  const currentPlayerIndex = players.value.findIndex(p => p.name === username.value)
  if (currentPlayerIndex >= 0) {
    players.value[currentPlayerIndex].isReady = newReadyState
  }
  
  console.log('準備切換後:', isReady.value)
  addSystemMessage(`${newReadyState ? '🟢 你已準備就緒' : '⏸️ 你取消了準備'}`)
}

/**
 * 開始遊戲
 */
const startGame = () => {
  console.log('嘗試開始遊戲:', {
    isReady: isReady.value,
    isConnected: isConnected.value,
    gameState: gameState.value,
    playerCount: players.value.length,
    readyCount: players.value.filter(p => p.isReady).length
  })
  
  if (!isReady.value) {
    alert('請先點擊準備按鈕')
    return
  }
  
  if (!isConnected.value) {
    alert('網路連接異常，請檢查連線狀態')
    return
  }
  
  if (gameState.value !== 'waiting') {
    alert('遊戲狀態異常，請重新整理頁面')
    return
  }
  
  //  檢查所有玩家是否都已準備
  const readyPlayers = players.value.filter(p => p.isReady)
  const totalPlayers = players.value.length
  
  if (readyPlayers.length < totalPlayers) {
    alert(`還有 ${totalPlayers - readyPlayers.length} 位玩家未準備就緒`)
    return
  }
  
  if (totalPlayers < 2) {
    alert('至少需要 2 位玩家才能開始遊戲')
    return
  }
  
  sendWebSocketMessage({
    type: 'start_game',
    message: 'start'
  })
  
  addSystemMessage('🎯 正在啟動遊戲...')
}

/**
 * 提交猜測
 */
const submitGuess = () => {
  //  檢查是否輪到當前玩家
  if (currentTurnPlayer.value !== username.value) {
    alert('還沒輪到你猜測！')
    return
  }
  
  const guess = parseInt(guessNumber.value)
  
  // 驗證輸入
  if (isNaN(guess) || guess < 1 || guess > 100) {
    alert('請輸入 1-100 之間的數字')
    return
  }
  
  sendWebSocketMessage({
    type: 'player_guess',
    message: guess.toString()
  })
  
  guessNumber.value = ''
}

/**
 * 滾動猜測記錄到底部
 */
const scrollHistoryToBottom = () => {
  nextTick(() => {
    const container = document.querySelector('.history-list')
    if (container) {
      container.scrollTop = container.scrollHeight
    }
  })
}

/**
 * 滾動猜測記錄到頂部
 */
// const scrollHistoryToTop = () => {
//   nextTick(() => {
//     const container = document.querySelector('.history-list')
//     if (container) {
//       container.scrollTop = 0
//     }
//   })
// }

/**
 * 處理猜測結果
 */
const handleGuessResult = (data) => {
  attempts.value++
  
  // 提取猜測資訊
  const playerName = data.playerName || data.from
  const message = data.message || ''
  
  console.log('🔍 handleGuessResult - 處理猜測結果')
  console.log('🔍 猜測玩家:', playerName)
  console.log('🔍 猜測結果:', message)
  
  // 更新該玩家的猜測狀態
  const playerIndex = players.value.findIndex(p => p.name === playerName)
  if (playerIndex >= 0) {
    players.value[playerIndex].guessed = true
    console.log('🔍 已標記玩家為已猜測:', players.value[playerIndex])
  }
  
  // 更新輪次順序中的猜測狀態
  const turnIndex = turnOrder.value.findIndex(p => p.name === playerName)
  if (turnIndex >= 0) {
    turnOrder.value[turnIndex].guessed = true
  }
  
  // 記錄猜測歷史
  guessHistory.value.push({
    attempt: attempts.value,
    player: playerName,
    message: message,
    result: message,
    timestamp: data.timestamp || new Date().toLocaleTimeString()
  })
  
  // 自動滾動到最新記錄
  nextTick(() => {
    scrollHistoryToBottom()
  })
  
  addSystemMessage(message)
  
  //  檢查是否遊戲結束
  if (message && (message.includes('猜中') || message.includes('恭喜'))) {
    gameResult.value = 'win'
    gameState.value = 'finished'
    addSystemMessage(`🎉 ${playerName} 獲勝！`)
  } else {
    //  檢查是否所有人都猜過了
    const allGuessed = players.value.every(player => player.guessed)
    if (allGuessed) {
      console.log('🔄 所有玩家都已猜過，等待後端決定是否開始新回合')
      addSystemMessage('📋 本輪所有玩家都已猜測完畢，等待結果...')
    } else {
      console.log('🔍 等待後端發送 player_turn 事件')
    }
  }
}

/**
 * 處理遊戲結束
 */
const handleGameEnd = (data) => {
  gameState.value = 'finished'
  gameResult.value = data.result || 'finished'
  targetNumber.value = data.answer || data.targetNumber
  addSystemMessage(`🎯 遊戲結束！答案是：${targetNumber.value}`)
}

/**
 * 處理猜測輸入的 Enter 鍵事件
 */
const handleGuessKeyPress = (event) => {
  if (event.key === 'Enter') {
    submitGuess()
  }
}

/**
 * 重新開始遊戲
 */
const restartGame = () => {
  //  修復：在發送重置請求前，先本地重置輪次狀態
  console.log('🔄 重新開始遊戲 - 預清理狀態')
  
  //  預先重置輪次相關狀態
  currentTurnPlayer.value = ''
  currentTurnIndex.value = 0
  turnOrder.value = []
  
  //  預先重置遊戲狀態
  gameState.value = 'waiting'
  isReady.value = false
  targetNumber.value = null
  guessNumber.value = ''
  guessHistory.value = []
  gameResult.value = ''
  attempts.value = 0
  
  //  預先重置遊戲結束相關變數
  gameWinner.value = ''
  gameAnswer.value = null
  
  //  預先重置玩家猜測狀態
  players.value.forEach(player => {
    player.guessed = false
    player.isReady = false  // 重置準備狀態
  })
  
  sendWebSocketMessage({
    type: 'game_reset',
    message: 'restart'
  })
  
  //  修改：只在主動重新開始時顯示訊息
  addSystemMessage('🔄 遊戲已重置，可以重新開始準備!')
  
  console.log('✅ 重新開始遊戲 - 狀態已預清理:', {
    gameState: gameState.value,
    currentTurnPlayer: currentTurnPlayer.value,
    players: players.value.map(p => ({ 
      name: p.name, 
      isReady: p.isReady, 
      guessed: p.guessed 
    }))
  })
}

/**
 * 離開房間，回到房間選擇頁面
 */
const leaveRoom = () => {
  if (confirm('確定要離開房間嗎？')) {
    // 發送離開遊戲訊息
    sendWebSocketMessage({
      type: 'left_game',
      message: 'leave_game'
    })
    

    
    // 等待短暫時間確保訊息發送完成
    setTimeout(() => {
      // 斷開 WebSocket 連接
      disconnectWebSocket()
      
      // 清除當前房間ID
      localStorage.removeItem('currentRoomId')
      
      // 跳轉回房間選擇頁面
      router.push('/room-selection')
    }, 100) // 等待 100ms 確保訊息發送
  }
}

/**
 * 獲取遊戲狀態文字
 */
const getGameStateText = () => {
  switch (gameState.value) {
    case 'waiting': return '等待中'
    case 'ready': return '準備中'
    case 'playing': return '遊戲中'
    case 'finished': return '已結束'
    default: return '未知'
  }
}

//  處理遊戲開始的詳細邏輯
const handleGameStarted = (data) => {
  console.log('🎯 handleGameStarted 被調用了！')
  console.log('處理遊戲開始資訊:', data)
  console.log('當前 gameState:', gameState.value)
  
  //  處理玩家和輪次資訊
  if (data.gameInfo && data.gameInfo.Players) {
    console.log('✅ 找到 Players 資料:', data.gameInfo.Players)
    const playersData = data.gameInfo.Players
    
    //  1. 按 TurnOrder 排序玩家
    const sortedPlayers = playersData.sort((a, b) => a.TurnOrder - b.TurnOrder)
    console.log('✅ 排序後的玩家:', sortedPlayers)
    
    //  2. 更新玩家列表
    players.value = sortedPlayers.map(player => ({
      name: player.Name || '',
      uuid: player.Uuid || '',
      isReady: player.Ready ?? false,
      turnOrder: player.TurnOrder ?? 0,
      guessed: player.Guessed ?? false,
      score: player.Score ?? 0
    }))
    
    //  3. 建立輪次順序陣列
    turnOrder.value = sortedPlayers.map((player, index) => ({
      uuid: player.Uuid,
      name: player.Name,
      turnOrder: player.TurnOrder,
      position: index + 1,  // 顯示位置（1-based）
      guessed: player.Guessed
    }))
    
    //  4. 修復：強制設定當前輪次為第一個玩家（turnOrder = 0）
    const firstPlayer = sortedPlayers.find(player => player.TurnOrder === 0)
    if (firstPlayer) {
      currentTurnIndex.value = 0
      currentTurnPlayer.value = firstPlayer.Name
      console.log('🎯 強制設定第一個玩家為當前輪次:', firstPlayer.Name)
    } else {
      //  備用方案：如果沒有找到 turnOrder = 0 的玩家，使用第一個玩家
      currentTurnIndex.value = 0
      currentTurnPlayer.value = sortedPlayers[0].Name
      console.log('🎯 備用方案：使用第一個玩家作為當前輪次:', sortedPlayers[0].Name)
    }
    
    console.log('✅ turnOrder.value:', turnOrder.value)
    console.log('✅ currentTurnIndex.value:', currentTurnIndex.value)
    console.log('✅ currentTurnPlayer.value:', currentTurnPlayer.value)
    console.log('✅ players.value:', players.value)
    console.log('✅ gameState.value:', gameState.value)
  } else {
    console.log('❌ 沒有找到 Players 資料')
  }
}


/**
 * 處理遊戲結束事件
 */
const handleGameOver = (data) => {
  console.log('🎉 遊戲結束:', data)
  
  //  更新遊戲狀態
  gameState.value = 'finished'
  gameResult.value = 'finished'
  
  //  設定獲勝者
  gameWinner.value = data.playerName || ''
  
  //  從訊息中提取答案（如："答案是 47"）
  const answerMatch = data.message.match(/答案是\s*(\d+)/)
  if (answerMatch) {
    gameAnswer.value = parseInt(answerMatch[1])
    targetNumber.value = parseInt(answerMatch[1])
  }
  
  //  重置當前輪次
  currentTurnPlayer.value = ''
  
  //  將所有玩家標記為已結束
  players.value.forEach(player => {
    player.guessed = true
  })
  
  //  添加系統訊息
  addSystemMessage(data.message)
  addSystemMessage(`🏆 獲勝者：${gameWinner.value}`)
  
  console.log('遊戲狀態已更新:', {
    gameState: gameState.value,
    winner: gameWinner.value,
    answer: gameAnswer.value
  })
}

/**
 * 處理新回合開始
 */
const handleNewRound = (data) => {
  console.log('🔄 新回合開始:', data)
  
  //  重置所有玩家的猜測狀態
  players.value.forEach(player => {
    player.guessed = false
  })
  
  //  重置輪次順序中的猜測狀態
  if (turnOrder.value.length > 0) {
    turnOrder.value.forEach(player => {
      player.guessed = false
    })
  }
  
  //  修復：重置當前輪次到第一個玩家（turnOrder = 0）
  currentTurnIndex.value = 0
  if (players.value.length > 0) {
    //  按 turnOrder 排序找到第一個玩家（turnOrder = 0）
    const sortedPlayers = [...players.value].sort((a, b) => (a.turnOrder || 0) - (b.turnOrder || 0))
    const firstPlayer = sortedPlayers.find(p => p.turnOrder === 0)
    
    if (firstPlayer) {
      currentTurnPlayer.value = firstPlayer.name
      console.log('🎯 新回合設定第一個玩家:', firstPlayer.name)
    } else {
      // 備用方案
      currentTurnPlayer.value = sortedPlayers[0].name
      console.log('🎯 新回合備用方案，使用第一個玩家:', sortedPlayers[0].name)
    }
  }
  
  //  添加系統訊息
  addSystemMessage(data.message || '🔄 新回合開始，所有玩家可以重新猜測！')
  
  console.log('✅ 新回合狀態已更新:', {
    players: players.value.map(p => ({ name: p.name, guessed: p.guessed, turnOrder: p.turnOrder })),
    currentTurnPlayer: currentTurnPlayer.value,
    currentTurnIndex: currentTurnIndex.value
  })
}

/**
 * 處理輪次重置
 */
const handleTurnsReset = (data) => {
  console.log('🔄 輪次重置:', data)
  
  //  重置所有玩家的猜測狀態
  resetAllPlayersGuessStatus()
  
  //  更新當前輪次玩家
  if (data.currentPlayer) {
    currentTurnPlayer.value = data.currentPlayer
    highlightCurrentPlayer(data.currentPlayer)
  } else if (data.gameInfo && data.gameInfo.CurrentTurn !== undefined) {
    const currentPlayer = players.value.find(p => p.turnOrder === data.gameInfo.CurrentTurn)
    if (currentPlayer) {
      currentTurnPlayer.value = currentPlayer.name
      currentTurnIndex.value = data.gameInfo.CurrentTurn
    }
  }
  
  addSystemMessage(data.message || '🔄 輪次已重置，開始新的回合！')
}

/**
 * 處理回合完成（所有人都猜過但沒人猜中）
 */
const handleRoundComplete = (data) => {
  console.log('🔄 回合完成，準備開始新回合:', data)
  
  //  重置所有玩家的猜測狀態
  resetAllPlayersGuessStatus()
  
  //  重置到第一個玩家開始新回合
  currentTurnIndex.value = 0
  if (players.value.length > 0) {
    //  按 turnOrder 排序找到第一個玩家
    const sortedPlayers = [...players.value].sort((a, b) => (a.turnOrder || 0) - (b.turnOrder || 0))
    currentTurnPlayer.value = sortedPlayers[0].name
  }
  
  //  添加系統訊息
  addSystemMessage(data.message || '📋 本回合結束，沒有人猜中！開始新的回合...')
  addSystemMessage(`🎯 輪到 ${currentTurnPlayer.value} 開始新回合的猜測`)
  
  console.log('✅ 新回合已開始:', {
    currentTurnPlayer: currentTurnPlayer.value,
    playersStatus: players.value.map(p => ({ name: p.name, guessed: p.guessed, turnOrder: p.turnOrder }))
  })
}

/**
 * 重置所有玩家的猜測狀態
 */
const resetAllPlayersGuessStatus = () => {
  console.log('🔄 重置所有玩家的猜測狀態')
  
  //  重置 players 列表中的猜測狀態
  players.value.forEach(player => {
    player.guessed = false
  })
  
  //  重置 turnOrder 列表中的猜測狀態
  turnOrder.value.forEach(player => {
    player.guessed = false
  })
  
  console.log('✅ 所有玩家猜測狀態已重置')
}

/**
 * 處理遊戲重置
 */
const handleGameReset = (data) => {
  console.log('🔄 處理遊戲重置:', data)
  
  //  重置遊戲狀態
  gameState.value = 'waiting'
  targetNumber.value = null
  guessNumber.value = ''
  guessHistory.value = []
  gameResult.value = ''
  attempts.value = 0
  
  //  重置遊戲結束相關變數
  gameWinner.value = ''
  gameAnswer.value = null
  
  //  修復：完全重置輪次相關變數
  currentTurnPlayer.value = ''
  currentTurnIndex.value = 0
  turnOrder.value = []
  
  //  修復：重置所有玩家的遊戲相關狀態
  players.value.forEach(player => {
    player.guessed = false
    player.score = player.score || 0  // 保持分數，只重置猜測狀態
  })
  
  //  修改：從後端數據中獲取當前玩家的準備狀態，而不是直接設為 false
  const currentPlayer = players.value.find(p => p.name === username.value)
  if (currentPlayer) {
    isReady.value = currentPlayer.isReady || false
    console.log('從玩家列表更新準備狀態:', isReady.value)
  } else {
    isReady.value = false
    console.log('玩家不在列表中，設置準備狀態為 false')
  }
  
  //  修改：移除系統訊息，避免重複顯示
  // addSystemMessage('🎮 遊戲已重置，可以重新開始準備!')
  
  console.log('✅ 遊戲重置完成:', {
    gameState: gameState.value,
    isReady: isReady.value,
    currentTurnPlayer: currentTurnPlayer.value,
    currentTurnIndex: currentTurnIndex.value,
    turnOrder: turnOrder.value,
    players: players.value.map(p => ({ 
      name: p.name, 
      isReady: p.isReady, 
      guessed: p.guessed,
      turnOrder: p.turnOrder
    }))
  })
}

// 在 <script setup> 區域新增一個方法用於提取猜測數字
// const extractGuessNum = (msg) => {
//   // 只抓 "猜測 NN" 的 NN
//   const match = msg && msg.match(/猜測\s*(\d+)/)
//   return match ? match[1] : ''
// }
// const extractResult = (msg) => {
//   // 只抓 "結果："後面的內容
//   const match = msg && msg.match(/結果：(.+)/)
//   return match ? match[1] : msg
// }

const onLeaveRoomClick = () => {
  if (gameState.value === 'playing') {
    alert('遊戲進行中，無法離開')
    return
  }
  leaveRoom()
}
</script>

<template>
  <div class="game-container">
    <!-- 頂部資訊欄 -->
    <div class="game-header">
      <div class="room-info">
        <h2>房間: {{ roomId }}</h2>
        <span class="player-name">玩家: {{ username }}</span>
        <span :class="['connection-status', { connected: isConnected }]">
          {{ isConnected ? '🟢 已連接' : '🔴 未連接' }}
        </span>
      </div>
      <button 
        @click="onLeaveRoomClick" 
        class="leave-button"
        :class="{ disabled: gameState === 'playing' }"
        tabindex="0"
      >
        離開房間
      </button>
    </div>

    <!-- 主遊戲區域 -->
    <div class="game-main">
      <!-- 左側聊天室 -->
      <div class="chat-section">
        <h3>💬 聊天室</h3>
        
        <!-- 聊天訊息區域 -->
        <div ref="chatContainer" class="chat-messages">
          <div
            v-for="message in messages"
            :key="message.id"
            :class="['message', message.type]"
          >
            <div v-if="message.type === 'system'" class="system-message">
              <span class="timestamp">{{ message.timestamp }}</span>
              {{ message.content }}
            </div>
            <div v-else class="user-message">
              <div class="message-header">
                <span class="username">{{ message.username }}</span>
                <span class="timestamp">{{ message.timestamp }}</span>
              </div>
              <div class="message-content">{{ message.content }}</div>
            </div>
          </div>
        </div>
        
        <!-- 訊息輸入區域 -->
        <div class="chat-input">
          <input
            v-model="newMessage"
            type="text"
            placeholder="輸入訊息..."
            maxlength="200"
            @keypress="handleMessageKeyPress"
            :disabled="!isConnected"
            class="message-input"
          />
          <button 
            @click="sendMessage" 
            :disabled="!newMessage.trim() || !isConnected" 
            class="send-button"
          >
            發送
          </button>
        </div>
      </div>

      <!--  中間玩家狀態面板 -->
      <div class="players-section">
        <h3>👥 玩家狀態</h3>
        
        <!-- 房間資訊 -->
        <div class="room-status">
          <div class="room-stats">
            <div class="stat-item">
              <span class="stat-label">房間人數</span>
              <span class="stat-value">{{ roomInfo.currentPlayers }}/{{ roomInfo.maxPlayers }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">遊戲狀態</span>
              <span :class="['stat-value', 'status', gameState]">
                {{ getGameStateText() }}
              </span>
            </div>
            
            <!--  優勝者顯示 -->
            <div v-if="gameState === 'finished' && gameWinner" class="stat-item">
              <span class="stat-label">優勝者</span>
              <span class="stat-value winner">
                🏆 {{ gameWinner }}
              </span>
            </div>
            
            <!--  答案顯示 -->
            <div v-if="gameState === 'finished' && gameAnswer" class="stat-item">
              <span class="stat-label">答案</span>
              <span class="stat-value answer">
                🎯 {{ gameAnswer }}
              </span>
            </div>
          </div>
        </div>



        <!-- 玩家列表 -->
        <div class="players-list">
          <div class="players-header">
            <span>玩家</span>
            <span>順位</span>
            <span>狀態</span>
          </div>
          
          <div 
            v-for="player in players" 
            :key="player.uuid"
            :class="['player-item', { 
              'current-player': player.name === username,
              'ready': player.isReady,
              'current-turn': gameState === 'playing' && player.name === currentTurnPlayer && !player.guessed,
              'guessed': player.guessed,
              'winner': gameState === 'finished' && player.name === gameWinner
            }]"
          >
            <div class="player-info">
              <span class="player-name">
                {{ player.name }}
                <span v-if="player.name === username" class="you-indicator">(你)</span>
                <!--  優勝者標記 -->
                <span v-if="gameState === 'finished' && player.name === gameWinner" class="winner-indicator">👑</span>
              </span>
            </div>
            <div class="turn-order-display">
              <span class="order-number">第{{ (player.turnOrder || 0) + 1 }}位</span>
            </div>
            <div class="player-status">
              <span v-if="gameState === 'playing'" :class="['game-status', {
                'current': player.name === currentTurnPlayer && !player.guessed,
                'guessed': player.guessed,
                'waiting': player.name !== currentTurnPlayer && !player.guessed
              }]">
                <span v-if="player.guessed">✅ 已猜測</span>
                <span v-else-if="player.name === currentTurnPlayer">🎯 輪次中</span>
                <span v-else>⏳ 等待中</span>
              </span>
              <!--  遊戲結束狀態 -->
              <span v-else-if="gameState === 'finished'" :class="['final-status', {
                'winner': player.name === gameWinner,
                'participant': player.name !== gameWinner
              }]">
                <span v-if="player.name === gameWinner">🏆 獲勝</span>
                <span v-else>🎮 參與</span>
              </span>
              <span v-else :class="['ready-indicator', { ready: player.isReady }]">
                {{ (player.isReady ?? false) ? '✅ 已準備' : '⏳ 未準備' }}
              </span>
            </div>
          </div>
        </div>

        <!-- 快速準備區域 -->
        <div class="quick-actions">
          <button
            @click="toggleReady"
            :class="['ready-toggle', { active: isReady }]"
            :disabled="!isConnected || gameState === 'playing' || gameState === 'finished'"
          >
            {{ isReady ? '取消準備' : '準備就緒' }}
          </button>
        </div>
      </div>

      <!-- 右側遊戲區域 -->
      <div class="game-section">
        <h3>🎯 猜數字遊戲</h3>
        
        <!-- 遊戲狀態顯示 -->
        <div class="game-status">
          <div v-if="gameState === 'waiting'" class="status-waiting">
            <p>等待遊戲開始...</p>
            <div class="game-controls">
              <button
                @click="startGame"
                :disabled="!isReady || !isConnected || gameState !== 'waiting' || players.filter(p => p.isReady).length < players.length || players.length < 2"
                class="start-button"
              >
                開始遊戲
              </button>
              
              <!--  準備狀態提示 -->
              <div v-if="gameState === 'waiting'" class="ready-status-hint">
                <p v-if="!isReady" class="hint-message">
                  ⚠️ 你還沒有準備就緒
                </p>
                <p v-else-if="players.filter(p => p.isReady).length < players.length" class="hint-message">
                  ⏳ 等待其他玩家準備中... ({{ players.filter(p => p.isReady).length }}/{{ players.length }})
                </p>
                <p v-else-if="players.length < 2" class="hint-message">
                  👥 至少需要 2 位玩家才能開始
                </p>
                <p v-else class="hint-success">
                  ✅ 所有玩家已準備就緒，可以開始遊戲！
                </p>
              </div>
            </div>
          </div>

          <div v-else-if="gameState === 'playing'" class="status-playing">
            <p>目標：猜一個 1-100 之間的數字</p>
            <!-- <p>剩餘嘗試次數: {{ maxAttempts - attempts }}</p> -->
            
            <!--  回合資訊顯示 -->
            <div class="round-info">
              <span class="round-indicator">
                回合進度: {{ players.filter(p => p.guessed).length }}/{{ players.length }} 位玩家已猜測
              </span>
            </div>
            
            <div class="guess-input">
              <input
                v-model="guessNumber"
                type="number"
                min="1"
                max="100"
                placeholder="輸入你的猜測"
                @keypress="handleGuessKeyPress"
                :disabled="!isConnected || currentTurnPlayer !== username"
                class="number-input"
              />
              <button
                @click="submitGuess"
                :disabled="!guessNumber || !isConnected || currentTurnPlayer !== username"
                class="guess-button"
              >
                猜測
              </button>
            </div>
            
            <!-- 輪次提示訊息 -->
            <div class="turn-hint">
              <span v-if="currentTurnPlayer === username" class="your-turn">
                🎯 輪到你猜測了！
              </span>
              <span v-else-if="currentTurnPlayer" class="waiting-turn">
                ⏳ 等待 {{ currentTurnPlayer }} 猜測中...
              </span>
              <span v-else class="waiting-turn">
                ⏳ 等待遊戲開始...
              </span>
            </div>
          </div>

          <div v-else-if="gameState === 'finished'" class="status-finished">
            <div :class="['game-result', gameResult]">
              <!--  修改：根據是否為獲勝者顯示不同訊息 -->
              <h4 v-if="gameWinner === username">🎉 恭喜你獲勝！</h4>
              <h4 v-else-if="gameWinner">🎊 遊戲結束</h4>
              <h4 v-else>😢 遊戲結束</h4>
              
              <!--  修改：顯示獲勝者和答案 -->
              <div class="game-summary">
                <p v-if="gameWinner" class="winner-info">
                  🏆 獲勝者：<span class="winner-name">{{ gameWinner }}</span>
                </p>
                <p v-if="gameAnswer" class="answer-info">
                  🎯 答案是：<span class="answer-number">{{ gameAnswer }}</span>
                </p>
                <!-- <p class="attempts-info">
                  📊 總共嘗試：{{ attempts }} 次
                </p> -->
              </div>
            </div>
            
            <!--  修改：添加重置提示 -->
            <div class="restart-controls">
              <button @click="restartGame" :disabled="!isConnected" class="restart-button">
                <span class="restart-icon">🔄</span>
                重新開始
              </button>
              <p class="restart-hint">
                點擊重新開始將重置所有遊戲狀態
              </p>
            </div>
          </div>
        </div>

        <!-- 猜測歷史 -->
        <!-- <div v-if="guessHistory.length > 0" class="guess-history">
          <h4>📝 猜測記錄</h4>
          <div class="history-list">
            <div
              v-for="record in guessHistory"
              :key="record.attempt"
              class="history-item"
            >
              <div class="attempt-header">
                <span class="attempt-number">第{{ record.attempt }}次</span>
                <span class="player-name">{{ record.player }}</span>
              </div>
              <div class="guess-detail">
                <div class="guess-row">
                  <span class="guess-label">玩家：</span>
                  <span class="guess-player">{{ record.player }}</span>
                </div>
                <div class="guess-row">
                  <span class="guess-label">猜測：</span>
                  <span class="guess-value">{{ extractGuessNum(record.message) }}</span>
                </div>
                <div class="guess-row">
                  <span class="result-label">結果：</span>
                  <span class="result-value">{{ extractResult(record.message) }}</span>
                </div>
              </div>
            </div>
          </div>
          <div class="statistics">
            <div class="stat-item">
              <span class="stat-label">總嘗試次數</span>
              <span class="stat-value">{{ attempts }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">已猜中次數</span>
              <span class="stat-value">
                {{ guessHistory.filter(record => record.result && record.result.includes('恭喜')).length }}
              </span>
            </div>
          </div>
          <div class="scroll-buttons">
            <button @click="scrollHistoryToTop" class="scroll-button">
              ⬆️
            </button>
            <button @click="scrollHistoryToBottom" class="scroll-button">
              ⬇️
            </button>
          </div>
        </div> -->

      </div>
    </div>
  </div>
</template>

<style scoped>
/* =============== 基礎樣式 =============== */
.game-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #1a1a1a;
  color: #e8eaed;
}

/* =============== 頂部資訊欄 =============== */
.game-header {
  background: #2d2d2d;
  padding: 1rem 2rem;
  border-bottom: 2px solid #404040;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  flex-shrink: 0;
}

.room-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.room-info h2 {
  margin: 0;
  color: #e8eaed;
  font-size: 1.5rem;
}

.player-name {
  color: #a8a8a8;
  font-size: 1rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.25rem; /*  縮小間距 */
  overflow: hidden; /*  防止溢出 */
}

.connection-status {
  font-size: 0.9rem;
  padding: 4px 8px;
  border-radius: 4px;
  margin-left: 1rem;
}

.connection-status.connected {
  background-color: #27ae60;
  color: white;
}

.connection-status:not(.connected) {
  background-color: #e74c3c;
  color: white;
}

.leave-button {
  background: linear-gradient(135deg, #dc3545, #bd2130);
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.95rem;
  transition: all 0.3s ease;
}

.leave-button:hover {
  background: linear-gradient(135deg, #c82333, #a71e2a);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(220, 53, 69, 0.3);
}

/* 新增：遊戲進行中無法點擊且顏色變暗 */
.leave-button.disabled,
.leave-button.disabled:hover {
  background: #888 !important;
  color: #eee !important;
  cursor: not-allowed !important;
  box-shadow: none !important;
  transform: none !important;
}

/* =============== 主遊戲區域佈局 =============== */
.game-main {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 380px 1fr; /*  從 300px 改為 380px */
  gap: 1.5rem;
  padding: 2rem;
  overflow: hidden;
  min-height: 0;
}

/* =============== 聊天室區域 =============== */
.chat-section {
  background: #2d2d2d;
  border-radius: 8px;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  border: 1px solid #404040;
  min-height: 0;
}

.chat-section h3 {
  margin: 0 0 1rem 0;
  color: #e8eaed;
  border-bottom: 2px solid #4db6e6;
  padding-bottom: 0.5rem;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  background: #1a1a1a;
  border: 1px solid #404040;
  border-radius: 6px;
  padding: 1rem;
  margin-bottom: 1rem;
  max-height: 400px;
}

.message {
  margin-bottom: 0.75rem;
  padding: 0.5rem;
  border-radius: 6px;
}

.message.system {
  background: rgba(77, 182, 230, 0.1);
  border-left: 3px solid #4db6e6;
}

.message.user {
  background: rgba(255, 255, 255, 0.05);
}

.system-message {
  color: #4db6e6;
  font-size: 0.9rem;
}

.user-message {
  color: #e8eaed;
}

.message-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.25rem;
  font-size: 0.85rem;
}

.username {
  color: #4db6e6;
  font-weight: 600;
}

.timestamp {
  color: #888;
  font-size: 0.8rem;
}

.message-content {
  color: #e8eaed;
}

.chat-input {
  display: flex;
  gap: 0.75rem;
  align-items: stretch;
}

.message-input {
  flex: 1;
  padding: 12px 16px;
  border: 2px solid #404040;
  border-radius: 8px;
  background: #1a1a1a;
  color: #e8eaed;
  font-size: 0.95rem;
  transition: all 0.3s ease;
  outline: none;
}

.message-input:focus {
  border-color: #4db6e6;
  box-shadow: 0 0 0 3px rgba(77, 182, 230, 0.1);
  background: #242424;
}

.message-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background: #0f0f0f !important;
  border-color: #333 !important;
}

.message-input::placeholder {
  color: #666;
}

.send-button {
  background: linear-gradient(135deg, #4db6e6, #3a9bc1);
  color: white;
  border: none;
  padding: 12px 20px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.95rem;
  cursor: pointer;
  transition: all 0.3s ease;
  min-width: 70px;
}

.send-button:hover:not(:disabled) {
  background: linear-gradient(135deg, #3a9bc1, #2980b9);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(77, 182, 230, 0.3);
}

.send-button:disabled {
  background: #666;
  cursor: not-allowed;
  opacity: 0.6;
}

/* =============== 玩家狀態面板 =============== */
.players-section {
  background: #2d2d2d;
  border-radius: 8px;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  border: 1px solid #404040;
  min-height: 0;
}

.players-section h3 {
  margin: 0 0 1rem 0;
  color: #e8eaed;
  border-bottom: 2px solid #4db6e6;
  padding-bottom: 0.5rem;
}

.room-status {
  background: #1a1a1a;
  border-radius: 6px;
  padding: 0.75rem; /*  縮小內邊距 */
  margin-bottom: 1rem;
  border: 1px solid #404040;
}

.room-stats {
  display: flex;
  flex-direction: column;
  gap: 0.5rem; /*  縮小間距 */
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-label {
  color: #a8a8a8;
  font-size: 0.8rem; /*  縮小字體 */
}

.stat-value {
  color: #e8eaed;
  font-weight: 600;
  font-size: 0.85rem; /*  縮小字體 */
}

/*  玩家列表樣式 */
.players-list {
  flex: 1;
  background: #1a1a1a;
  border-radius: 6px;
  border: 1px solid #404040;
  overflow: hidden;
}

.players-header {
  display: grid;
  grid-template-columns: 1.5fr 70px 90px; /*  調整比例，給玩家名稱更多空間 */
  gap: 0.75rem; /*  縮小間距 */
  padding: 0.75rem 1rem;
  background: #333;
  border-bottom: 1px solid #404040;
  font-weight: 600;
  font-size: 0.85rem; /*  稍微縮小字體 */
  color: #a8a8a8;
}

.player-item {
  display: grid;
  grid-template-columns: 1.5fr 70px 90px; /*  與 header 保持一致 */
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #333;
  transition: all 0.3s ease;
}

.player-item:last-child {
  border-bottom: none;
}

.player-item.current-player {
  background: rgba(77, 182, 230, 0.1);
  border-left: 3px solid #4db6e6;
}

.player-item.ready {
  background: rgba(40, 167, 69, 0.1);
}

.player-item.current-turn {
  background: rgba(253, 126, 20, 0.15);
  border-left: 3px solid #fd7e14;
  animation: pulse 2s infinite;
}

.player-item.guessed {
  background: rgba(108, 117, 125, 0.1);
  opacity: 0.7;
}

.player-item.winner {
  background: linear-gradient(135deg, rgba(255, 215, 0, 0.2), rgba(255, 215, 0, 0.1));
  border-left: 3px solid #ffd700;
  box-shadow: 0 0 12px rgba(255, 215, 0, 0.2);
}

.player-info {
  display: flex;
  align-items: center;
}

.player-name {
  font-weight: 600;
  color: #e8eaed;
  display: flex;
  align-items: center;
  gap: 0.25rem; /*  縮小間距 */
  overflow: hidden; /*  防止溢出 */
}

.you-indicator {
  color: #4db6e6;
  font-size: 0.7rem; /*  縮小 "(你)" 標識 */
  font-weight: 500;
  white-space: nowrap; /*  防止換行 */
}

.winner-indicator {
  color: #ffd700;
  font-size: 0.9rem; /*  稍微縮小皇冠 */
  text-shadow: 0 0 8px rgba(255, 215, 0, 0.5);
  animation: sparkle 2s infinite;
}

.turn-order-display {
  display: flex;
  align-items: center;
  justify-content: center;
}

.order-number {
  background: #404040;
  color: #e8eaed;
  padding: 1px 6px; /*  縮小內邊距 */
  border-radius: 10px; /*  稍微調整圓角 */
  font-size: 0.75rem; /*  縮小字體 */
  font-weight: 600;
  white-space: nowrap;
}

.player-status {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem; /*  統一縮小狀態文字 */
}

.game-status {
  font-size: 0.75rem; /*  縮小字體 */
  padding: 1px 4px; /*  縮小內邊距 */
  border-radius: 3px;
  font-weight: 600;
  white-space: nowrap; /*  防止換行 */
}

.final-status {
  font-size: 0.75rem; /*  縮小字體 */
  padding: 1px 4px; /*  縮小內邊距 */
  border-radius: 3px;
  font-weight: 600;
  white-space: nowrap;
}

.ready-indicator {
  font-size: 0.75rem; /*  縮小字體 */
  padding: 1px 4px; /*  縮小內邊距 */
  border-radius: 3px;
  font-weight: 600;
  white-space: nowrap;
}

/*  快速準備區域樣式 */
.quick-actions {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #404040;
}

.ready-toggle {
  width: 100%;
  padding: 12px;
  border: 2px solid #404040;
  border-radius: 8px;
  background: #1a1a1a;
  color: #e8eaed;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.ready-toggle:hover:not(:disabled) {
  border-color: #4db6e6;
  background: #242424;
}

.ready-toggle.active {
  background: linear-gradient(135deg, #28a745, #20c997);
  border-color: #28a745;
  color: white;
}

.ready-toggle:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background: #0f0f0f !important;
  border-color: #333 !important;
}

/* =============== 遊戲結束狀態樣式 =============== */
.stat-value.winner {
  color: #ffd700;
  font-weight: 700;
  text-shadow: 0 0 8px rgba(255, 215, 0, 0.3);
}

.stat-value.answer {
  color: #4db6e6;
  font-weight: 700;
}

/* =============== 右側遊戲區域 =============== */
.game-section {
  background: #2d2d2d;
  border-radius: 8px;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  border: 1px solid #404040;
  min-height: 0;
}

.game-section h3 {
  margin: 0 0 1rem 0;
  color: #e8eaed;
  border-bottom: 2px solid #4db6e6;
  padding-bottom: 0.5rem;
}

.game-status {
  flex: 1;
  display: flex;
  flex-direction: column;
}

/*  等待狀態樣式 */
.status-waiting {
  text-align: center;
  padding: 2rem 1rem;
}

.status-waiting p {
  font-size: 1.1rem;
  color: #a8a8a8;
  margin-bottom: 2rem;
}

.game-controls {
  display: flex;
  justify-content: center;
}

.start-button {
  background: linear-gradient(135deg, #28a745, #20c997);
  color: white;
  border: none;
  padding: 15px 30px;
  border-radius: 8px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.start-button:hover:not(:disabled) {
  background: linear-gradient(135deg, #20c997, #17a2b8);
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(40, 167, 69, 0.3);
}

.start-button:disabled {
  background: #6c757d !important;
  cursor: not-allowed !important;
  opacity: 0.6 !important;
  transform: none !important;
  box-shadow: none !important;
}

/*  遊戲進行中樣式 */
.status-playing {
  padding: 1rem;
}

.status-playing p {
  margin: 0.5rem 0;
  color: #e8eaed;
  font-size: 1rem;
}

.guess-input {
  display: flex;
  gap: 1rem;
  margin: 1.5rem 0;
  align-items: stretch;
}

.number-input {
  flex: 1;
  padding: 15px 20px;
  border: 2px solid #404040;
  border-radius: 8px;
  background: #1a1a1a;
  color: #e8eaed;
  font-size: 1.1rem;
  text-align: center;
  transition: all 0.3s ease;
  outline: none;
}

.number-input:focus {
  border-color: #4db6e6;
  box-shadow: 0 0 0 3px rgba(77, 182, 230, 0.1);
  background: #242424;
}

.number-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background: #0f0f0f !important;
  border-color: #333 !important;
}

.guess-button {
  background: linear-gradient(135deg, #fd7e14, #e9500f);
  color: white;
  border: none;
  padding: 15px 25px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 1.1rem;
  cursor: pointer;
  transition: all 0.3s ease;
  min-width: 100px;
}

.guess-button:hover:not(:disabled) {
  background: linear-gradient(135deg, #e9500f, #dc3545);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(253, 126, 20, 0.3);
}

.guess-button:disabled {
  background: #6c757d;
  cursor: not-allowed;
  opacity: 0.6;
}

/*  輪次提示樣式 */
.turn-hint {
  margin-top: 1rem;
  padding: 0.75rem;
  border-radius: 6px;
  text-align: center;
  font-weight: 600;
}

.your-turn {
  background: rgba(253, 126, 20, 0.15);
  color: #fd7e14;
  border: 1px solid #fd7e14;
  animation: pulse 2s infinite;
}

.waiting-turn {
  background: rgba(255, 193, 7, 0.1);
  color: #ffc107;
  border: 1px solid #ffc107;
}

/*  遊戲結束樣式 */
.status-finished {
  text-align: center;
  padding: 1rem;
}

.game-result {
  margin-bottom: 2rem;
}

.game-result h4 {
  font-size: 1.5rem;
  margin-bottom: 1rem;
  color: #e8eaed;
}

.game-summary {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 8px;
  padding: 1rem;
  margin: 1rem 0;
  border: 1px solid #404040;
}

.winner-info,
.answer-info,
.attempts-info {
  margin: 0.5rem 0;
  font-size: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.winner-name {
  color: #ffd700;
  font-weight: 700;
  text-shadow: 0 0 8px rgba(255, 215, 0, 0.3);
}

.answer-number {
  color: #4db6e6;
  font-weight: 700;
  font-size: 1.2rem;
}

.attempts-info {
  color: #a8a8a8;
}

/*  重置控制區域樣式 */
.restart-controls {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
}

.restart-button {
  background: linear-gradient(135deg, #4db6e6, #3a9bc1);
  color: white;
  border: none;
  padding: 15px 30px;
  border-radius: 8px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.restart-button:hover:not(:disabled) {
  background: linear-gradient(135deg, #3a9bc1, #2980b9);
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(77, 182, 230, 0.3);
}

.restart-button:disabled {
  background: #6c757d;
  cursor: not-allowed;
  opacity: 0.6;
}

.restart-icon {
  font-size: 1.2rem;
  animation: rotate 2s linear infinite;
}

.restart-button:hover .restart-icon {
  animation-duration: 0.5s;
}

.restart-hint {
  font-size: 0.85rem;
  color: #a8a8a8;
  margin: 0;
  text-align: center;
}

/*  旋轉動畫 */
@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* =============== 動画效果 =============== */
@keyframes sparkle {
  0%, 100% {
    opacity:  1;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.1);
  }
}

@keyframes pulse {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
  100% {
    opacity: 1;
  }
}

@keyframes bounce {
  0%, 20%, 60%, 100% {
    transform: translateY(0);
  }
  40% {
    transform: translateY(-3px);
  }
  60% {
    transform: translateY(-2px);
  }
}

/* =============== 響應式設計 =============== */
@media (max-width: 1400px) {
  .game-main {
    grid-template-columns: 1fr 350px 1fr; /*  中等屏幕調整 */
  }
}

@media (max-width: 1200px) {
  .game-main {
    grid-template-columns: 1fr;
    grid-template-rows: auto auto auto;
    gap: 1rem;
  }
  
  .players-section {
    order: -1;
  }
  
  /*  小屏幕時恢復較大的字體 */
  .players-header,
  .player-item {
    grid-template-columns: 2fr 80px 100px;
    font-size: 0.9rem;
  }
}

@media (max-width: 768px) {
  .game-main {
    padding: 1rem;
  }
  
  .players-header,
  .player-item {
    grid-template-columns: 1fr;
    text-align: center;
    gap: 0.5rem;
  }
  
  .players-header {
    display: none; /*  手機版隱藏表頭 */
  }
  
  .player-item {
    display: flex;
    flex-direction: column;
    padding: 1rem;
    text-align: left;
  }
}
</style>

