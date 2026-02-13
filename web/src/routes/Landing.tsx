import { Link } from 'react-router-dom'

export default function Landing() {
  return (
    <div className="min-h-screen bg-dark-bg-base">
      {/* Navigation */}
      <nav className="fixed top-0 left-0 right-0 z-50 bg-dark-bg-base/80 backdrop-blur-lg border-b border-dark-border-subtle">
        <div className="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2.5">
            <img src="/logo.svg" alt="TaskAI" className="w-7 h-7" />
            <span className="text-base font-semibold text-dark-text-primary tracking-tight">TaskAI</span>
          </Link>
          <div className="flex items-center gap-4">
            <Link
              to="/login"
              className="text-sm text-dark-text-tertiary hover:text-dark-text-primary transition-colors duration-150"
            >
              Sign in
            </Link>
            <Link
              to="/signup"
              className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-500 hover:bg-primary-600 rounded-md shadow-linear-sm transition-all duration-150"
            >
              Get started
            </Link>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="relative pt-32 pb-20 md:pt-44 md:pb-32 overflow-hidden">
        {/* Gradient orb background */}
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[600px] bg-primary-500/[0.07] rounded-full blur-[120px]" />
          <div className="absolute top-1/3 left-1/3 w-[400px] h-[400px] bg-secondary-500/[0.04] rounded-full blur-[100px]" />
        </div>

        <div className="relative max-w-4xl mx-auto px-6 text-center">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 mb-8 text-xs font-medium text-primary-400 bg-primary-500/10 border border-primary-500/20 rounded-full">
            <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            MCP Server built-in — Let AI manage your projects
          </div>
          <h1 className="text-5xl md:text-6xl font-bold text-dark-text-primary tracking-tight">
            AI-native
            <br />
            <span className="bg-gradient-to-r from-primary-400 to-secondary-400 bg-clip-text text-transparent">
              project management
            </span>
          </h1>
          <p className="mt-6 text-lg md:text-xl text-dark-text-tertiary max-w-2xl mx-auto leading-relaxed">
            TaskAI is the first project management tool built for AI agents.
            Connect via MCP and let LLMs create tasks, manage sprints, and
            ship autonomously — or use the visual board yourself.
          </p>
          <div className="mt-10 flex items-center justify-center gap-4">
            <Link
              to="/signup"
              className="inline-flex items-center px-6 py-3 text-base font-medium text-white bg-primary-500 hover:bg-primary-600 rounded-md shadow-linear transition-all duration-150"
            >
              Start building with AI
              <svg className="ml-2 w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
              </svg>
            </Link>
            <Link
              to="/login"
              className="inline-flex items-center px-6 py-3 text-base font-medium text-dark-text-secondary border border-dark-border-medium hover:border-dark-border-strong hover:bg-dark-bg-secondary rounded-md transition-all duration-150"
            >
              Sign in
            </Link>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 md:py-28 border-t border-dark-border-subtle">
        <div className="max-w-6xl mx-auto px-6">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold text-dark-text-primary tracking-tight">
              Built for AI, designed for humans
            </h2>
            <p className="mt-4 text-base text-dark-text-tertiary max-w-xl mx-auto">
              The only project management tool where AI agents are first-class citizens.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {/* Feature 1: MCP Server */}
            <div className="group p-6 rounded-xl bg-dark-bg-primary border border-dark-border-subtle hover:border-dark-border-medium transition-all duration-200">
              <div className="w-12 h-12 bg-primary-500/10 rounded-lg flex items-center justify-center mb-4 group-hover:bg-primary-500/15 transition-colors duration-200">
                <svg className="w-6 h-6 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9.75 3.104v5.714a2.25 2.25 0 01-.659 1.591L5 14.5M9.75 3.104c-.251.023-.501.05-.75.082m.75-.082a24.301 24.301 0 014.5 0m0 0v5.714a2.25 2.25 0 00.659 1.591L19 14.5M14.25 3.104c.251.023.501.05.75.082M19 14.5l-2.47 2.47a2.25 2.25 0 01-1.59.659H9.06a2.25 2.25 0 01-1.591-.659L5 14.5m14 0V7.088a1.5 1.5 0 00-.444-1.067l-1.6-1.6A1.5 1.5 0 0015.888 4H8.112a1.5 1.5 0 00-1.068.443l-1.6 1.6A1.5 1.5 0 005 7.088V14.5" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-dark-text-primary mb-2 tracking-tight">MCP Server</h3>
              <p className="text-sm text-dark-text-tertiary leading-relaxed">
                AI agents create tasks, manage sprints, and update statuses via the Model Context Protocol. LLMs work independently with your projects through natural language.
              </p>
            </div>

            {/* Feature 2: API-First */}
            <div className="group p-6 rounded-xl bg-dark-bg-primary border border-dark-border-subtle hover:border-dark-border-medium transition-all duration-200">
              <div className="w-12 h-12 bg-success-500/10 rounded-lg flex items-center justify-center mb-4 group-hover:bg-success-500/15 transition-colors duration-200">
                <svg className="w-6 h-6 text-success-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17.25 6.75L22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3l-4.5 16.5" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-dark-text-primary mb-2 tracking-tight">API-First Architecture</h3>
              <p className="text-sm text-dark-text-tertiary leading-relaxed">
                50+ REST endpoints with full OpenAPI spec. Every action available programmatically. Build custom integrations or connect any AI agent in minutes.
              </p>
            </div>

            {/* Feature 3: Visual Project Management */}
            <div className="group p-6 rounded-xl bg-dark-bg-primary border border-dark-border-subtle hover:border-dark-border-medium transition-all duration-200">
              <div className="w-12 h-12 bg-secondary-500/10 rounded-lg flex items-center justify-center mb-4 group-hover:bg-secondary-500/15 transition-colors duration-200">
                <svg className="w-6 h-6 text-secondary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-dark-text-primary mb-2 tracking-tight">Visual Project Management</h3>
              <p className="text-sm text-dark-text-tertiary leading-relaxed">
                Kanban boards with drag-and-drop, customizable swim lanes, sprint planning, team collaboration, and real-time sync across devices.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* How it works Section */}
      <section className="py-20 md:py-28 border-t border-dark-border-subtle">
        <div className="max-w-4xl mx-auto px-6">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold text-dark-text-primary tracking-tight">
              How AI agents use TaskAI
            </h2>
            <p className="mt-4 text-base text-dark-text-tertiary max-w-xl mx-auto">
              Connect any LLM via MCP and it can manage your projects autonomously.
            </p>
          </div>

          <div className="space-y-6">
            <div className="flex items-start gap-4 p-5 rounded-lg bg-dark-bg-primary border border-dark-border-subtle">
              <div className="flex-shrink-0 w-8 h-8 bg-primary-500/10 rounded-full flex items-center justify-center text-primary-400 text-sm font-bold">1</div>
              <div>
                <h3 className="font-semibold text-dark-text-primary mb-1">Connect via MCP</h3>
                <p className="text-sm text-dark-text-tertiary">Point your AI agent (Claude, GPT, or any MCP-compatible LLM) at your TaskAI instance. Authenticate with an API key.</p>
              </div>
            </div>
            <div className="flex items-start gap-4 p-5 rounded-lg bg-dark-bg-primary border border-dark-border-subtle">
              <div className="flex-shrink-0 w-8 h-8 bg-primary-500/10 rounded-full flex items-center justify-center text-primary-400 text-sm font-bold">2</div>
              <div>
                <h3 className="font-semibold text-dark-text-primary mb-1">AI discovers available tools</h3>
                <p className="text-sm text-dark-text-tertiary">The MCP server exposes 50+ tools — create projects, manage tasks, assign team members, plan sprints, and more.</p>
              </div>
            </div>
            <div className="flex items-start gap-4 p-5 rounded-lg bg-dark-bg-primary border border-dark-border-subtle">
              <div className="flex-shrink-0 w-8 h-8 bg-primary-500/10 rounded-full flex items-center justify-center text-primary-400 text-sm font-bold">3</div>
              <div>
                <h3 className="font-semibold text-dark-text-primary mb-1">LLM works independently</h3>
                <p className="text-sm text-dark-text-tertiary">Your AI agent breaks down requirements into tasks, assigns priorities, creates sprints, and tracks progress — all without human intervention.</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 md:py-28 border-t border-dark-border-subtle">
        <div className="max-w-3xl mx-auto px-6 text-center">
          <h2 className="text-3xl md:text-4xl font-bold text-dark-text-primary tracking-tight">
            Ready to let AI manage your projects?
          </h2>
          <p className="mt-4 text-base text-dark-text-tertiary">
            Get started for free. Connect your AI agent in minutes.
          </p>
          <div className="mt-8 flex items-center justify-center gap-4">
            <Link
              to="/signup"
              className="inline-flex items-center px-8 py-3 text-base font-medium text-white bg-primary-500 hover:bg-primary-600 rounded-md shadow-linear transition-all duration-150"
            >
              Start building with AI
              <svg className="ml-2 w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
              </svg>
            </Link>
            <a
              href="/api/docs"
              className="inline-flex items-center px-6 py-3 text-base font-medium text-dark-text-secondary border border-dark-border-medium hover:border-dark-border-strong hover:bg-dark-bg-secondary rounded-md transition-all duration-150"
            >
              View API docs
            </a>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-12 border-t border-dark-border-subtle">
        <div className="max-w-6xl mx-auto px-6 flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <img src="/logo.svg" alt="TaskAI" className="w-5 h-5 opacity-60" />
            <span className="text-sm text-dark-text-quaternary">TaskAI</span>
          </div>
          <p className="text-sm text-dark-text-quaternary">
            AI-native project management. Ship with confidence.
          </p>
        </div>
      </footer>
    </div>
  )
}
