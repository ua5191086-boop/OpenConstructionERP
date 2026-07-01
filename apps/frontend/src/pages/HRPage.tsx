import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

const STATUS_COLORS: Record<string, string> = {
  active: '#22c55e', on_leave: '#f97316', terminated: '#ef4444', suspended: '#64748b',
}

export default function HRPage() {
  const [employees, setEmployees] = useState<any[]>([])
  const [departments, setDepartments] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      fetch('/api/hr/employees').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/hr/departments').then(r => r.ok ? r.json() : Promise.reject()),
    ])
      .then(([emps, deps]) => { setEmployees(emps); setDepartments(deps) })
      .catch(() => setError('API unavailable'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading HR data...</div>

  const statusCounts: Record<string, number> = {}
  employees.forEach((e: any) => { statusCounts[e.status || 'active'] = (statusCounts[e.status || 'active'] || 0) + 1 })
  const statusChartData = Object.entries(statusCounts).map(([name, value]) => ({ name, value }))

  const deptCounts: Record<string, number> = {}
  employees.forEach((e: any) => { const d = e.department || 'Unknown'; deptCounts[d] = (deptCounts[d] || 0) + 1 })
  const deptChartData = Object.entries(deptCounts).map(([name, value]) => ({ name, value }))

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">👥 HR — Human Resources</h1>
        <p className="text-[#94a3b8] mt-1">Employee Management</p>
      </div>

      {error ? (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">👥</div>
          <h2 className="text-xl font-bold text-white mb-2">HR Module</h2>
          <p className="text-[#94a3b8]">{error}</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Total Employees</div>
              <div className="text-2xl font-bold text-white mt-2">{employees.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Departments</div>
              <div className="text-2xl font-bold text-[#3b82f6] mt-2">{departments.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Active</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">{statusCounts['active'] || 0}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">On Leave</div>
              <div className="text-2xl font-bold text-[#f97316] mt-2">{statusCounts['on_leave'] || 0}</div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Employees by Status</h3>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={statusChartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                  <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
                  <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
                  <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
                  <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                    {statusChartData.map((d) => <Cell key={d.name} fill={STATUS_COLORS[d.name] || '#64748b'} />)}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Employees by Department</h3>
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie data={deptChartData} cx="50%" cy="50%" outerRadius={100} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                    {deptChartData.map((_, i) => <Cell key={i} fill={['#3b82f6','#22c55e','#a855f7','#f97316','#ef4444','#14b8a6','#f59e0b','#6366f1'][i % 8]} />)}
                  </Pie>
                  <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
            <div className="p-4 border-b border-[#334155]">
              <h3 className="text-white font-semibold">Employees</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                    <th className="p-3 pl-4">Name</th>
                    <th className="p-3">Position</th>
                    <th className="p-3">Department</th>
                    <th className="p-3">Status</th>
                    <th className="p-3">Email</th>
                    <th className="p-3 pr-4">Phone</th>
                  </tr>
                </thead>
                <tbody>
                  {employees.map((e: any) => (
                    <tr key={e.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                      <td className="p-3 pl-4 text-white font-medium">{e.full_name}</td>
                      <td className="p-3 text-[#94a3b8]">{e.position}</td>
                      <td className="p-3 text-[#94a3b8]">{e.department}</td>
                      <td className="p-3">
                        <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold"
                          style={{ background: `${STATUS_COLORS[e.status] || '#64748b'}22`, color: STATUS_COLORS[e.status] || '#64748b' }}>
                          {e.status}
                        </span>
                      </td>
                      <td className="p-3 text-[#94a3b8] text-xs">{e.email}</td>
                      <td className="p-3 pr-4 text-[#94a3b8] text-xs">{e.phone || '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
