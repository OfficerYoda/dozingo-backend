# Load .env file so just has access to env-variables
set dotenv-load

# Default: list available commands
default:
    @just --list

# === Setup ===

# Install all project tools via mise
tools:
    mise install

# First-time project setup (run once after cloning)
setup:
    @echo "Installing tools via mise..."
    just tools
    @echo "Copying .env.example → .env (if .env doesn't exist)..."
    cp -n .env.example .env || true
    @echo "Starting infrastructure..."
    just infra-up
    @echo "Waiting for postgres to be ready..."
    @sleep 3
    @echo "Running migrations..."
    just migrate-up
    @echo "Generating code..."
    just generate
    @echo "Installing Go dependencies..."
    go mod tidy
    @echo ""
    @echo "Setup complete. Run 'just run' to start the server."

# === Run ===

# Start the Go server
run:
    go run ./cmd/api

# === Infrastructure ===

# Start postgres
infra-up:
    docker compose up -d

# Stop postgres
infra-down:
    docker compose down

# Wipe database and restart from scratch
infra-reset:
    docker compose down -v
    docker compose up -d
    @sleep 3
    just migrate-up

# === Database Migrations ===

# Apply all pending migrations
migrate-up:
    migrate -path internal/db/migrations -database "$DATABASE_URL" up

# Roll back the last migration
migrate-down:
    migrate -path internal/db/migrations -database "$DATABASE_URL" down 1

# Create a new migration (usage: just migrate-create name_of_migration)
migrate-create name:
    migrate create -ext sql -dir internal/db/migrations -seq {{name}}

# === Code Generation ===

# Regenerate Go code from SQL queries
generate:
    sqlc generate

# === Quality ===

# Run all tests
test:
    go test ./...

# Run linter
lint:
    golangci-lint run

# === Helpers ===

# Open a psql shell to the local database
db-shell:
    docker exec -it dozingo-postgres psql -U dozingo_user -d dozingo_db
