# Deploy to Production

Deploy SprintSpark to production server with automated verification.

## Usage

```bash
./script/server deploy [commit_message]
```

## What it does

1. **Commits changes** - Adds all uncommitted files and creates a commit
2. **Pushes to GitHub** - Pushes the main branch to origin
3. **Waits for webhook** - Gives GitHub webhook time to trigger (30s)
4. **Verifies rebuild** - Checks if containers are running, manually rebuilds if needed
5. **Health checks** - Verifies API, Web UI, and database are responding
6. **Reports status** - Shows deployment success with URLs or error details

## Examples

```bash
# Deploy with auto-generated message
./script/server deploy

# Deploy with custom message
./script/server deploy "feat: add new feature"

# Deploy bug fix
./script/server deploy "fix: resolve login issue"
```

## Default Commit Format

If no message provided, uses:
```
Deploy: automated deployment

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Verification Process

The script performs comprehensive checks:

- ✅ SSH connectivity to production server
- ✅ Container rebuild completion
- ✅ API health endpoint responding
- ✅ Web UI accessible
- ✅ Database connectivity
- ✅ Service health status (up to 20 attempts, 5s intervals)

## Production URLs

After successful deployment:

- **Web**: https://sprintspark.biswas.me
- **API**: https://sprintspark.biswas.me/api/health
- **Docs**: https://sprintspark.biswas.me/api/openapi

## Troubleshooting

If deployment fails:

```bash
# Check container status
./script/server status

# View API logs
./script/server logs api

# View web logs
./script/server logs web

# Manual restart
./script/server restart

# Force rebuild
ssh ubuntu@biswas.me "cd /home/ubuntu/projects/sprintspark && docker-compose down && docker-compose up -d --build"
```

## Related Commands

```bash
# Check health without deploying
./script/server health

# Restart services
./script/server restart

# View real-time logs
./script/server logs api
./script/server logs web

# Check service status
./script/server status

# Query production database
./script/server db-query "SELECT COUNT(*) FROM users;"
```

## Deployment Workflow

1. **Local Development**
   ```bash
   # Make changes
   # Test locally
   ./script/server test
   ```

2. **Deploy to Production**
   ```bash
   ./script/server deploy "feat: your feature description"
   ```

3. **Verify Deployment**
   ```bash
   ./script/server health
   ```

## Best Practices

✅ **DO:**
- Test locally before deploying (`./script/server test`)
- Use descriptive commit messages
- Check health after deployment
- Monitor logs for errors

❌ **DON'T:**
- Deploy with failing tests
- Deploy without reviewing changes
- Ignore health check failures
- Skip verification step

## Architecture

```
Local Machine → GitHub → Production Server
     ↓             ↓            ↓
   Commit       Webhook    Auto-rebuild
                            (or manual)
                               ↓
                          Health Check
                               ↓
                        ✅ Deployment OK
```

## Environment

- **Remote Server**: ubuntu@biswas.me
- **Deploy Path**: /home/ubuntu/projects/sprintspark
- **Domain**: sprintspark.biswas.me
- **Services**: Docker Compose (api + web)
- **Deployment**: GitHub webhook or manual trigger
