# SSH Access to Production Server

## Server Details
- **Host**: `ubuntu@biswas.me`
- **SprintSpark Project Path**: `/home/ubuntu/projects/sprintspark`

## Common SSH Commands

### Check Container Status
```bash
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose ps"
```

### View Logs
```bash
# API logs
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose logs api --tail 50"

# Web logs
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose logs web --tail 50"

# Follow logs in real-time
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose logs -f api"
```

### Rebuild and Restart Services
```bash
# Rebuild without cache
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose build --no-cache"

# Start services
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose up -d"

# Restart services
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose restart"

# Stop services
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose down"
```

### Check Git Status
```bash
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && git log --oneline -5"
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && git status"
```

### Database Operations
```bash
# Access SQLite database
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose exec api sqlite3 /data/sprintspark.db"

# Run a query
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose exec api sqlite3 /data/sprintspark.db 'SELECT * FROM swim_lanes LIMIT 5;'"
```

### Health Checks
```bash
# Check API health
curl -s https://sprintspark.biswas.me/api/health

# Check healthz endpoint
curl -s https://sprintspark.biswas.me/healthz

# Check web UI
curl -I https://sprintspark.biswas.me
```

## Deployment Workflow

### Full Deployment Process

Use this complete workflow to deploy changes from local to production:

1. **Build locally** (if needed)
   ```bash
   # For frontend changes
   cd web && npm run build

   # For backend changes
   cd api && go build ./cmd/api
   ```

2. **Commit and push changes**
   ```bash
   git add -A
   git commit -m "$(cat <<'EOF'
   <type>(<scope>): <subject>

   - Detailed change 1
   - Detailed change 2

   🤖 Generated with [Claude Code](https://claude.com/claude-code)

   Co-Authored-By: Claude <noreply@anthropic.com>
   EOF
   )"
   git push origin main
   ```

3. **Pull latest code on production**
   ```bash
   ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && git pull origin main"
   ```

4. **Rebuild affected containers**
   ```bash
   # For API changes
   ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose build api"

   # For web changes
   ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose build web"

   # For both or unclear changes
   ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose build --no-cache"
   ```

5. **Restart services**
   ```bash
   ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose up -d"
   ```

6. **Verify deployment**
   ```bash
   # Check container health
   ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose ps"

   # Check API logs
   ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose logs api --tail 30"

   # Test health endpoints
   curl -s https://sprintspark.biswas.me/api/health
   curl -I https://sprintspark.biswas.me
   ```

### Quick Deploy (when webhook fails)

When webhook deployment fails, use this streamlined process:

```bash
# 1. Check status
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose ps"

# 2. Pull, rebuild, and restart
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && git pull origin main && docker compose build --no-cache && docker compose up -d"

# 3. Verify
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose ps"
curl -s https://sprintspark.biswas.me/api/health
```

## Troubleshooting

### Containers not running
```bash
# Check logs for errors
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose logs --tail 100"

# Try rebuilding
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose down && docker compose build --no-cache && docker compose up -d"
```

### Migrations not applied
- Migrations run automatically on API startup
- Check logs to confirm: `docker compose logs api | grep -i migration`
- Migrations are located in `/app/internal/db/migrations` inside the container

### 502 Bad Gateway
- Usually means containers aren't running
- Check with: `docker compose ps`
- Restart with: `docker compose restart`

## Webhook Information

The webhook server runs on the production server and automatically:
1. Pulls latest code from GitHub when commits are pushed to `main`
2. Rebuilds Docker containers
3. Restarts services

**Webhook health**: https://webhook.biswas.me/health

If webhook deployment fails, use the manual deployment workflow above.
