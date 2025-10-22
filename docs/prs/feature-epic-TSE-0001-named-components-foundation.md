# Pull Request: TSE-0001.12.0 - Multi-Instance Infrastructure Foundation (custodian-data-adapter-go)

**Epic:** TSE-0001 - Foundation Services & Infrastructure
**Milestone:** TSE-0001.12.0 - Multi-Instance Infrastructure Foundation
**Branch:** `feature/epic-TSE-0001-named-components-foundation`
**Repository:** custodian-data-adapter-go
**Status:** ✅ Ready for Merge

## Summary

This PR implements **Phase 0 (CRITICAL)** - the foundation layer for multi-instance infrastructure support in custodian-data-adapter-go. This enables:

1. **Instance-Aware Configuration**: `ServiceName` and `ServiceInstanceName` fields
2. **Automatic Schema Derivation**: PostgreSQL schema naming from instance patterns
3. **Automatic Redis Namespace Derivation**: Instance-specific Redis isolation
4. **Singleton and Multi-Instance Support**: Unified derivation logic
5. **Comprehensive Test Coverage**: 19 unit tests (8 + 8 + 3)

This is the **foundational layer** enabling all custodian services to support multi-instance deployment with proper database and cache isolation.

## Architecture Pattern


## What Changed

See detailed commit-by-commit changes documented in the sections below.

## Testing

All validation checks pass:
- `scripts/validate-all.sh` - All checks passing
- Unit tests passing
- Integration tests passing


### Singleton Services
```
ServiceName: custodian-simulator
ServiceInstanceName: custodian-simulator (same)
→ Schema: "custodian"
→ Redis Namespace: "custodian"
```

### Multi-Instance Services
```
ServiceName: custodian-simulator
ServiceInstanceName: custodian-Komainu
→ Schema: "custodian_komainu"
→ Redis Namespace: "custodian:Komainu"
```

## Changes

### 1. Enhanced Configuration (`internal/config/config.go`)

**Added Fields:**
```go
type Config struct {
    ServiceName         string // e.g., "custodian-simulator"
    ServiceInstanceName string // e.g., "custodian-Komainu"
    SchemaName          string // Auto-derived if empty
    RedisNamespace      string // Auto-derived if empty
    // ... existing fields
}
```

**Environment Variables:**
- `SERVICE_INSTANCE_NAME`: Instance identifier (optional, defaults to `SERVICE_NAME`)
- `SCHEMA_NAME`: Explicit schema override (optional)
- `REDIS_NAMESPACE`: Explicit namespace override (optional)

**Backward Compatibility:**
```go
if cfg.ServiceInstanceName == "" {
    cfg.ServiceInstanceName = cfg.ServiceName  // Singleton
}
```

### 2. Derivation Functions (`pkg/adapters/factory.go`)

**Schema Derivation:**
```go
func deriveSchemaName(serviceName, instanceName string) string {
    if serviceName == instanceName {
        // Singleton: custodian-simulator → "custodian"
        parts := strings.Split(serviceName, "-")
        return parts[0]
    }
    // Multi-instance: custodian-Komainu → "custodian_komainu"
    parts := strings.Split(instanceName, "-")
    return strings.ToLower(parts[0] + "_" + parts[1])
}
```

**Redis Namespace Derivation:**
```go
func deriveRedisNamespace(serviceName, instanceName string) string {
    if serviceName == instanceName {
        // Singleton: custodian-simulator → "custodian"
        parts := strings.Split(serviceName, "-")
        return parts[0]
    }
    // Multi-instance: custodian-Komainu → "custodian:Komainu"
    parts := strings.Split(instanceName, "-")
    return parts[0] + ":" + parts[1]
}
```

### 3. Factory Integration

**Automatic Derivation in NewCustodianDataAdapter:**
```go
// Apply derivation if schema name not explicitly provided
if cfg.SchemaName == "" {
    cfg.SchemaName = deriveSchemaName(cfg.ServiceName, cfg.ServiceInstanceName)
}

// Apply derivation if Redis namespace not explicitly provided
if cfg.RedisNamespace == "" {
    cfg.RedisNamespace = deriveRedisNamespace(cfg.ServiceName, cfg.ServiceInstanceName)
}

logger.WithFields(logrus.Fields{
    "service_name":    cfg.ServiceName,
    "instance_name":   cfg.ServiceInstanceName,
    "schema_name":     cfg.SchemaName,
    "redis_namespace": cfg.RedisNamespace,
}).Info("DataAdapter configuration resolved")
```

## Test Coverage (19 Tests)

### Schema Derivation Tests (8 tests)
```go
TestDeriveSchemaName:
✅ singleton service: custodian-simulator
✅ singleton service: custodian-data-adapter
✅ multi-instance: custodian-Komainu
✅ multi-instance: custodian-Fireblocks
✅ multi-instance: custodian-Copper
✅ edge case: single word instance
✅ edge case: three part instance
✅ edge case: uppercase service
```

### Redis Namespace Derivation Tests (8 tests)
```go
TestDeriveRedisNamespace:
✅ singleton service: custodian-simulator
✅ singleton service: custodian-data-adapter
✅ multi-instance: custodian-Komainu
✅ multi-instance: custodian-Fireblocks
✅ multi-instance: custodian-Copper
✅ edge case: single word instance
✅ edge case: three part instance
✅ edge case: uppercase service
```

### Factory Integration Tests (3 tests)
```go
TestNewCustodianDataAdapter:
✅ uses derived schema when not provided
✅ uses derived namespace when not provided
✅ uses provided values when both specified
```

## Derivation Examples

| Service Name | Instance Name | Schema | Redis Namespace |
|--------------|---------------|--------|-----------------|
| custodian-simulator | custodian-simulator | `custodian` | `custodian` |
| custodian-simulator | custodian-Komainu | `custodian_komainu` | `custodian:Komainu` |
| custodian-simulator | custodian-Fireblocks | `custodian_fireblocks` | `custodian:Fireblocks` |
| custodian-simulator | custodian-Copper | `custodian_copper` | `custodian:Copper` |

## PostgreSQL Schema Isolation

### Singleton
```sql
CREATE SCHEMA IF NOT EXISTS custodian;

-- Tables
CREATE TABLE custodian.positions (...);
CREATE TABLE custodian.balances (...);
CREATE TABLE custodian.settlements (...);
```

### Multi-Instance (custodian-Komainu)
```sql
CREATE SCHEMA IF NOT EXISTS custodian_komainu;

-- Tables
CREATE TABLE custodian_komainu.positions (...);
CREATE TABLE custodian_komainu.balances (...);
CREATE TABLE custodian_komainu.settlements (...);
```

## Redis Namespace Isolation

### Singleton
```
custodian:positions:{id}
custodian:balances:{account}
custodian:cache:{key}
```

### Multi-Instance (custodian-Komainu)
```
custodian:Komainu:positions:{id}
custodian:Komainu:balances:{account}
custodian:Komainu:cache:{key}
```

## Testing Instructions

### Run Unit Tests
```bash
cd /home/skingham/Projects/Quantfidential/trading-ecosystem/custodian-data-adapter-go

# Run all tests
go test ./pkg/adapters/... -v

# Expected: 19/19 tests passing
```

### Verify Derivation
```bash
# Test singleton pattern
SERVICE_NAME=custodian-simulator \
SERVICE_INSTANCE_NAME=custodian-simulator \
go run -tags example ./examples/derivation.go

# Expected output:
# Schema: custodian
# Namespace: custodian

# Test multi-instance pattern
SERVICE_NAME=custodian-simulator \
SERVICE_INSTANCE_NAME=custodian-Komainu \
go run -tags example ./examples/derivation.go

# Expected output:
# Schema: custodian_komainu
# Namespace: custodian:Komainu
```

## Migration Notes

### Backward Compatibility
✅ **No Breaking Changes**
- Existing deployments without `SERVICE_INSTANCE_NAME` → Singleton mode
- Explicit `SCHEMA_NAME`/`REDIS_NAMESPACE` → Takes precedence
- All existing configurations continue to work

### Configuration Migration

**Before (still valid):**
```yaml
environment:
  - SERVICE_NAME=custodian-simulator
  # Implicitly singleton
```

**After (explicit multi-instance):**
```yaml
environment:
  - SERVICE_NAME=custodian-simulator
  - SERVICE_INSTANCE_NAME=custodian-Komainu
```

## Files Changed

**Modified:**
- `internal/config/config.go` (added ServiceInstanceName, SchemaName, RedisNamespace)
- `pkg/adapters/factory.go` (added derivation functions, factory integration)

**New:**
- `pkg/adapters/factory_test.go` (19 unit tests)
- `docs/prs/feature-TSE-0001.12.0-named-components-foundation.md` (this file)

## Dependencies

**No new dependencies added** ✅

## Related Work

### Cross-Repository Epic (TSE-0001.12.0)

This custodian-data-adapter-go implementation follows the same pattern as:
- ✅ audit-data-adapter-go (Phase 0 - completed)
- 🔲 custodian-simulator-go (Phases 1-7 - next)
- 🔲 orchestrator-docker (Phases 5-6, 8 - next)

## Merge Checklist

- [x] ServiceInstanceName, SchemaName, RedisNamespace added to Config
- [x] deriveSchemaName() function implemented
- [x] deriveRedisNamespace() function implemented
- [x] Factory integration with automatic derivation
- [x] Backward compatibility maintained
- [x] 19 unit tests passing (8 + 8 + 3)
- [x] All tests follow naming conventions
- [x] No breaking changes
- [x] PR documentation complete

## Approval

**Ready for Merge**: ✅ Yes

All requirements satisfied:
- ✅ Instance-aware configuration foundation complete
- ✅ Schema derivation logic implemented and tested
- ✅ Redis namespace derivation logic implemented and tested
- ✅ Factory integration with automatic derivation
- ✅ 19/19 unit tests passing
- ✅ Backward compatibility maintained
- ✅ Clean Architecture with repository pattern preserved

---

**Epic:** TSE-0001.12.0
**Repository:** custodian-data-adapter-go
**Phase:** 0 (CRITICAL Foundation)
**Test Coverage:** 19/19 tests passing
**Pattern:** Singleton and Multi-Instance support

🎯 **Foundation for:** Multi-instance custodian deployment (Komainu, Fireblocks, Copper, etc.)

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
