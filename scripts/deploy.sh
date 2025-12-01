#!/bin/bash
# =============================================================================
# DEPLOYMENT SCRIPT - Matrix MUD
# =============================================================================
# Usage: ./scripts/deploy.sh [environment]
#   environment: local | fly | docker (default: local)
# =============================================================================

set -e  # Exit on error

VERSION="1.31.0"
APP_NAME="matrix-mud"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

ENVIRONMENT=${1:-local}

# =============================================================================
# PRE-FLIGHT CHECKS
# =============================================================================
preflight() {
    log_info "Running pre-flight checks..."
    
    # Check Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi
    
    # Run tests
    log_info "Running tests..."
    if ! go test ./... > /dev/null 2>&1; then
        log_error "Tests failed! Fix tests before deploying."
        go test ./...
        exit 1
    fi
    log_success "All tests passing"
    
    # Build check
    log_info "Verifying build..."
    if ! go build -o /tmp/${APP_NAME}-check . 2>&1; then
        log_error "Build failed!"
        exit 1
    fi
    rm -f /tmp/${APP_NAME}-check
    log_success "Build verified"
}

# =============================================================================
# LOCAL DEPLOYMENT
# =============================================================================
deploy_local() {
    log_info "Deploying locally..."
    
    # Build
    go build -o ./bin/${APP_NAME} .
    
    log_success "Built: ./bin/${APP_NAME}"
    log_info "Run with: ./bin/${APP_NAME}"
    log_info "Or use: make run"
    
    echo ""
    log_info "Services will be available at:"
    echo "  Telnet: telnet localhost 2323"
    echo "  Web:    http://localhost:8080"
    echo "  Admin:  http://localhost:9090 (localhost only)"
}

# =============================================================================
# DOCKER DEPLOYMENT
# =============================================================================
deploy_docker() {
    log_info "Deploying with Docker..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    # Build image
    log_info "Building Docker image..."
    docker build -t ${APP_NAME}:${VERSION} -t ${APP_NAME}:latest .
    
    # Create volume if not exists
    docker volume create ${APP_NAME}_data 2>/dev/null || true
    
    # Stop existing container
    docker stop ${APP_NAME} 2>/dev/null || true
    docker rm ${APP_NAME} 2>/dev/null || true
    
    # Run new container
    log_info "Starting container..."
    docker run -d \
        --name ${APP_NAME} \
        --restart unless-stopped \
        -p 2323:2323 \
        -p 8080:8080 \
        -v ${APP_NAME}_data:/app/data \
        -e ADMIN_PASS="${ADMIN_PASS:-}" \
        -e ALLOWED_ORIGINS="${ALLOWED_ORIGINS:-*}" \
        ${APP_NAME}:latest
    
    log_success "Container started: ${APP_NAME}"
    log_info "View logs: docker logs -f ${APP_NAME}"
    
    echo ""
    log_info "Services available at:"
    echo "  Telnet: telnet localhost 2323"
    echo "  Web:    http://localhost:8080"
    echo "  Health: http://localhost:8080/health"
}

# =============================================================================
# FLY.IO DEPLOYMENT
# =============================================================================
deploy_fly() {
    log_info "Deploying to Fly.io..."
    
    if ! command -v fly &> /dev/null; then
        log_error "Fly CLI is not installed"
        log_info "Install with: brew install flyctl"
        exit 1
    fi
    
    # Check if logged in
    if ! fly auth whoami &> /dev/null; then
        log_warn "Not logged in to Fly.io"
        log_info "Run: fly auth login"
        exit 1
    fi
    
    # Check if app exists
    if ! fly apps list | grep -q ${APP_NAME}; then
        log_info "Creating Fly.io app..."
        fly apps create ${APP_NAME}
        
        # Create volume
        log_info "Creating persistent volume..."
        fly volumes create matrix_mud_data --size 1 --region ewr
        
        # Set secrets
        log_warn "Set admin password with: fly secrets set ADMIN_PASS=your-secure-password"
    fi
    
    # Deploy
    log_info "Deploying to Fly.io..."
    fly deploy
    
    log_success "Deployed to Fly.io!"
    
    echo ""
    log_info "Services available at:"
    echo "  Web:    https://${APP_NAME}.fly.dev"
    echo "  Telnet: ${APP_NAME}.fly.dev:2323"
    echo "  Health: https://${APP_NAME}.fly.dev/health"
    echo ""
    log_info "View logs: fly logs"
    log_info "SSH access: fly ssh console"
}

# =============================================================================
# MAIN
# =============================================================================
echo "================================================"
echo "  Matrix MUD Deployment - v${VERSION}"
echo "================================================"
echo ""

preflight

case $ENVIRONMENT in
    local)
        deploy_local
        ;;
    docker)
        deploy_docker
        ;;
    fly)
        deploy_fly
        ;;
    *)
        log_error "Unknown environment: $ENVIRONMENT"
        echo "Usage: $0 [local|docker|fly]"
        exit 1
        ;;
esac

echo ""
log_success "Deployment complete!"
