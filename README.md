# custodian-data-adapter-go

Clean Architecture data adapter for custodian domain operations in the Trading Ecosystem.

## Overview

The custodian-data-adapter-go provides a repository pattern interface for managing custodian operations including positions, settlements, and balance tracking.

## Architecture

```
custodian-data-adapter-go/
├── pkg/interfaces/        # Repository interfaces
├── pkg/adapters/          # PostgreSQL and Redis implementations
├── pkg/models/            # Domain models
├── internal/config/       # Configuration
├── internal/database/     # PostgreSQL connection
└── internal/cache/        # Redis connection
```

## Features

- Position Management with locking
- Settlement Processing (DEPOSIT, WITHDRAWAL, TRANSFER)
- Balance Tracking with atomic updates
- Service Discovery (Redis)
- Caching with TTL
- Connection Pooling
- Graceful Degradation

## Installation

```bash
cp .env.example .env
go mod download
```

## Testing

```bash
make test-all
make test-coverage
```

## Status

✅ READY FOR INTEGRATION
