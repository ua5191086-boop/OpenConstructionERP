import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

const TYPE_COLORS = ['#3b82f6','#22c55e','#a855f7','#f97316','#ef4444','#14b8a6','#f59e0b','#ec4899','#6366f1','#84cc16']

export default function EquipmentPage() {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [projectFilter, setProjectFilter] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [categories, equipment, maintenance, maintenanceSchedules, telemetry, fuel, downtime, spareParts] = await Promise.all([
          fetch('/api/v1/equipment/categories').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/equipment/items').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/equipment/maintenance').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/equipment/maintenance-schedules').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/equipment/telemetry').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/equipment/fuel').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/equipment/downtime').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/equipment/spare-parts').then(r => r.ok ? r.json() : { data: [] }),
        ])
        setData({
          categories: categories.data || [],
          equipment: equipment.data || [],
          maintenance: maintenance.data || [],
          maintenanceSchedules: maintenanceSchedules.data || [],
          telemetry: telemetry.data || [],
          fuel: fuel.data || [],
          downtime: downtime.data || [],
          spareParts: spareParts.data || [],
        })
      } catch {
        setError('API unavailable — start the Go backend on port 8081')
      }
      setLoading(false)
    }
    load()
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Equipment data...</div>
  if (error) {
    return (
      <div className="p-8">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">🏗️</div>
          <h2 className="text-xl font-bold text-white mb-2">Equipment Management</h2>
          <p className="text-[#94a3b8] mb-4">{error}</p>
          <p className="text-sm text-[#64748b]">Start the Go API server on port 8081 to see live data</p>
        </div>
      </div>
    )
  }

  const equipment = data.equipment || []
  const maintenance = data.maintenance || []
  const downtime = data.downtime || []
  const categories = data.categories || []

  // Filters
  const projects = [...new Set(equipment.map((e: any) => e.project_name).filter(Boolean))] as string[]
  let filtered = equipment
  if (projectFilter) filtered = filtered.filter((e: any) => e.project_name === projectFilter)
  const eqIds = new Set(filtered.map((e: any) => e.id))
  const filteredMaint = maintenance.filter((m: any) => eqIds.has(m.equipment_id))
  const filteredDowntime = downtime.filter((d: any) => eqIds.has(d.equipment_id))

  // Stats
  const total = filtered.length
  const available = filtered.filter((e: any) => e.status === 'available').length
  const inUse = filtered.filter((e: any) => e.status === 'in_use').length
  const underMaint = filtered.filter((e: any) => e.status === 'under_maintenance').length
  const outOfService = filtered.filter((e: any) => e.status === 'out_of_service').length
  const overdue = filtered.filter((e: any) => e.next_service_date && e.next_service_date < '2025-07-02').length

  // Chart data
  const statusCounts: Record<string, number> = { available: 0, in_use: 0, under_maintenance: 0, out_of_service: 0 }
  filtered.forEach((e: any) => { if (statusCounts[e.status] !== undefined) statusCounts[e.status]++; else statusCounts.other = (statusCounts.other || 0) + 1 })
  const statusData = Object.entries(statusCounts).map(([name, value]) => ({ name, value }))

  const byCat: Record<string, number> = {}
  filtered.forEach((e: any) => {
    const cat = categories.find((c: any) => c.id === e.category_id)
    const code = cat ? cat.category_code : e.category_id
    byCat[code] = (byCat[code] || 0) + 1
  })
  const catData = Object.entries(byCat).map(([name, value]) => ({ name, value }))

  const withFuel = filtered.filter((e: any) => (e.fuel_capacity || 0) > 0).slice(0, 15)
  const fuelData = withFuel.map((e: any) => ({ name: e.equipment_code, value: e.fuel_capacity || 0 }))

  const byEq: Record<string, number> = {}
  filteredDowntime.forEach((d: any) => { byEq[d.equipment_id] = (byEq[d.equipment_id] || 0) + (d.duration_hours || 0) })
  const topDowntime = Object.entries(byEq).sort((a: any, b: any) => b[1] - a[1]).slice(0, 10)
  const downtimeData = topDowntime.map(([id, hrs]) => {
    const eq = filtered.find((e: any) => e.id === id)
    return { name: eq ? eq.equipment_code : id, value: hrs as number }
  })

  const withMeter = filtered.filter((e: any) => (e.meter_reading || 0) > 0).slice(0, 15)
  const meterData = withMeter.map((e: any) => ({ name: e.equipment_code, value: e.meter_reading || 0 }))

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">🏗️ Equipment Management</h1>
        <p className="text-[#94a3b8] mt-1">Fleet & Equipment Dashboard — Status, Maintenance, Downtime, Telemetry</p>
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
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Total</div>
          <div className="text-2xl font-bold text-[#3b82f6] mt-1">{total}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Available</div>
          <div className="text-2xl font-bold text-[#22c55e] mt-1">{available}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">In Use</div>
          <div className="text-2xl font-bold text-[#a855f7] mt-1">{inUse}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Under Maint</div>
          <div className="text-2xl font-bold text-[#eab308] mt-1">{underMaint}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Out of Service</div>
          <div className="text-2xl font-bold text-[#ef4444] mt-1">{outOfService}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Overdue Service</div>
          <div className="text-2xl font-bold text-[#f97316] mt-1">{overdue}</div>
        </div>
      </div>

      {/* Charts Row 1 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Equipment Status</h3>
          <ResponsiveContainer width="100%" height={260}>
            <PieChart>
              <Pie data={statusData} cx="50%" cy="50%" outerRadius={90} dataKey="value"
                label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {statusData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
              </Pie>
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">By Category</h3>
          <ResponsiveContainer width="100%" height={260}>
            <BarChart data={catData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
              <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                {catData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Fuel Capacity</h3>
          <ResponsiveContainer width="100%" height={260}>
            <BarChart data={fuelData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 10 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
              <Bar dataKey="value" radius={[4, 4, 0, 0]} fill="#f97316" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Charts Row 2 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Downtime by Equipment (hours)</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={downtimeData} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis type="number" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis dataKey="name" type="category" tick={{ fill: '#94a3b8', fontSize: 10 }} width={80} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
              <Bar dataKey="value" radius={[0, 4, 4, 0]} fill="#ef4444" />
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Meter Readings</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={meterData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 10 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
              <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                {meterData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Equipment Table */}
      <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden mb-6">
        <div className="p-4 border-b border-[#334155]">
          <h3 className="text-white font-semibold">🔧 Equipment ({filtered.length})</h3>
        </div>
        <div className="overflow-x-auto max-h-96 overflow-y-auto">
          <table className="w-full">
            <thead>
              <tr className="text-[#64748b] text-xs uppercase tracking-wider">
                <th className="text-left p-3 border-b border-[#334155]">Code</th>
                <th className="text-left p-3 border-b border-[#334155]">Name</th>
                <th className="text-left p-3 border-b border-[#334155]">Manufacturer</th>
                <th className="text-left p-3 border-b border-[#334155]">Status</th>
                <th className="text-left p-3 border-b border-[#334155]">Location</th>
                <th className="text-left p-3 border-b border-[#334155]">Meter</th>
                <th className="text-left p-3 border-b border-[#334155]">Next Service</th>
                <th className="text-left p-3 border-b border-[#334155]">Rate</th>
              </tr>
            </thead>
            <tbody>
              {filtered.slice(0, 100).map((e: any) => (
                <tr key={e.id} className="hover:bg-[#334155] border-b border-[#1e293b]">
                  <td className="p-3 font-mono text-xs text-[#64748b]">{e.equipment_code}</td>
                  <td className="p-3 text-sm text-[#e2e8f0]">{e.equipment_name}</td>
                  <td className="p-3 text-xs text-[#94a3b8]">{e.manufacturer || '—'}</td>
                  <td className="p-3">
                    <span className={`px-2 py-0.5 rounded text-xs font-semibold ${
                      e.status === 'available' ? 'bg-[#1e3a2f] text-[#34d399]' :
                      e.status === 'in_use' ? 'bg-[#1e3a5f] text-[#60a5fa]' :
                      e.status === 'under_maintenance' ? 'bg-[#3b2f1e] text-[#fbbf24]' :
                      'bg-[#3b1e1e] text-[#f87171]'
                    }`}>{e.status}</span>
                  </td>
                  <td className="p-3 text-xs text-[#94a3b8]">{e.location || '—'}</td>
                  <td className="p-3 text-sm">{(e.meter_reading || 0).toFixed(1)} {e.meter_type || ''}</td>
                  <td className="p-3 text-xs text-[#94a3b8]">{e.next_service_date || '—'}</td>
                  <td className="p-3 text-sm">${(e.hourly_rate || 0).toFixed(2)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}