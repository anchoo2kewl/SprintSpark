#!/usr/bin/env python3
import json
import hashlib
import hmac
import subprocess
import sys
import logging
import yaml
import os
from http.server import HTTPServer, BaseHTTPRequestHandler
from pathlib import Path
import threading
import time
from datetime import datetime

class WebhookFramework:
    def __init__(self, config_path="/home/ubuntu/webhooks/config.yaml"):
        self.config_path = config_path
        self.config = self.load_config()
        self.setup_logging()
        
    def load_config(self):
        """Load configuration from YAML file"""
        try:
            with open(self.config_path, 'r') as f:
                return yaml.safe_load(f)
        except Exception as e:
            print(f"Error loading config: {e}")
            sys.exit(1)
    
    def setup_logging(self):
        """Setup logging configuration"""
        log_file = self.config['server'].get('log_file', '/home/ubuntu/webhooks/logs/webhook.log')
        log_level = getattr(logging, self.config['server'].get('log_level', 'INFO'))
        
        # Ensure log directory exists
        os.makedirs(os.path.dirname(log_file), exist_ok=True)
        
        logging.basicConfig(
            level=log_level,
            format='%(asctime)s - [%(project)s] - %(levelname)s - %(message)s',
            handlers=[
                logging.FileHandler(log_file),
                logging.StreamHandler()
            ]
        )
    
    def get_project_config(self, project_id):
        """Get configuration for a specific project"""
        if project_id not in self.config['projects']:
            return None
            
        project = self.config['projects'][project_id].copy()
        
        # Apply defaults for missing values
        defaults = self.config.get('defaults', {})
        for key, value in defaults.items():
            if key not in project:
                project[key] = value
                
        return project
    
    def verify_signature(self, payload, signature_header, secret):
        """Verify GitHub webhook signature"""
        if not signature_header:
            return False
            
        try:
            sha_name, signature = signature_header.split('=')
            if sha_name != 'sha256':
                return False
                
            mac = hmac.new(secret.encode(), payload, digestmod=hashlib.sha256)
            expected_signature = mac.hexdigest()
            
            return hmac.compare_digest(signature, expected_signature)
        except:
            return False
    
    def execute_action(self, action, project_config, logger):
        """Execute a project action"""
        action_type = action.get('type')
        
        try:
            if action_type == 'git_pull':
                return self.git_pull(project_config, logger)
            elif action_type == 'fix_permissions':
                return self.execute_command(action.get('command'), logger, project_config.get('timeout', 30))
            elif action_type == 'custom_command':
                return self.execute_command(action.get('command'), logger, project_config.get('timeout', 30))
            elif action_type == 'build':
                return self.execute_command(action.get('command'), logger, project_config.get('timeout', 120))
            elif action_type == 'restart_service':
                return self.execute_command(action.get('command'), logger, project_config.get('timeout', 30))
            else:
                logger.error(f"Unknown action type: {action_type}")
                return False
        except Exception as e:
            logger.error(f"Error executing action {action_type}: {str(e)}")
            return False
    
    def git_pull(self, project_config, logger):
        """Execute git pull for a project"""
        local_path = project_config.get('local_path')
        branch = project_config.get('branch', 'main')
        
        if not local_path or not os.path.exists(local_path):
            logger.error(f"Local path does not exist: {local_path}")
            return False
            
        command = f'cd {local_path} && git pull origin {branch}'
        return self.execute_command(command, logger, project_config.get('timeout', 30))
    
    def execute_command(self, command, logger, timeout=30):
        """Execute a shell command"""
        if not command:
            return True
            
        try:
            logger.info(f"Executing: {command}")
            result = subprocess.run([
                'bash', '-c', command
            ], capture_output=True, text=True, timeout=timeout)
            
            if result.returncode == 0:
                logger.info("Command executed successfully")
                if result.stdout.strip():
                    logger.info(f"Output: {result.stdout.strip()}")
                return True
            else:
                logger.error(f"Command failed with code {result.returncode}")
                if result.stderr.strip():
                    logger.error(f"Error: {result.stderr.strip()}")
                return False
                
        except subprocess.TimeoutExpired:
            logger.error(f"Command timed out after {timeout} seconds")
            return False
        except Exception as e:
            logger.error(f"Error executing command: {str(e)}")
            return False
    
    def process_webhook(self, project_id, payload):
        """Process webhook for a specific project"""
        # Create project-specific logger
        logger = logging.getLogger()
        logger = logging.LoggerAdapter(logger, {'project': project_id})
        
        project_config = self.get_project_config(project_id)
        if not project_config:
            logger.error(f"Project {project_id} not found in configuration")
            return False, "Project not found"
            
        if not project_config.get('enabled', False):
            logger.warning(f"Project {project_id} is disabled")
            return False, "Project disabled"
        
        # Check if this is the right repository and branch
        repo_name = payload.get('repository', {}).get('full_name')
        ref = payload.get('ref', '')
        expected_ref = f"refs/heads/{project_config.get('branch', 'main')}"
        
        if repo_name != project_config.get('repository'):
            logger.warning(f"Repository mismatch: got {repo_name}, expected {project_config.get('repository')}")
            return False, "Repository mismatch"
            
        if ref != expected_ref:
            logger.info(f"Ignoring push to {ref}, waiting for {expected_ref}")
            return False, "Branch mismatch"
        
        logger.info(f"Processing webhook for {project_config.get('name', project_id)}")
        logger.info(f"Repository: {repo_name}, Branch: {ref}")
        
        # Execute all actions
        success_count = 0
        total_actions = len(project_config.get('actions', []))
        
        for action in project_config.get('actions', []):
            if self.execute_action(action, project_config, logger):
                success_count += 1
            else:
                logger.error(f"Action failed: {action}")
        
        if success_count == total_actions:
            logger.info(f"All {total_actions} actions completed successfully")
            return True, "All actions completed successfully"
        else:
            logger.warning(f"Only {success_count}/{total_actions} actions completed successfully")
            return False, f"Only {success_count}/{total_actions} actions completed"

class WebhookHandler(BaseHTTPRequestHandler):
    def __init__(self, *args, webhook_framework=None, **kwargs):
        self.webhook_framework = webhook_framework
        super().__init__(*args, **kwargs)
    
    def log_message(self, format, *args):
        """Override to use our logging system"""
        pass
    
    def do_POST(self):
        # Parse the path to get project ID
        path_parts = self.path.strip('/').split('/')
        if len(path_parts) < 2 or path_parts[0] != 'webhook':
            self.send_error(404, "Invalid webhook path")
            return
            
        project_id = path_parts[1]
        
        # Get project configuration
        project_config = self.webhook_framework.get_project_config(project_id)
        if not project_config:
            self.send_error(404, f"Project '{project_id}' not found")
            return
            
        # Read payload
        content_length = int(self.headers.get('Content-Length', 0))
        post_data = self.rfile.read(content_length)
        
        # Verify GitHub signature
        signature = self.headers.get('X-Hub-Signature-256')
        secret = project_config.get('secret')
        
        if not self.webhook_framework.verify_signature(post_data, signature, secret):
            logging.warning(f'[{project_id}] Invalid signature received')
            self.send_error(403, "Invalid signature")
            return
            
        try:
            payload = json.loads(post_data.decode('utf-8'))
            
            # Process the webhook
            success, message = self.webhook_framework.process_webhook(project_id, payload)
            
            if success:
                self.send_response(200)
                self.end_headers()
                self.wfile.write(message.encode('utf-8'))
            else:
                self.send_response(200)  # Still 200 to avoid GitHub retries
                self.end_headers()
                self.wfile.write(f"Webhook received but no action taken: {message}".encode('utf-8'))
                
        except json.JSONDecodeError:
            logging.error(f'[{project_id}] Invalid JSON payload')
            self.send_error(400, "Invalid JSON payload")
        except Exception as e:
            logging.error(f'[{project_id}] Error processing webhook: {str(e)}')
            self.send_error(500, "Internal server error")
    
    def do_GET(self):
        if self.path == '/health':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            
            status = {
                "status": "healthy",
                "timestamp": datetime.now().isoformat(),
                "projects": {}
            }
            
            for project_id, project in self.webhook_framework.config['projects'].items():
                status["projects"][project_id] = {
                    "name": project.get('name', project_id),
                    "enabled": project.get('enabled', False),
                    "repository": project.get('repository')
                }
            
            self.wfile.write(json.dumps(status, indent=2).encode('utf-8'))
            
        elif self.path.startswith('/projects'):
            # List all configured projects
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            
            projects = {}
            for project_id, project in self.webhook_framework.config['projects'].items():
                projects[project_id] = {
                    "name": project.get('name', project_id),
                    "repository": project.get('repository'),
                    "branch": project.get('branch'),
                    "enabled": project.get('enabled', False),
                    "webhook_url": f"https://webhook.biswas.me/webhook/{project_id}"
                }
            
            self.wfile.write(json.dumps(projects, indent=2).encode('utf-8'))
        else:
            self.send_error(404, "Not Found")

def create_handler(webhook_framework):
    """Create a handler class with the webhook framework"""
    def handler(*args, **kwargs):
        return WebhookHandler(*args, webhook_framework=webhook_framework, **kwargs)
    return handler

if __name__ == '__main__':
    # Check if PyYAML is installed
    try:
        import yaml
    except ImportError:
        print("PyYAML is required. Installing...")
        subprocess.run([sys.executable, "-m", "pip", "install", "PyYAML"])
        import yaml
    
    # Initialize webhook framework
    framework = WebhookFramework()
    
    # Create server
    server_config = framework.config['server']
    handler = create_handler(framework)
    server = HTTPServer((server_config['host'], server_config['port']), handler)
    
    print(f"Multi-project webhook server starting on {server_config['host']}:{server_config['port']}")
    print(f"Configured projects: {list(framework.config['projects'].keys())}")
    
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nWebhook server stopped")
        server.server_close()
