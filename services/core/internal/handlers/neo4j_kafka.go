package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/openconstructionerp/oce/services/core/internal/models"
)

// Neo4jKafkaHandler handles Neo4j+Kafka module endpoints (V028)
type Neo4jKafkaHandler struct {
	db *sql.DB
}

func NewNeo4jKafkaHandler(db *sql.DB) *Neo4jKafkaHandler {
	return &Neo4jKafkaHandler{db: db}
}

func (h *Neo4jKafkaHandler) RegisterRoutes(r chi.Router) {
	r.Route("/knowledge", func(r chi.Router) {
		r.Get("/graph", h.GetFullGraph)
		r.Get("/nodes/{type}", h.GetNodesByType)
		r.Get("/relationships/{type}", h.GetEdgesByType)
		r.Post("/sync", h.TriggerSync)

		r.Get("/nodes", h.ListNodes)
		r.Post("/nodes", h.CreateNode)
		r.Get("/nodes/detail/{id}", h.GetNode)

		r.Get("/edges", h.ListEdges)
		r.Post("/edges", h.CreateEdge)
	})

	r.Route("/events", func(r chi.Router) {
		r.Get("/topics", h.ListTopics)
		r.Post("/publish", h.PublishEvent)
		r.Get("/events", h.ListEvents)
		r.Get("/consumers", h.ListConsumers)
	})
}

// --- Knowledge Graph Nodes ---

func (h *Neo4jKafkaHandler) ListNodes(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id,node_type,node_label,node_properties,neo4j_id,is_synced,is_active,created_at,updated_at FROM knowledge_graph_nodes ORDER BY created_at DESC`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.KnowledgeGraphNode, 0)
	for rows.Next() {
		var m models.KnowledgeGraphNode
		if err := rows.Scan(&m.ID, &m.NodeType, &m.NodeLabel, &m.NodeProperties, &m.Neo4jID, &m.IsSynced, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *Neo4jKafkaHandler) CreateNode(w http.ResponseWriter, r *http.Request) {
	var input struct {
		NodeType       string  `json:"node_type"`
		NodeLabel      *string `json:"node_label"`
		NodeProperties *string `json:"node_properties"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO knowledge_graph_nodes (id,node_type,node_label,node_properties,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6)`,
		id, input.NodeType, input.NodeLabel, input.NodeProperties, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *Neo4jKafkaHandler) GetNode(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.KnowledgeGraphNode
	err := h.db.QueryRow(`SELECT id,node_type,node_label,node_properties,neo4j_id,is_synced,is_active,created_at,updated_at FROM knowledge_graph_nodes WHERE id=$1`, id).Scan(
		&m.ID, &m.NodeType, &m.NodeLabel, &m.NodeProperties, &m.Neo4jID, &m.IsSynced, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, m)
}

func (h *Neo4jKafkaHandler) GetNodesByType(w http.ResponseWriter, r *http.Request) {
	nodeType := chi.URLParam(r, "type")
	rows, err := h.db.Query(`SELECT id,node_type,node_label,node_properties,neo4j_id,is_synced,is_active,created_at,updated_at FROM knowledge_graph_nodes WHERE node_type=$1 AND is_active=TRUE`, nodeType)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.KnowledgeGraphNode, 0)
	for rows.Next() {
		var m models.KnowledgeGraphNode
		if err := rows.Scan(&m.ID, &m.NodeType, &m.NodeLabel, &m.NodeProperties, &m.Neo4jID, &m.IsSynced, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

// --- Knowledge Graph Edges ---

func (h *Neo4jKafkaHandler) ListEdges(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id,edge_type,source_node_id,target_node_id,edge_properties,neo4j_id,is_synced,is_active,created_at FROM knowledge_graph_edges ORDER BY created_at DESC`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.KnowledgeGraphEdge, 0)
	for rows.Next() {
		var m models.KnowledgeGraphEdge
		if err := rows.Scan(&m.ID, &m.EdgeType, &m.SourceNodeID, &m.TargetNodeID, &m.EdgeProperties, &m.Neo4jID, &m.IsSynced, &m.IsActive, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *Neo4jKafkaHandler) CreateEdge(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EdgeType       string  `json:"edge_type"`
		SourceNodeID   string  `json:"source_node_id"`
		TargetNodeID   string  `json:"target_node_id"`
		EdgeProperties *string `json:"edge_properties"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO knowledge_graph_edges (id,edge_type,source_node_id,target_node_id,edge_properties,created_at) VALUES($1,$2,$3,$4,$5,NOW())`,
		id, input.EdgeType, input.SourceNodeID, input.TargetNodeID, input.EdgeProperties)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *Neo4jKafkaHandler) GetEdgesByType(w http.ResponseWriter, r *http.Request) {
	edgeType := chi.URLParam(r, "type")
	rows, err := h.db.Query(`SELECT id,edge_type,source_node_id,target_node_id,edge_properties,neo4j_id,is_synced,is_active,created_at FROM knowledge_graph_edges WHERE edge_type=$1 AND is_active=TRUE`, edgeType)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.KnowledgeGraphEdge, 0)
	for rows.Next() {
		var m models.KnowledgeGraphEdge
		if err := rows.Scan(&m.ID, &m.EdgeType, &m.SourceNodeID, &m.TargetNodeID, &m.EdgeProperties, &m.Neo4jID, &m.IsSynced, &m.IsActive, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

// --- Full Graph ---

func (h *Neo4jKafkaHandler) GetFullGraph(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.db.Query(`SELECT id,node_type,node_label,node_properties,neo4j_id,is_synced,is_active,created_at,updated_at FROM knowledge_graph_nodes WHERE is_active=TRUE`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer nodes.Close()

	nodeList := make([]models.KnowledgeGraphNode, 0)
	for nodes.Next() {
		var m models.KnowledgeGraphNode
		if err := nodes.Scan(&m.ID, &m.NodeType, &m.NodeLabel, &m.NodeProperties, &m.Neo4jID, &m.IsSynced, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		nodeList = append(nodeList, m)
	}

	edges, err := h.db.Query(`SELECT id,edge_type,source_node_id,target_node_id,edge_properties,neo4j_id,is_synced,is_active,created_at FROM knowledge_graph_edges WHERE is_active=TRUE`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer edges.Close()

	edgeList := make([]models.KnowledgeGraphEdge, 0)
	for edges.Next() {
		var m models.KnowledgeGraphEdge
		if err := edges.Scan(&m.ID, &m.EdgeType, &m.SourceNodeID, &m.TargetNodeID, &m.EdgeProperties, &m.Neo4jID, &m.IsSynced, &m.IsActive, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		edgeList = append(edgeList, m)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"nodes": nodeList,
		"edges": edgeList,
	})
}

// --- Graph Sync ---

func (h *Neo4jKafkaHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO graph_sync_queue (id,operation,entity_type,entity_id,payload,status,created_at) VALUES($1,'sync_all','graph','all','{}','pending',NOW())`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"sync_id": id, "status": "queued"})
}

// --- Kafka Topics ---

func (h *Neo4jKafkaHandler) ListTopics(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id,topic_name,description,partitions,replication_factor,config,is_internal,is_active,created_at FROM kafka_topics ORDER BY topic_name`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.KafkaTopic, 0)
	for rows.Next() {
		var m models.KafkaTopic
		if err := rows.Scan(&m.ID, &m.TopicName, &m.Description, &m.Partitions, &m.ReplicationFactor, &m.Config, &m.IsInternal, &m.IsActive, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

// --- Kafka Events ---

func (h *Neo4jKafkaHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	topicName := r.URL.Query().Get("topic")
	query := `SELECT id,topic_id,topic_name,event_type,event_key,event_value,headers,partition_id,offset_id,producer,status,created_at FROM kafka_events`
	var rows *sql.Rows
	var err error
	if topicName != "" {
		rows, err = h.db.Query(query+` WHERE topic_name=$1 ORDER BY created_at DESC LIMIT 100`, topicName)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY created_at DESC LIMIT 100`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.KafkaEvent, 0)
	for rows.Next() {
		var m models.KafkaEvent
		if err := rows.Scan(&m.ID, &m.TopicID, &m.TopicName, &m.EventType, &m.EventKey, &m.EventValue, &m.Headers, &m.Partition, &m.Offset, &m.Producer, &m.Status, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *Neo4jKafkaHandler) PublishEvent(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TopicName  string  `json:"topic_name"`
		EventType  *string `json:"event_type"`
		EventKey   *string `json:"event_key"`
		EventValue *string `json:"event_value"`
		Headers    *string `json:"headers"`
		Producer   *string `json:"producer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO kafka_events (id,topic_name,event_type,event_key,event_value,headers,producer,status,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,'published',NOW())`,
		id, input.TopicName, input.EventType, input.EventKey, input.EventValue, input.Headers, input.Producer)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"event_id": id, "status": "published"})
}

// --- Kafka Consumers ---

func (h *Neo4jKafkaHandler) ListConsumers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id,consumer_name,group_id,topic_pattern,description,status,config,is_active,created_at,updated_at FROM kafka_consumers ORDER BY consumer_name`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.KafkaConsumer, 0)
	for rows.Next() {
		var m models.KafkaConsumer
		if err := rows.Scan(&m.ID, &m.ConsumerName, &m.GroupID, &m.TopicPattern, &m.Description, &m.Status, &m.Config, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}