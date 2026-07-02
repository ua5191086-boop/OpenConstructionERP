// @ts-nocheck
import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, LineChart, Line } from 'recharts'

// Colors for severity and categories
const SEV_COLORS: Record<string, string> = { critical: '#ef4444', warning: '#f97316', info: '#3b82f6', emergency: '#a855f7' }

export default function TBMPage() {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [telemetry, alarms, operators, shifts, consumables, performance, summary] = await Promise.all([
          fetch('/api/v1/tbm/telemetry?limit=100').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/tbm/alarms').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/tbm/operators').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/tbm/shifts').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/tbm/consumables').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/tbm/performance').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/tbm/summary').then(r => r.ok ? r.json() : {}),
        ])
        setData({ telemetry, alarms, operators, shifts, consumables, performance, summary })
      } catch {
        setError('API unavailable — start the Go backend on port 8081')
      }
      setLoading(false)
    }
    load()
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading TBM data...</div>
  if (error) {
    return (
      <div className="p-8">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">🛠️</div>
          <h2 className="text-xl font-bold text-white mb-2">TBM Management Module</h2>
          <p className="text-[#94a3b8] mb-4">{error}</p>
          <p className="text-sm text-[#64748b]">Start the Go API server on port 8081 to see live data</p>
        </div>
      </div>
    )
  }

  const tel = data.telemetry || []
  const alarms = data.alarms || []
  const shifts = data.shifts || []
  const consumables = data.consumables || []
  const perf = data.performance || []
  const sm = data.summary || {}

  // Stats
  const stats = [
    { label: 'Telemetry Points', value: tel.length, color: 'text-blue-400' },
    { label: 'Active Alarms', value: alarms.filter((a: any) => a.is_active).length, color: 'text-red-400' },
    { label: 'Total Rings', value: sm.total_rings || 0, color: 'text-green-400' },
    { label: 'Advance (m)', value: Math.round((sm.total_advance_mm || 0) / 1000), color: 'text-purple-400' },
    { label: 'Avg Utilisation', value: (sm.avg_utilisation_pct || 0).toFixed(1) + '%', color: 'text-green-400' },
    { label: 'Operators', value: (data.operators || []).length, color: 'text-blue-400' },
  ]

  // Alarm severity pie
  const sevCount: Record<string, number> = {}
  alarms.forEach((a: any) => { sevCount[a.alarm_severity] = (sevCount[a.alarm_severity] || 0) + 1 })

  // Consumables
  const conByType: Record<string, number> = {}
  consumables.forEach((c: any) => { conByType[c.consumable_type] = (conByType[c.consumable_type] || 0) + (c.quantity_used || 0) })

  // Recent telemetry
  const recentTel = tel.slice(0, 50).reverse()

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-bold text-white">🛠️ TBM Management</h1>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        {stats.map((s, i) => (
          <div key={i} className="bg-[#1e293b] border border-[#334155] rounded-xl p-4 text-center">
            <div className="text-xs text-[#94a3b8] uppercase tracking-wider">{s.label}</div>
            <div className={`text-2xl font-bold mt-2 ${s.color}`}>{s.value}</div>
          </div>
        ))}
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <h3 className="text-white font-medium mb-4">📊 Thrust & Torque</h3>
          <ResponsiveContainer width="100%" height={250}>
            <LineChart data={recentTel.map((t: any, i: number) => ({ i, thrust: t.thrust_force_kN || 0, torque: t.torque_kNm || 0 }))}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="i" stroke="#64748b" tick={false} />
              <YAxis stroke="#64748b" />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155' }} />
              <Line type="monotone" dataKey="thrust" stroke="#3b82f6" dot={false} name="Thrust (kN)" />
              <Line type="monotone" dataKey="torque" stroke="#f97316" dot={false} name="Torque (kNm)" />
            </LineChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <h3 className="text-white font-medium mb-4">🔔 Alarms by Severity</h3>
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie data={Object.entries(sevCount).map(([k, v]) => ({ name: k, value: v }))} cx="50%" cy="50%" outerRadius={80} dataKey="value" label={({ name, value }) => `${name}: ${value}`}>
                {Object.entries(sevCount).map(([k]) => <Cell key={k} fill={SEV_COLORS[k] || '#64748b'} />)}
              </Pie>
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155' }} />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Consumables */}
      <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
        <h3 className="text-white font-medium mb-4">⛏️ Consumables Usage</h3>
        <ResponsiveContainer width="100%" height={250}>
          <BarChart data={Object.entries(conByType).map(([k, v]) => ({ name: k, value: v }))}>
            <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
            <XAxis dataKey="name" stroke="#64748b" />
            <YAxis stroke="#64748b" />
            <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155' }} />
            <Bar dataKey="value" fill="#3b82f6" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* Active Alarms Table */}
      <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
        <div className="p-4 border-b border-[#334155]">
          <h3 className="text-white font-medium">🔔 Active Alarms ({alarms.filter((a: any) => a.is_active).length})</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead><tr className="text-xs text-[#64748b] uppercase"><th className="p-3 text-left">Code</th><th className="p-3 text-left">Name</th><th className="p-3 text-left">Severity</th><th className="p-3 text-left">Triggered</th></tr></thead>
            <tbody>
              {alarms.filter((a: any) => a.is_active).slice(0, 20).map((a: any) => (
                <tr key={a.id} className="border-t border-[#1e293b] hover:bg-[#334155]">
                  <td className="p-3 text-[#e2e8f0]">{a.alarm_code}</td>
                  <td className="p-3 text-[#94a3b8]">{a.alarm_name}</td>
                  <td className="p-3"><span className={`px-2 py-0.5 rounded text-xs font-semibold bg-opacity-20 ${a.alarm_severity === 'critical' ? 'bg-red-900 text-red-300' : a.alarm_severity === 'warning' ? 'bg-orange-900 text-orange-300' : 'bg-blue-900 text-blue-300'}`}>{a.alarm_severity}</span></td>
                  <td className="p-3 text-[#94a3b8]">{a.triggered_at ? new Date(a.triggered_at).toLocaleString() : ''}</td>
                </tr>
              ))}
              {alarms.filter((a: any) => a.is_active).length === 0 && <tr><td colSpan={4} className="p-6 text-center text-[#64748b]">No active alarms</td></tr>}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}