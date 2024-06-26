package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"

	"maos-cloud-project-api/config"
	models "maos-cloud-project-api/models"
	utils "maos-cloud-project-api/utils"

	gin "github.com/gin-gonic/gin"
)

type ErrorResponse struct {
    Error string `json:"error"`
}

func getCfg() *config.Config {
	if os.Getenv("ENV") == "production" || os.Getenv("ENV") == "staging" {
		return config.LoadConfig()
	}
	return config.LoadConfig()
}

func Signup( emailSender utils.EmailSender ) gin.HandlerFunc {
		return func(c *gin.Context) {

		var user models.Users

		cfg := getCfg()

		email_verification_token, err := utils.GenerateToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not generate token"})
			return
		}
		
		db, err := config.InitDB(cfg)
		if err != nil {
			// Handle error
			logrus.Error(err)
			panic(err)
		}

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
			return
		}

		// Check for empty email
		if user.Email == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "email must be provided"})
			return
		}

		if !govalidator.IsEmail(user.Email){
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid email address"})
			return
		}

		var existingUser models.Users


		db.Where("email = ?", user.Email).First(&existingUser)

		if existingUser.ID != 0 {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "user already exists"})
			return
		}

		hashedPassword, err := utils.GenerateHashPassword(user.Password)

		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not generate password hash"})
			return
		}

		user.Password = hashedPassword
		user.EmailVerificationToken = email_verification_token
		user.IsEmailVerified = false

		db.Create(&user)

		// Send email verification link
		subject := "Email Verification"
		route := "verify-email"
		body := "Click the link below to verify your email\n\n" + utils.CreateVerificationLink(route, email_verification_token)
		emailSendStatusChan := make(chan error)
		go func ()  {
			
			err := emailSender.SendEmail(cfg, user.Email, subject, body)
			emailSendStatusChan <- err
			close(emailSendStatusChan)
		}()
		// TODO: Handle email sending error
		err = <-emailSendStatusChan
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not send email"})
			logrus.Error("Error sending email: ", err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"success": "user created"})
		return 
	}
}

func SendEmailVerificationLink(emailSender utils.EmailSender) gin.HandlerFunc {
	return func(c *gin.Context) {
		
		var user models.Users
		cfg := config.LoadConfig()

		// check if user is logged in
		cookie, err := c.Cookie("token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, err := utils.ParseToken(cookie)
		if err != nil {
			logrus.Error(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		db, err := config.InitDB(cfg)
		if err != nil {
			// Handle error
			logrus.Error(err)
			panic(err)
		}

		db.Where("email = ?", claims.Subject).First(&user)
		
		//check if user is already verified
		if user.IsEmailVerified {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "email already verified"})
			return
		}

		
		//generate token
		email_verification_token, err := utils.GenerateToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not generate token"})
			return
		}

		//save token to user
		user.EmailVerificationToken = email_verification_token

		//update user
		db.Model(&user).Update("email_verification_token", email_verification_token)
		//send email
		// Send email verification link
		subject := "Email Verification"
		route := "verify-email"
		body := "Click the link below to verify your email\n\n" + utils.CreateVerificationLink(route, email_verification_token)
		emailSendStatusChan := make(chan error)
		go func () {
			err := emailSender.SendEmail(cfg, claims.Subject, subject, body)
			emailSendStatusChan <- err
			close(emailSendStatusChan)
		}()

		err = <-emailSendStatusChan
	
		// Wait for the email sending result
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email."})
			return
		}
		db.Save(&user)
		c.JSON(http.StatusOK, gin.H{"success": "email verification link sent"})
		return
	}
}

func VerifyEmail(c *gin.Context) {

	var user models.Users
	cfg := config.LoadConfig()

	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid token"})
		return
	}

	db, err := config.InitDB(cfg)
	if err != nil {
		// Handle error
		logrus.Error(err)
		panic(err)
	}
	db.Where("email_verification_token = ?", token).First(&user)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid token"})
		return
	}

	user.IsEmailVerified = true
	db.Save(&user)

	// Redirect to a success page
	c.JSON(http.StatusOK, gin.H{"success": "Email verified"})

	// http.Redirect(c.Writer, c.Request, "/api/v1/dashboard", http.StatusSeeOther)
}

func Login(c *gin.Context) {

	var user models.Users
	cfg := config.LoadConfig()

	db, err := config.InitDB(cfg)
	if err != nil {
		// Handle error
		panic(err)
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.Users
	db.Where("email = ?", user.Email).First(&existingUser)

	if existingUser.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	errHash := utils.CompareHashPassword(user.Password, existingUser.Password)

	if !errHash {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)

	claims := &models.Claims{
		Role: existingUser.Role,
		StandardClaims: jwt.StandardClaims{
			Subject:   existingUser.Email,
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey, err := utils.GetSecretKey()
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	jwtKey := []byte(secretKey)

	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		c.JSON(500, gin.H{"error": "could not generate token"})
		return
	}

	c.SetCookie("token", tokenString, int(expirationTime.Unix()), "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"success": "user logged in"})
}

func ResetPassword( emailSender utils.EmailSender  )  gin.HandlerFunc {
	return func(c *gin.Context) {

	var user models.Users
	logrus.Info("ENVVVVVVVVVVV: ", os.Getenv("ENV"))
	cfg := config.LoadConfig()

	db, err := config.InitDB(cfg)
	if err != nil {
		// Handle error
		panic(err)
	}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.Users

	db.Where("email = ?", user.Email).First(&existingUser)

	if existingUser.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user does not exist"})
		return
	}

	//generate token
	reset_password_token, err := utils.GenerateToken()
	//save token to user
	user.ResetPasswordToken = reset_password_token

	//update user
	db.Model(&user).Update("reset_password_token", reset_password_token)
	//send email
	// Send email verification link
	subject := "Password Reset"
	route   := "updatepassword"
	body    := "Click the link below to reset your password\n\n" + utils.CreateVerificationLink(route, reset_password_token)
	emailSendStatusChan := make(chan error)
	go func () {
		err := emailSender.SendEmail(cfg, existingUser.Email, subject, body)
		emailSendStatusChan <- err
		close(emailSendStatusChan)
	}()

	err = <-emailSendStatusChan

	// Wait for the email sending result
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email."})
		return
	}
	db.Save(&user)
	c.JSON(http.StatusOK, gin.H{"success": "password reset link sent to email"})
	return

   }
}

func UpdatePassword(c *gin.Context) {
	
	var user models.Users
	cfg := getCfg()

	db, err := config.InitDB(cfg)
	if err != nil {
		// Handle error
		panic(err)
	}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid token"})
		return
	}
	db.Where("reset_password_token = ?", token).First(&user)

	var existingUser models.Users


	var errHashNew error
	user.Password, errHashNew = utils.GenerateHashPassword(user.Password)

	if errHashNew != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate password hash"})
		return
	}

	db.Model(&existingUser).Update("password", user.Password)
	db.Save(&user)

	c.JSON(http.StatusOK, gin.H{"success": "password updated"})
}

func Dashboard(c *gin.Context) {

	cookie, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	claims, err := utils.ParseToken(cookie)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if claims.Role != "admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "customer dashboard", "role": claims.Role})

}

func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"success": "user logged out"})
}
