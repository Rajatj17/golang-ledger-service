package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang-exercise/config"
	"golang-exercise/internal/database"
	"golang-exercise/internal/database/model"
	requestdto "golang-exercise/internal/dto/request"
	"golang-exercise/internal/middleware"
	"golang-exercise/internal/router"
	"golang-exercise/tests/helpers"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
	router *gin.Engine
	testDB *helpers.TestDatabase
}

func (suite *APITestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// Load test configuration
	config.Load("../config/test_config.yaml")

	// Setup test database
	testDB, err := helpers.NewTestDatabase()
	if err != nil {
		suite.T().Skip("Skipping E2E tests: Test database not available")
		return
	}
	suite.testDB = testDB

	// Connect to test databases
	database.ConnectDB()

	// Setup router with middleware
	suite.router = gin.New()
	suite.router.Use(middleware.Logger())
	router.SetupRouter(suite.router)
}

func (suite *APITestSuite) SetupTest() {
	if suite.testDB != nil {
		suite.testDB.CleanupTestData()
	}
}

func (suite *APITestSuite) TearDownSuite() {
	if suite.testDB != nil {
		suite.testDB.Close()
	}
}

func (suite *APITestSuite) TestCreateAccount_E2E() {
	if suite.testDB == nil {
		suite.T().Skip("Test database not available")
	}

	tests := []struct {
		name           string
		request        requestdto.CreateAccount
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "create checking account successfully",
			request: requestdto.CreateAccount{
				FirstName:      "John",
				LastName:       "Doe",
				AccountType:    model.AccountTypeChecking,
				Currency:       "USD",
				InitialBalance: decimal.NewFromInt(1500),
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.True(t, response["success"].(bool))
				assert.Contains(t, response["message"], "successfully")

				data, ok := response["data"].(map[string]interface{})
				assert.True(t, ok)

				account, ok := data["account"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "John", account["first_name"])
				assert.Equal(t, "Doe", account["last_name"])
				assert.Contains(t, account["account_number"].(string), "CHE")
			},
		},
		{
			name: "create savings account successfully",
			request: requestdto.CreateAccount{
				FirstName:      "Jane",
				LastName:       "Smith",
				AccountType:    model.AccountTypeSaving,
				Currency:       "USD",
				InitialBalance: decimal.NewFromInt(2000),
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				data := response["data"].(map[string]interface{})
				account := data["account"].(map[string]interface{})
				assert.Contains(t, account["account_number"].(string), "SAV")
			},
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/v1/accounts", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func (suite *APITestSuite) TestGetAccount_E2E() {
	if suite.testDB == nil {
		suite.T().Skip("Test database not available")
	}

	// Create a test account first
	testAccount := suite.testDB.CreateTestAccount("TEST123456")
	if testAccount == nil {
		suite.T().Fatal("Failed to create test account")
	}

	tests := []struct {
		name           string
		accountNumber  string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "get existing account",
			accountNumber:  "TEST123456",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				account := data["account"].(map[string]interface{})
				assert.Equal(t, "TEST123456", account["account_number"])
				assert.Equal(t, "Test", account["first_name"])
			},
		},
		{
			name:           "get non-existent account",
			accountNumber:  "NONEXISTENT",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/accounts/"+tt.accountNumber, nil)
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func (suite *APITestSuite) TestProcessFunds_E2E() {
	if suite.testDB == nil {
		suite.T().Skip("Test database not available")
	}

	// Create a test account
	testAccount := suite.testDB.CreateTestAccount("PROCESS123")
	if testAccount == nil {
		suite.T().Fatal("Failed to create test account")
	}

	tests := []struct {
		name           string
		request        requestdto.MoveMoneyFromAccount
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful deposit",
			request: requestdto.MoveMoneyFromAccount{
				AccountNumber: "PROCESS123",
				Amount:        decimal.NewFromInt(500),
				Type:          model.TransactionTypeDeposit,
				Memo:          "Test deposit",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				assert.NotEmpty(t, data["TransactionID"])
				assert.Equal(t, "IN_PROGRESS", data["Status"])
			},
		},
		{
			name: "successful withdrawal",
			request: requestdto.MoveMoneyFromAccount{
				AccountNumber: "PROCESS123",
				Amount:        decimal.NewFromInt(200),
				Type:          model.TransactionTypeWithdrawal,
				Memo:          "Test withdrawal",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
			},
		},
		{
			name: "invalid transaction type",
			request: requestdto.MoveMoneyFromAccount{
				AccountNumber: "PROCESS123",
				Amount:        decimal.NewFromInt(100),
				Type:          "INVALID",
				Memo:          "Invalid transaction",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/v1/accounts/process", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func (suite *APITestSuite) TestGetAccountBalance_E2E() {
	if suite.testDB == nil {
		suite.T().Skip("Test database not available")
	}

	// Create test account
	testAccount := suite.testDB.CreateTestAccount("BALANCE123")
	if testAccount == nil {
		suite.T().Fatal("Failed to create test account")
	}

	req, _ := http.NewRequest("GET", "/api/v1/accounts/BALANCE123/balance", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "BALANCE123", data["account_number"])
	assert.Equal(suite.T(), "1000", data["balance"])
	assert.Equal(suite.T(), "USD", data["currency"])
}

func (suite *APITestSuite) TestHealthCheck_E2E() {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["message"], "healthy")
}

func TestAPIEndToEndSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
