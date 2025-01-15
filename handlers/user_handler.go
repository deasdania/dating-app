package handlers

import (
	"net/http"

	"github.com/deasdania/dating-app/models"
	"github.com/deasdania/dating-app/status"
	smodels "github.com/deasdania/dating-app/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

func (h *Handlers) SignUp(c echo.Context) error {
	var user smodels.User
	if err := c.Bind(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}
	ctx := c.Request().Context()
	if err := validateStruct(h.validate, user); err != nil {
		h.log.Error("err validator:", err)
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusBadRequest, status.UserErrCode_InvalidRequest, err.Error()))
		return err
	}

	if err := h.core.SignUp(ctx, &user); err != nil {
		c.JSON(http.StatusBadRequest, models.NewResponseError(http.StatusInternalServerError, status.SystemErrCode_Generic, err.Error()))
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
	return nil
}

// Login route
// func (h *Handlers) Login(c *gin.Context) {
// 	var user smodels.User
// 	var input struct {
// 		Username string `json:"username"`
// 		Password string `json:"password"`
// 	}

// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
// 		return
// 	}

// 	// Check if user exists
// 	if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
// 		return
// 	}

// 	// Check password
// 	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
// 		return
// 	}

// 	// Generate JWT
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		"sub": user.ID.String(), // Store the UUID as a string in the token
// 		"exp": time.Now().Add(time.Hour * 24).Unix(),
// 	})
// 	tokenString, err := token.SignedString([]byte("secret"))

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"token": tokenString})
// }

// Get profile route
// func getProfile(c *gin.Context) {
// 	var profiles []models.Profile
// 	db.Limit(10).Find(&profiles)

// 	c.JSON(http.StatusOK, gin.H{"profiles": profiles})
// }
