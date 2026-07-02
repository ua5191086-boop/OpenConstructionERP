// @ts-nocheck
import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, LineChart, Line } from 'recharts'

const TYPE_COLORS = ['#3b82f6','#22c55e','#a855f7','#f97316','#ef4444','#14b8a6','#f59e0b','#ec4899','#6366f1','#84cc16']

export default function SchedulePage() {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [projectFilter, setProjectFilter] = useState('')
  const [selectedSchedule, setSelectedSchedule] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [schedules, activities, relationships, resources, baselines, changes] = await Promise.all([
          fetch('/api/v1/schedule/schedules').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/schedule/activities').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/schedule/relationships').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/schedule/resources').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/schedule/baselines').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/schedule/changes').then(r => r.ok ? r.json() : { data: [] }),
        ])
        setData({
          schedules: schedules.data || [],
          activities: activities.data || [],
          relationships: relationships.data || [],
          resources: resources.data || [],
          baselines: baselines.data || [],
          changes: changes.data || [],
        })
      } catch {
        setError('API unavailable — start the Go backend on port 8081')
      }
      setLoading(false)
    }
    load()
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Schedule data...</div>
  if (error) {
    return (
      <div className="p-8">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">📅</div>
          <h2 className="text-xl font-bold text-white mb-2">Schedule Management</h2>
          <p className="text-[#94a3b8] mb-4">{error}</p>
          <p className="text-sm text-[#64748b]">Start the Go API server on port 8081 to see live data</p>
        </div>
      </div>
    )
  }

  const schedules = data.schedules || []
  const activities = data.activities || []
  const resources = data.resources || []

  // Filters
  const projects = [...new Set(schedules.map((s: any) => s.project_name).filter(Boolean))] as string[]
  let filteredScheds = schedules
  if (projectFilter) filteredScheds = filteredScheds.filter((s: any) => s.project_name === projectFilter)
  const schedIds = new Set(filteredScheds.map((s: any) => s.id))
  const filteredActs = activities.filter((a: any) => schedIds.has(a.schedule_id))
  const filteredRes = resources.filter((r: any) => schedIds.has(r.schedule_id))

  // Stats
  const totalActs = filteredActs.length
  const criticalActs = filteredActs.filter((a: any) => a.is_critical).length
  const completedActs = filteredActs.filter((a: any) => a.status === 'completed').length
  const inProgActs = filteredActs.filter((a: any) => a.status === 'in_progress').length
  const avgFloat = filteredActs.length > 0
    ? (filteredActs.reduce((s: number, a: any) => s + (a.float_total || 0), 0) / filteredActs.length).toFixed(1)
    : '0'

  // Chart data
  const actStatusCounts: Record<string, number> = {}
  filteredActs.forEach((a: any) => { const s = a.status || 'not_started'; actStatusCounts[s] = (actStatusCounts[s] || 0) + 1 })
  const actStatusData = Object.entries(actStatusCounts).map(([name, value]) => ({ name, value }))

  const schedStatusCounts: Record<string, number> = {}
  filteredScheds.forEach((s: any) => { const st = s.status || 'unknown'; schedStatusCounts[st] = (schedStatusCounts[st] || 0) + 1 })
  const schedStatusData = Object.entries(schedStatusCounts).map(([name, value]) => ({ name, value }))

  const floatData = filteredScheds.map((s: any) => ({
    name: (s.schedule_code || '').substring(0, 12),
    value: s.total_float_pct || 0,
  }))

  const ganttData = [...filteredActs.filter((a: any) => a.is_critical).slice(0, 10), ...filteredActs.filter((a: any) => !a.is_critical).slice(0, 10)]

  const resByType: Record<string, number> = {}
  filteredRes.forEach((r: any) => { const t = r.resource_type || 'other'; resByType[t] = (resByType[t] || 0) + (r.quantity || 1) })
  const resData = Object.entries(resByType).map(([name, value]) => ({ name, value }))

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">📅 Schedule Management</h1>
        <p className="text-[#94a3b8] mt-1">Construction Project Schedules — Activities, Critical Path, Resource Loading, Baselines</p>
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
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Schedules</div>
          <div className="text-2xl font-bold text-[#3b82f6] mt-1">{filteredScheds.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Activities</div>
          <div className="text-2xl font-bold text-[#22c55e] mt-1">{totalActs}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Critical Path</div>
          <div className="text-2xl font-bold text-[#ef4444] mt-1">{criticalActs}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Completed</div>
          <div className="text-2xl font-bold text-[#a855f7] mt-1">{completedActs}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">In Progress</div>
          <div className="text-2xl font-bold text-[#eab308] mt-1">{inProgActs}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Avg Float (d)</div>
          <div className="text-2xl font-bold text-[#f97316] mt-1">{avgFloat}</div>
        </div>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Schedule Status</h3>
          <ResponsiveContainer width="100%" height={260}>
            <PieChart>
              <Pie data={schedStatusData} cx="50%" cy="50%" outerRadius={90} dataKey="value"
                label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {schedStatusData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
              </Pie>
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Activity Status</h3>
          <ResponsiveContainer width="100%" height={260}>
            <BarChart data={actStatusData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
              <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                {actStatusData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Total Float %</h3>
          <ResponsiveContainer width="100%" height={260}>
            <BarChart data={floatData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 10 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} tickFormatter={v => v + '%'} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
              <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                {floatData.map((_, i) => <Cell key={i} fill={floatData[i].value < 3 ? '#ef4444' : floatData[i].value < 6 ? '#f97316' : '#22c55e'} />)}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Critical Path Gantt (Duration in days)</h3>
          <ResponsiveContainer width="100%" height={350}>
            <BarChart data={ganttData} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis type="number" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis dataKey="activity_name" type="category" tick={{ fill: '#94a3b8', fontSize: 10 }} width={180} tickFormatter={v => v.substring(0, 20)} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} formatter={(v: any, _: any, props: any) => [`${v}d — ${props.payload.activity_name}`, 'Duration']} />
              <Bar dataKey="original_duration" radius={[0, 4, 4, 0]}>
                {ganttData.map((a: any) => <Cell key={a.id} fill={a.is_critical ? '#ef4444' : a.status === 'completed' ? '#22c55e' : a.status === 'in_progress' ? '#3b82f6' : '#64748b'} />)}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Resource Loading</h3>
          <ResponsiveContainer width="100%" height={350}>
            <PieChart>
              <Pie data={resData} cx="50%" cy="50%" outerRadius={110} dataKey="value"
                label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {resData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
              </Pie>
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Activities Table */}
      <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden mb-6">
        <div className="p-4 border-b border-[#334155]">
          <h3 className="text-white font-semibold">📋 Activities ({filteredActs.length})</h3>
        </div>
        <div className="overflow-x-auto max-h-96 overflow-y-auto">
          <table className="w-full">
            <thead>
              <tr className="text-[#64748b] text-xs uppercase tracking-wider">
                <th className="text-left p-3 border-b border-[#334155]">Activity ID</th>
                <th className="text-left p-3 border-b border-[#334155]">Name</th>
                <th className="text-left p-3 border-b border-[#334155]">WBS</th>
                <th className="text-left p-3 border-b border-[#334155]">Type</th>
                <th className="text-left p-3 border-b border-[#334155]">Duration</th>
                <th className="text-left p-3 border-b border-[#334155]">%</th>
                <th className="text-left p-3 border-b border-[#334155]">Float</th>
                <th className="text-left p-3 border-b border-[#334155]">Critical</th>
                <th className="text-left p-3 border-b border-[#334155]">Status</th>
              </tr>
            </thead>
            <tbody>
              {filteredActs.slice(0, 100).map((a: any) => (
                <tr key={a.id} className="hover:bg-[#334155] border-b border-[#1e293b]">
                  <td className="p-3 font-mono text-xs text-[#64748b]">{a.activity_id}</td>
                  <td className="p-3 text-sm text-[#e2e8f0] max-w-[200px] truncate">{a.activity_name}</td>
                  <td className="p-3 text-xs text-[#94a3b8]">{a.wbs_code || '—'}</td>
                  <td className="p-3 text-xs text-[#94a3b8]">{a.activity_type}</td>
                  <td className="p-3 text-sm">{a.original_duration}d</td>
                  <td className="p-3 text-sm">{a.percent_complete}%</td>
                  <td className="p-3 text-sm">{a.float_total}</td>
                  <td className="p-3">{a.is_critical ? <span className="px-2 py-0.5 rounded text-xs font-semibold bg-[#3b1e1e] text-[#f87171]">CRITICAL</span> : '—'}</td>
                  <td className="p-3"><span className={`px-2 py-0.5 rounded text-xs font-semibold ${a.status === 'completed' ? 'bg-[#1e3a2f] text-[#22c55e]' : a.status === 'in_progress' ? 'bg-[#1e3a5f] text-[#60a5fa]' : 'bg-[#3b2f1e] text-[#fbbf24]'}`}>{a.status}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}