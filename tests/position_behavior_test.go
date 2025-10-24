package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/quantfidential/trading-ecosystem/custodian-data-adapter-go/pkg/models"
)

// PositionBehaviorTestSuite tests the behavior of position repository operations
type PositionBehaviorTestSuite struct {
	BehaviorTestSuite
}

// TestPositionBehaviorSuite runs the position behavior test suite
func TestPositionBehaviorSuite(t *testing.T) {
	suite.Run(t, new(PositionBehaviorTestSuite))
}

// TestPositionCRUDOperations tests basic position CRUD operations
func (suite *PositionBehaviorTestSuite) TestPositionCRUDOperations() {
	var positionID = GenerateTestUUID()

	suite.Given("a new position to create", func() {
		// Position defined below
	}).When("creating the position", func() {
		position := suite.CreateTestPosition(positionID, func(p *models.Position) {
			p.Symbol = "BTC-USD"
			p.Quantity = "2.5"
			p.AveragePrice = "45000.00"
			p.Side = "long"
			p.Status = "open"
		})

		err := suite.adapter.CreatePosition(suite.ctx, position)
		suite.Require().NoError(err)
		suite.trackCreatedPosition(positionID)
	}).Then("the position should be retrievable", func() {
		retrieved, err := suite.adapter.GetPosition(suite.ctx, positionID)
		suite.Require().NoError(err)
		suite.Equal(positionID, retrieved.ID)
		suite.Equal("BTC-USD", retrieved.Symbol)
		suite.Equal("2.5", retrieved.Quantity)
	}).And("the position can be updated", func() {
		err := suite.adapter.UpdatePosition(suite.ctx, positionID, func(p *models.Position) {
			p.Quantity = "3.0"
			p.AveragePrice = "46000.00"
		})
		suite.Require().NoError(err)

		updated, err := suite.adapter.GetPosition(suite.ctx, positionID)
		suite.Require().NoError(err)
		suite.Equal("3.0", updated.Quantity)
	})
}

// TestPositionQueryByAccount tests querying positions by account
func (suite *PositionBehaviorTestSuite) TestPositionQueryByAccount() {
	var (
		accountID  = "test-account-" + GenerateTestID("query")
		positionID1 = GenerateTestUUID()
		positionID2 = GenerateTestUUID()
	)

	suite.Given("multiple positions for an account", func() {
		// Create first position
		pos1 := suite.CreateTestPosition(positionID1, func(p *models.Position) {
			p.AccountID = accountID
			p.Symbol = "BTC-USD"
			p.Quantity = "1.0"
		})
		err := suite.adapter.CreatePosition(suite.ctx, pos1)
		suite.Require().NoError(err)
		suite.trackCreatedPosition(positionID1)

		// Create second position
		pos2 := suite.CreateTestPosition(positionID2, func(p *models.Position) {
			p.AccountID = accountID
			p.Symbol = "ETH-USD"
			p.Quantity = "10.0"
		})
		err = suite.adapter.CreatePosition(suite.ctx, pos2)
		suite.Require().NoError(err)
		suite.trackCreatedPosition(positionID2)
	}).When("querying positions by account", func() {
		positions, err := suite.adapter.GetPositionsByAccount(suite.ctx, accountID)
		suite.Require().NoError(err)

		suite.Then("all account positions should be returned", func() {
			suite.GreaterOrEqual(len(positions), 2)
			
			// Verify both symbols are present
			symbols := make(map[string]bool)
			for _, pos := range positions {
				symbols[pos.Symbol] = true
			}
			suite.True(symbols["BTC-USD"])
			suite.True(symbols["ETH-USD"])
		})
	})
}

// TestPositionStatusTransitions tests position status lifecycle
func (suite *PositionBehaviorTestSuite) TestPositionStatusTransitions() {
	var positionID = GenerateTestUUID()

	suite.Given("an open position", func() {
		position := suite.CreateTestPosition(positionID, func(p *models.Position) {
			p.Status = "open"
		})
		err := suite.adapter.CreatePosition(suite.ctx, position)
		suite.Require().NoError(err)
		suite.trackCreatedPosition(positionID)
	}).When("closing the position", func() {
		err := suite.adapter.UpdatePosition(suite.ctx, positionID, func(p *models.Position) {
			p.Status = "closed"
			p.ClosedAt = &time.Time{}
			*p.ClosedAt = time.Now()
		})
		suite.Require().NoError(err)
	}).Then("the position status should be updated", func() {
		position, err := suite.adapter.GetPosition(suite.ctx, positionID)
		suite.Require().NoError(err)
		suite.Equal("closed", position.Status)
		suite.NotNil(position.ClosedAt)
	})
}
