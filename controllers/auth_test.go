package controllers // Replace with your actual package name

import (
	"bytes"
	"os"
	"testing"

	"encoding/json"
	"net/http"
	"net/http/httptest"
	// "strings"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"maos-cloud-project-api/config"
	"maos-cloud-project-api/mocks"
	"maos-cloud-project-api/models"
	utils "maos-cloud-project-api/utils"

	"github.com/stretchr/testify/mock"
)

type SignupTestSuite struct {
	suite.Suite
	router *gin.Engine
	w      *httptest.ResponseRecorder
	c      *gin.Context
	db     *gorm.DB
	mockEmailSender *mocks.MockEmailSender
	
}

func (s *SignupTestSuite) SetupTest() {

	
	cfg := config.LoadConfig()

	db, err := config.InitDB(cfg)
	if err != nil {
		s.T().Fatal("Error initializing database connection", err)
	}
	// TODO: Move this to teardown function
	result := db.Exec("TRUNCATE TABLE users RESTART IDENTITY")
	if result.Error != nil {
		s.T().Fatal("Failed to truncate table:", result.Error)
	}
	
	db.AutoMigrate(&models.Users{})

    s.router = utils.SetUpRouter()
	s.router.POST("/api/v1/signup", func (c *gin.Context ) {
		Signup(s.mockEmailSender)(c)
	})
	s.db = db

	s.w = httptest.NewRecorder()
	s.c, _ = gin.CreateTestContext(s.w)
	s.mockEmailSender = new(mocks.MockEmailSender)
}

// Simulate a HTTP request with a user body
func (s *SignupTestSuite) prepareTestContext(userBody []byte) (*gin.Context, *httptest.ResponseRecorder) {
    // Initialize the response recorder
    w := httptest.NewRecorder()

    // Create a new HTTP request with the user body
    req := httptest.NewRequest("POST", "/api/v1/signup", bytes.NewBuffer(userBody))
    req.Header.Add("Content-Type", "application/json")

    // Create a new gin context from the request
    c, _ := gin.CreateTestContext(w)
    c.Request = req

    return c, w
}


func (s *SignupTestSuite) Test_ValidSignup() {

	user := map[string]interface{}{ 
		"name":     "test",
		"email":    "test@gmail.com",
		"password": "plainPassword123", 
		"role":     "admin", 
	  }
	userBody, _ := json.Marshal(user)
	s.T().Log("USER BODY REQ: ", bytes.NewBuffer(userBody))
	s.mockEmailSender.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	ctx, w := s.prepareTestContext(userBody)
	Signup(s.mockEmailSender)(ctx)

	s.T().Log("RESPONSE BODY: ", w.Body.String())
	assert.Equal(s.T(), http.StatusCreated, w.Code)
	assert.Contains(s.T(), w.Body.String(), "user created")

}

func TestSignupSuite(t *testing.T) {
	suite.Run(t, new(SignupTestSuite))
}

func (s *SignupTestSuite) TearDownSuite() {

    if err := s.db.Migrator().DropTable(&models.Users{}); err != nil {
        s.T().Error("Failed to drop table:", err)
    } else {
        s.T().Log("TearDownSuite: Users table dropped")
    }

}

type LoginTestSuite struct {
	suite.Suite
	router *gin.Engine
	w      *httptest.ResponseRecorder
	c      *gin.Context
	db     *gorm.DB
	mockEmailSender *mocks.MockEmailSender
}

func (s *LoginTestSuite) SetupTest() {

	os.Setenv("SECRET_KEY", "testkey")

	cfg := config.LoadConfig()
	db, err := config.InitDB(cfg)
	if err != nil {
		// Handle error
		s.T().Fatal("Error initializing database connection")
	}
	db.AutoMigrate(&models.Users{})

	s.router = utils.SetUpRouter()
	s.router.POST("/api/v1/login", Login)
	s.db = db

	s.w = httptest.NewRecorder()
	s.c, _ = gin.CreateTestContext(s.w)
	s.mockEmailSender = new(mocks.MockEmailSender)
	
}

func (s *LoginTestSuite) prepareTestContext(userBody []byte) (*gin.Context, *httptest.ResponseRecorder) {
	// Initialize the response recorder
	w := httptest.NewRecorder()

	// Create a new HTTP request with the user body
	req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(userBody))
	req.Header.Add("Content-Type", "application/json")

	// Create a new gin context from the request
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	return c, w
}

func (s *LoginTestSuite) Test_ValidLogin() {

	signupUser := map[string]interface{}{ 
		"name":     "test",
		"email":    "test@gmail.com",
		"password": "plainPassword123", 
		"role":     "admin", 
	  }
	userBody, _ := json.Marshal(signupUser)
	s.mockEmailSender.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	ctx, w := s.prepareTestContext(userBody)
	Signup(s.mockEmailSender)(ctx)
	s.T().Log("LOGIN_SUITE: Signup RESPONSE BODY: ", w.Body.String())

	loginUser := map[string]interface{}{	
		"email":    "test@gmail.com",
		"password": "plainPassword123",
	}
	loginBody, _ := json.Marshal(loginUser)
	s.T().Log("USER BODY REQ: ", bytes.NewBuffer(loginBody))

	ctx, w = s.prepareTestContext(loginBody)
	Login(ctx)

	s.T().Log("Login RESPONSE BODY: ", w.Body.String())
	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.Contains(s.T(), w.Body.String(), "user logged in")
}

func (s *LoginTestSuite) Test_InvalidEmailLogin() {

	signupUser := map[string]interface{}{ 
		"name":     "test",
		"email":    "test@gmail.com",
		"password": "plainPassword123", 
		"role":     "admin", 
	  }
	userBody, _ := json.Marshal(signupUser)
	s.T().Log("USER BODY REQ: ", bytes.NewBuffer(userBody))
	s.mockEmailSender.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	ctx, w := s.prepareTestContext(userBody)
	Signup(s.mockEmailSender)(ctx)
	s.T().Log("Signup RESPONSE BODY: ", w.Body.String())

	LoginUser := map[string]interface{}{ 
		"email":    "email@me",
		"password": "plainPassword123",
	}
	loginBody, _ := json.Marshal(LoginUser)
	s.T().Log("USER BODY REQ: ", bytes.NewBuffer(loginBody))

	ctx, w = s.prepareTestContext(loginBody)
	Login(ctx)

	s.T().Log("RESPONSE BODY: ", w.Body.String())
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(s.T(), w.Body.String(), "invalid username or password")

}

func (s *LoginTestSuite) Test_InvalidPasswordLogin() {

	signupUser := map[string]interface{}{ 
		"name":     "test",
		"email":    "test@gmail.com",
		"password": "plainPassword123", 
		"role":     "admin", 
	  }

	userBody, _ := json.Marshal(signupUser)
	s.T().Log("USER BODY REQ: ", bytes.NewBuffer(userBody))
	s.mockEmailSender.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	ctx, w := s.prepareTestContext(userBody)
	Signup(s.mockEmailSender)(ctx)
	s.T().Log("Signup RESPONSE BODY: ", w.Body.String())
	
	LoginUser := map[string]interface{}{
		"email":    "test@gmail.com",
		"password": "plainPassw",
	}
	loginBody, _ := json.Marshal(LoginUser)
	s.T().Log("USER BODY REQ: ", bytes.NewBuffer(loginBody))
		
	ctx, w = s.prepareTestContext(loginBody)
	Login(ctx)

	s.T().Log("RESPONSE BODY: ", w.Body.String())
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(s.T(), w.Body.String(), "invalid username or password")

}

func TestLoginSuite(t *testing.T) {
	suite.Run(t, new(LoginTestSuite))
}

func (s *LoginTestSuite) TearDownSuite() {

	s.db.Exec("DROP TABLE users")
	s.T().Log("TearDownSuite")

}

type EmailVerficationLinkTestSuite struct {
	suite.Suite
	router          *gin.Engine
	db 			    *gorm.DB
	w               *httptest.ResponseRecorder
	c               *gin.Context
	mockEmailSender *mocks.MockEmailSender
}

func (s *EmailVerficationLinkTestSuite) prepareTestContext(userBody []byte) (*gin.Context, *httptest.ResponseRecorder) {
	// Initialize the response recorder
	w := httptest.NewRecorder()

	// Create a new HTTP request with the user body
	req := httptest.NewRequest("POST", "/api/v1/send-verification-email", bytes.NewBuffer(userBody))
	req.Header.Add("Content-Type", "application/json")

	// Create a new gin context from the request
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	return c, w
}

func (s *EmailVerficationLinkTestSuite) SetupTest() {

	cfg := config.LoadConfig()
	db, err := config.InitDB(cfg)
	if err != nil {
		// Handle error
		s.T().Fatal("Error initializing database connection")
	}
	db.AutoMigrate(&models.Users{})
	s.mockEmailSender = new(mocks.MockEmailSender)

	s.router = utils.SetUpRouter()
	s.router.POST("/api/v1/send-verification-email", func(c *gin.Context) {
		SendEmailVerificationLink(s.mockEmailSender)(c)
	})
	s.db = db

	s.w = httptest.NewRecorder()
	s.c, _ = gin.CreateTestContext(s.w)
	s.mockEmailSender = new(mocks.MockEmailSender)

}

func (suite *EmailVerficationLinkTestSuite) Test_SendEmailVerificationLinkSuccess() {

	signupUser := map[string]interface{}{ 
		"name":     "test",
		"email":    "test@gmail.com",
		"password": "plainPassword123", 
		"role":     "admin", 
	  }
	userBody, _ := json.Marshal(signupUser)
	suite.mockEmailSender.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// suite.mockEmailSender.On("SendEmail",
	// 	mock.MatchedBy(func(email string) bool { return email == "test@gmail.com" }),
	// 	mock.MatchedBy(func(subject string) bool { return subject == "Email Verification" }),
	// 	mock.MatchedBy(func(body string) bool { return strings.Contains(body, "Click the link below to verify your email") }),
	// ).Return(nil)

	ctx, w := suite.prepareTestContext(userBody)
	Signup(suite.mockEmailSender)(ctx)
	suite.T().Log("EMAIL_VERIFICATION_LINK_SUITE: Signup RESPONSE BODY: ", w.Body.String())

	loginUser := map[string]interface{}{	
		"email":    "test@gmail.com",
		"password": "plainPassword123",
	}
	loginBody, _ := json.Marshal(loginUser)

	ctx, w = suite.prepareTestContext(loginBody)
	Login(ctx)

	var token string
	cookies := w.Result().Cookies()

	for _, cookie := range cookies {
		if cookie.Name == "token" {
			token = cookie.Value
			break
		}
	}

	if token == "" {
		suite.T().Fatal("Token not found in cookies")
	}
		
	emailPayload := map[string]string{
		"email": "test@gmail.com",
	}
	emailBody, err := json.Marshal(emailPayload)
	if err != nil {
		suite.T().Fatal("Failed to marshal JSON:", err)
	}
	
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/send-verification-email", bytes.NewBuffer(emailBody))
	req.AddCookie(&http.Cookie{Name: "token", Value: token})

	resp := httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)

	suite.T().Log("EMAIL_VERIFICATION_LINK_SUITE RESPONSE BODY: ", resp.Body.String())

	suite.Equal(http.StatusOK, resp.Code)
	// suite.mockEmailSender.AssertExpectations(suite.T())

}

// func (suite *EmailVerficationLinkTestSuite) Test_SendEmailVerificationAlreadyVerified(){
// 	suite.user.IsEmailVerified = true
// 	req, _ := http.NewRequest(http.MethodPost, "/api/v1/send-verification-email", nil)
// 	resp := httptest.NewRecorder()
// 	suite.router.ServeHTTP(resp, req)

// 	suite.Equal(http.StatusBadRequest, resp.Code)
// }

func TestEmailVerficationLinkTestSuite(t *testing.T) {
	suite.Run(t, new(EmailVerficationLinkTestSuite))
}

func (s *EmailVerficationLinkTestSuite) TearDownSuite() {

	// s.db.Exec("DROP TABLE users")
	s.T().Log("TearDownSuite")

}