<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { API_HOST } from '../config'

// 響應式資料
const email = ref('')
const password = ref('')
const username = ref('')
const confirmPassword = ref('')
const isLoading = ref(false)
const errorMessage = ref('')
const isRegisterMode = ref(false)
const router = useRouter()

// 驗證規則
const emailPattern = /^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$/
const passwordPattern = /^[A-Za-z0-9]{6,20}$/
const usernamePattern = /^[\u4e00-\u9fa5A-Za-z0-9]{3,16}$/

// 驗證碼相關
const captchaId = ref('')
const captchaImg = ref('')
const captchaValue = ref('')

/**
 * 檢查是否已經登入
 */
const checkExistingAuth = () => {
  const token = localStorage.getItem('token')
  if (token) {
    router.push('/room-selection')
  }
}

/**
 * 取得驗證碼
 */
const fetchCaptcha = async () => {
  try {
    const res = await fetch(`${API_HOST}/api/v1/auth/captcha`, {
      method: 'POST'
    })
    const data = await res.json()
    captchaId.value = data.captcha_id
    captchaImg.value = data.captcha_img
    captchaValue.value = ''
  } catch (e) {
    console.error('取得驗證碼失敗:', e)
    captchaId.value = ''
    captchaImg.value = ''
  }
}

onMounted(() => {
  checkExistingAuth()
  fetchCaptcha()
})

/**
 * 處理使用者登入
 */
const handleLogin = async () => {
  if (!email.value.trim()) {
    errorMessage.value = '請輸入電子郵件'
    return
  }
  if (!password.value) {
    errorMessage.value = '請輸入密碼'
    return
  }
  if (!captchaValue.value.trim()) {
    errorMessage.value = '請輸入驗證碼'
    return
  }
  if (!emailPattern.test(email.value)) {
    errorMessage.value = '請輸入正確的電子郵件格式'
    return
  }
  if (!passwordPattern.test(password.value)) {
    errorMessage.value = '密碼僅能輸入英文與數字，長度6-20字元'
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    const res = await fetch(`${API_HOST}/api/v1/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        email: email.value.trim(),
        password: password.value,
        captcha_id: captchaId.value,
        captcha_value: captchaValue.value.trim()
      })
    })
    const response = await res.json()
    isLoading.value = false

    if (res.ok && response.token) {
      localStorage.setItem('token', response.token)
      localStorage.setItem('username', response.username || '')
      localStorage.setItem('email', response.email || email.value.trim())
      window.location.href = '/room-selection'
      return
    } else {
      // 若 response.error 是字串，顯示 error；若是物件且有 error 屬性則顯示 error.error
      if (typeof response.error === 'string') {
        errorMessage.value = response.error
      } else if (response.error && typeof response.error === 'object' && response.error.error) {
        errorMessage.value = response.error.error
      } else {
        errorMessage.value = response.message || '登入失敗，請檢查帳號密碼或驗證碼'
      }
      fetchCaptcha()
    }
  } catch (error) {
    isLoading.value = false
    console.error('登入請求失敗:', error)
    errorMessage.value = '伺服器連線失敗，請稍後再試'
    fetchCaptcha()
  }
}

/**
 * 處理使用者註冊
 */
const handleRegister = async () => {
  if (!username.value.trim()) {
    errorMessage.value = '請輸入使用者名稱'
    return
  }
  if (!email.value.trim()) {
    errorMessage.value = '請輸入電子郵件'
    return
  }
  if (!password.value) {
    errorMessage.value = '請輸入密碼'
    return
  }
  if (!confirmPassword.value) {
    errorMessage.value = '請再次輸入密碼'
    return
  }
  if (!usernamePattern.test(username.value)) {
    errorMessage.value = '使用者名稱可輸入中文、英文與數字，長度3-16字元'
    return
  }
  if (!emailPattern.test(email.value)) {
    errorMessage.value = '請輸入正確的電子郵件格式'
    return
  }
  if (!passwordPattern.test(password.value)) {
    errorMessage.value = '密碼僅能輸入英文與數字，長度6-20字元'
    return
  }
  if (password.value !== confirmPassword.value) {
    errorMessage.value = '兩次輸入的密碼不一致'
    return
  }
  if (!captchaValue.value.trim()) {
    errorMessage.value = '請輸入驗證碼'
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    const res = await fetch(`${API_HOST}/api/v1/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        username: username.value.trim(),
        email: email.value.trim(),
        password: password.value,
        confirm_password: confirmPassword.value,
        captcha_id: captchaId.value,
        captcha_value: captchaValue.value.trim()
      })
    })
    const response = await res.json()
    isLoading.value = false

    if (res.ok) {
      errorMessage.value = '註冊成功，請登入'
      isRegisterMode.value = false
      username.value = ''
      email.value = ''
      password.value = ''
      confirmPassword.value = ''
      fetchCaptcha()
    } else {
      // 若 response.error 是字串，顯示 error；若是物件且有 error 屬性則顯示 error.error
      if (typeof response.error === 'string') {
        errorMessage.value = response.error
      } else if (response.error && typeof response.error === 'object' && response.error.error) {
        errorMessage.value = response.error.error
      } else {
        errorMessage.value = response.message || '註冊失敗，請檢查輸入資料或驗證碼'
      }
      fetchCaptcha()
    }
  } catch (error) {
    isLoading.value = false
    console.error('註冊請求失敗:', error)
    errorMessage.value = '伺服器連線失敗，請稍後再試'
    fetchCaptcha()
  }
}

/**
 * 切換登入/註冊模式
 */
const toggleMode = () => {
  isRegisterMode.value = !isRegisterMode.value
  errorMessage.value = ''
  username.value = ''
  email.value = ''
  password.value = ''
  confirmPassword.value = ''
  fetchCaptcha()
}
</script>

<template>
  <div class="login-container">
    <div class="login-form">
      <h1 v-if="!isRegisterMode">🎯 猜數字遊戲登入</h1>
      <h1 v-else>📝 註冊新帳號</h1>
      <p v-if="!isRegisterMode">請輸入電子郵件與密碼登入</p>
      <p v-else>請輸入使用者名稱、電子郵件、密碼與確認密碼註冊</p>
      
      <!-- 錯誤訊息 -->
      <div v-if="errorMessage" class="error-message">
        {{ errorMessage }}
      </div>
      
      <!-- 登入表單 -->
      <input
        v-if="!isRegisterMode"
        v-model="email"
        type="email"
        placeholder="請輸入電子郵件"
        :disabled="isLoading"
        class="username-input"
        @keydown.enter="handleLogin()"
        maxlength="50"
        autocomplete="email"
      />
      <input
        v-if="!isRegisterMode"
        v-model="password"
        type="password"
        placeholder="請輸入密碼"
        :disabled="isLoading"
        class="username-input"
        @keydown.enter="handleLogin()"
        maxlength="20"
        autocomplete="current-password"
      />
      <!-- 驗證碼欄位（登入） -->
      <div v-if="!isRegisterMode" class="captcha-row">
        <input
          v-model="captchaValue"
          type="text"
          placeholder="請輸入驗證碼"
          :disabled="isLoading"
          class="username-input"
          maxlength="8"
          style="width: 60%; display: inline-block;"
          @keydown.enter="handleLogin()"
        />
        <img
          :src="captchaImg"
          alt="驗證碼"
          style="height: 40px; vertical-align: middle; cursor: pointer; margin-left: 8px;"
          @click="fetchCaptcha"
        />
      </div>

      <!-- 註冊表單 -->
      <input
        v-if="isRegisterMode"
        v-model="username"
        type="text"
        placeholder="請輸入使用者名稱（中文、英文與數字，3-16字元）"
        :disabled="isLoading"
        class="username-input"
        @keydown.enter="handleRegister()"
        maxlength="16"
        autocomplete="username"
      />
      <input
        v-if="isRegisterMode"
        v-model="email"
        type="email"
        placeholder="請輸入電子郵件"
        :disabled="isLoading"
        class="username-input"
        @keydown.enter="handleRegister()"
        maxlength="50"
        autocomplete="email"
      />
      <input
        v-if="isRegisterMode"
        v-model="password"
        type="password"
        placeholder="請輸入密碼（英文與數字，6-20字元）"
        :disabled="isLoading"
        class="username-input"
        @keydown.enter="handleRegister()"
        maxlength="20"
        autocomplete="new-password"
      />
      <input
        v-if="isRegisterMode"
        v-model="confirmPassword"
        type="password"
        placeholder="請再次輸入密碼"
        :disabled="isLoading"
        class="username-input"
        @keydown.enter="handleRegister()"
        maxlength="20"
        autocomplete="new-password"
      />
      <!-- 驗證碼欄位（註冊） -->
      <div v-if="isRegisterMode" class="captcha-row">
        <input
          v-model="captchaValue"
          type="text"
          placeholder="請輸入驗證碼"
          :disabled="isLoading"
          class="username-input"
          maxlength="8"
          style="width: 60%; display: inline-block;"
          @keydown.enter="handleRegister()"
        />
        <img
          :src="captchaImg"
          alt="驗證碼"
          style="height: 40px; vertical-align: middle; cursor: pointer; margin-left: 8px;"
          @click="fetchCaptcha"
        />
      </div>
      
      <button 
        v-if="!isRegisterMode"
        @click="handleLogin" 
        :disabled="isLoading || !email.trim() || !password"
        class="login-button"
      >
        <span v-if="isLoading">登入中...</span>
        <span v-else>登入</span>
      </button>
      <button 
        v-else
        @click="handleRegister" 
        :disabled="isLoading || !username.trim() || !email.trim() || !password || !confirmPassword"
        class="login-button"
      >
        <span v-if="isLoading">註冊中...</span>
        <span v-else>註冊</span>
      </button>
      <div style="margin-top: 1rem;">
        <a href="#" @click.prevent="toggleMode" style="color:#4db6e6;">
          {{ isRegisterMode ? '已有帳號？點此登入' : '沒有帳號？點此註冊' }}
        </a>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 70vh;
  padding: 2rem;
  background: #1a1a1a;
}

.login-form {
  background: #2d2d2d;
  border-radius: 12px;
  padding: 3rem;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  text-align: center;
  min-width: 400px;
  border: 1px solid #404040;
}

h1 {
  color: #e8eaed;
  margin-bottom: 1rem;
  font-size: 2.5rem;
}

p {
  color: #a8a8a8;
  margin-bottom: 2rem;
  font-size: 1.1rem;
}

.username-input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #404040;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.3s ease;
  background: #1a1a1a;
  color: #e8eaed;
  margin-bottom: 1rem;
}

.username-input:focus {
  outline: none;
  border-color: #4db6e6;
}

.login-button {
  background: #4db6e6;
  color: #1a1a1a;
  border: none;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.3s ease;
  min-width: 120px;
}

.login-button:hover:not(:disabled) {
  background: #3a9bc1;
}

.login-button:disabled {
  background-color: #666;
  cursor: not-allowed;
}

.error-message {
  background-color: #ff4444;
  color: white;
  padding: 10px;
  border-radius: 5px;
  margin-bottom: 15px;
  text-align: center;
}

.captcha-row {
  display: flex;
  align-items: center;
  margin-bottom: 1rem;
}
</style>
