#!/usr/bin/env python3
import json
import hashlib
import hmac
import requests
import sys

def test_webhook(project_id, webhook_url, secret):
    """Test webhook endpoint for a specific project"""
    
    # Sample GitHub push payload
    payload = {
        "ref": "refs/heads/main",
        "repository": {
            "full_name": "NayantaraB/academic-website" if project_id == "nayantara" else f"user/{project_id}"
        },
        "pusher": {
            "name": "test-user"
        }
    }

    payload_json = json.dumps(payload)
    payload_bytes = payload_json.encode('utf-8')

    # Create signature
    signature = hmac.new(
        secret.encode('utf-8'),
        payload_bytes,
        digestmod=hashlib.sha256
    ).hexdigest()

    headers = {
        'Content-Type': 'application/json',
        'X-Hub-Signature-256': f'sha256={signature}'
    }

    print(f"Testing webhook for {project_id}...")
    print(f"URL: {webhook_url}")
    print(f"Secret: {secret}")
    
    try:
        response = requests.post(webhook_url, data=payload_json, headers=headers)
        print(f"Status Code: {response.status_code}")
        print(f"Response: {response.text}")
        return response.status_code == 200
    except Exception as e:
        print(f"Error: {e}")
        return False

if __name__ == "__main__":
    # Test Nayantara's webhook
    success = test_webhook(
        "nayantara", 
        "https://webhook.biswas.me/webhook/nayantara",
        "nayantara-website-hook-2025"
    )
    
    print(f"\nTest {'PASSED' if success else 'FAILED'}!")
