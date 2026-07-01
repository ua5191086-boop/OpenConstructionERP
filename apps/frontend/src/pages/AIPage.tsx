import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

const AGENT_COLORS: Record<string, string> = {
  assistant: '#3b82f6', analyst: '#22c55e', estimator: '#f97316',
  scheduler: '#a855f7', qa: '#14b8a6', document: '#f59e0b',
}

export default function AIPage() {
  const [agents, setAgents] = useState<any[]>([])
  const [tasks, setTasks] = useState<any[]>([])
  const [conversations, setConversations] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      fetch('/api/ai/agents').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/ai/tasks').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/ai/conversations').then(r => r.ok ? r.json() : Promise.reject()),
    ])
      .then(([a, t, c]) => { setAgents(a); setTasks(t); setConversations(c) })
      .catch(() => setError('API unavailable'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading AI data...</div>

  const typeCounts: Record<string, number> = {}
  agents.forEach((a: any) => { typeCounts[a.agent_type || 'assistant'] = (typeCounts[a.agent_type || 'assistant'] || 0) + 1 })
  const typeChartData = Object.entries(typeCounts).map(([name, value]) => ({ name, value }))

  const activeAgents = agents.filter((a: any) => a.is_active).length
  const taskStatusCounts: Record<string, number> = {}
  tasks.forEach((t: any) => { taskStatusCounts[t.status || 'pending'] = (taskStatusCounts[t.status || 'pending'] || 0) + 1 })

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">🤖 AI — AI Assistant</h1>
        <p className="text-[#94a3b8] mt-1">AI Agents & Task Management</p>
      </div>

      {error ? (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">🤖</div>
          <h2 className="text-xl font-bold text-white mb-2">AI Module</h2>
          <p className="text-[#94a3b8]">{error}</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">AI Agents</div>
              <div className="text-2xl font-bold text-white mt-2">{agents.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Active</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">{activeAgents}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Tasks</div>
              <div className="text-2xl font-bold text-[#3b82f6] mt-2">{tasks.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Conversations</div>
              <div className="text-2xl font-bold text-[#a855f7] mt-2">{conversations.length}</div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Agents by Type</h3>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={typeChartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                  <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
                  <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
                  <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
                  <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                    {typeChartData.map((d) => <Cell key={d.name} fill={AGENT_COLORS[d.name] || '#64748b'} />)}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Task Status</h3>
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie data={Object.entries(taskStatusCounts).map(([name, value]) => ({ name, value }))}
                    cx="50%" cy="50%" outerRadius={100} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                    {Object.entries(taskStatusCounts).map(([name], i) => (
                      <Cell key={name} fill={['#3b82f6','#22c55e','#f97316','#ef4444','#64748b','#a855f7'][i % 6]} />
                    ))}
                  </Pie>
                  <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">AI Agents</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Agent</th>
                      <th className="p-3">Type</th>
                      <th className="p-3">Model</th>
                      <th className="p-3 pr-4">Active</th>
                    </tr>
                  </thead>
                  <tbody>
                    {agents.map((a: any) => (
                      <tr key={a.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 text-white">{a.agent_name}</td>
                        <td className="p-3">
                          <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold"
                            style={{ background: `${AGENT_COLORS[a.agent_type] || '#64748b'}22`, color: AGENT_COLORS[a.agent_type] || '#64748b' }}>
                            {a.agent_type}
                          </span>
                        </td>
                        <td className="p-3 text-[#94a3b8] text-xs">{a.model_name || '—'}</td>
                        <td className="p-3 pr-4">
                          <span className={`inline-block w-2 h-2 rounded-full ${a.is_active ? 'bg-[#22c55e]' : 'bg-[#64748b]'}`} />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">AI Tasks</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Task</th>
                      <th className="p-3">Type</th>
                      <th className="p-3">Status</th>
                      <th className="p-3 pr-4">Agent</th>
                    </tr>
                  </thead>
                  <tbody>
                    {tasks.map((t: any) => (
                      <tr key={t.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 text-white">{t.task_name}</td>
                        <td className="p-3 text-[#94a3b8]">{t.task_type}</td>
                        <td className="p-3">
                          <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold text-[#3b82f6] bg-[#3b82f6]/20">
                            {t.status}
                          </span>
                        </td>
                        <td className="p-3 pr-4 text-[#94a3b8] text-xs">{t.agent_id?.substring(0, 8) || '—'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
