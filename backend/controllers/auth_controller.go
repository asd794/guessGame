package controllers

import (
	"game/middleware"
	"game/services"
	"log"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

type AuthController struct {
	gameManager *services.GameManagerMysql
}

type LoginRequest struct {
	Email      string `json:"email" binding:"required"`
	Password   string `json:"password" binding:"required"`
	CaptchaId  string `json:"captcha_id"`
	CaptchaVal string `json:"captcha_value"`
}

type RegisterRequest struct {
	Email           string `json:"email" binding:"required"`
	Username        string `json:"username" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	CaptchaId       string `json:"captcha_id"`
	CaptchaVal      string `json:"captcha_value"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Message  string `json:"message"`
}

func NewAuthController(gameManager *services.GameManagerMysql) *AuthController {
	return &AuthController{
		gameManager: gameManager,
	}
}

// 建立 email 驗證的正規表達式
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func (ac *AuthController) LoginController(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Email或Password不得為空"})
		return
	}

	// 驗證碼校驗
	if !base64Captcha.DefaultMemStore.Verify(req.CaptchaId, req.CaptchaVal, true) {
		c.JSON(400, gin.H{"error": "驗證碼錯誤"})
		return
	}
	// 檢查email長度
	if len(req.Email) < 6 || len(req.Email) > 100 {
		c.JSON(400, gin.H{"error": "Email長度必須在6到100個字元之間"})
		return
	}

	// 檢查email格式
	if !emailRegex.MatchString(req.Email) {
		c.JSON(400, gin.H{"error": "Email格式不正確"})
		return
	}

	// 檢查密碼長度
	if len(req.Password) < 6 || len(req.Password) > 15 {
		c.JSON(400, gin.H{"error": "Email或Password長度必須在6到15個字元之間"})
		return
	}

	// 檢查 email 和密碼是否正確
	uuid, username, err := ac.gameManager.Login(req.Email, req.Password)
	if err != nil {
		log.Println("Login error:", err)
		c.JSON(401, gin.H{"error": "Email或Password錯誤"})
		return
	}

	// 生成 JWT Token
	token, err := middleware.GenerateJWTGame(req.Email, username, uuid)
	if err != nil {
		c.JSON(500, gin.H{"error": "生成token失敗"})
		return
	}

	c.JSON(200, LoginResponse{
		Token:    token,
		Username: username,
		Message:  "登入成功",
	})
}

func (ac *AuthController) RegisterController(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "註冊資訊不得為空"})
		return
	}
	// 驗證碼校驗
	if !base64Captcha.DefaultMemStore.Verify(req.CaptchaId, req.CaptchaVal, true) {
		c.JSON(400, gin.H{"error": "驗證碼錯誤"})
		return
	}
	// 檢查email名稱長度
	if len(req.Email) < 6 || len(req.Email) > 100 {
		c.JSON(400, gin.H{"error": "Email長度必須在6到100個字元之間"})
		return
	}

	// 檢查email格式
	if !emailRegex.MatchString(req.Email) {
		c.JSON(400, gin.H{"error": "Email格式不正確"})
		return
	}

	// 檢查username名稱長度
	if len(req.Username) < 3 || len(req.Username) > 15 {
		c.JSON(400, gin.H{"error": "Username長度必須在3到15個字元之間"})
		return
	}
	// 檢查密碼長度
	if len(req.Password) < 6 || len(req.Password) > 15 {
		c.JSON(400, gin.H{"error": "Password長度必須在6到15個字元之間"})
		return
	}
	if req.Password != req.ConfirmPassword {
		c.JSON(400, gin.H{"error": "Password不一致"})
		return
	}

	err := ac.gameManager.Register(req.Username, req.Email, req.Password)
	if err != nil {
		log.Println("Register error:", err)
		c.JSON(500, gin.H{"error": "註冊失敗，Email已被使用"})
		return
	}

	c.JSON(200, gin.H{"message": "註冊成功"})
}

// 產生驗證碼
func (ac *AuthController) GetCaptchaController(c *gin.Context) {
	driver := base64Captcha.NewDriverDigit(80, 350, 5, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, base64Captcha.DefaultMemStore)
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		c.JSON(500, gin.H{"error": "生成驗證碼失敗"})
		return
	}
	c.JSON(200, gin.H{
		"captcha_id":  id,
		"captcha_img": b64s,
	})
}
