package adapters

import (
	"io"
	"testing"

	"github.com/quantfidential/trading-ecosystem/custodian-data-adapter-go/internal/config"
	"github.com/sirupsen/logrus"
)

// TestDeriveSchemaName tests PostgreSQL schema name derivation
func TestDeriveSchemaName(t *testing.T) {
	tests := []struct {
		name         string
		serviceName  string
		instanceName string
		expected     string
	}{
		// Singleton service tests
		{
			name:         "singleton service: custodian-simulator",
			serviceName:  "custodian-simulator",
			instanceName: "custodian-simulator",
			expected:     "custodian",
		},
		{
			name:         "singleton service: custodian-data-adapter",
			serviceName:  "custodian-data-adapter",
			instanceName: "custodian-data-adapter",
			expected:     "custodian",
		},

		// Multi-instance service tests
		{
			name:         "multi-instance: custodian-Komainu",
			serviceName:  "custodian-simulator",
			instanceName: "custodian-Komainu",
			expected:     "custodian_komainu",
		},
		{
			name:         "multi-instance: custodian-Fireblocks",
			serviceName:  "custodian-simulator",
			instanceName: "custodian-Fireblocks",
			expected:     "custodian_fireblocks",
		},
		{
			name:         "multi-instance: custodian-Copper",
			serviceName:  "custodian-simulator",
			instanceName: "custodian-Copper",
			expected:     "custodian_copper",
		},

		// Edge cases
		{
			name:         "edge case: single word instance",
			serviceName:  "custodian-simulator",
			instanceName: "custodian",
			expected:     "custodian",
		},
		{
			name:         "edge case: three part instance",
			serviceName:  "custodian-simulator",
			instanceName: "custodian-Komainu-Primary",
			expected:     "custodian_komainu",
		},
		{
			name:         "edge case: uppercase service",
			serviceName:  "CUSTODIAN-SIMULATOR",
			instanceName: "CUSTODIAN-SIMULATOR",
			expected:     "CUSTODIAN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deriveSchemaName(tt.serviceName, tt.instanceName)
			if result != tt.expected {
				t.Errorf("deriveSchemaName(%s, %s) = %s, expected %s",
					tt.serviceName, tt.instanceName, result, tt.expected)
			}
		})
	}
}

// TestDeriveRedisNamespace tests Redis namespace derivation
func TestDeriveRedisNamespace(t *testing.T) {
	tests := []struct {
		name         string
		serviceName  string
		instanceName string
		expected     string
	}{
		// Singleton service tests
		{
			name:         "singleton service: custodian-simulator",
			serviceName:  "custodian-simulator",
			instanceName: "custodian-simulator",
			expected:     "custodian",
		},
		{
			name:         "singleton service: custodian-data-adapter",
			serviceName:  "custodian-data-adapter",
			instanceName: "custodian-data-adapter",
			expected:     "custodian",
		},

		// Multi-instance service tests
		{
			name:         "multi-instance: custodian-Komainu",
			serviceName:  "custodian-simulator",
			instanceName: "custodian-Komainu",
			expected:     "custodian:Komainu",
		},
		{
			name:         "multi-instance: custodian-Fireblocks",
			serviceName:  "custodian-simulator",
			instanceName: "custodian-Fireblocks",
			expected:     "custodian:Fireblocks",
		},
		{
			name:         "multi-instance: custodian-Copper",
			serviceName:  "custodian-simulator",
			instanceName: "custodian-Copper",
			expected:     "custodian:Copper",
		},

		// Edge cases
		{
			name:         "edge case: single word instance",
			serviceName:  "custodian-simulator",
			instanceName: "custodian",
			expected:     "custodian",
		},
		{
			name:         "edge case: three part instance",
			serviceName:  "custodian-simulator",
			instanceName: "custodian-Komainu-Primary",
			expected:     "custodian:Komainu",
		},
		{
			name:         "edge case: uppercase service",
			serviceName:  "CUSTODIAN-SIMULATOR",
			instanceName: "CUSTODIAN-SIMULATOR",
			expected:     "CUSTODIAN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deriveRedisNamespace(tt.serviceName, tt.instanceName)
			if result != tt.expected {
				t.Errorf("deriveRedisNamespace(%s, %s) = %s, expected %s",
					tt.serviceName, tt.instanceName, result, tt.expected)
			}
		})
	}
}

// TestNewCustodianDataAdapter tests the factory with derivation
func TestNewCustodianDataAdapter(t *testing.T) {
	tests := []struct {
		name              string
		config            *config.Config
		expectedSchema    string
		expectedNamespace string
	}{
		{
			name: "uses derived schema when not provided",
			config: &config.Config{
				ServiceName:         "custodian-simulator",
				ServiceInstanceName: "custodian-Komainu",
				// SchemaName empty - should be derived
				RedisNamespace: "custodian:Komainu",
			},
			expectedSchema:    "custodian_komainu",
			expectedNamespace: "custodian:Komainu",
		},
		{
			name: "uses derived namespace when not provided",
			config: &config.Config{
				ServiceName:         "custodian-simulator",
				ServiceInstanceName: "custodian-Fireblocks",
				SchemaName:          "custodian_fireblocks",
				// RedisNamespace empty - should be derived
			},
			expectedSchema:    "custodian_fireblocks",
			expectedNamespace: "custodian:Fireblocks",
		},
		{
			name: "uses provided values when both specified",
			config: &config.Config{
				ServiceName:         "custom-service",
				ServiceInstanceName: "custom-instance",
				SchemaName:          "explicit_schema",
				RedisNamespace:      "explicit:namespace",
			},
			expectedSchema:    "explicit_schema",
			expectedNamespace: "explicit:namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetOutput(io.Discard) // Suppress log output in tests

			adapter, err := NewCustodianDataAdapter(tt.config, logger)

			if err != nil {
				t.Fatalf("NewCustodianDataAdapter failed: %v", err)
			}

			// Verify schema name was applied
			if tt.config.SchemaName != tt.expectedSchema {
				t.Errorf("Schema = %s, expected %s",
					tt.config.SchemaName, tt.expectedSchema)
			}

			// Verify Redis namespace was applied
			if tt.config.RedisNamespace != tt.expectedNamespace {
				t.Errorf("Namespace = %s, expected %s",
					tt.config.RedisNamespace, tt.expectedNamespace)
			}

			_ = adapter // Suppress unused variable warning
		})
	}
}
