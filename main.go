package main

import (
	"log"
	"strings"

	cfg "github.com/deasdania/dating-app/config"
	"github.com/deasdania/dating-app/handlers"
	"github.com/deasdania/dating-app/storage/postgresql"
	"github.com/deasdania/dating-app/storage/postgresutil"
	"github.com/faiface/mainthread"
	"github.com/jmoiron/sqlx"
	"github.com/ory/viper"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

const (
	serviceName = "dating-app"
)

var (
	logger      *logrus.Entry
	config      *viper.Viper
	appMetadata = &config.AppMetadata{}
	dbCon       *sqlx.DB
)

func initLogger(config *viper.Viper) (*logrus.Entry, error) {
	l := cfg.NewLogger()
	var logLevel logrus.Level

	llStr := config.GetString("server.logLevel")
	appEnvStr := config.GetString("server.appEnv")
	if appEnvStr == "" {
		logger.Fatal("no configured app environment")
	}
	if llStr == "fromenv" {
		switch config.GetString("runtime.environment") {
		case "staging", "development":
			logLevel = logrus.DebugLevel // to simplify debugging
		default: // including production
			logLevel = logrus.InfoLevel
		}
	} else {
		var err error
		logLevel, err = logrus.ParseLevel(llStr)
		if err != nil {
			return nil, err
		}
	}

	l.SetLevel(logLevel)
	return l.WithFields(logrus.Fields{
		"service": serviceName,
		"app_env": appEnvStr,
	}), nil
}

func init() {
	config = viper.NewWithOptions(
		viper.EnvKeyReplacer(
			strings.NewReplacer(".", "_"),
		),
	)
	config.SetConfigFile("env/config")
	config.SetConfigType("ini")
	config.AutomaticEnv()
	if err := config.ReadInConfig(); err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}

	var err error
	logger, err = initLogger(config)
	if err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}

	appEnvStr := config.GetString("server.appEnv")
	if appEnvStr == "" {
		logger.Fatal("no configured app environment")
	}
	appEnvStr = strings.Title(strings.ToLower(appEnvStr))

	e := strings.ToLower(config.GetString("runtime.environment"))
	switch e {
	case "staging":
		appMetadata.Env = cfg.Env_Staging
	case "production":
		appMetadata.Env = cfg.Env_Production
	default:
		appMetadata.Env = cfg.Env_Development
	}
}

func main() {
	mainthread.Run(runServer)
}

func runserver() {
	validate := cfg.NewValidator()
	e := cfg.NewEcho(config, validate)

	// Create a new postgres storage
	var err error
	dbCon, err = postgresutil.NewStorageWithTracing(logger, config)
	if err != nil {
		cfg.WithError(err, logger).Fatal("error initializing postgres connection")
	}
	defer dbCon.Close()

	appEnvStr := config.GetString("server.appEnv")
	logger.Info("appEnv:", appEnvStr)
	store, err := postgresql.NewStorageFromConn(logger, dbCon, appEnvStr)
	if err != nil {
		cfg.WithError(err, logger).Fatal("error initializing postgres connection")
	}

	handlers.Bootstrap(&handlers.API{
		App:      e,
		Log:      logger,
		Validate: validate.Validator,
		Config:   config,
		Storage:  store,
	})

	if config.GetBool(`debug`) {
		logger.Info("Service RUN on DEBUG mode")
	}

}

// Sign-up route
// func signUp(c *gin.Context) {
// 	var user models.User
// 	if err := c.ShouldBindJSON(&user); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
// 		return
// 	}

// 	// Hash password
// 	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
// 	user.Password = string(hashedPassword)

// 	// Save user to database
// 	if err := db.Create(&user).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
// }

// Login route
// func login(c *gin.Context) {
// 	var user models.User
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

// // Get profile route
// func getProfile(c *gin.Context) {
// 	var profiles []models.Profile
// 	db.Limit(10).Find(&profiles)

// 	c.JSON(http.StatusOK, gin.H{"profiles": profiles})
// }

// // Swipe route
// func swipe(c *gin.Context) {
// 	// var input struct {
// 	// 	ProfileID uuid.UUID `json:"profile_id"`
// 	// 	Direction string    `json:"direction"` // "like" or "pass"
// 	// }

// 	// if err := c.ShouldBindJSON(&input); err != nil {
// 	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
// 	// 	return
// 	// }

// 	// // Extract user ID from JWT
// 	// userID := utils.GetUserIDFromJWT(c)
// 	// if userID == uuid.Nil {
// 	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 	// 	return
// 	// }

// 	// // Check daily swipe limit
// 	// var user models.User
// 	// if err := db.First(&user, "id = ?", userID).Error; err != nil {
// 	// 	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 	// 	return
// 	// }

// 	// if user.DailySwipeCount >= 10 && !user.IsPremium {
// 	// 	c.JSON(http.StatusForbidden, gin.H{"error": "Daily swipe limit reached"})
// 	// 	return
// 	// }

// 	// // Save the swipe
// 	// sw := models.Swipe{
// 	// 	UserID:    &userID,
// 	// 	ProfileID: &input.ProfileID,
// 	// 	Direction: input.Direction,
// 	// }
// 	// if err := db.Create(&sw).Error; err != nil {
// 	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record swipe"})
// 	// 	return
// 	// }

// 	// // Increment the daily swipe count
// 	// if !user.IsPremium {
// 	// 	user.DailySwipeCount++
// 	// 	if err := db.Save(&user).Error; err != nil {
// 	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update swipe count"})
// 	// 		return
// 	// 	}
// 	// }

// 	c.JSON(http.StatusOK, gin.H{"message": "Swipe recorded"})
// }

// // Purchase premium route
// func purchasePremium(c *gin.Context) {
// 	// userID := utils.GetUserIDFromJWT(c)
// 	// if userID == uuid.Nil {
// 	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 	// 	return
// 	// }

// 	// var user models.User
// 	// if err := db.First(&user, "id = ?", userID).Error; err != nil {
// 	// 	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 	// 	return
// 	// }

// 	// // Update the user to premium
// 	// user.IsPremium = true
// 	// if err := db.Save(&user).Error; err != nil {
// 	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlock premium"})
// 	// 	return
// 	// }

// 	c.JSON(http.StatusOK, gin.H{"message": "Premium features unlocked"})
// }

// func updateProfile(c *gin.Context) {
// 	// Extract user ID from JWT token (you should already have the getUserIDFromJWT function)
// 	// userID := utils.GetUserIDFromJWT(c)
// 	// if userID == uuid.Nil {
// 	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 	// 	return
// 	// }

// 	// // Retrieve the user's profile from the database
// 	// var profile models.Profile
// 	// if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
// 	// 	c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
// 	// 	return
// 	// }

// 	// // Define a struct for the input payload
// 	// var input struct {
// 	// 	Username    string `json:"username"`
// 	// 	Description string `json:"description"`
// 	// 	ImageURL    string `json:"image_url"`
// 	// }

// 	// // Bind the incoming JSON to the input struct
// 	// if err := c.ShouldBindJSON(&input); err != nil {
// 	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
// 	// 	return
// 	// }

// 	// // Update profile fields
// 	// profile.Username = input.Username
// 	// profile.Description = input.Description
// 	// profile.ImageURL = input.ImageURL
// 	// profile.UpdatedAt = time.Now()

// 	// // Save the updated profile back to the database
// 	// if err := db.Save(&profile).Error; err != nil {
// 	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
// 	// 	return
// 	// }

// 	// c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully", "profile": profile})
// }
