#!/usr/bin/env python3
"""
Generator: Neo4j + Kafka Module (V028)
Generates: knowledge_graph_nodes, knowledge_graph_edges, sync_queue,
           kafka_topics, kafka_events, kafka_consumers
"""
import json, random, sys
from datetime import datetime, timedelta

NODE_COUNT = 30
EDGE_COUNT = 45
QUEUE_COUNT = 12
TOPIC_COUNT = 6
EVENT_COUNT = 20
CONSUMER_COUNT = 4

NODE_TYPES = [
    'project', 'contract', 'document', 'equipment', 'employee',
    'hse_incident', 'risk', 'boq_item', 'budget', 'invoice',
    'supplier', 'material', 'tbm', 'schedule_activity', 'change_order',
]
EDGE_TYPES = [
    'belongs_to', 'references', 'depends_on', 'assigned_to',
    'contains', 'impacts', 'approves', 'relates_to',
    'generates', 'uses', 'located_at', 'owns',
]
EVENT_TYPES = [
    'project.created', 'contract.signed', 'payment.processed',
    'incident.reported', 'milestone.achieved', 'change.approved',
    'risk.identified', 'equipment.deployed', 'document.uploaded',
    'invoice.submitted',
]

def rand(min_v, max_v):
    return round(random.uniform(min_v, max_v), 2)

def pick(values):
    return random.choice(values)

def generate_data():
    nodes = []
    for i in range(NODE_COUNT):
        nt = pick(NODE_TYPES)
        props = {
            'name': f'{nt.title()} #{i+1}',
            'code': f'{nt[:3].upper()}-{i+1:04d}',
            'status': pick(['active', 'pending', 'completed', 'cancelled']),
            'value': rand(1000, 10000000),
        }
        nodes.append({
            'id': f'kg{i:03d}',
            'node_type': nt,
            'node_label': f'{nt.replace("_"," ").title()} {i+1}',
            'node_properties': json.dumps(props),
            'neo4j_id': random.randint(100000, 999999),
            'is_synced': random.random() < 0.85,
            'is_active': True,
        })

    edges = []
    for i in range(EDGE_COUNT):
        src = pick(nodes)
        tgt = pick(nodes)
        while tgt['id'] == src['id']:
            tgt = pick(nodes)
        edges.append({
            'id': f'ke{i:03d}',
            'edge_type': pick(EDGE_TYPES),
            'source_node_id': src['id'],
            'target_node_id': tgt['id'],
            'edge_properties': json.dumps({'weight': rand(0.1, 1.0), 'label': f'{src["node_type"]}-{tgt["node_type"]}'}),
            'neo4j_id': random.randint(100000, 999999),
            'is_synced': random.random() < 0.85,
            'is_active': True,
        })

    queue = []
    for i in range(QUEUE_COUNT):
        queue.append({
            'id': f'q{i:03d}',
            'operation': pick(['CREATE', 'UPDATE', 'DELETE', 'SYNC']),
            'entity_type': pick(NODE_TYPES),
            'entity_id': f'kg{random.randint(0,NODE_COUNT-1):03d}',
            'payload': json.dumps({'change': pick(['node_added', 'node_updated', 'edge_added'])}),
            'status': pick(['pending', 'processing', 'completed', 'failed']),
            'retry_count': random.randint(0, 2),
            'max_retries': 3,
            'created_at': (datetime.now() - timedelta(hours=random.randint(1, 72))).strftime('%Y-%m-%dT%H:%M:%SZ'),
        })

    topics = []
    for i in range(TOPIC_COUNT):
        topics.append({
            'id': f'kt{i:03d}',
            'topic_name': f'oce.{pick(["project","finance","hse","quality","contract","change"])}.{pick(["event","command","notification"])}',
            'description': f'Kafka topic #{i+1}',
            'partitions': random.choice([1, 3, 6, 12]),
            'replication_factor': random.choice([1, 2, 3]),
            'config': json.dumps({'cleanup.policy': 'delete', 'retention.ms': 604800000}),
            'is_internal': random.random() < 0.3,
            'is_active': True,
        })

    events = []
    for i in range(EVENT_COUNT):
        t = pick(topics) if topics else None
        events.append({
            'id': f'ev{i:03d}',
            'topic_id': t['id'] if t else None,
            'topic_name': t['topic_name'] if t else 'oce.default',
            'event_type': pick(EVENT_TYPES),
            'event_key': f'key-{random.randint(1,100)}',
            'event_value': json.dumps({'id': f'obj-{i:04d}', 'action': pick(['created','updated','deleted'])}),
            'headers': json.dumps({'source': 'core-api', 'user': 'system'}),
            'partition': random.randint(0, 5),
            'offset': random.randint(1, 99999),
            'created_at': (datetime.now() - timedelta(hours=random.randint(1, 168))).strftime('%Y-%m-%dT%H:%M:%SZ'),
        })

    consumers = []
    for i in range(CONSUMER_COUNT):
        consumers.append({
            'id': f'kc{i:03d}',
            'consumer_name': pick(['funding-sync', 'notification-svc', 'analytics-pipeline', 'audit-logger', 'dashboard-updater']),
            'consumer_group': f'oce-{pick(["funding","notify","analytics","audit"])}-group',
            'topic_name': f'oce.{pick(["project","finance","hse","quality","change"])}.{pick(["event","command"])}',
            'status': pick(['running', 'stopped', 'lagging']),
            'last_heartbeat': (datetime.now() - timedelta(seconds=random.randint(1, 300))).strftime('%Y-%m-%dT%H:%M:%SZ'),
            'lag_count': random.randint(0, 500),
            'config': json.dumps({'auto.offset.reset': 'earliest', 'enable.auto.commit': True}),
            'is_active': True,
        })

    return {
        'knowledge_graph_nodes': nodes,
        'knowledge_graph_edges': edges,
        'graph_sync_queue': queue,
        'kafka_topics': topics,
        'kafka_events': events,
        'kafka_consumers': consumers,
    }

def main():
    data = generate_data()
    if '--json' in sys.argv:
        print(json.dumps(data, ensure_ascii=False, indent=2))
    elif '--sql' in sys.argv:
        for table, rows in data.items():
            for row in rows:
                cols = ', '.join(row.keys())
                vals = ', '.join(f"'{v}'" if isinstance(v, str) else str(v) for v in row.values())
                print(f"INSERT INTO {table} ({cols}) VALUES ({vals});")
    else:
        print(json.dumps({k: len(v) for k, v in data.items()}, ensure_ascii=False))

if __name__ == '__main__':
    main()