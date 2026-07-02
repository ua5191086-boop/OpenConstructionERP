// @ts-nocheck
import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

const STATUS_COLORS: Record<string, string> = {
  lead: '#64748b', tender: '#f97316', planning: '#a855f7', design: '#fbbf24',
  construction: '#3b82f6', commissioning: '#14b8a6', operation: '#22c55e', closed: '#64748b',
}

const TYPE_COLORS = ['#3b82f6','#22c55e','#a855f7','#f97316','#ef4444','#14b8a6','#f59e0b']

export default function PMProjectPage() {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [typeFilter, setTypeFilter] = useState('')
  const [searchFilter, setSearchFilter] = useState('')
  const [selectedProject, setSelectedProject] = useState<string | null>(null)

  useEffect(() => {
    async function load() {
      try {
        const [projects, wbs, milestones, phases, risks, changes, lessons, portfolios] = await Promise.all([
          fetch('/api/pm/projects').then(r => r.ok ? r.json() : []),
          fetch('/api/pm/wbs-items').then(r => r.ok ? r.json() : []),
          fetch('/api/pm/milestones').then(r => r.ok ? r.json() : []),
          fetch('/api/pm/phases').then(r => r.ok ? r.json() : []),
          fetch('/api/pm/risks').then(r => r.ok ? r.json() : []),
          fetch('/api/pm/changes').then(r => r.ok ? r.json() : []),
          fetch('/api/pm/lessons').then(r => r.ok ? r.json() : []),
          fetch('/api/pm/portfolios').then(r => r.ok ? r.json() : []),
        ])
        setData({ projects, wbs, milestones, phases, risks, changes, lessons, portfolios })
      } catch {
        setError('API unavailable — start the Go backend on port 8081')
      }
      setLoading(false)
    }
    load()
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Project Management data...</div>
  if (error) {
    return (
      <div className="p-8">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">📁</div>
          <h2 className="text-xl font-bold text-white mb-2">Project Management Module</h2>
          <p className="text-[#94a3b8] mb-4">{error}</p>
          <p className="text-sm text-[#64748b]">Start the Go API server on port 8081 to see live data</p>
        </div>
      </div>
    )
  }

  const projects = data.projects || []
  const wbs = data.wbs || []
  const milestones = data.milestones || []
  const phases = data.phases || []
  const risks = data.risks || []
  const changes = data.changes || []
  const lessons = data.lessons || []

  // Filters
  let filtered = [...projects]
  if (statusFilter) filtered = filtered.filter((p: any) => p.status === statusFilter)
  if (typeFilter) filtered = filtered.filter((p: any) => p.project_type === typeFilter)
  if (searchFilter) filtered = filtered.filter((p: any) =>
    (p.name || '').toLowerCase().includes(searchFilter.toLowerCase()) ||
    (p.code || '').toLowerCase().includes(searchFilter.toLowerCase())
  )

  const projectIds = new Set(filtered.map((p: any) => p.id))
  const filteredRisks = risks.filter((r: any) => projectIds.has(r.project_id))
  const filteredMilestones = milestones.filter((m: any) => projectIds.has(m.project_id))

  // Stats
  const totalBudget = filtered.reduce((s: number, p: any) => s + (p.budget_total || 0), 0)
  const activeProjects = filtered.filter((p: any) => !['closed', 'lead'].includes(p.status)).length

  // Budget chart
  const budgetChartData = filtered.map((p: any) => ({
    name: p.code || p.id,
    value: p.budget_total || 0,
  }))

  // Risk chart
  const riskByCat: Record<string, number> = {}
  filteredRisks.forEach((r: any) => {
    const cat = r.risk_category || 'other'
    riskByCat[cat] = (riskByCat[cat] || 0) + 1
  })
  const riskChartData = Object.entries(riskByCat).map(([name, value]) => ({ name, value }))

  // WBS tree for selected project
  const selectedWBS = selectedProject
    ? wbs.filter((w: any) => w.project_id === parseInt(selectedProject))
    : []

  // Build WBS tree
  function buildWBSTree(items: any[], parentId: number | null = null, depth = 0): any[] {
    return items
      .filter((i: any) => i.parent_id === parentId)
      .sort((a: any, b: any) => (a.sort_order || 0) - (b.sort_order || 0))
      .map((i: any) => ({
        ...i,
        children: buildWBSTree(items, i.id, depth + 1),
      }))
  }

  const wbsTree = buildWBSTree(selectedWBS, null)

  function renderWBSNode(node: any, depth = 0): JSX.Element {
    const progressColor = node.progress_pct > 80 ? '#22c55e' : node.progress_pct > 40 ? '#3b82f6' : node.progress_pct > 0 ? '#f97316' : '#64748b'
    return (
      <div key={node.id} style={{ marginLeft: depth * 20 }}>
        <div className="flex items-center gap-2 py-1.5 px-2 rounded hover:bg-[#334155] cursor-pointer">
          <span className="font-mono text-xs text-[#64748b] min-w-[50px]">{node.wbs_code}</span>
          <span className="text-sm text-[#e2e8f0]">{node.name}</span>
          <span className="ml-auto text-xs text-[#94a3b8]">{node.progress_pct}%</span>
        </div>
        <div className="h-1.5 bg-[#334155] rounded-full overflow-hidden mx-2 mb-1">
          <div className="h-full rounded-full transition-all" style={{ width: `${node.progress_pct}%`, background: progressColor }} />
        </div>
        {node.children?.map((child: any) => renderWBSNode(child, depth + 1))}
      </div>
    )
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">📁 Project Management</h1>
        <p className="text-[#94a3b8] mt-1">Portfolio & Project Management Dashboard</p>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-3 mb-6">
        <select
          className="bg-[#0f172a] text-[#e2e8f0] border border-[#334155] rounded-lg px-3 py-1.5 text-sm"
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
        >
          <option value="">All Statuses</option>
          {Object.keys(STATUS_COLORS).map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
        <select
          className="bg-[#0f172a] text-[#e2e8f0] border border-[#334155] rounded-lg px-3 py-1.5 text-sm"
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
        >
          <option value="">All Types</option>
          <option value="metro">Metro</option>
          <option value="tunnel">Tunnel</option>
          <option value="bridge">Bridge</option>
          <option value="road">Road</option>
          <option value="building">Building</option>
          <option value="industrial">Industrial</option>
          <option value="water">Water</option>
        </select>
        <input
          type="text"
          placeholder="Search project..."
          className="bg-[#0f172a] text-[#e2e8f0] border border-[#334155] rounded-lg px-3 py-1.5 text-sm flex-1 min-w-[200px]"
          value={searchFilter}
          onChange={(e) => setSearchFilter(e.target.value)}
        />
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Projects</div>
          <div className="text-2xl font-bold text-[#3b82f6] mt-2">{filtered.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Total Budget</div>
          <div className="text-2xl font-bold text-[#22c55e] mt-2">${(totalBudget / 1e6).toFixed(1)}M</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Active</div>
          <div className="text-2xl font-bold text-[#eab308] mt-2">{activeProjects}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Risks</div>
          <div className="text-2xl font-bold text-[#ef4444] mt-2">{filteredRisks.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Milestones</div>
          <div className="text-2xl font-bold text-[#a855f7] mt-2">{filteredMilestones.length}</div>
        </div>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
          <h3 className="text-white font-semibold mb-4">Budget by Project</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={budgetChartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} tickFormatter={(v) => `$${(v / 1e6).toFixed(0)}M`} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
              <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                {budgetChartData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
          <h3 className="text-white font-semibold mb-4">Risk Distribution</h3>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie data={riskChartData} cx="50%" cy="50%" outerRadius={100} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {riskChartData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
              </Pie>
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Project Cards */}
      <div className="space-y-4 mb-6">
        <h3 className="text-white font-semibold text-lg">Projects</h3>
        {filtered.map((project: any) => {
          const pPhases = phases.filter((ph: any) => ph.project_id === project.id)
          const pRisks = risks.filter((r: any) => r.project_id === project.id)
          const pMilestones = milestones.filter((m: any) => m.project_id === project.id)
          const avgProgress = pPhases.length > 0
            ? Math.round(pPhases.reduce((s: number, ph: any) => s + (ph.completion_pct || 0), 0) / pPhases.length)
            : 0
          const highRisks = pRisks.filter((r: any) => (r.risk_score || 0) >= 12).length
          const progressColor = avgProgress > 80 ? '#22c55e' : avgProgress > 40 ? '#3b82f6' : avgProgress > 0 ? '#f97316' : '#64748b'

          return (
            <div
              key={project.id}
              className={`bg-[#1e293b] border rounded-xl p-5 cursor-pointer transition-all ${
                selectedProject === String(project.id) ? 'border-[#3b82f6]' : 'border-[#334155]'
              }`}
              onClick={() => setSelectedProject(selectedProject === String(project.id) ? null : String(project.id))}
            >
              <div className="flex justify-between items-start mb-3">
                <div>
                  <h4 className="text-white font-semibold">{project.name}</h4>
                  <span className="text-xs text-[#64748b] font-mono">{project.code}</span>
                </div>
                <span
                  className="px-2 py-0.5 rounded text-xs font-semibold"
                  style={{
                    background: `${STATUS_COLORS[project.status] || '#64748b'}20`,
                    color: STATUS_COLORS[project.status] || '#64748b',
                  }}
                >
                  {project.status}
                </span>
              </div>
              <div className="flex flex-wrap gap-4 text-xs text-[#94a3b8] mb-3">
                <span>🏗️ {project.project_type}</span>
                <span>💰 ${(project.budget_total / 1e6).toFixed(1)}M</span>
                <span>⚠️ {highRisks} high risks</span>
                <span>🏁 {pMilestones.length} milestones</span>
              </div>
              <div className="flex justify-between text-xs text-[#94a3b8] mb-1">
                <span>Progress</span>
                <span>{avgProgress}%</span>
              </div>
              <div className="h-1.5 bg-[#334155] rounded-full overflow-hidden">
                <div className="h-full rounded-full transition-all" style={{ width: `${avgProgress}%`, background: progressColor }} />
              </div>

              {/* WBS Tree (expandable) */}
              {selectedProject === String(project.id) && (
                <div className="mt-4 pt-4 border-t border-[#334155]">
                  <h5 className="text-sm text-white font-medium mb-3">🌳 WBS — Work Breakdown Structure</h5>
                  {wbsTree.length > 0 ? (
                    <div className="max-h-80 overflow-y-auto">
                      {wbsTree.map((node: any) => renderWBSNode(node))}
                    </div>
                  ) : (
                    <p className="text-xs text-[#64748b]">No WBS data available</p>
                  )}

                  {/* Milestones */}
                  {pMilestones.length > 0 && (
                    <div className="mt-4">
                      <h5 className="text-sm text-white font-medium mb-2">🏁 Milestones</h5>
                      <div className="space-y-1">
                        {pMilestones.slice(0, 5).map((m: any) => (
                          <div key={m.id} className="flex items-center gap-2 text-xs">
                            <span className="font-mono text-[#64748b]">{m.milestone_code}</span>
                            <span className="text-[#e2e8f0]">{m.name}</span>
                            <span className="ml-auto" style={{ color: m.status === 'achieved' ? '#22c55e' : m.status === 'delayed' ? '#ef4444' : '#64748b' }}>
                              {m.status}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Risks */}
                  {pRisks.length > 0 && (
                    <div className="mt-4">
                      <h5 className="text-sm text-white font-medium mb-2">⚠️ Top Risks</h5>
                      <div className="space-y-1">
                        {pRisks.sort((a: any, b: any) => (b.risk_score || 0) - (a.risk_score || 0)).slice(0, 3).map((r: any) => (
                          <div key={r.id} className="flex items-center gap-2 text-xs">
                            <span className="font-mono text-[#64748b]">{r.risk_code}</span>
                            <span className="text-[#e2e8f0]">{r.name}</span>
                            <span className="ml-auto font-bold" style={{ color: r.risk_score >= 12 ? '#ef4444' : r.risk_score >= 8 ? '#f97316' : '#22c55e' }}>
                              Score: {r.risk_score}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          )
        })}
      </div>

      {/* Changes Table */}
      {changes.length > 0 && (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden mb-6">
          <div className="p-4 border-b border-[#334155]">
            <h3 className="text-white font-semibold">📝 Changes & Variations</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                  <th className="p-3 pl-4">#</th>
                  <th className="p-3">Type</th>
                  <th className="p-3">Source</th>
                  <th className="p-3">Description</th>
                  <th className="p-3 text-right">Cost Impact</th>
                  <th className="p-3 text-right">Schedule</th>
                  <th className="p-3 pr-4">Status</th>
                </tr>
              </thead>
              <tbody>
                {changes.filter((c: any) => projectIds.has(c.project_id)).map((c: any) => (
                  <tr key={c.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                    <td className="p-3 pl-4 font-mono text-xs text-[#64748b]">{c.change_number}</td>
                    <td className="p-3 text-[#94a3b8] text-xs">{c.change_type}</td>
                    <td className="p-3 text-[#94a3b8] text-xs">{c.source}</td>
                    <td className="p-3 text-white">{c.description}</td>
                    <td className={`p-3 text-right ${(c.cost_impact || 0) >= 0 ? 'text-[#22c55e]' : 'text-[#ef4444]'}`}>
                      ${(c.cost_impact || 0).toLocaleString()}
                    </td>
                    <td className="p-3 text-right text-[#94a3b8]">
                      {(c.schedule_impact || 0) > 0 ? '+' : ''}{c.schedule_impact || 0}d
                    </td>
                    <td className="p-3 pr-4 text-[#94a3b8] text-xs">{c.status}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Lessons Learned */}
      {lessons.length > 0 && (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
          <div className="p-4 border-b border-[#334155]">
            <h3 className="text-white font-semibold">📚 Lessons Learned</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                  <th className="p-3 pl-4">Category</th>
                  <th className="p-3">Title</th>
                  <th className="p-3">Type</th>
                  <th className="p-3">Severity</th>
                  <th className="p-3 pr-4">Status</th>
                </tr>
              </thead>
              <tbody>
                {lessons.filter((l: any) => projectIds.has(l.project_id)).map((l: any) => (
                  <tr key={l.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                    <td className="p-3 pl-4 text-[#94a3b8] text-xs">{l.category}</td>
                    <td className="p-3 text-white">{l.title}</td>
                    <td className="p-3">
                      <span className={`text-xs font-semibold ${l.is_positive ? 'text-[#22c55e]' : 'text-[#ef4444]'}`}>
                        {l.is_positive ? '✅ Success' : '⚠️ Lesson'}
                      </span>
                    </td>
                    <td className="p-3 text-[#94a3b8] text-xs">{l.severity}</td>
                    <td className="p-3 pr-4 text-[#94a3b8] text-xs">{l.status}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
