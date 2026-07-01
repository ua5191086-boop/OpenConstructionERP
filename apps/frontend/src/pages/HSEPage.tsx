import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, LineChart, Line } from 'recharts'

const TYPE_COLORS = ['#3b82f6','#22c55e','#a855f7','#f97316','#ef4444','#14b8a6','#f59e0b','#ec4899','#6366f1','#84cc16']

export default function HSEPage() {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [projectFilter, setProjectFilter] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [incidents, permits, audits, inspections, training, ppe, drills, statistics, emergencyPlans, chemicals] = await Promise.all([
          fetch('/api/v1/hse/incidents').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/hse/permits').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/hse/audits').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/hse/inspections').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/hse/training').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/hse/ppe').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/hse/drills').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/hse/statistics').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/hse/emergency-plans').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/hse/chemicals').then(r => r.ok ? r.json() : { data: [] }),
        ])
        setData({
          incidents: incidents.data || [],
          permits: permits.data || [],
          audits: audits.data || [],
          inspections: inspections.data || [],
          training: training.data || [],
          ppe: ppe.data || [],
          drills: drills.data || [],
          statistics: statistics.data || [],
          emergencyPlans: emergencyPlans.data || [],
          chemicals: chemicals.data || [],
        })
      } catch {
        setError('API unavailable — start the Go backend on port 8081')
      }
      setLoading(false)
    }
    load()
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading HSE data...</div>
  if (error) {
    return (
      <div className="p-8">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">🛡️</div>
          <h2 className="text-xl font-bold text-white mb-2">HSE Module</h2>
          <p className="text-[#94a3b8] mb-4">{error}</p>
          <p className="text-sm text-[#64748b]">Start the Go API server on port 8081 to see live data</p>
        </div>
      </div>
    )
  }

  const incidents = data.incidents || []
  const permits = data.permits || []
  const audits = data.audits || []
  const training = data.training || []

  // Filters
  const projects = [...new Set(incidents.map((i: any) => i.project_name).filter(Boolean))] as string[]
  let filteredIncidents = incidents
  if (projectFilter) filteredIncidents = filteredIncidents.filter((i: any) => i.project_name === projectFilter)
  const projectNames = new Set(filteredIncidents.map((i: any) => i.project_name))
  const filteredPermits = projectFilter ? permits.filter((p: any) => p.project_name === projectFilter) : permits
  const filteredAudits = projectFilter ? audits.filter((a: any) => a.project_name === projectFilter) : audits
  const filteredTraining = projectFilter ? training.filter((t: any) => t.project_name === projectFilter) : training

  // Stats
  const openInc = filteredIncidents.filter((i: any) => i.status === 'open').length
  const criticalInc = filteredIncidents.filter((i: any) => i.severity === 'critical').length
  const reportable = filteredIncidents.filter((i: any) => i.is_reportable).length
  const activePermits = filteredPermits.filter((p: any) => p.status === 'active' || p.status === 'issued').length

  // Chart data
  const incTypeCounts: Record<string, number> = {}
  filteredIncidents.forEach((i: any) => { const t = i.incident_type || 'other'; incTypeCounts[t] = (incTypeCounts[t] || 0) + 1 })
  const incTypeData = Object.entries(incTypeCounts).map(([name, value]) => ({ name, value }))

  const incSevCounts: Record<string, number> = {}
  filteredIncidents.forEach((i: any) => { const s = i.severity || 'minor'; incSevCounts[s] = (incSevCounts[s] || 0) + 1 })
  const incSevData = Object.entries(incSevCounts).map(([name, value]) => ({ name, value }))

  const monthly: Record<string, number> = {}
  filteredIncidents.forEach((i: any) => {
    const d = i.incident_date || ''
    if (d) { const m = d.substring(0, 7); monthly[m] = (monthly[m] || 0) + 1 }
  })
  const trendData = Object.entries(monthly).sort(([a], [b]) => a.localeCompare(b)).map(([name, value]) => ({ name, value }))

  const permitCounts: Record<string, number> = {}
  filteredPermits.forEach((p: any) => { const s = p.status || 'unknown'; permitCounts[s] = (permitCounts[s] || 0) + 1 })
  const permitData = Object.entries(permitCounts).map(([name, value]) => ({ name, value }))

  const trainingByType: Record<string, number> = {}
  filteredTraining.forEach((t: any) => { const tp = t.training_type || 'other'; trainingByType[tp] = (trainingByType[tp] || 0) + 1 })
  const trainingData = Object.entries(trainingByType).map(([name, value]) => ({ name, value }))

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">🛡️ HSE Dashboard</h1>
        <p className="text-[#94a3b8] mt-1">Health, Safety & Environment — Incidents, Permits, Audits, Training</p>
      </div>

      {/* Project filter */}
      <div className="flex gap-3 mb-6">
        <select className="bg-[#0f172a] text-[#e2e8f0] border border-[#334155] rounded-lg px-3 py-1.5 text-sm"
          value={projectFilter} onChange={e => setProjectFilter(e.target.value)}>
          <option value="">All Projects</option>
          {projects.map(p => <option key={p} value={p}>{p}</option>)}
        </select>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-6 gap-4 mb-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Incidents</div>
          <div className="text-2xl font-bold text-[#3b82f6] mt-1">{filteredIncidents.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Open</div>
          <div className="text-2xl font-bold text-[#ef4444] mt-1">{openInc}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Critical</div>
          <div className="text-2xl font-bold text-[#ef4444] mt-1">{criticalInc}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Reportable</div>
          <div className="text-2xl font-bold text-[#f97316] mt-1">{reportable}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Permits</div>
          <div className="text-2xl font-bold text-[#22c55e] mt-1">{filteredPermits.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Active Permits</div>
          <div className="text-2xl font-bold text-[#a855f7] mt-1">{activePermits}</div>
        </div>
      </div>

      {/* Charts Row 1 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Incidents by Type</h3>
          <ResponsiveContainer width="100%" height={260}>
            <PieChart>
              <Pie data={incTypeData} cx="50%" cy="50%" outerRadius={90} dataKey="value"
                label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {incTypeData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
              </Pie>
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Severity Distribution</h3>
          <ResponsiveContainer width="100%" height={260}>
            <BarChart data={incSevData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
              <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                {incSevData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Monthly Incident Trend</h3>
          <ResponsiveContainer width="100%" height={260}>
            <LineChart data={trendData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
              <Line type="monotone" dataKey="value" stroke="#ef4444" strokeWidth={2} dot={{ fill: '#ef4444', r: 4 }} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Charts Row 2 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Permits by Status</h3>
          <ResponsiveContainer width="100%" height={280}>
            <PieChart>
              <Pie data={permitData} cx="50%" cy="50%" outerRadius={100} dataKey="value"
                label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {permitData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
              </Pie>
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Training by Type</h3>
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={trainingData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
              <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                {trainingData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Incidents Table */}
      <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden mb-6">
        <div className="p-4 border-b border-[#334155]">
          <h3 className="text-white font-semibold">⚠️ Incidents ({filteredIncidents.length})</h3>
        </div>
        <div className="overflow-x-auto max-h-96 overflow-y-auto">
          <table className="w-full">
            <thead>
              <tr className="text-[#64748b] text-xs uppercase tracking-wider">
                <th className="text-left p-3 border-b border-[#334155]">Code</th>
                <th className="text-left p-3 border-b border-[#334155]">Title</th>
                <th className="text-left p-3 border-b border-[#334155]">Type</th>
                <th className="text-left p-3 border-b border-[#334155]">Severity</th>
                <th className="text-left p-3 border-b border-[#334155]">Date</th>
                <th className="text-left p-3 border-b border-[#334155]">Location</th>
                <th className="text-left p-3 border-b border-[#334155]">Cost</th>
                <th className="text-left p-3 border-b border-[#334155]">Status</th>
              </tr>
            </thead>
            <tbody>
              {filteredIncidents.slice(0, 100).map((i: any) => (
                <tr key={i.id} className="hover:bg-[#334155] border-b border-[#1e293b]">
                  <td className="p-3 font-mono text-xs text-[#64748b]">{i.incident_code}</td>
                  <td className="p-3 text-sm text-[#e2e8f0] max-w-[200px] truncate">{i.title}</td>
                  <td className="p-3 text-xs text-[#94a3b8]">{i.incident_type}</td>
                  <td className="p-3">
                    <span className={`px-2 py-0.5 rounded text-xs font-semibold ${
                      i.severity === 'critical' ? 'bg-[#7f1d1d] text-[#fca5a5]' :
                      i.severity === 'major' ? 'bg-[#9a3412] text-[#fdba74]' :
                      i.severity === 'moderate' ? 'bg-[#854d0e] text-[#fde047]' :
                      'bg-[#1e3a2f] text-[#86efac]'
                    }`}>{i.severity}</span>
                  </td>
                  <td className="p-3 text-xs text-[#94a3b8]">{i.incident_date || '—'}</td>
                  <td className="p-3 text-xs text-[#94a3b8]">{i.location || '—'}</td>
                  <td className="p-3 text-sm">${(i.total_cost || 0).toLocaleString()}</td>
                  <td className="p-3">
                    <span className={`px-2 py-0.5 rounded text-xs font-semibold ${
                      i.status === 'closed' ? 'bg-[#1e2f3b] text-[#38bdf8]' : 'bg-[#3b1e1e] text-[#f87171]'
                    }`}>{i.status}</span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Permits + Audits summary */}
      {filteredPermits.length > 0 && (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden mb-6">
          <div className="p-4 border-b border-[#334155]">
            <h3 className="text-white font-semibold">📋 Permits ({filteredPermits.length})</h3>
          </div>
          <div className="overflow-x-auto max-h-64 overflow-y-auto">
            <table className="w-full">
              <thead>
                <tr className="text-[#64748b] text-xs uppercase tracking-wider">
                  <th className="text-left p-3 border-b border-[#334155]">Permit #</th>
                  <th className="text-left p-3 border-b border-[#334155]">Type</th>
                  <th className="text-left p-3 border-b border-[#334155]">Status</th>
                  <th className="text-left p-3 border-b border-[#334155]">Issued</th>
                  <th className="text-left p-3 border-b border-[#334155]">Expires</th>
                  <th className="text-left p-3 border-b border-[#334155]">Area</th>
                </tr>
              </thead>
              <tbody>
                {filteredPermits.slice(0, 50).map((p: any) => (
                  <tr key={p.id} className="hover:bg-[#334155] border-b border-[#1e293b]">
                    <td className="p-3 font-mono text-xs text-[#64748b]">{p.permit_number || p.id}</td>
                    <td className="p-3 text-xs text-[#94a3b8]">{p.permit_type}</td>
                    <td className="p-3"><span className="px-2 py-0.5 rounded text-xs font-semibold bg-[#1e3a2f] text-[#34d399]">{p.status}</span></td>
                    <td className="p-3 text-xs text-[#94a3b8]">{p.issue_date || '—'}</td>
                    <td className="p-3 text-xs text-[#94a3b8]">{p.expiry_date || '—'}</td>
                    <td className="p-3 text-xs text-[#94a3b8]">{p.work_area || '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Training summary */}
      {filteredTraining.length > 0 && (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
          <div className="p-4 border-b border-[#334155]">
            <h3 className="text-white font-semibold">🎓 Training ({filteredTraining.length})</h3>
          </div>
          <div className="overflow-x-auto max-h-64 overflow-y-auto">
            <table className="w-full">
              <thead>
                <tr className="text-[#64748b] text-xs uppercase tracking-wider">
                  <th className="text-left p-3 border-b border-[#334155]">Title</th>
                  <th className="text-left p-3 border-b border-[#334155]">Type</th>
                  <th className="text-left p-3 border-b border-[#334155]">Date</th>
                  <th className="text-left p-3 border-b border-[#334155]">Instructor</th>
                  <th className="text-left p-3 border-b border-[#334155]">Attendees</th>
                  <th className="text-left p-3 border-b border-[#334155]">Status</th>
                </tr>
              </thead>
              <tbody>
                {filteredTraining.slice(0, 50).map((t: any) => (
                  <tr key={t.id} className="hover:bg-[#334155] border-b border-[#1e293b]">
                    <td className="p-3 text-sm text-[#e2e8f0] max-w-[200px] truncate">{t.title || t.training_title || '—'}</td>
                    <td className="p-3 text-xs text-[#94a3b8]">{t.training_type || '—'}</td>
                    <td className="p-3 text-xs text-[#94a3b8]">{t.training_date || t.date || '—'}</td>
                    <td className="p-3 text-xs text-[#94a3b8]">{t.instructor || t.trainer || '—'}</td>
                    <td className="p-3 text-sm">{t.attendee_count || t.participants || 0}</td>
                    <td className="p-3"><span className="px-2 py-0.5 rounded text-xs font-semibold bg-[#1e3a2f] text-[#34d399]">{t.status}</span></td>
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