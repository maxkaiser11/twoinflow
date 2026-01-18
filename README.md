#Workshop Booking System

## Setup

1. Install Go 1.21 or higher
2. Clone the repository
3. Copy `.env.example` to `.env` and configure
4. Run setup: `cd setup && go run setup.go && cd ..`
5. Start server: `go run .`

## Environment Variables

- `GIN_MODE=release` - Set to release for production
- `PORT=8080` - Port to run server on
- `DEFAULT_ADMIN_USERNAME=admin` - Initial admin username
- `DEFAULT_ADMIN_PASSWORD=changeme` - Initial admin password (change immediately!)

## Deployment

The app is ready to deploy to platforms like:
- Fly.io
- Railway
- Render
- DigitalOcean App Platform

Make sure to set environment variables in your hosting platform.
```

## 7. Create `.env.example`

Create `.env.example` (this CAN be committed to git):
```
# Server Configuration
GIN_MODE=release
PORT=8080
