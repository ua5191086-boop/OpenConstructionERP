import { useState, useEffect } from 'react'

const NODE_COLORS: Record<string, string> = {
  project: '#3b82f6', contract: '#22c55e', document: '#a855f7',
  equipment: '#f97316', employee: '#06b6d4', hse_incident: '#ef4444',
  risk: '#eab308', boq_item: '#84cc16', budget: '#6366f1', invoice: '#ec4899',
}

export default function KnowledgeGraphPage() {
  const [nodes, setNodes] = useState<any[]>([])
  const [edges, setEdges] = useState<any[]>([])
  const [events, setEvents] = useState<any[]>([])
  const [topics, setTopics] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      fetch('/api/v1/protected/knowledge/nodes').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/v1/protected/knowledge/edges').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/v1/protected/events/events').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/v1/protected/events/topics').then(r => r.ok ? r.json() : Promise.reject()),
    ])
      .then(([n, e, ev, t]) => { setNodes(n); setEdges(e); setEvents(ev); setTopics(t) })
      .catch(() => setError('API unavailable'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Knowledge Graph data...</div>

  const nodeCounts: Record<string, number> = {}
  nodes.forEach(n => { nodeCounts[n.node_type] = (nodeCounts[n.node_type] || 0) + 1 })
  const synced = nodes.filter(n => n.is_synced).length
  const runningConsumers = topics.filter(t => t.is_active !== false).length

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">🕸️ Knowledge Graph</h1>
        <p className="text-[#94a3b8] mt-1">V028 — Neo4j + Kafka | Knowledge Graph &amp; Event Streaming</p>
      </div>

      {error ? (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">🕸️</div>
          <h2 className="text-xl font-bold text-white mb-2">Knowledge Graph Module</h2>
          <p className="text-[#94a3b8]">{error}</p>
        </div>
      ) : (
        <>
          {/* KPI Cards */}
          <div className="grid grid-cols-2 md:grid-cols-6 gap-4 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Nodes</div>
              <div className="text-2xl font-bold text-[#3b82f6] mt-2">{nodes.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Edges</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">{edges.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Synced</div>
              <div className="text-2xl font-bold text-[#a855f7] mt-2">{synced}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Topics</div>
              <div className="text-2xl font-bold text-[#f97316] mt-2">{topics.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Events</div>
              <div className="text-2xl font-bold text-[#eab308] mt-2">{events.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Consumers</div>
              <div className="text-2xl font-bold text-[#06b6d4] mt-2">{runningConsumers}</div>
            </div>
          </div>

          {/* Nodes Table */}
          <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5 mb-6">
            <h3 className="text-white font-semibold mb-3">🧠 Graph Nodes</h3>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-[#94a3b8] text-xs uppercase border-b border-[#334155]">
                    <th className="text-left py-2 pr-4">Type</th>
                    <th className="text-left py-2 pr-4">Label</th>
                    <th className="text-left py-2">Synced</th>
                  </tr>
                </thead>
                <tbody>
                  {nodes.slice(0, 15).map(n => (
                    <tr key={n.id} className="border-b border-[#1e293b]">
                      <td className="py-2 pr-4">
                        <span className="px-2 py-0.5 rounded-full text-xs font-medium"
                          style={{ backgroundColor: (NODE_COLORS[n.node_type] || '#64748b') + '22', color: NODE_COLORS[n.node_type] || '#64748b' }}>
                          {n.node_type}
                        </span>
                      </td>
                      <td className="py-2 pr-4 text-[#e2e8f0]">{n.node_label || n.node_type}</td>
                      <td className="py-2">{n.is_synced ? '✅' : '❌'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Topics & Events */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <h3 className="text-white font-semibold mb-3">📨 Kafka Topics</h3>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="text-[#94a3b8] text-xs uppercase border-b border-[#334155]">
                      <th className="text-left py-2 pr-4">Topic</th>
                      <th className="text-left py-2 pr-4">Partitions</th>
                      <th className="text-left py-2">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {topics.slice(0, 6).map(t => (
                      <tr key={t.id || t.topic_name} className="border-b border-[#1e293b]">
                        <td className="py-2 pr-4 text-[#e2e8f0]">{t.topic_name}</td>
                        <td className="py-2 pr-4 text-[#94a3b8]">{t.partitions || '-'}</td>
                        <td className="py-2">
                          <span className={`px-2 py-0.5 rounded-full text-xs ${t.is_active !== false ? 'bg-green-900/50 text-green-400' : 'bg-red-900/50 text-red-400'}`}>
                            {t.is_active !== false ? 'active' : 'inactive'}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <h3 className="text-white font-semibold mb-3">📨 Recent Events</h3>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="text-[#94a3b8] text-xs uppercase border-b border-[#334155]">
                      <th className="text-left py-2 pr-4">Type</th>
                      <th className="text-left py-2">Topic</th>
                    </tr>
                  </thead>
                  <tbody>
                    {events.slice(0, 10).map(e => (
                      <tr key={e.id} className="border-b border-[#1e293b]">
                        <td className="py-2 pr-4 text-[#e2e8f0]">{e.event_type}</td>
                        <td className="py-2 text-[#94a3b8]">{e.topic_name}</td>
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