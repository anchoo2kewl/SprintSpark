import { Link } from 'react-router-dom'

export default function Landing() {
  return (
    <div className="min-h-screen bg-dark-bg-base">
      {/* Navigation */}
      <nav className="fixed top-0 left-0 right-0 z-50 bg-dark-bg-base/80 backdrop-blur-lg border-b border-dark-border-subtle">
        <div className="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2.5">
            <img src="/logo.svg" alt="SprintSpark" className="w-7 h-7" />
            <span className="text-base font-semibold text-dark-text-primary tracking-tight">SprintSpark</span>
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
          <h1 className="text-5xl md:text-6xl font-bold text-dark-text-primary tracking-tight">
            Build better products,
            <br />
            <span className="bg-gradient-to-r from-primary-400 to-secondary-400 bg-clip-text text-transparent">
              ship faster
            </span>
          </h1>
          <p className="mt-6 text-lg md:text-xl text-dark-text-tertiary max-w-2xl mx-auto leading-relaxed">
            SprintSpark helps teams organize, track, and ship projects with speed and clarity.
            Streamline your workflow with powerful project management.
          </p>
          <div className="mt-10 flex items-center justify-center gap-4">
            <Link
              to="/signup"
              className="inline-flex items-center px-6 py-3 text-base font-medium text-white bg-primary-500 hover:bg-primary-600 rounded-md shadow-linear transition-all duration-150"
            >
              Get started free
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
              Everything you need to ship
            </h2>
            <p className="mt-4 text-base text-dark-text-tertiary max-w-xl mx-auto">
              Powerful tools designed for modern teams who want to move fast without losing clarity.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {/* Feature 1: Kanban */}
            <div className="group p-6 rounded-xl bg-dark-bg-primary border border-dark-border-subtle hover:border-dark-border-medium transition-all duration-200">
              <div className="w-12 h-12 bg-primary-500/10 rounded-lg flex items-center justify-center mb-4 group-hover:bg-primary-500/15 transition-colors duration-200">
                <svg className="w-6 h-6 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-dark-text-primary mb-2 tracking-tight">Kanban Boards</h3>
              <p className="text-sm text-dark-text-tertiary leading-relaxed">
                Drag-and-drop task management with customizable swim lanes. See your work flow from start to finish.
              </p>
            </div>

            {/* Feature 2: Real-time Sync */}
            <div className="group p-6 rounded-xl bg-dark-bg-primary border border-dark-border-subtle hover:border-dark-border-medium transition-all duration-200">
              <div className="w-12 h-12 bg-success-500/10 rounded-lg flex items-center justify-center mb-4 group-hover:bg-success-500/15 transition-colors duration-200">
                <svg className="w-6 h-6 text-success-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-dark-text-primary mb-2 tracking-tight">Real-time Sync</h3>
              <p className="text-sm text-dark-text-tertiary leading-relaxed">
                Offline-first architecture with instant sync. Your work is always saved, even without a connection.
              </p>
            </div>

            {/* Feature 3: Sprint Planning */}
            <div className="group p-6 rounded-xl bg-dark-bg-primary border border-dark-border-subtle hover:border-dark-border-medium transition-all duration-200">
              <div className="w-12 h-12 bg-secondary-500/10 rounded-lg flex items-center justify-center mb-4 group-hover:bg-secondary-500/15 transition-colors duration-200">
                <svg className="w-6 h-6 text-secondary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-dark-text-primary mb-2 tracking-tight">Sprint Planning</h3>
              <p className="text-sm text-dark-text-tertiary leading-relaxed">
                Plan sprints, set milestones, and track progress. Keep your team aligned and shipping on schedule.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 md:py-28 border-t border-dark-border-subtle">
        <div className="max-w-3xl mx-auto px-6 text-center">
          <h2 className="text-3xl md:text-4xl font-bold text-dark-text-primary tracking-tight">
            Ready to ship faster?
          </h2>
          <p className="mt-4 text-base text-dark-text-tertiary">
            Get started for free. No credit card required.
          </p>
          <div className="mt-8">
            <Link
              to="/signup"
              className="inline-flex items-center px-8 py-3 text-base font-medium text-white bg-primary-500 hover:bg-primary-600 rounded-md shadow-linear transition-all duration-150"
            >
              Start building today
              <svg className="ml-2 w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
              </svg>
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-12 border-t border-dark-border-subtle">
        <div className="max-w-6xl mx-auto px-6 flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <img src="/logo.svg" alt="SprintSpark" className="w-5 h-5 opacity-60" />
            <span className="text-sm text-dark-text-quaternary">SprintSpark</span>
          </div>
          <p className="text-sm text-dark-text-quaternary">
            Built with care. Ship with confidence.
          </p>
        </div>
      </footer>
    </div>
  )
}
