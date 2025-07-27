package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("my_golang_jwt_secret_key") // JWT 密鑰

// JWTClaims 自訂的 JWT 內容結構
type JWTClaims struct {
	UserUUID string `json:"user_uuid"` // 使用者UUID
	UserID   uint64 `json:"id"`        // 使用者ID
	Username string `json:"username"`  // 使用者名稱
	Email    string `json:"email"`     // 使用者電子郵件
	jwt.RegisteredClaims
	// jwt.RegisteredClaims `json:"registered_claims"` // 註冊的聲明
}

// 生成 JWT Token
func GenerateToken(userUUID string, userID uint64, username string, email string) (string, error) {
	// log.Println(userUUID,userID, username, email,"--------")
	// 設定 JWT 的標準聲明
	claim := JWTClaims{
		UserUUID: userUUID,
		UserID:   userID, // 將 userID 轉換為字串
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "TransferTest", // 設定發行者
			// ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // 設定過期時間為1小時
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Minute)), // 設定過期時間為2分鐘
			IssuedAt:  jwt.NewNumericDate(time.Now()),                      // 設定簽發時間為當前時間
			NotBefore: jwt.NewNumericDate(time.Now()),                      // 設定生效時間為當前時間
			Subject:   email,                                               // 設定主題為email
		},
	}

	// 使用 HMAC SHA256 加密演算法簽名 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	// 簽名並返回 JWT 字串
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// 解析 JWT Token
func ParseToken(tokenString string) (*JWTClaims, error) {
	// 解析 JWT Token，並驗證簽名
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非預期的簽名方法: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// 驗證 token 是否有效
	// 如果 token 有效，則將其轉換為自訂的 JWTClaims 結構
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("無效的 token")
}

// JWT 中間件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		log.Println("Authorization Header:", authHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "需要授權",
			})
			c.Abort()
			return
		}

		// 檢查是否為 Bearer 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "授權格式錯誤",
			})
			c.Abort()
			return
		}

		// 解析 token
		claims, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "無效的 token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// log.Println(claims)

		// 將用戶信息存儲在 context 中
		c.Set("userUUID", claims.UserUUID)
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// Game JWT 密鑰
var jwtSecretGame = []byte("my_game_jwt_secret_key")

type JWTClaimsGame struct {
	Email    string `json:"email"`    // 使用者電子郵件
	Username string `json:"username"` // 使用者名稱
	jwt.RegisteredClaims
}

// Game WT 生成函數
func GenerateJWTGame(email string, username string, uuid string) (string, error) {

	claimsGame := JWTClaimsGame{
		Email:    email,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "my-hub.site", // 發行者
			// ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // 設定過期時間為1小時
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)), // 設定過期時間為60分鐘
			IssuedAt:  jwt.NewNumericDate(time.Now()),                       // 設定簽發時間為當前時間
			NotBefore: jwt.NewNumericDate(time.Now()),                       // 設定生效時間為當前時間
			Subject:   uuid,                                                 // 設定主題為UUID
		},
	}

	// // 設定 JWT 的標準聲明
	// claims := jwt.MapClaims{
	// 	"username": username,
	// 	"exp":      time.Now().Add(time.Hour * 1).Unix(), // 1小時過期
	// 	"iat":      time.Now().Unix(),
	// }

	// 使用 HMAC SHA256 加密演算法簽名 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsGame)

	tokenString, err := token.SignedString(jwtSecretGame)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GAME JWT 中間件
func JWTAuthGame() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		//  檢查是否為 WebSocket 升級請求
		if c.GetHeader("Upgrade") == "websocket" {
			// WebSocket 請求：從 URL 參數獲取 token
			tokenString = c.Query("token")
			log.Printf("WebSocket 請求，從 URL 獲取 token")
		} else {
			// 普通 HTTP 請求：從 Authorization Header 獲取
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "需要授權",
				})
				c.Abort()
				return
			}

			// 檢查是否為 Bearer 格式
			parts := strings.SplitN(authHeader, " ", 2)
			if !(len(parts) == 2 && parts[0] == "Bearer") {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "授權格式錯誤",
				})
				c.Abort()
				return
			}
			tokenString = parts[1]
		}

		//  檢查 token 是否存在
		if tokenString == "" {
			log.Printf("缺少 token - WebSocket: %v", c.GetHeader("Upgrade") == "websocket")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供 token",
			})
			c.Abort()
			return
		}

		// 解析 token
		claimsGame, err := ParseTokenGame(tokenString)
		if err != nil {
			log.Printf("JWT 解析失敗: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "無效的 token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 將用戶信息存儲在 context 中
		c.Set("email", claimsGame.Email)                   // 使用者電子郵件
		c.Set("username", claimsGame.Username)             // 使用者名稱
		c.Set("uuid", claimsGame.RegisteredClaims.Subject) // 使用者 UUID

		log.Printf("JWT 認證成功 - email: %s, username: %s, uuid: %s",
			claimsGame.Email, claimsGame.Username, claimsGame.RegisteredClaims.Subject)

		c.Next()
	}
}

// 解析 Game JWT Token
func ParseTokenGame(tokenString string) (*JWTClaimsGame, error) {
	// 解析 JWT Token，並驗證簽名
	// to := tokenString
	// to = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50IjoiZHNhIiwiaXNzIjoibXktaHViLnNpdGUiLCJzdWIiOiJkc2EiLCJleHAiOjE3NTIzNTkwMzEsIm5iZiI6MTc1MjM1NTQzMSwiaWF0IjoxNzUyMzU1NDMxfQ.WOpl_K20rVGu2T8vcgauEf3NlwqyYz3KFBq0QriIDpQ111"
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaimsGame{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非預期的簽名方法: %v", token.Header["alg"])
		}
		return jwtSecretGame, nil
	})

	if err != nil {
		return nil, err
	}

	// 驗證 token 是否有效
	// 如果 token 有效，則將其轉換為自訂的 JWTClaims 結構
	if claimsGame, ok := token.Claims.(*JWTClaimsGame); ok && token.Valid {
		return claimsGame, nil
	}

	return nil, errors.New("無效的 token")
}
