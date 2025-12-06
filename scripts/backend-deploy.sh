#!/bin/bash
set -e

# ────── Config ────────────────────────────────────────────────
# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Try to find the SSH key in common locations
KEY_PATH=""
for key_location in \
    "$HOME/Keys/pupo_laptop.pem" \
    "$HOME/.ssh/pupo_laptop.pem" \
    "$HOME/pupo_laptop.pem" \
    "$SCRIPT_DIR/pupo_laptop.pem"
do
    if [[ -f "$key_location" ]]; then
        KEY_PATH="$key_location"
        echo "Found SSH key at: $KEY_PATH"
        break
    fi
done

if [[ -z "$KEY_PATH" ]]; then
    echo "Error: SSH key not found. Please place your .pem key in one of these locations:"
    echo "  - $HOME/Keys/pupo_laptop.pem"
    echo "  - $HOME/.ssh/pupo_laptop.pem"
    echo "  - $HOME/pupo_laptop.pem"
    echo "  - $SCRIPT_DIR/pupo_laptop.pem"
    exit 1
fi

# Find the project directory with enhanced detection
PROJECT_DIR=""
if [[ "$SCRIPT_DIR" == "/usr/local/bin" ]] || [[ "$SCRIPT_DIR" == "/usr/bin" ]]; then
    # When running from system bin directories, search in common project locations
    for project_path in \
        "$HOME/projects/ProNexus/backend" \
        "$HOME/ProNexus/backend" \
        "$HOME/workspace/ProNexus/backend" \
        "$HOME/dev/ProNexus/backend" \
        "/opt/ProNexus/backend" \
        "/usr/local/ProNexus/backend" \
        "$(pwd)/backend" \
        "$(pwd)"
    do
        if [[ -f "$project_path/go.mod" ]]; then
            PROJECT_DIR="$project_path"
            echo "Found project directory: $PROJECT_DIR"
            break
        fi
    done
    
    # If still not found, try to find it relative to current working directory
    if [[ -z "$PROJECT_DIR" ]]; then
        current_dir="$(pwd)"
        while [[ "$current_dir" != "/" ]]; do
            if [[ -f "$current_dir/go.mod" ]]; then
                PROJECT_DIR="$current_dir"
                echo "Found project directory from current working directory: $PROJECT_DIR"
                break
            fi
            current_dir="$(dirname "$current_dir")"
        done
    fi
else
    # If we're in a subdirectory, find the parent directory with go.mod
    current_dir="$SCRIPT_DIR"
    while [[ "$current_dir" != "/" ]]; do
        if [[ -f "$current_dir/go.mod" ]]; then
            PROJECT_DIR="$current_dir"
            echo "Found project directory: $PROJECT_DIR"
            break
        fi
        current_dir="$(dirname "$current_dir")"
    done
    
    # If not found from script location, try from current working directory
    if [[ -z "$PROJECT_DIR" ]]; then
        current_dir="$(pwd)"
        while [[ "$current_dir" != "/" ]]; do
            if [[ -f "$current_dir/go.mod" ]]; then
                PROJECT_DIR="$current_dir"
                echo "Found project directory from current working directory: $PROJECT_DIR"
                break
            fi
            current_dir="$(dirname "$current_dir")"
        done
    fi
fi

if [[ -z "$PROJECT_DIR" ]]; then
    echo "Error: Could not find project directory with go.mod"
    echo "Current script location: $SCRIPT_DIR"
    echo "Current working directory: $(pwd)"
    echo "Please ensure you're running this script from within the ProNexus project directory"
    echo "or that the project is located in one of the common paths."
    exit 1
fi

EC2_USER=ec2-user
EC2_HOST=${EC2_HOST:-18.219.205.161}
REMOTE_BINARY_PATH=/home/ec2-user/pronexus-backend
LOCAL_BINARY_PATH="$PROJECT_DIR/pronexus-backend"
SERVICE_NAME=pronexus-backend.service

# ────── Colors ────────────────────────────────────────────────
GREEN="\e[32m"
RED="\e[31m"
YELLOW="\e[33m"
RESET="\e[0m"

log_success() { echo -e "${GREEN}[✓] $1${RESET}"; }
log_error()   { echo -e "${RED}[✗] $1${RESET}"; }
log_info()    { echo -e "${YELLOW}[→] $1${RESET}"; }

# ────── Build ────────────────────────────────────────────────
log_info "Building Go binary…"
cd "$PROJECT_DIR"
if go build -o "$(basename "$LOCAL_BINARY_PATH")"; then
  log_success "Go binary built successfully."
else
  log_error "Go build failed."
  exit 1
fi

# ────── Deploy ────────────────────────────────────────────────
log_info "Starting Go backend binary upload and restart…"

log_info "Uploading binary via rsync…"
if rsync -avz -e "ssh -i $KEY_PATH" "$LOCAL_BINARY_PATH" "$EC2_USER@$EC2_HOST:$REMOTE_BINARY_PATH"; then
  log_success "Binary uploaded successfully."
else
  log_error "Binary upload failed."
  exit 1
fi

log_info "Restarting remote systemd service…"
ssh -i "$KEY_PATH" "$EC2_USER@$EC2_HOST" bash <<EOF
  set -e
  GREEN="\e[32m"; RED="\e[31m"; YELLOW="\e[33m"; RESET="\e[0m"
  log_remote_ok() { echo -e "\$GREEN[✓] \$1\$RESET"; }
  log_remote_fail() { echo -e "\$RED[✗] \$1\$RESET"; exit 1; }
  log_info() { echo -e "\$YELLOW[→] \$1\$RESET"; }

  log_info "Stopping service..."
  sudo systemctl stop "$SERVICE_NAME" || true
  
  log_info "Starting service..."
  if sudo systemctl daemon-reload && sudo systemctl start "$SERVICE_NAME"; then
    log_remote_ok "Service restarted."
  else
    log_remote_fail "Failed to restart service."
  fi
  
  log_info "Checking service status..."
  sudo systemctl status "$SERVICE_NAME" --no-pager -l
  
  log_info "Checking binary modification time..."
  ls -la "$REMOTE_BINARY_PATH"
  
  log_info "Checking if service is using the new binary..."
  sudo systemctl show "$SERVICE_NAME" --property=ExecStart --no-pager
EOF

log_info "Waiting for service to fully start..."
sleep 5

log_info "Running health check on the deployed service..."
ssh -i "$KEY_PATH" "$EC2_USER@$EC2_HOST" bash <<'HEALTH_EOF'
  set -euo pipefail
  
  GREEN="\e[32m"; RED="\e[31m"; YELLOW="\e[33m"; RESET="\e[0m"
  log_remote_ok() { echo -e "$GREEN[✓] $1$RESET"; }
  log_remote_fail() { echo -e "$RED[✗] $1$RESET"; exit 1; }
  log_info() { echo -e "$YELLOW[→] $1$RESET"; }
  
  # Check if service is active
  log_info "Checking if service is active..."
  if ! sudo systemctl is-active pronexus-backend.service --quiet; then
    log_remote_fail "Service is not active!"
    sudo systemctl status pronexus-backend.service --no-pager -l
    exit 1
  fi
  log_remote_ok "Service is active."
  
  # Check if service is listening on port 8080
  log_info "Checking if service is listening on port 8080..."
  if ! ss -tuln | grep -q ":8080"; then
    log_remote_fail "Service is not listening on port 8080!"
    log_info "Current listening ports:"
    ss -tuln
    exit 1
  fi
  log_remote_ok "Service is listening on port 8080."
  
  # Test basic health endpoint
  log_info "Testing basic health endpoint..."
  for i in {1..30}; do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:8080/health || echo "000")
    if [ "$STATUS" = "200" ]; then
      log_remote_ok "Basic health check passed (HTTP $STATUS)"
      break
    else
      log_info "Health check attempt $i/30 - got HTTP $STATUS, retrying..."
      sleep 2
    fi
  done
  
  if [ "$STATUS" != "200" ]; then
    log_remote_fail "Basic health check failed after 30 attempts (HTTP $STATUS)"
    exit 1
  fi
  
  # Test comprehensive health endpoint
  log_info "Testing comprehensive health endpoint..."
  HEALTH_RESPONSE=$(curl -s http://127.0.0.1:8080/health/check || echo "")
  if [ -n "$HEALTH_RESPONSE" ]; then
    log_remote_ok "Comprehensive health check passed"
    log_info "Health response preview:"
    echo "$HEALTH_RESPONSE" | head -10
  else
    log_remote_fail "Comprehensive health check failed - no response"
    exit 1
  fi
  
  # Test root endpoint
  log_info "Testing root endpoint..."
  ROOT_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:8080/ || echo "000")
  if [ "$ROOT_STATUS" = "200" ]; then
    log_remote_ok "Root endpoint check passed (HTTP $ROOT_STATUS)"
  else
    log_remote_fail "Root endpoint check failed (HTTP $ROOT_STATUS)"
    exit 1
  fi
  
  log_remote_ok "All health checks passed!"
HEALTH_EOF

log_success "Deployment completed and health checks passed!"