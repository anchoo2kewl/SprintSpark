# Webhook Auto-Deployment System

Multi-project webhook framework for GitHub auto-deployment with HMAC security.

## Features

- 🔒 Secure HMAC signature verification
- 🚀 Multi-project support
- 📝 YAML configuration
- 🔄 Automatic git pull and deployment
- 📊 Real-time logging
- ✅ Production-ready

## Setup

### 1. Install Dependencies

```bash
pip install pyyaml
```

### 2. Configure Projects

```bash
cp config.sample.yaml config.yaml
# Edit config.yaml with your project settings
```

### 3. Start Webhook Server

```bash
python3 webhook_server.py
```

The server will listen on port 9876 by default.

### 4. Configure Nginx (Optional)

For HTTPS and domain routing:

```nginx
server {
    server_name webhook.yourdomain.com;

    location /webhook/ {
        proxy_pass http://127.0.0.1:9876;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        proxy_pass http://127.0.0.1:9876/health;
    }
}
```

### 5. Configure GitHub Webhook

1. Go to your repository settings
2. Navigate to Webhooks → Add webhook
3. Set the Payload URL: `https://webhook.yourdomain.com/webhook/project-id`
4. Content type: `application/json`
5. Secret: Use the same secret from your `config.yaml`
6. Events: Select "Just the push event"

## Configuration

### Project Structure

```yaml
projects:
  project-id:
    name: "Project Display Name"
    repository: "username/repository"
    secret: "your-webhook-secret"
    branch: "main"
    local_path: "/path/to/local/project"
    actions:
      - type: "git_pull"
      - type: "custom_command"
        command: "docker-compose up --build -d"
    enabled: true
```

### Action Types

- **`git_pull`**: Automatically pulls latest changes from repository
- **`custom_command`**: Executes any shell command
- **`fix_permissions`**: Changes file permissions (requires command)

## Endpoints

- **`POST /webhook/{project-id}`** - Receive GitHub webhook
- **`GET /health`** - Health check
- **`GET /projects`** - List configured projects

## Testing

```bash
# Test webhook framework
python3 test_framework.py

# Check webhook health
curl https://webhook.yourdomain.com/health
```

## Security

- All webhooks are verified using HMAC-SHA256 signatures
- Secrets should be strong and unique per project
- Use HTTPS in production (configure via Nginx with Let's Encrypt)
- Log files should be regularly rotated

## Logs

Logs are written to the path specified in `config.yaml`:

```bash
tail -f /path/to/webhooks/logs/webhook.log
```

## Troubleshooting

### Webhook returns 404
- Check that the project ID in the URL matches your config
- Verify the project is enabled in `config.yaml`

### Webhook returns 403
- Check that the webhook secret matches your GitHub settings
- Verify HMAC signature verification is working

### Deployment not triggering
- Check webhook logs for errors
- Verify the repository path exists and has correct permissions
- Ensure git is configured on the server
