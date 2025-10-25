.PHONY: help dev dev-all lint test build run clean docker-build docker-run cleanup-commands cleanup-commands-list cleanup-commands-delete

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Run the application in development mode with live reload
	@echo "Starting development server..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	@air

dev-all: ## Run both bot and web server simultaneously
	@echo "Starting bot and web server..."
	@echo "Bot will run on port 8080, web dev server will run on port 5173"
	@trap 'kill %1; kill %2' INT; \
		cd web && bun run dev & \
		which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest); \
		air

lint: ## Run linter
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

build: ## Build the application
	@echo "Building application..."
	@go build -o bin/leetbot ./cmd/bot

build-server: ## Build the HTTP server
	@echo "Building HTTP server..."
	@go build -o bin/server ./cmd/server

build-web: ## Build the React frontend
	@echo "Building React frontend..."
	@cd web && bun run build

build-all: build-server build-web ## Build both server and frontend

run: build ## Build and run the application
	@echo "Running application..."
	@./bin/leetbot

run-server: build-server ## Build and run the HTTP server
	@echo "Running HTTP server..."
	@./bin/server

run-web: build-web ## Build and run the web server
	@echo "Building and running web server..."
	@cd web && bun run dev

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t leetbot .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 \
		--env-file .env \
		-e GOOGLE_APPLICATION_CREDENTIALS=/app/service-account.json \
		-v $(PWD)/service-account.json:/app/service-account.json \
		leetbot

validate: lint test ## Run linting and tests
	@echo "Validation complete!"

setup: ## Setup development environment
	@echo "Setting up development environment..."
	@go mod tidy
	@go mod download
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Setup complete!"

generate-embedded: ## Generate embedded CSV data from actual files
	@echo "Generating embedded data..."
	@go run scripts/generate_embedded/main.go data
	@echo "Copy the generated content to internal/data/parser.go"

validate-data: ## Validate all CSV files in data directory
	@echo "Validating CSV data..."
	@go run scripts/validate_data/main.go data

demo: ## Run the bot demo
	@echo "Running bot demo..."
	@go run scripts/demo_bot/main.go

cleanup-commands: ## List or delete Discord slash commands
	@echo "Discord Command Cleanup Tool"
	@echo "Usage: make cleanup-commands-list or make cleanup-commands-delete command=shutdown"
	@if [ "$(command)" = "" ]; then \
		go run scripts/cleanup_commands/main.go -list; \
	else \
		go run scripts/cleanup_commands/main.go -command $(command); \
	fi

cleanup-commands-list: ## List all registered Discord slash commands
	@echo "Listing registered Discord commands..."
	@go run scripts/cleanup_commands/main.go -list

cleanup-commands-delete: ## Delete a specific Discord slash command (requires command= parameter)
	@if [ "$(command)" = "" ]; then \
		echo "Error: Please specify command=shutdown (or other command name)"; \
		echo "Usage: make cleanup-commands-delete command=shutdown"; \
		exit 1; \
	fi
	@echo "Deleting Discord command: $(command)"
	@go run scripts/cleanup_commands/main.go -command $(command)
