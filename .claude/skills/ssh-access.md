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

When webhook deployment fails:

1. **SSH to server and check status**
   ```bash
   ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose ps"
   ```

2. **Pull latest code** (if needed)
   ```bash
   ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && git pull origin main"
   ```

3. **Rebuild containers**
   ```bash
   ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose build --no-cache"
   ```

4. **Start services**
   ```bash
   ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose up -d"
   ```

5. **Verify deployment**
   ```bash
   ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker compose logs api --tail 30"
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
