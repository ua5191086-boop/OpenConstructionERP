import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, LineChart, Line } from 'recharts'

const TYPE_COLORS = ['#3b82f6','#22c55e','#a855f7','#f97316','#ef4444','#14b8a6','#f59e0b','#ec4899','#6366f1','#84cc16']

export default function ChangePage() {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [projectFilter, setProjectFilter] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [requests, orders, impactAnalysis, approvalWorkflow, changeLog] = await Promise.all([
          fetch('/api/v1/change/requests').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/change/orders').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/change/impact-analysis').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/change/approval-workflow').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/change/change-log').then(r => r.ok ? r.json() : { data: [] }),
        ])
        setData({
          requests: requests.data || [],
          orders: orders.data || [],
          impactAnalysis: impactAnalysis.data || [],
          approvalWorkflow: approvalWorkflow.data || [],
          changeLog: changeLog.data || [],
        })
      } catch {
        setError('API unavailable — start the Go backend on port 8081')
      }
      setLoading(false)
    }
    load()
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Change data...</div>
  if (error) {
    return (
      <div className="p-8">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">🔄</div>
          <h2 className="text-xl font-bold text-white mb-2">Change Management Module</h2>
          <p className="text-[#94a3b8] mb-4">{error}</p>
          <p className="text-sm text-[#64748b]">Start the Go API server on port 8081 to see live data</p>
        </div>
      </div>
    )
  }

  const requests = data.requests || []
  const orders = data.orders || []
  const impactAnalysis = data.impactAnalysis || []
  const approvalWorkflow = data.approvalWorkflow || []
  const changeLog = data.changeLog || []

  // Filters
  const projects = [...new Set(requests.map((r: any) => r.project_name).filter(Boolean))] as string[]
  let filteredRequests = requests
  let filteredOrders = orders
  let filteredImpact = impactAnalysis
  if (projectFilter) {
    filteredRequests = filteredRequests.filter((r: any) => r.project_name === projectFilter)
    filteredOrders = filteredOrders.filter((o: any) => o.project_name === projectFilter)
    filteredImpact = filteredImpact.filter((ia: any) => ia.project_name === projectFilter)
  }

  // Stats
  const openRequests = filteredRequests.filter((r: any) => r.status === 'draft' || r.status === 'submitted' || r.status === 'under_review').length
  const approvedRequests = filteredRequests.filter((r: any) => r.status === 'approved').length
  const implementedRequests = filteredRequests.filter((r: any) => r.status === 'implemented').length
  const totalOrderCost = filteredOrders.reduce((sum: number, o: any) => sum + (o.cost_change || 0), 0)

  // Request status pie
  const statusCounts: Record<string, number> = {}
  filteredRequests.forEach((r: any) => { statusCounts[r.status] = (statusCounts[r.status] || 0) + 1 })
  const statusPie = Object.entries(statusCounts).map(([k, v], i) => ({
    name: k,
    value: v as number,
    color: TYPE_COLORS[i % TYPE_COLORS.length],
  }))

  // CR type pie
  const typeCounts: Record<string, number> = {}
  filteredRequests.forEach((r: any) => { typeCounts[r.cr_type] = (typeCounts[r.cr_type] || 0) + 1 })
  const typePie = Object.entries(typeCounts).map(([k, v], i) => ({
    name: k,
    value: v as number,
    color: TYPE_COLORS[(i + 3) % TYPE_COLORS.length],
  }))

  // Order cost impact
  const topOrders = [...filteredOrders].sort((a: any, b: any) => Math.abs(b.cost_change) - Math.abs(a.cost_change)).slice(0, 10)

  // Impact analysis level distribution
  const impactLevels: Record<string, number> = {}
  filteredImpact.forEach((ia: any) => { impactLevels[ia.impact_level] = (impactLevels[ia.impact_level] || 0) + 1 })
  const impactData = Object.entries(impactLevels).map(([k, v]) => ({ name: k.replace('_', ' '), count: v }))

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">🔄 Change Management</h1>
          <p className="text-sm text-[#94a3b8] mt-1">Change Requests, Orders, Impact Analysis, Approval Workflow & Change Log</p>
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
      <div className="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-6 gap-4">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Total CRs</div>
          <div className="text-2xl font-bold text-[#3b82f6]">{requests.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Open/Pending</div>
          <div className="text-2xl font-bold text-[#f97316]">{openRequests}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Approved</div>
          <div className="text-2xl font-bold text-[#22c55e]">{approvedRequests}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Implemented</div>
          <div className="text-2xl font-bold text-[#14b8a6]">{implementedRequests}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Change Orders</div>
          <div className="text-2xl font-bold text-[#a855f7]">{orders.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Total Cost Impact</div>
          <div className="text-lg font-bold text-[#ef4444]">${(totalOrderCost / 1000).toFixed(0)}k</div>
        </div>
      </div>

      {/* Charts Row 1 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Change Request Status</h3>
          <ResponsiveContainer width="100%" height={240}>
            <PieChart>
              <Pie data={statusPie} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {statusPie.map((e, i) => <Cell key={i} fill={e.color} />)}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">CR Types</h3>
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
          <h3 className="text-sm font-semibold text-white mb-4">Impact Level Distribution</h3>
          <ResponsiveContainer width="100%" height={240}>
            <PieChart>
              <Pie data={impactData.map(d => ({ ...d, color: d.name === 'very high' ? '#ef4444' : d.name === 'high' ? '#f97316' : d.name === 'medium' ? '#f59e0b' : d.name === 'low' ? '#22c55e' : d.name === 'very low' ? '#64748b' : '#3b82f6' }))} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="count" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {impactData.map((e, i) => <Cell key={i} fill={['#ef4444','#f97316','#f59e0b','#22c55e','#64748b','#3b82f6'][i % 6]} />)}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Change Orders Cost Impact */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Top Change Orders by Cost Impact</h3>
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={topOrders} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis type="number" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis type="category" dataKey="co_code" width={70} tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip formatter={(value: any) => `$${(Number(value) / 1000).toFixed(0)}k`} />
              <Bar dataKey="cost_change" fill="#a855f7" radius={[0,4,4,0]} name="Cost Change ($)" />
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Change Orders by Type</h3>
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={(() => {
              const byType: Record<string, number> = {}
              orders.forEach((o: any) => { byType[o.co_type] = (byType[o.co_type] || 0) + 1 })
              return Object.entries(byType).map(([k, v]) => ({ name: k, count: v }))
            })()}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip />
              <Bar dataKey="count" fill="#f97316" radius={[4,4,0,0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Change Requests Table */}
      <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
        <h3 className="text-sm font-semibold text-white mb-4">Change Requests</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-[#64748b] border-b border-[#334155]">
                <th className="text-left py-2 px-3">Code</th>
                <th className="text-left py-2 px-3">Name</th>
                <th className="text-left py-2 px-3">Project</th>
                <th className="text-left py-2 px-3">Type</th>
                <th className="text-left py-2 px-3">Source</th>
                <th className="text-left py-2 px-3">Priority</th>
                <th className="text-left py-2 px-3">Proposed By</th>
                <th className="text-left py-2 px-3">Status</th>
              </tr>
            </thead>
            <tbody>
              {filteredRequests.slice(0, 10).map((cr: any) => (
                <tr key={cr.id} className="border-b border-[#1e293b] hover:bg-[#334155]/40">
                  <td className="py-2 px-3 text-white font-mono text-xs">{cr.cr_code}</td>
                  <td className="py-2 px-3 text-white">{cr.cr_name}</td>
                  <td className="py-2 px-3 text-[#94a3b8]">{cr.project_name}</td>
                  <td className="py-2 px-3 text-[#94a3b8]">{cr.cr_type}</td>
                  <td className="py-2 px-3 text-[#94a3b8]">{cr.source}</td>
                  <td className="py-2 px-3">
                    <span className={`px-2 py-0.5 rounded-full text-xs ${
                      cr.priority === 'emergency' ? 'bg-red-500/20 text-red-400' :
                      cr.priority === 'high' ? 'bg-orange-500/20 text-orange-400' :
                      cr.priority === 'medium' ? 'bg-yellow-500/20 text-yellow-400' :
                      'bg-green-500/20 text-green-400'
                    }`}>{cr.priority}</span>
                  </td>
                  <td className="py-2 px-3 text-[#94a3b8]">{cr.proposed_by}</td>
                  <td className="py-2 px-3">
                    <span className={`px-2 py-0.5 rounded-full text-xs ${
                      cr.status === 'approved' ? 'bg-green-500/20 text-green-400' :
                      cr.status === 'implemented' ? 'bg-blue-500/20 text-blue-400' :
                      cr.status === 'rejected' ? 'bg-red-500/20 text-red-400' :
                      cr.status === 'closed' ? 'bg-gray-500/20 text-gray-400' :
                      'bg-yellow-500/20 text-yellow-400'
                    }`}>{cr.status}</span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Approval Workflow & Change Log */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Approval Workflow</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-[#64748b] border-b border-[#334155]">
                  <th className="text-left py-2 px-3">Step</th>
                  <th className="text-left py-2 px-3">Step Name</th>
                  <th className="text-left py-2 px-3">Approver</th>
                  <th className="text-left py-2 px-3">Status</th>
                </tr>
              </thead>
              <tbody>
                {approvalWorkflow.slice(0, 8).map((aw: any) => (
                  <tr key={aw.id} className="border-b border-[#1e293b] hover:bg-[#334155]/40">
                    <td className="py-2 px-3 text-white">{aw.step_order}</td>
                    <td className="py-2 px-3 text-white">{aw.step_name}</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{aw.approver_role}</td>
                    <td className="py-2 px-3">
                      <span className={`px-2 py-0.5 rounded-full text-xs ${
                        aw.status === 'approved' ? 'bg-green-500/20 text-green-400' :
                        aw.status === 'rejected' ? 'bg-red-500/20 text-red-400' :
                        'bg-yellow-500/20 text-yellow-400'
                      }`}>{aw.status}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Recent Change Log</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-[#64748b] border-b border-[#334155]">
                  <th className="text-left py-2 px-3">Type</th>
                  <th className="text-left py-2 px-3">Description</th>
                  <th className="text-left py-2 px-3">Changed By</th>
                  <th className="text-left py-2 px-3">Date</th>
                </tr>
              </thead>
              <tbody>
                {changeLog.slice(0, 8).map((cl: any) => (
                  <tr key={cl.id} className="border-b border-[#1e293b] hover:bg-[#334155]/40">
                    <td className="py-2 px-3">
                      <span className={`px-2 py-0.5 rounded-full text-xs ${
                        cl.log_type === 'approval' ? 'bg-green-500/20 text-green-400' :
                        cl.log_type === 'rejection' ? 'bg-red-500/20 text-red-400' :
                        cl.log_type === 'status_change' ? 'bg-blue-500/20 text-blue-400' :
                        'bg-yellow-500/20 text-yellow-400'
                      }`}>{cl.log_type}</span>
                    </td>
                    <td className="py-2 px-3 text-white">{cl.description}</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{cl.changed_by}</td>
                    <td className="py-2 px-3 text-[#94a3b8] text-xs">{cl.changed_at}</td>
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