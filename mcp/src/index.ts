import express from "express";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
import { z } from "zod";
import { TaskAIClient } from "./api.js";

const TASKAI_API_URL = process.env.TASKAI_API_URL || "https://taskai.cc";
const PORT = parseInt(process.env.PORT || "3000", 10);

/**
 * Create and configure the MCP server with all TaskAI tools.
 */
function createServer(client: TaskAIClient): McpServer {
  const server = new McpServer({
    name: "taskai",
    version: "1.0.0",
  });

  // --- get_me ---
  server.tool("get_me", "Get current authenticated user info", {}, async () => {
    const user = await client.getMe();
    return { content: [{ type: "text", text: JSON.stringify(user, null, 2) }] };
  });

  // --- list_projects ---
  server.tool(
    "list_projects",
    "List all projects",
    { page: z.number().optional(), limit: z.number().optional() },
    async ({ page, limit }) => {
      const result = await client.listProjects(page, limit);
      return { content: [{ type: "text", text: JSON.stringify(result, null, 2) }] };
    }
  );

  // --- get_project ---
  server.tool(
    "get_project",
    "Get project details by ID",
    { project_id: z.string().describe("Project ID") },
    async ({ project_id }) => {
      const project = await client.getProject(project_id);
      return { content: [{ type: "text", text: JSON.stringify(project, null, 2) }] };
    }
  );

  // --- list_swim_lanes ---
  server.tool(
    "list_swim_lanes",
    "List swim lanes (columns) for a project",
    { project_id: z.string().describe("Project ID") },
    async ({ project_id }) => {
      const lanes = await client.listSwimLanes(project_id);
      return { content: [{ type: "text", text: JSON.stringify(lanes, null, 2) }] };
    }
  );

  // --- list_tasks ---
  server.tool(
    "list_tasks",
    "List tasks in a project (optional status/search filter)",
    {
      project_id: z.string().describe("Project ID"),
      query: z.string().optional().describe("Search query"),
      status: z.string().optional().describe("Filter by status (e.g. todo, in_progress, done)"),
      page: z.number().optional(),
      limit: z.number().optional(),
    },
    async ({ project_id, query, status, page, limit }) => {
      const result = await client.listTasks(project_id, { query, status, page, limit });
      return { content: [{ type: "text", text: JSON.stringify(result, null, 2) }] };
    }
  );

  // --- create_task ---
  server.tool(
    "create_task",
    "Create a new task in a project",
    {
      project_id: z.string().describe("Project ID"),
      title: z.string().describe("Task title"),
      description: z.string().optional().describe("Task description"),
      status: z.string().optional().describe("Task status (default: todo)"),
      priority: z.string().optional().describe("Priority: low, medium, high, critical"),
      assigned_to: z.string().optional().describe("User ID to assign"),
      swim_lane_id: z.number().optional().describe("Swim lane ID (use list_swim_lanes to get valid IDs)"),
    },
    async ({ project_id, title, description, status, priority, assigned_to, swim_lane_id }) => {
      const task = await client.createTask(project_id, { title, description, status, priority, assigned_to, swim_lane_id });
      return { content: [{ type: "text", text: JSON.stringify(task, null, 2) }] };
    }
  );

  // --- update_task ---
  server.tool(
    "update_task",
    "Update an existing task",
    {
      task_id: z.string().describe("Task ID"),
      title: z.string().optional().describe("New title"),
      description: z.string().optional().describe("New description"),
      status: z.string().optional().describe("New status"),
      priority: z.string().optional().describe("New priority"),
      assigned_to: z.string().optional().describe("New assignee user ID"),
      swim_lane_id: z.number().optional().describe("Swim lane ID (use list_swim_lanes to get valid IDs)"),
    },
    async ({ task_id, title, description, status, priority, assigned_to, swim_lane_id }) => {
      const task = await client.updateTask(task_id, { title, description, status, priority, assigned_to, swim_lane_id });
      return { content: [{ type: "text", text: JSON.stringify(task, null, 2) }] };
    }
  );

  // --- list_comments ---
  server.tool(
    "list_comments",
    "List comments on a task",
    { task_id: z.string().describe("Task ID") },
    async ({ task_id }) => {
      const result = await client.listComments(task_id);
      return { content: [{ type: "text", text: JSON.stringify(result, null, 2) }] };
    }
  );

  // --- add_comment ---
  server.tool(
    "add_comment",
    "Add a comment to a task",
    {
      task_id: z.string().describe("Task ID"),
      content: z.string().describe("Comment text"),
    },
    async ({ task_id, content }) => {
      const comment = await client.addComment(task_id, content);
      return { content: [{ type: "text", text: JSON.stringify(comment, null, 2) }] };
    }
  );

  return server;
}

// --- Express app ---
const app = express();
app.use(express.json());

// Health endpoint
app.get("/health", (_req, res) => {
  res.json({ status: "ok", service: "taskai-mcp" });
});

// MCP endpoint — stateless: one transport per request
app.post("/mcp", async (req, res) => {
  // Extract API key from X-API-Key header
  const apiKey = req.headers["x-api-key"] as string | undefined;
  if (!apiKey) {
    res.status(401).json({ error: "Missing X-API-Key header" });
    return;
  }

  // Validate the API key by calling /api/me
  const client = new TaskAIClient(TASKAI_API_URL, apiKey);
  try {
    await client.getMe();
  } catch {
    res.status(403).json({ error: "Invalid API key" });
    return;
  }

  // Create MCP server with authenticated client
  const server = createServer(client);

  // Stateless transport — no session persistence
  const transport = new StreamableHTTPServerTransport({
    sessionIdGenerator: undefined,
  });
  await server.connect(transport);
  await transport.handleRequest(req, res, req.body);
});

// Handle GET and DELETE on /mcp for protocol compliance (stateless = 405)
app.get("/mcp", (_req, res) => {
  res.status(405).json({ error: "Method not allowed — stateless server, use POST" });
});

app.delete("/mcp", (_req, res) => {
  res.status(405).json({ error: "Method not allowed — stateless server, use POST" });
});

app.listen(PORT, () => {
  console.log(`TaskAI MCP server listening on port ${PORT}`);
  console.log(`API backend: ${TASKAI_API_URL}`);
});
