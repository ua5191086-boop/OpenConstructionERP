import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, LineChart, Line } from 'recharts'

const TYPE_COLORS = ['#3b82f6','#22c55e','#a855f7','#f97316','#ef4444','#14b8a6','#f59e0b','#ec4899','#6366f1','#84cc16']

export default function RiskPage() {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [projectFilter, setProjectFilter] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [categories, registers, scenarios, mitigations, escalations, monteCarloRuns, dashboard] = await Promise.all([
          fetch('/api/v1/risk/categories').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/risk/registers').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/risk/scenarios').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/risk/mitigations').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/risk/escalations').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/risk/monte-carlo').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/risk/dashboard').then(r => r.ok ? r.json() : { data: [] }),
        ])
        setData({
          categories: categories.data || [],
          registers: registers.data || [],
          scenarios: scenarios.data || [],
          mitigations: mitigations.data || [],
          escalations: escalations.data || [],
          monteCarloRuns: monteCarloRuns.data || [],
          dashboardData: dashboard.data || [],
        })
      } catch {
        setError('API unavailable — start the Go backend on port 8081')
      }
      setLoading(false)
    }
    load()
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Risk data...</div>
  if (error) {
    return (
      <div className="p-8">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">⚠️</div>
          <h2 className="text-xl font-bold text-white mb-2">Risk Management Module</h2>
          <p className="text-[#94a3b8] mb-4">{error}</p>
          <p className="text-sm text-[#64748b]">Start the Go API server on port 8081 to see live data</p>
        </div>
      </div>
    )
  }

  const categories = data.categories || []
  const registers = data.registers || []
  const scenarios = data.scenarios || []
  const mitigations = data.mitigations || []
  const escalations = data.escalations || []
  const monteCarloRuns = data.monteCarloRuns || []
  const dashboardData = data.dashboardData || []

  // Filters
  const projects = [...new Set(registers.map((r: any) => r.project_name).filter(Boolean))] as string[]
  let filteredRegisters = registers
  let filteredScenarios = scenarios
  let filteredMitigations = mitigations
  if (projectFilter) {
    filteredRegisters = filteredRegisters.filter((r: any) => r.project_name === projectFilter)
    filteredScenarios = filteredScenarios.filter((s: any) => s.project_name === projectFilter)
    filteredMitigations = filteredMitigations.filter((m: any) => m.project_name === projectFilter)
  }

  // Stats
  const openRisks = filteredRegisters.filter((r: any) => r.status === 'identified' || r.status === 'analyzed' || r.status === 'response_planned' || r.status === 'monitoring').length
  const extremeRisks = filteredRegisters.filter((r: any) => r.risk_rating === 'extreme').length
  const highRisks = filteredRegisters.filter((r: any) => r.risk_rating === 'high').length
  const activeMitigations = filteredMitigations.filter((m: any) => m.status === 'in_progress' || m.status === 'planned').length

  // Risk rating pie
  const ratingCounts: Record<string, number> = {}
  filteredRegisters.forEach((r: any) => { ratingCounts[r.risk_rating] = (ratingCounts[r.risk_rating] || 0) + 1 })
  const ratingPie = Object.entries(ratingCounts).map(([k, v]) => ({
    name: k.replace('_', ' '),
    value: v,
    color: k === 'extreme' ? '#ef4444' : k === 'high' ? '#f97316' : k === 'medium' ? '#f59e0b' : k === 'low' ? '#22c55e' : '#64748b',
  }))

  // Risk type pie
  const typeCounts: Record<string, number> = {}
  filteredRegisters.forEach((r: any) => { typeCounts[r.risk_type] = (typeCounts[r.risk_type] || 0) + 1 })
  const typePie = Object.entries(typeCounts).map(([k, v]) => ({
    name: k,
    value: v,
    color: k === 'threat' ? '#ef4444' : '#22c55e',
  }))

  // Top risks by score
  const topRisks = [...filteredRegisters].sort((a: any, b: any) => b.risk_score - a.risk_score).slice(0, 10)

  // Mitigation status
  const mitStatusCounts: Record<string, number> = {}
  filteredMitigations.forEach((m: any) => { mitStatusCounts[m.status] = (mitStatusCounts[m.status] || 0) + 1 })
  const mitStatusData = Object.entries(mitStatusCounts).map(([k, v]) => ({ name: k, count: v }))

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">⚠️ Risk Management</h1>
          <p className="text-sm text-[#94a3b8] mt-1">Risk Register, Scenarios, Mitigation, Monte Carlo & Escalations</p>
        </div>
        <select
          className="bg-[#1e293b] border border-[#334155] rounded-lg px-3 py-2 text-sm text-white"
          value={projectFilter}
          onChange={e => setProjectFilter(e.target.value)}
        >
          <option value="">All Projects</option>
          {projects.map(p => <option key={p} value={p}>{p}</option>)}
        </select>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-7 gap-4">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Total Risks</div>
          <div className="text-2xl font-bold text-white">{registers.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Open</div>
          <div className="text-2xl font-bold text-[#f97316]">{openRisks}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Extreme</div>
          <div className="text-2xl font-bold text-[#ef4444]">{extremeRisks}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">High</div>
          <div className="text-2xl font-bold text-[#f97316]">{highRisks}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Mitigations</div>
          <div className="text-2xl font-bold text-[#a855f7]">{mitigations.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Scenarios</div>
          <div className="text-2xl font-bold text-[#14b8a6]">{scenarios.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Escalations</div>
          <div className="text-2xl font-bold text-[#f59e0b]">{escalations.length}</div>
        </div>
      </div>

      {/* Charts Row 1 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Risk Rating Distribution</h3>
          <ResponsiveContainer width="100%" height={240}>
            <PieChart>
              <Pie data={ratingPie} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {ratingPie.map((e, i) => <Cell key={i} fill={e.color} />)}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Threats vs Opportunities</h3>
          <ResponsiveContainer width="100%" height={240}>
            <PieChart>
              <Pie data={typePie} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {typePie.map((e, i) => <Cell key={i} fill={e.color} />)}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Mitigation Status</h3>
          <ResponsiveContainer width="100%" height={240}>
            <PieChart>
              <Pie data={mitStatusData.map(d => ({ ...d, color: d.name === 'completed' ? '#22c55e' : d.name === 'in_progress' ? '#3b82f6' : d.name === 'planned' ? '#f59e0b' : d.name === 'overdue' ? '#ef4444' : '#64748b' }))} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="count" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {mitStatusData.map((e, i) => <Cell key={i} fill={['#22c55e','#3b82f6','#f59e0b','#ef4444','#64748b'][i % 5]} />)}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Risk Score Chart & Monte Carlo */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Top 10 Risks by Score</h3>
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={topRisks} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis type="number" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis type="category" dataKey="risk_code" width={80} tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip />
              <Bar dataKey="risk_score" fill="#ef4444" radius={[0,4,4,0]} name="Risk Score" />
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Monte Carlo Analysis (P10 / P50 / P90)</h3>
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={monteCarloRuns.slice(0, 10)}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="run_label" tick={{ fill: '#94a3b8', fontSize: 10 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip />
              <Bar dataKey="p10_value" fill="#22c55e" name="P10" />
              <Bar dataKey="p50_value" fill="#3b82f6" name="P50" />
              <Bar dataKey="p90_value" fill="#ef4444" name="P90" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Risk Register Table */}
      <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
        <h3 className="text-sm font-semibold text-white mb-4">Risk Register</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-[#64748b] border-b border-[#334155]">
                <th className="text-left py-2 px-3">Code</th>
                <th className="text-left py-2 px-3">Risk Name</th>
                <th className="text-left py-2 px-3">Project</th>
                <th className="text-left py-2 px-3">Type</th>
                <th className="text-left py-2 px-3">Score</th>
                <th className="text-left py-2 px-3">Rating</th>
                <th className="text-left py-2 px-3">Owner</th>
                <th className="text-left py-2 px-3">Response</th>
                <th className="text-left py-2 px-3">Status</th>
              </tr>
            </thead>
            <tbody>
              {topRisks.map((r: any) => (
                <tr key={r.id} className="border-b border-[#1e293b] hover:bg-[#334155]/40">
                  <td className="py-2 px-3 text-white font-mono text-xs">{r.risk_code}</td>
                  <td className="py-2 px-3 text-white">{r.risk_name}</td>
                  <td className="py-2 px-3 text-[#94a3b8]">{r.project_name}</td>
                  <td className="py-2 px-3 text-[#94a3b8]">{r.risk_type}</td>
                  <td className="py-2 px-3 text-white font-medium">{r.risk_score}</td>
                  <td className="py-2 px-3">
                    <span className={`px-2 py-0.5 rounded-full text-xs ${
                      r.risk_rating === 'extreme' ? 'bg-red-500/20 text-red-400' :
                      r.risk_rating === 'high' ? 'bg-orange-500/20 text-orange-400' :
                      r.risk_rating === 'medium' ? 'bg-yellow-500/20 text-yellow-400' :
                      'bg-green-500/20 text-green-400'
                    }`}>{r.risk_rating}</span>
                  </td>
                  <td className="py-2 px-3 text-[#94a3b8]">{r.risk_owner}</td>
                  <td className="py-2 px-3 text-[#94a3b8]">{r.risk_response}</td>
                  <td className="py-2 px-3 text-[#94a3b8]">{r.status}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Scenarios & Escalations */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Scenario Analysis</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-[#64748b] border-b border-[#334155]">
                  <th className="text-left py-2 px-3">Code</th>
                  <th className="text-left py-2 px-3">Name</th>
                  <th className="text-left py-2 px-3">Type</th>
                  <th className="text-left py-2 px-3">Cost ML</th>
                  <th className="text-left py-2 px-3">Sched ML</th>
                  <th className="text-left py-2 px-3">Prob %</th>
                </tr>
              </thead>
              <tbody>
                {filteredScenarios.slice(0, 8).map((s: any) => (
                  <tr key={s.id} className="border-b border-[#1e293b] hover:bg-[#334155]/40">
                    <td className="py-2 px-3 text-white font-mono text-xs">{s.scenario_code}</td>
                    <td className="py-2 px-3 text-white">{s.scenario_name}</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{s.scenario_type}</td>
                    <td className="py-2 px-3 text-[#94a3b8]">${(s.cost_impact_ml / 1000).toFixed(0)}k</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{s.schedule_impact_ml}d</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{s.probability_pct}%</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Escalations</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-[#64748b] border-b border-[#334155]">
                  <th className="text-left py-2 px-3">Code</th>
                  <th className="text-left py-2 px-3">Title</th>
                  <th className="text-left py-2 px-3">Escalated To</th>
                  <th className="text-left py-2 px-3">Decision</th>
                  <th className="text-left py-2 px-3">Status</th>
                </tr>
              </thead>
              <tbody>
                {escalations.slice(0, 8).map((e: any) => (
                  <tr key={e.id} className="border-b border-[#1e293b] hover:bg-[#334155]/40">
                    <td className="py-2 px-3 text-white font-mono text-xs">{e.escalation_code}</td>
                    <td className="py-2 px-3 text-white">{e.title}</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{e.escalated_to}</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{e.decision || '—'}</td>
                    <td className="py-2 px-3">
                      <span className={`px-2 py-0.5 rounded-full text-xs ${
                        e.status === 'closed' ? 'bg-green-500/20 text-green-400' :
                        e.status === 'acknowledged' || e.status === 'responded' ? 'bg-blue-500/20 text-blue-400' :
                        'bg-yellow-500/20 text-yellow-400'
                      }`}>{e.status}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}