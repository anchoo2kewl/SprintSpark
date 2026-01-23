# User Management - Local Development

## 📋 Quick Reference

### Login Credentials

**Primary Admin User:**
```
Email: a@b.me
Password: password123
```

**Test Users:**
```
testuser2@example.com / TestPass123!
demo.user1@sprintspark.app / DemoPass123!
demo.user2@sprintspark.app / DemoPass223!
demo.user3@sprintspark.app / DemoPass323!
demo.user4@sprintspark.app / DemoPass423!
demo.user5@sprintspark.app / DemoPass523!
```

## 🛠️ User Management Commands

### List All Users (Local Only)
```bash
./script/server local users:list
```

**Output:**
- User ID
- Email
- Admin status
- Created date
- Password hash preview (first 20 characters)

**Note:** Passwords are bcrypt-hashed and **CANNOT** be decrypted.

### Reset User Password (Local Only)
```bash
# Reset with default password (password123)
./script/server local users:password <email>

# Reset with custom password
./script/server local users:password <email> <new_password>
```

**Examples:**
```bash
# Reset a@b.me to default password
./script/server local users:password a@b.me

# Reset with custom password
./script/server local users:password a@b.me MySecurePass123!
```

**What it does:**
1. Checks if user exists
2. Saves admin status
3. Deletes the user from database
4. Recreates user via API (which hashes the password)
5. Restores admin status if needed

**⚠️ Important:**
- Only works in **LOCAL** environment
- Requires API container to be running (`./script/server local:start`)
- Will **delete and recreate** the user (preserves admin status only)
- All user's projects/tasks remain intact (foreign key constraints)

## 🔐 Security Notes

### Password Storage
- All passwords are hashed using **bcrypt** (cost factor: 12)
- Password hashes are stored in `password_hash` column
- **Passwords CANNOT be retrieved** - only reset

### Why Can't We List Plain Passwords?
Passwords are one-way hashed for security:
```
User enters: "password123"
       ↓
bcrypt hash: "$2a$12$B6j7mW59H6l4K..."
       ↓
Stored in database (irreversible)
```

When user logs in:
```
User enters: "password123"
       ↓
bcrypt.Compare(entered, stored_hash) → true/false
```

## 📦 Creating Test Data

### Populate Demo Data
```bash
# Populate production (remote)
./script/populate_demo_data.sh

# Populate local
API_URL=http://localhost:8083 ./script/populate_demo_data.sh
```

Creates:
- 5 demo users
- 3 projects
- 9 sprints (3 per project)
- 12 tags
- 90 tasks (30 per project)

## 🔧 Admin Management

### List Admins
```bash
# Remote (production)
./script/server admin list

# Local
./script/server local admin list
```

### Promote User to Admin
```bash
# Remote
./script/server admin create user@example.com

# Local
./script/server local admin create user@example.com
```

### Revoke Admin Privileges
```bash
# Remote
./script/server admin revoke user@example.com

# Local
./script/server local admin revoke user@example.com
```

## 🚨 Troubleshooting

### "User not found in database"
```bash
# Check all users
./script/server local users:list

# Or query database directly
./script/server db-query "SELECT * FROM users;"
```

### "Local API container is not running"
```bash
# Start local development server
./script/server local:start

# Check if running
docker-compose ps
```

### "Failed to reset password"
Check the API logs:
```bash
docker-compose logs api | tail -50
```

Common issues:
- API not running
- Database locked
- Invalid email format
- Network issues

## 📝 Database Schema

### Users Table
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_admin INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Direct Database Access
```bash
# Execute SQL query
./script/server db-query "SELECT * FROM users LIMIT 5;"

# Or use Docker
docker-compose exec -u root api sh -c "sqlite3 /data/sprintspark.db 'SELECT * FROM users;'"
```

## 🔒 Production vs Local

| Feature | Local | Production |
|---------|-------|------------|
| List users | ✅ `local users:list` | ❌ Security risk |
| Reset password | ✅ `local users:password` | ❌ Use forgot password flow |
| View password hashes | ✅ Debugging only | ❌ Never expose |
| Direct DB access | ✅ Via Docker | ✅ Via SSH + Docker |

**Never expose user credentials in production!**

## 🎯 Common Workflows

### I forgot my local admin password
```bash
./script/server local users:password a@b.me password123
```

### I want to test with a fresh user
```bash
# Option 1: Reset existing user
./script/server local users:password demo.user1@sprintspark.app NewPass123!

# Option 2: Create new user via signup
curl -X POST http://localhost:8083/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"newuser@test.com","password":"Test123!"}'
```

### I want to promote a user to admin
```bash
./script/server local admin create user@example.com
```

### I want to see all users and their data
```bash
# List users
./script/server local users:list

# Check specific user
./script/server db-query "SELECT * FROM users WHERE email = 'a@b.me';"
```

---

**Last Updated:** 2026-01-22
**Environment:** Local Development Only
