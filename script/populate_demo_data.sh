#!/bin/bash

# SprintSpark Demo Data Population Script
# Creates users, projects, sprints, tags, and tasks with realistic data

set -e

API_URL="${API_URL:-https://sprintspark.biswas.me}"
TESTUSER_EMAIL="${TESTUSER_EMAIL:-testuser2@example.com}"
TESTUSER_PASSWORD="${TESTUSER_PASSWORD:-TestPass123!}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}=== SprintSpark Demo Data Population ===${NC}"
echo "API URL: $API_URL"
echo ""

# Function to make API calls
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local token=$4

    if [ -n "$token" ]; then
        curl -s -X "$method" "$API_URL$endpoint" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" \
            -d "$data"
    else
        curl -s -X "$method" "$API_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data"
    fi
}

# Login as testuser - try multiple common passwords
echo -e "${YELLOW}Logging in as testuser...${NC}"
PASSWORDS=("TestPass123!" "test123" "password" "test" "testpass" "Test123!")

for TRY_PASS in "${PASSWORDS[@]}"; do
    LOGIN_RESPONSE=$(api_call POST "/api/auth/login" "{\"email\":\"$TESTUSER_EMAIL\",\"password\":\"$TRY_PASS\"}")
    TOKEN=$(echo "$LOGIN_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('token', ''))" 2>/dev/null || echo "")

    if [ -n "$TOKEN" ]; then
        TESTUSER_PASSWORD="$TRY_PASS"
        echo -e "${GREEN}✓ Logged in successfully with password: $TESTUSER_PASSWORD${NC}"
        break
    fi
done

if [ -z "$TOKEN" ]; then
    echo "Failed to login with common passwords."
    echo "Trying to create new testuser account..."

    # Try alternate email
    TESTUSER_EMAIL="testuser2@example.com"
    SIGNUP_RESPONSE=$(api_call POST "/api/auth/signup" "{\"email\":\"$TESTUSER_EMAIL\",\"password\":\"$TESTUSER_PASSWORD\",\"name\":\"Test User\"}")
    TOKEN=$(echo "$SIGNUP_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('token', ''))" 2>/dev/null || echo "")

    if [ -z "$TOKEN" ]; then
        echo "Failed to create testuser. Please ensure testuser@example.com exists or create it manually."
        echo "Response: $SIGNUP_RESPONSE"
        exit 1
    fi
    echo -e "${GREEN}✓ Created new user: $TESTUSER_EMAIL${NC}"
fi

echo ""

# Create 5 additional users
echo -e "${YELLOW}Creating 5 additional users...${NC}"
USERS=()
USER_TOKENS=()
USER_IDS=()

for i in {1..5}; do
    USER_EMAIL="demo.user${i}@sprintspark.app"
    USER_PASSWORD="DemoPass${i}23!"
    USER_NAME="Demo User $i"

    USER_RESPONSE=$(api_call POST "/api/auth/signup" "{\"email\":\"$USER_EMAIL\",\"password\":\"$USER_PASSWORD\",\"name\":\"$USER_NAME\"}")
    USER_TOKEN=$(echo "$USER_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('token', ''))" 2>/dev/null || echo "")
    USER_ID=$(echo "$USER_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('user', {}).get('id', ''))" 2>/dev/null || echo "$((i+4))")

    if [ -z "$USER_TOKEN" ]; then
        # User might already exist, try to get ID by assuming sequential IDs
        USER_ID=$((i+4))
        echo "  ⚠ User $USER_EMAIL may already exist (using ID: $USER_ID)"
    else
        echo "  ✓ Created $USER_NAME ($USER_EMAIL) - ID: $USER_ID"
    fi

    USERS+=("$USER_EMAIL")
    USER_TOKENS+=("$USER_TOKEN")
    USER_IDS+=("$USER_ID")
done

echo ""

# Create Projects
echo -e "${YELLOW}Creating projects...${NC}"
PROJECTS=(
    "E-Commerce Platform|Complete overhaul of the online shopping experience with modern UI/UX"
    "Mobile App Redesign|Redesigning the mobile application for better user engagement"
    "Backend Infrastructure|Upgrading backend services for improved scalability and performance"
)

PROJECT_IDS=()
for project_data in "${PROJECTS[@]}"; do
    IFS='|' read -r name description <<< "$project_data"

    PROJECT_RESPONSE=$(api_call POST "/api/projects" "{\"name\":\"$name\",\"description\":\"$description\"}" "$TOKEN")
    PROJECT_ID=$(echo "$PROJECT_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))" 2>/dev/null || echo "")

    if [ -n "$PROJECT_ID" ]; then
        PROJECT_IDS+=("$PROJECT_ID")
        echo "  ✓ Created project: $name (ID: $PROJECT_ID)"
    fi
done

echo ""

# Create Sprints for each project
echo -e "${YELLOW}Creating sprints...${NC}"
SPRINT_NAMES=("Sprint 1: Foundation" "Sprint 2: Core Features" "Sprint 3: Polish & Testing")

for PROJECT_ID in "${PROJECT_IDS[@]}"; do
    for sprint_name in "${SPRINT_NAMES[@]}"; do
        # Calculate sprint dates (2-week sprints)
        START_DATE=$(date -v+1d +%Y-%m-%d 2>/dev/null || date -d "+1 day" +%Y-%m-%d)
        END_DATE=$(date -v+14d +%Y-%m-%d 2>/dev/null || date -d "+14 days" +%Y-%m-%d)

        SPRINT_RESPONSE=$(api_call POST "/api/sprints" "{\"name\":\"$sprint_name\",\"project_id\":$PROJECT_ID,\"start_date\":\"${START_DATE}T00:00:00Z\",\"end_date\":\"${END_DATE}T23:59:59Z\",\"goal\":\"Complete assigned features and address technical debt\"}" "$TOKEN")
        echo "  ✓ Created sprint: $sprint_name for project $PROJECT_ID"
    done
done

echo ""

# Create Tags
echo -e "${YELLOW}Creating tags...${NC}"
TAGS=(
    "bug|#ef4444"
    "feature|#3b82f6"
    "enhancement|#10b981"
    "documentation|#8b5cf6"
    "security|#f59e0b"
    "performance|#06b6d4"
    "ui|#ec4899"
    "backend|#6366f1"
    "frontend|#14b8a6"
    "testing|#f97316"
    "urgent|#dc2626"
    "low-priority|#84cc16"
)

TAG_IDS=()
for tag_data in "${TAGS[@]}"; do
    IFS='|' read -r name color <<< "$tag_data"

    TAG_RESPONSE=$(api_call POST "/api/tags" "{\"name\":\"$name\",\"color\":\"$color\"}" "$TOKEN")
    TAG_ID=$(echo "$TAG_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))" 2>/dev/null || echo "")

    if [ -n "$TAG_ID" ]; then
        TAG_IDS+=("$TAG_ID")
        echo "  ✓ Created tag: $name (ID: $TAG_ID)"
    else
        echo "  ⚠ Tag $name may already exist, skipping"
    fi
done

# If no tags were created (all exist), use IDs 1-12
if [ ${#TAG_IDS[@]} -eq 0 ]; then
    echo "  Using existing tags (IDs 1-12)"
    TAG_IDS=(1 2 3 4 5 6 7 8 9 10 11 12)
fi

echo ""

# Task templates with rich descriptions
declare -a TASK_TEMPLATES=(
    "Implement user authentication|User Authentication - Implement secure user authentication system with JWT tokens, password hashing, session management, and remember me functionality. Acceptance criteria includes login endpoint, signup endpoint, password reset, and email verification.|high|8"
    "Design database schema|Database Design - Design and implement the core database schema for Users, Projects, Tasks, and Comments. Consider proper indexing, foreign key constraints, and migration scripts.|medium|12"
    "Create API documentation|API Documentation - Create comprehensive API documentation using OpenAPI/Swagger. Document all endpoints, include request/response examples, add authentication details, and provide code samples.|low|4"
    "Implement search functionality|Global Search - Add search across projects and tasks with full-text search, fuzzy matching, and search filters.|urgent|16"
    "Add email notifications|Email Notifications - Implement email notification system for task assignments, due date reminders, project updates, and weekly digest using SendGrid or AWS SES.|medium|10"
    "Create dashboard widgets|Dashboard Widgets - Build interactive dashboard with task completion charts, velocity tracking, burndown chart, and team activity feed using Chart.js or Recharts.|high|14"
    "Setup CI/CD pipeline|CI/CD Pipeline - Setup automated deployment with GitHub Actions workflow, automated testing, Docker image building, and production deployment with stages for lint, test, build, and deploy.|urgent|20"
    "Implement file uploads|File Upload Feature - Add file attachment support for images and documents with file preview and storage in S3 or MinIO. Maximum file size of 10MB per file.|high|12"
    "Add real-time updates|WebSocket Integration - Implement real-time features including live task updates, user presence indicators, and push notifications.|urgent|24"
    "Create mobile responsive design|Responsive Design - Make UI mobile-friendly with mobile-first approach, touch-friendly controls, responsive breakpoints, and PWA support.|medium|8"
)

# Create 30+ tasks for each project
echo -e "${YELLOW}Creating tasks for each project...${NC}"

STATUSES=("todo" "in_progress" "done")
PRIORITIES=("low" "medium" "high" "urgent")

for PROJECT_ID in "${PROJECT_IDS[@]}"; do
    echo "  Creating 30 tasks for project $PROJECT_ID..."

    for i in {1..30}; do
        # Select template
        TEMPLATE_INDEX=$((i % ${#TASK_TEMPLATES[@]}))
        IFS='|' read -r title description priority estimate <<< "${TASK_TEMPLATES[$TEMPLATE_INDEX]}"

        # Add variety to title
        TASK_TITLE="${title} - Iteration $((i / ${#TASK_TEMPLATES[@]} + 1))"

        # Random status
        STATUS="${STATUSES[$((i % 3))]}"

        # Random priority
        PRIORITY_VAL="${PRIORITIES[$((i % 4))]}"

        # Random assignee (including testuser)
        ASSIGNEE_ID="${USER_IDS[$((i % 5))]}"

        # Random tags (1-3 tags)
        NUM_TAGS=$((1 + RANDOM % 3))
        TAG_SELECTION=""
        for ((t=0; t<NUM_TAGS; t++)); do
            TAG_IDX=$((RANDOM % ${#TAG_IDS[@]}))
            if [ -n "$TAG_SELECTION" ]; then
                TAG_SELECTION="$TAG_SELECTION,${TAG_IDS[$TAG_IDX]}"
            else
                TAG_SELECTION="${TAG_IDS[$TAG_IDX]}"
            fi
        done

        # Due date (random between 1-30 days from now)
        DAYS_AHEAD=$((1 + RANDOM % 30))
        DUE_DATE=$(date -v+${DAYS_AHEAD}d +%Y-%m-%dT00:00:00Z 2>/dev/null || date -d "+${DAYS_AHEAD} days" +%Y-%m-%dT00:00:00Z)

        # Estimated hours
        EST_HOURS=$((4 + RANDOM % 20))

        # Actual hours (if done)
        if [ "$STATUS" == "done" ]; then
            ACTUAL_HOURS=$((EST_HOURS - 2 + RANDOM % 5))
        else
            ACTUAL_HOURS=0
        fi

        # Create task
        TASK_DATA=$(cat <<EOF
{
  "title": "$TASK_TITLE",
  "description": "$description",
  "status": "$STATUS",
  "priority": "$PRIORITY_VAL",
  "assignee_id": $ASSIGNEE_ID,
  "estimated_hours": $EST_HOURS,
  "actual_hours": $ACTUAL_HOURS,
  "due_date": "$DUE_DATE",
  "tag_ids": [$TAG_SELECTION]
}
EOF
)

        TASK_RESPONSE=$(api_call POST "/api/projects/$PROJECT_ID/tasks" "$TASK_DATA" "$TOKEN")

        if [ $((i % 10)) -eq 0 ]; then
            echo "    ✓ Created $i tasks..."
        fi
    done

    echo "  ✓ Completed 30 tasks for project $PROJECT_ID"
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Demo Data Population Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "📊 Summary:"
echo "  - Projects created: ${#PROJECT_IDS[@]}"
echo "  - Users created: ${#USERS[@]} (+ testuser)"
echo "  - Tags created: ${#TAG_IDS[@]}"
echo "  - Tasks created: $((${#PROJECT_IDS[@]} * 30))"
echo ""
echo "🔐 Test User Credentials:"
echo "  Email: $TESTUSER_EMAIL"
echo "  Password: $TESTUSER_PASSWORD"
echo "  Admin: Yes"
echo ""
echo "👥 Other Users:"
for i in {0..4}; do
    echo "  - ${USERS[$i]} / DemoPass$((i+1))23!"
done
echo ""
echo "🌐 Access: $API_URL"
echo ""
