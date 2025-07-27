## 後端 API 服務說明

本專案後端以 Golang + Gin 框架實作，提供 RESTful API 與 WebSocket 即時互動，涵蓋帳號管理、遊戲房間管理、排行榜、即時聊天室等功能。以下為主要 API 端點與用途說明：

---

### 1. 帳號與驗證相關

- **POST `/api/v1/auth/register`**  
  使用者註冊。  
  參數：`email`, `password`, `captcha`  
  回傳：註冊成功訊息或錯誤原因。

- **POST `/api/v1/auth/login`**  
  使用者登入。  
  參數：`email`, `password`, `captcha`  
  回傳：JWT Token、用戶資訊。

- **GET `/api/v1/auth/captcha`**  
  取得圖片驗證碼。  
  回傳：圖片驗證碼（Base64）與驗證碼 ID。

---

### 2. 遊戲房間管理

- **GET `/api/v1/auth/allGames`**  
  header: `Authorization: Bearer <token>`
  查詢目前所有公開遊戲房間。  
  回傳：房間列表（房號、玩家數、狀態等）。

- **POST `/api/v1/auth/createGame`**  
  header: `Authorization: Bearer <token>`
  建立新遊戲房間，並自動加入該房間。  
  參數：`roomName`  
  回傳：房間資訊。

- **POST `/api/v1/auth/joinGame`**  
  header: `Authorization: Bearer <token>`
  加入指定遊戲房間。  
  參數：`roomId`  
  回傳：加入結果與房間資訊。

---

### 3. 排行榜與歷史紀錄

- **GET `/api/v1/auth/leaderboard`**  
  header: `Authorization: Bearer <token>`
  查詢玩家排行榜（依勝場數排序）。  
  需帶 JWT Token。  
  回傳：玩家列表（名稱、勝場數等）。

- **GET `/api/v1/auth/history`** 
  header: `Authorization: Bearer <token>` 
  查詢個人遊戲歷史紀錄。  
  需帶 JWT Token。  
  回傳：歷史對戰紀錄列表。

---

### 4. 即時互動（WebSocket）

- **GET `/api/v1/auth/wsGame?token={{token}}`**  
  header: `Authorization: Bearer <token>`
  遊戲房間 WebSocket 連線端點。  
  需帶 JWT Token。  
  功能：
  - 即時聊天室
  - 玩家進出房間通知
  - 遊戲開始/結束通知
  - 猜數字遊戲互動（出題、猜測、勝負判斷）



---

## 資料庫說明

### MySQL

- 使用 [GORM](https://gorm.io/) 作為 ORM 框架，操作 MySQL 資料庫。
- 所有用戶資料、遊戲紀錄、排行榜等皆以 GORM Model 定義並持久化於 MySQL。
- 例如：用戶註冊、登入、遊戲歷史查詢、排行榜等功能，皆透過 GORM 進行資料查詢與寫入。

### Redis

- 使用 [go-redis](https://github.com/go-redis/redis) 套件連接 Redis，作為快取與即時狀態管理。
- 房間狀態、玩家即時列表、遊戲進行中資料等，皆存放於 Redis，以提升查詢效能與即時互動體驗。
- 例如：房間內玩家進出、遊戲狀態同步、WebSocket 廣播等，皆透過 Redis 快取與 Pub/Sub 機制實現。

---

### Mysql 資料表結構

| 資料表名稱      | 欄位名稱             | 型態           | 說明                         | 限制/關聯                      |
|----------------|---------------------|----------------|------------------------------|-------------------------------|
| **users**      | id                  | VARCHAR(36)    | 使用者ID，UUID               | PRIMARY KEY                   |
|                | username            | VARCHAR(100)   | 使用者名稱                   | UNIQUE, NOT NULL              |
|                | password_hash       | VARCHAR(255)   | 加密後密碼                   | NOT NULL                      |
|                | email               | VARCHAR(100)   | 電子郵件                     | UNIQUE                        |
|                | created_at          | TIMESTAMP      | 註冊時間                     | 預設 CURRENT_TIMESTAMP        |
||||||
| **game_results** | id                | VARCHAR(36)    | 遊戲結果ID                   | PRIMARY KEY                   |
|                  | game_id           | VARCHAR(36)    | 遊戲ID                       | NOT NULL                      |
|                  | winner_id         | VARCHAR(36)    | 獲勝者的 user_id             | 可為 NULL, 外鍵 users(id)     |
|                  | round             | INT            | 此房間已完第幾輪             | NOT NULL                      |
|                  | answer            | INT            | 當場答案                     | NOT NULL                      |
|                  | total_turns       | INT            | 總猜測回合數                 | 可為 NULL                     |
|                  | total_players     | INT            | 玩家人數                     | 可為 NULL                     |
|                  | finished_at       | TIMESTAMP      | 遊戲結束時間                 | 預設 CURRENT_TIMESTAMP        |
|                  |                   |                |                              | UNIQUE KEY (game_id, round)   |
|                  |                   |                |                              | FOREIGN KEY (winner_id)       |
||||||
| **game_players** | id                | VARCHAR(36)    | 玩家參與記錄ID               | PRIMARY KEY                   |
|                  | game_id           | VARCHAR(36)    | 遊戲ID                       | NOT NULL, 外鍵                |
|                  | user_id           | VARCHAR(36)    | 使用者ID                     | NOT NULL, 外鍵 users(id)      |
|                  | game_results_round| INT            | 此房間已完第幾輪             | NOT NULL, 外鍵                |
|                  | turn_order        | INT            | 輪次順序                     | NOT NULL                      |
|                  | score             | INT            | 分數                         | 預設 0                        |
|                  | guess_count       | INT            | 猜測次數                     | 預設 0                        |
|                  |                   |                |                              | UNIQUE KEY (game_id, game_results_round, user_id) |
|                  |                   |                |                              | FOREIGN KEY (game_id, game_results_round) 參考 game_results(game_id, round) |

---

