# Production Management Commands

Quick reference for managing SprintSpark production environment.

## Health & Status

```bash
# Full health check (API, Web, Database)
./script/server health

# Service status
./script/server status

# Quick API health
curl https://sprintspark.biswas.me/api/health
```

## Service Control

```bash
# Restart all services
./script/server restart

# Stop services
./script/server stop

# Start services
./script/server start
```

## Logs

```bash
# View API logs (last 50 lines)
./script/server logs api

# View web logs
./script/server logs web

# View all logs
./script/server logs

# Real-time logs
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker-compose logs -f api"
```

## Database Operations

```bash
# Query users
./script/server db-query "SELECT id, email, is_admin, created_at FROM users ORDER BY created_at DESC LIMIT 10;"

# Count users
./script/server db-query "SELECT COUNT(*) as total_users FROM users;"

# View API keys
./script/server db-query "SELECT id, user_id, name, key_prefix, created_at, last_used_at FROM api_keys ORDER BY created_at DESC;"

# Check projects
./script/server db-query "SELECT id, name, owner_id, created_at FROM projects ORDER BY created_at DESC LIMIT 10;"

# View tasks
./script/server db-query "SELECT id, title, status, project_id FROM tasks ORDER BY created_at DESC LIMIT 10;"
```

## Admin Management

```bash
# List all admins
./script/server admin list

# Make user admin
./script/server admin create user@example.com

# Revoke admin
./script/server admin revoke user@example.com
```

## Quick Troubleshooting

### API Not Responding

```bash
# Check logs
./script/server logs api

# Check container status
./script/server status

# Restart
./script/server restart

# Check migration status
./script/server db-query "SELECT version FROM schema_migrations ORDER BY version DESC;"
```

### Database Issues

```bash
# Check database file
ssh ubuntu@biswas.me "ls -lh /home/ubuntu/projects/sprintspark/data/sprintspark.db"

# Check tables
./script/server db-query ".tables"

# Check schema
./script/server db-query ".schema users"
```

### Container Issues

```bash
# SSH into production
ssh ubuntu@biswas.me

# Navigate to project
cd /home/ubuntu/projects/sprintspark

# Check containers
docker-compose ps

# View logs
docker-compose logs --tail=100 api
docker-compose logs --tail=100 web

# Rebuild
docker-compose down
docker-compose up -d --build

# Clean restart
docker-compose down -v
docker-compose up -d --build
```

## Performance Monitoring

```bash
# Check response time
time curl https://sprintspark.biswas.me/api/health

# Database size
./script/server db-query "SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size();"

# User activity
./script/server db-query "SELECT COUNT(*) as active_users FROM users WHERE created_at > datetime('now', '-7 days');"

# Recent activity
./script/server db-query "SELECT action, COUNT(*) as count FROM user_activity WHERE created_at > datetime('now', '-24 hours') GROUP BY action ORDER BY count DESC;"
```

## Production URLs

- **Web UI**: https://sprintspark.biswas.me
- **API Health**: https://sprintspark.biswas.me/api/health
- **API Docs**: https://sprintspark.biswas.me/api/openapi
- **Direct API**: https://sprintspark.biswas.me/api

## Emergency Procedures

### Complete Service Restart

```bash
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker-compose restart"
```

### Force Rebuild

```bash
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker-compose down && docker-compose build --no-cache && docker-compose up -d"
```

### Check Disk Space

```bash
ssh ubuntu@biswas.me "df -h"
```

### Clean Docker

```bash
ssh ubuntu@biswas.me "docker system prune -a --volumes -f"
```

## Security

```bash
# Check failed login attempts
./script/server db-query "SELECT action, user_id, created_at FROM user_activity WHERE action LIKE '%failed%' ORDER BY created_at DESC LIMIT 20;"

# Active API keys
./script/server db-query "SELECT COUNT(*) as active_keys FROM api_keys WHERE expires_at IS NULL OR expires_at > datetime('now');"

# Recently used API keys
./script/server db-query "SELECT name, key_prefix, last_used_at FROM api_keys WHERE last_used_at > datetime('now', '-24 hours') ORDER BY last_used_at DESC;"
```

## Server Details

- **Server**: ubuntu@biswas.me
- **Deploy Path**: /home/ubuntu/projects/sprintspark
- **Database**: /home/ubuntu/projects/sprintspark/data/sprintspark.db
- **Docker Compose**: /home/ubuntu/projects/sprintspark/docker-compose.yml
- **Containers**: sprintspark_api_1, sprintspark_web_1
