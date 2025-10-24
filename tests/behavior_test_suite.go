package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/quantfidential/trading-ecosystem/custodian-data-adapter-go/pkg/adapters"
	"github.com/quantfidential/trading-ecosystem/custodian-data-adapter-go/pkg/interfaces"
	"github.com/quantfidential/trading-ecosystem/custodian-data-adapter-go/pkg/models"
)

// BehaviorTestSuite provides the base test suite for custodian behavior tests
type BehaviorTestSuite struct {
	suite.Suite
	ctx     context.Context
	adapter *adapters.CustodianDataAdapter
	config  *adapters.Config

	// Tracking for cleanup
	createdPositions   []string
	createdSettlements []string
	createdBalances    []string
	createdServices    []string
}

// SetupSuite runs once before all tests
func (suite *BehaviorTestSuite) SetupSuite() {
	suite.T().Log("Setting up behavior test suite")

	// Get configuration from environment
	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		postgresURL = "postgres://custodian_adapter:custodian-adapter-db-pass@localhost:5432/trading_ecosystem?sslmode=disable"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://custodian-adapter:custodian-pass@localhost:6379/0"
	}

	// Create configuration
	suite.config = &adapters.Config{
		PostgresURL: postgresURL,
		RedisURL:    redisURL,
		ServiceName: "custodian-test",
		ServiceInstanceName: "custodian-test-" + GenerateTestID("instance"),
		Environment: "testing",
	}

	// Create adapter
	adapter, err := adapters.NewCustodianDataAdapter(suite.config)
	suite.Require().NoError(err, "Failed to create custodian data adapter")
	suite.adapter = adapter

	suite.T().Logf("Custodian Data Adapter initialized with schema: %s, namespace: %s",
		suite.config.SchemaName, suite.config.RedisNamespace)
}

// SetupTest runs before each test
func (suite *BehaviorTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.createdPositions = []string{}
	suite.createdSettlements = []string{}
	suite.createdBalances = []string{}
	suite.createdServices = []string{}
}

// TearDownTest runs after each test
func (suite *BehaviorTestSuite) TearDownTest() {
	suite.T().Log("Cleaning up test data")

	// Clean up created test data
	for _, posID := range suite.createdPositions {
		if err := suite.adapter.DeletePosition(suite.ctx, posID); err != nil {
			suite.T().Logf("Warning: Failed to delete position %s: %v", posID, err)
		}
	}

	for _, settlementID := range suite.createdSettlements {
		if err := suite.adapter.DeleteSettlement(suite.ctx, settlementID); err != nil {
			suite.T().Logf("Warning: Failed to delete settlement %s: %v", settlementID, err)
		}
	}

	for _, accountID := range suite.createdBalances {
		if err := suite.adapter.DeleteBalance(suite.ctx, accountID); err != nil {
			suite.T().Logf("Warning: Failed to delete balance %s: %v", accountID, err)
		}
	}

	for _, serviceID := range suite.createdServices {
		if err := suite.adapter.DeregisterService(suite.ctx, serviceID); err != nil {
			suite.T().Logf("Warning: Failed to deregister service %s: %v", serviceID, err)
		}
	}
}

// TearDownSuite runs once after all tests
func (suite *BehaviorTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down behavior test suite")

	if suite.adapter != nil {
		err := suite.adapter.Disconnect(suite.ctx)
		if err != nil {
			suite.T().Logf("Warning: Error disconnecting adapter: %v", err)
		}
	}
}

// BDD-style test helpers

// Given starts a BDD scenario with a given condition
func (suite *BehaviorTestSuite) Given(description string, fn func()) *BDDScenario {
	scenario := &BDDScenario{suite: suite}
	scenario.givens = append(scenario.givens, BDDStep{description, fn})
	return scenario
}

// CreateTestPosition creates a position for testing
func (suite *BehaviorTestSuite) CreateTestPosition(positionID string, modifiers ...func(*models.Position)) *models.Position {
	position := &models.Position{
		ID:          positionID,
		AccountID:   "test-account-" + GenerateTestID("account"),
		Symbol:      "BTC-USD",
		Quantity:    "1.5",
		AveragePrice: "50000.00",
		Side:        "long",
		Status:      "open",
		Timestamp:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	for _, modifier := range modifiers {
		modifier(position)
	}

	return position
}

// CreateTestSettlement creates a settlement for testing
func (suite *BehaviorTestSuite) CreateTestSettlement(settlementID string, modifiers ...func(*models.Settlement)) *models.Settlement {
	settlement := &models.Settlement{
		ID:            settlementID,
		AccountID:     "test-account-" + GenerateTestID("account"),
		Symbol:        "USD",
		Amount:        "10000.00",
		SettlementType: "deposit",
		Status:        "pending",
		Timestamp:     time.Now(),
		CompletedAt:   nil,
	}

	for _, modifier := range modifiers {
		modifier(settlement)
	}

	return settlement
}

// CreateTestBalance creates a balance for testing
func (suite *BehaviorTestSuite) CreateTestBalance(accountID string, modifiers ...func(*models.Balance)) *models.Balance {
	balance := &models.Balance{
		AccountID:       accountID,
		Symbol:          "USD",
		Available:       "100000.00",
		Reserved:        "0.00",
		Total:           "100000.00",
		LastUpdated:     time.Now(),
	}

	for _, modifier := range modifiers {
		modifier(balance)
	}

	return balance
}

// CreateTestServiceRegistration creates a service registration for testing
func (suite *BehaviorTestSuite) CreateTestServiceRegistration(serviceID string, modifiers ...func(*models.ServiceRegistration)) *models.ServiceRegistration {
	service := &models.ServiceRegistration{
		ID:       serviceID,
		Name:     "test-service",
		Version:  "1.0.0",
		Host:     "localhost",
		Port:     8080,
		Protocol: "http",
		Status:   "healthy",
		Metadata: map[string]string{
			"environment": "testing",
		},
		RegisteredAt: time.Now(),
		LastHeartbeat: time.Now(),
	}

	for _, modifier := range modifiers{
		modifier(service)
	}

	return service
}

// Track created entities for cleanup
func (suite *BehaviorTestSuite) trackCreatedPosition(positionID string) {
	suite.createdPositions = append(suite.createdPositions, positionID)
}

func (suite *BehaviorTestSuite) trackCreatedSettlement(settlementID string) {
	suite.createdSettlements = append(suite.createdSettlements, settlementID)
}

func (suite *BehaviorTestSuite) trackCreatedBalance(accountID string) {
	suite.createdBalances = append(suite.createdBalances, accountID)
}

func (suite *BehaviorTestSuite) trackCreatedService(serviceID string) {
	suite.createdServices = append(suite.createdServices, serviceID)
}

// BDDScenario represents a BDD-style test scenario
type BDDScenario struct {
	suite  *BehaviorTestSuite
	givens []BDDStep
	whens  []BDDStep
	thens  []BDDStep
}

// BDDStep represents a single step in a BDD scenario
type BDDStep struct {
	description string
	fn          func()
}

// When adds a when condition
func (s *BDDScenario) When(description string, fn func()) *BDDScenario {
	s.whens = append(s.whens, BDDStep{description, fn})
	return s
}

// Then adds a then assertion
func (s *BDDScenario) Then(description string, fn func()) *BDDScenario {
	s.thens = append(s.thens, BDDStep{description, fn})
	s.execute()
	return s
}

// And adds another condition to the current step type
func (s *BDDScenario) And(description string, fn func()) *BDDScenario {
	if len(s.thens) > 0 {
		// Add to thens
		s.thens = append(s.thens, BDDStep{description, fn})
	} else if len(s.whens) > 0 {
		// Add to whens
		s.whens = append(s.whens, BDDStep{description, fn})
	} else {
		// Add to givens
		s.givens = append(s.givens, BDDStep{description, fn})
	}
	return s
}

// execute runs all steps in the scenario
func (s *BDDScenario) execute() {
	// Execute givens
	for _, given := range s.givens {
		s.suite.T().Logf("  Given %s", given.description)
		given.fn()
	}

	// Execute whens
	for _, when := range s.whens {
		s.suite.T().Logf("  When %s", when.description)
		when.fn()
	}

	// Execute thens
	for _, then := range s.thens {
		s.suite.T().Logf("  Then %s", then.description)
		then.fn()
	}
}

// RunScenario runs a predefined scenario
func (suite *BehaviorTestSuite) RunScenario(scenarioName string) {
	log.Printf("Running scenario: %s", scenarioName)
	// Placeholder for scenario execution
	// Real implementation would load from behavior_scenarios.go
}
