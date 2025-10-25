![Leetbot Frontend](./public/ui.webp)

# LeetBot Web

The frontend interface for the leetbot project.

### Prerequisites

- Go 1.24+
- Bun (for frontend development)
- Node.js 20+ (if not using Bun)

### Running the Web Server

1. **Build and start everything:**
   ```bash
   ./start-web.sh
   ```

2. **Or manually:**
   ```bash
   # Build the server
   make build-server

   # Build the frontend
   make build-web

   # Start the server
   make run-server
   ```

3. **Open your browser:**
   ```
   http://localhost:8080
   ```

### Development Mode

For frontend development with hot reload:

```bash
# Terminal 1: Start the Go server
make run-server

# Terminal 2: Start the React dev server
cd web && bun run dev
```

## API Endpoints

The server exposes the following REST API:

- `GET /api/companies` - Get list of available companies
- `GET /api/companies/{company}/timeframes` - Get available timeframes for a company
- `GET /api/companies/{company}/problems` - Get problems for a company (uses priority timeframe)
- `GET /api/companies/{company}/timeframes/{timeframe}/problems` - Get problems for specific company and timeframe

> [!NOTE]
> The interface supports all companies from the leetbot data!

## Available Timeframes

- **All Time**: Complete historical data
- **Last 30 Days**: Recent problems (30d)
- **Last 3 Months**: Recent problems (3mo)
- **Last 6 Months**: Recent problems (6mo)
- **More than 6 Months**: Older problems (>6mo)

## Building for Production

```bash
# Build everything
make build-all

# The built files will be in:
# - bin/server (Go binary)
# - web/dist/ (React build)
```

## Deployment

The web interface is designed to be deployed as a single Go binary that serves both the API and the React frontend:

1. Build the server: `make build-server`
2. Build the frontend: `make build-web`
3. Deploy the `bin/server` binary
4. The binary will serve the API and static files

Same as the main leetbot project (GNU General Public License v3.0).
