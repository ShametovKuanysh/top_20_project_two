package api

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
		// panic("Error loading .env file")
	}

	dsn := os.Getenv("DB_URL")

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database", err)
		// panic("Failed to connect to database")
	}

	DB.AutoMigrate(&User{}, &Task{})
}

func GetTasks(c *gin.Context) {
	var tasks []Task

	status := c.Query("status")
	if status != "" {
		DB.Where("status =?", status).Find(&tasks)
	} else {
		DB.Find(&tasks)
	}

	ResponseJSON(c, http.StatusOK, "Tasks retrieved successfully", tasks)
}

func GetTask(c *gin.Context) {
	id := c.Param("id")
	var task Task
	if err := DB.Where("id = ?", id).First(&task).Error; err != nil {
		ResponseJSON(c, http.StatusNotFound, "Task not found", nil)
		return
	}
	ResponseJSON(c, http.StatusOK, "Task retrieved successfully", task)
}

func CreateTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		ResponseJSON(c, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	user_id, exist := c.Get("user_id")
	if !exist {
		ResponseJSON(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	task.UserID = user_id.(uint)

	DB.Create(&task)
	ResponseJSON(c, http.StatusCreated, "Task created successfully", task)
}

func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var task Task
	if err := DB.Where("id =?", id).First(&task).Error; err != nil {
		ResponseJSON(c, http.StatusNotFound, "Task not found", nil)
		return
	}

	var updatedTask Task
	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		ResponseJSON(c, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	updatedTask.ID = task.ID
	DB.Save(&updatedTask)
	ResponseJSON(c, http.StatusOK, "Task updated successfully", updatedTask)
}

func DeleteTask(c *gin.Context) {
	id := c.Param("id")
	var task Task
	if err := DB.Where("id =?", id).First(&task).Error; err != nil {
		ResponseJSON(c, http.StatusNotFound, "Task not found", nil)
		return
	}

	DB.Delete(&task)
	ResponseJSON(c, http.StatusOK, "Task deleted successfully", nil)
}

func Register(c *gin.Context) {
	var request RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		ResponseJSON(c, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if request.Password != request.ConfirmPassword {
		ResponseJSON(c, http.StatusBadRequest, "Passwords do not match", nil)
		return
	}

	var user User
	if err := DB.Where("email =?", request.Email).First(&user).Error; err == nil {
		ResponseJSON(c, http.StatusConflict, "Email already exists", nil)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		ResponseJSON(c, http.StatusInternalServerError, "Error hashing password", nil)
		return
	}

	user.Name = request.Name
	user.Email = request.Email
	user.Password = string(hashedPassword)
	DB.Create(&user)
	ResponseJSON(c, http.StatusCreated, "User registered successfully", user)
}

func Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		ResponseJSON(c, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	var user User
	if err := DB.Where("email =?", request.Email).First(&user).Error; err != nil {
		ResponseJSON(c, http.StatusNotFound, "User not found", nil)
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		ResponseJSON(c, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		ResponseJSON(c, http.StatusInternalServerError, "Error generating token", nil)
		return
	}

	ResponseJSON(c, http.StatusOK, "Logged in successfully", gin.H{"token": token})
}

func GenerateToken(id uint) (string, error) {
	// Create a new JWT token with claims
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": id,                               // Subject (user identifier)
		"iss": "todo-app",                       // Issuer
		"exp": time.Now().Add(time.Hour).Unix(), // Expiration time
		"iat": time.Now().Unix(),                // Issued at
	})

	// Print information about the created token
	log.Printf("Token claims added: %+v\n", claims)

	tokenString, err := claims.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
