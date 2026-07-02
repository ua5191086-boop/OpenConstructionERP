// OpenAPI 3.0 specification for OpenConstructionERP
// Generated from Go handlers
const swaggerSpec = {
  openapi: "3.0.3",
  info: {
    title: "OpenConstructionERP API",
    description: "Open-source platform for managing the full lifecycle of large infrastructure construction projects — Metro, Tunnels, Railways, Hydraulic Structures, EPC/EPCM",
    version: "0.1.0",
    contact: {
      name: "OpenConstructionERP Team",
      url: "https://github.com/ua5191086-boop/OpenConstructionERP"
    }
  },
  servers: [
    { url: "http://localhost:8085", description: "Local development" },
    { url: "https://api.openconstructionerp.com", description: "Production" }
  ],
  paths: {
    "/health": {
      get: {
        summary: "Health check",
        operationId: "healthCheck",
        responses: {
          "200": {
            description: "Service is healthy",
            content: { "application/json": { schema: { type: "object", properties: { status: { type: "string" }, service: { type: "string" }, version: { type: "string" }, time: { type: "string", format: "date-time" } } } } }
          }
        }
      }
    },
    "/api/v1/projects": {
      get: {
        summary: "List all projects",
        operationId: "listProjects",
        parameters: [
          { name: "status", in: "query", schema: { type: "string" }, description: "Filter by status (lead, tender, mobilization, execution, commissioning, dlp, closed)" },
          { name: "limit", in: "query", schema: { type: "integer", default: 50 } },
          { name: "offset", in: "query", schema: { type: "integer", default: 0 } }
        ],
        responses: {
          "200": { description: "List of projects", content: { "application/json": { schema: { type: "array", items: { "$ref": "#/components/schemas/Project" } } } } }
        }
      },
      post: {
        summary: "Create a new project",
        operationId: "createProject",
        requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/ProjectInput" } } } },
        responses: { "201": { description: "Project created" } }
      }
    },
    "/api/v1/projects/{id}": {
      get: {
        summary: "Get project by ID",
        operationId: "getProject",
        parameters: [{ name: "id", in: "path", required: true, schema: { type: "string", format: "uuid" } }],
        responses: { "200": { description: "Project details" } }
      },
      put: {
        summary: "Update project",
        operationId: "updateProject",
        parameters: [{ name: "id", in: "path", required: true, schema: { type: "string", format: "uuid" } }],
        requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/ProjectInput" } } } },
        responses: { "200": { description: "Project updated" } }
      },
      delete: {
        summary: "Delete project",
        operationId: "deleteProject",
        parameters: [{ name: "id", in: "path", required: true, schema: { type: "string", format: "uuid" } }],
        responses: { "204": { description: "Project deleted" } }
      }
    },
    "/api/v1/boq-items": {
      get: { summary: "List BOQ items", operationId: "listBoqItems", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "BOQ items list" } } },
      post: { summary: "Create BOQ item", operationId: "createBoqItem", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/BoqItemInput" } } } }, responses: { "201": { description: "BOQ item created" } } }
    },
    "/api/v1/contracts": {
      get: { summary: "List contracts", operationId: "listContracts", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "Contracts list" } } },
      post: { summary: "Create contract", operationId: "createContract", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/ContractInput" } } } }, responses: { "201": { description: "Contract created" } } }
    },
    "/api/v1/tbm": {
      get: { summary: "List TBM machines", operationId: "listTbm", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "TBM list" } } },
      post: { summary: "Register TBM", operationId: "createTbm", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/TbmInput" } } } }, responses: { "201": { description: "TBM registered" } } }
    },
    "/api/v1/tunnel-rings": {
      get: { summary: "List tunnel rings", operationId: "listTunnelRings", parameters: [{ name: "drive_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }, { name: "limit", in: "query", schema: { type: "integer", default: 100 } }], responses: { "200": { description: "Tunnel rings list" } } },
      post: { summary: "Record tunnel ring", operationId: "createTunnelRing", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/TunnelRingInput" } } } }, responses: { "201": { description: "Ring recorded" } } }
    },
    "/api/v1/hse-incidents": {
      get: { summary: "List HSE incidents", operationId: "listHseIncidents", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "HSE incidents list" } } },
      post: { summary: "Report HSE incident", operationId: "createHseIncident", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/HseIncidentInput" } } } }, responses: { "201": { description: "Incident reported" } } }
    },
    "/api/v1/ncrs": {
      get: { summary: "List NCRs", operationId: "listNcrs", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "NCRs list" } } },
      post: { summary: "Create NCR", operationId: "createNcr", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/NcrInput" } } } }, responses: { "201": { description: "NCR created" } } }
    },
    "/api/v1/risks": {
      get: { summary: "List project risks", operationId: "listRisks", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "Risks list" } } },
      post: { summary: "Create risk", operationId: "createRisk", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/RiskInput" } } } }, responses: { "201": { description: "Risk created" } } }
    },
    "/api/v1/change-orders": {
      get: { summary: "List change orders", operationId: "listChangeOrders", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "Change orders list" } } },
      post: { summary: "Create change order", operationId: "createChangeOrder", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/ChangeOrderInput" } } } }, responses: { "201": { description: "Change order created" } } }
    },
    "/api/v1/physical-progress": {
      get: { summary: "List physical progress (IPC)", operationId: "listPhysicalProgress", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "Physical progress list" } } },
      post: { summary: "Record physical progress", operationId: "createPhysicalProgress", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/PhysicalProgressInput" } } } }, responses: { "201": { description: "Progress recorded" } } }
    },
    "/api/v1/evm": {
      get: { summary: "Get EVM data", operationId: "getEvmData", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "EVM data" } } }
    },
    "/api/v1/milestones": {
      get: { summary: "List project milestones", operationId: "listMilestones", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "Milestones list" } } },
      post: { summary: "Create milestone", operationId: "createMilestone", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/MilestoneInput" } } } }, responses: { "201": { description: "Milestone created" } } }
    },
    "/api/v1/documents": {
      get: { summary: "List documents", operationId: "listDocuments", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "Documents list" } } },
      post: { summary: "Upload document", operationId: "createDocument", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/DocumentInput" } } } }, responses: { "201": { description: "Document created" } } }
    },
    "/api/v1/equipment": {
      get: { summary: "List equipment", operationId: "listEquipment", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "Equipment list" } } },
      post: { summary: "Register equipment", operationId: "createEquipment", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/EquipmentInput" } } } }, responses: { "201": { description: "Equipment registered" } } }
    },
    "/api/v1/stakeholders": {
      get: { summary: "List stakeholders", operationId: "listStakeholders", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "Stakeholders list" } } },
      post: { summary: "Add stakeholder", operationId: "createStakeholder", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/StakeholderInput" } } } }, responses: { "201": { description: "Stakeholder added" } } }
    },
    "/api/v1/settlement-monitoring": {
      get: { summary: "List settlement points", operationId: "listSettlementPoints", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "Settlement points list" } } },
      post: { summary: "Add settlement point", operationId: "createSettlementPoint", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/SettlementPointInput" } } } }, responses: { "201": { description: "Settlement point added" } } }
    },
    "/api/v1/tbm-telemetry": {
      get: { summary: "Get TBM telemetry", operationId: "getTbmTelemetry", parameters: [{ name: "tbm_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }, { name: "limit", in: "query", schema: { type: "integer", default: 100 } }], responses: { "200": { description: "TBM telemetry data" } } },
      post: { summary: "Record TBM telemetry", operationId: "createTbmTelemetry", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/TbmTelemetryInput" } } } }, responses: { "201": { description: "Telemetry recorded" } } }
    },
    "/api/v1/cost-transactions": {
      get: { summary: "List cost transactions", operationId: "listCostTransactions", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "Cost transactions list" } } },
      post: { summary: "Record cost transaction", operationId: "createCostTransaction", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/CostTransactionInput" } } } }, responses: { "201": { description: "Transaction recorded" } } }
    },
    "/api/v1/budget": {
      get: { summary: "Get project budget", operationId: "getBudget", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "Budget data" } } },
      post: { summary: "Create budget version", operationId: "createBudgetVersion", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/BudgetVersionInput" } } } }, responses: { "201": { description: "Budget version created" } } }
    },
    "/api/v1/wbs": {
      get: { summary: "Get WBS structure", operationId: "getWbs", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "WBS structure" } } },
      post: { summary: "Create WBS item", operationId: "createWbsItem", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/WbsItemInput" } } } }, responses: { "201": { description: "WBS item created" } } }
    },
    "/api/v1/lessons-learned": {
      get: { summary: "List lessons learned", operationId: "listLessonsLearned", parameters: [{ name: "project_id", in: "query", required: true, schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "Lessons learned list" } } },
      post: { summary: "Add lesson learned", operationId: "createLessonLearned", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/LessonLearnedInput" } } } }, responses: { "201": { description: "Lesson learned added" } } }
    },
    "/api/v1/employees": {
      get: { summary: "List employees", operationId: "listEmployees", parameters: [{ name: "project_id", in: "query", schema: { type: "string", format: "uuid" } }], responses: { "200": { description: "Employees list" } } },
      post: { summary: "Create employee", operationId: "createEmployee", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/EmployeeInput" } } } }, responses: { "201": { description: "Employee created" } } }
    },
    "/api/v1/organizations": {
      get: { summary: "List organizations", operationId: "listOrganizations", responses: { "200": { description: "Organizations list" } } },
      post: { summary: "Create organization", operationId: "createOrganization", requestBody: { required: true, content: { "application/json": { schema: { "$ref": "#/components/schemas/OrganizationInput" } } } }, responses: { "201": { description: "Organization created" } } }
    }
  },
  components: {
    schemas: {
      Project: {
        type: "object",
        properties: {
          id: { type: "string", format: "uuid" },
          code: { type: "string" },
          name: { type: "string" },
          name_ru: { type: "string" },
          project_type: { type: "string", enum: ["metro", "tunnel", "railway", "hydro", "microtunnel", "road", "other"] },
          status: { type: "string", enum: ["lead", "tender", "mobilization", "execution", "commissioning", "dlp", "closed", "cancelled"] },
          country: { type: "string" },
          currency: { type: "string" },
          start_date: { type: "string", format: "date" },
          finish_date: { type: "string", format: "date" },
          contract_value: { type: "number" },
          priority: { type: "string", enum: ["low", "medium", "high"] }
        }
      },
      ProjectInput: {
        type: "object",
        required: ["code", "name", "project_type"],
        properties: {
          code: { type: "string" },
          name: { type: "string" },
          name_ru: { type: "string" },
          project_type: { type: "string" },
          status: { type: "string" },
          country: { type: "string" },
          currency: { type: "string" },
          start_date: { type: "string", format: "date" },
          finish_date: { type: "string", format: "date" },
          contract_value: { type: "number" }
        }
      },
      BoqItemInput: {
        type: "object",
        required: ["project_id", "code", "name", "unit", "quantity", "unit_price"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          code: { type: "string" },
          name: { type: "string" },
          description: { type: "string" },
          unit: { type: "string" },
          quantity: { type: "number" },
          unit_price: { type: "number" },
          currency: { type: "string" }
        }
      },
      ContractInput: {
        type: "object",
        required: ["project_id", "code", "name", "contract_type"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          code: { type: "string" },
          name: { type: "string" },
          contract_type: { type: "string", enum: ["main", "subcontract", "supply", "services", "framework"] },
          contract_form: { type: "string", enum: ["fidic_red", "fidic_yellow", "fidic_silver", "epc", "epcm", "bespoke"] },
          currency: { type: "string" },
          contract_value: { type: "number" },
          start_date: { type: "string", format: "date" },
          end_date: { type: "string", format: "date" }
        }
      },
      TbmInput: {
        type: "object",
        required: ["project_id", "code", "tbm_type", "diameter_mm"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          code: { type: "string" },
          manufacturer: { type: "string" },
          model: { type: "string" },
          tbm_type: { type: "string", enum: ["EPB", "SLURRY", "OPEN", "MIXSHIELD", "GRIPPER", "MTBM"] },
          diameter_mm: { type: "integer" }
        }
      },
      TunnelRingInput: {
        type: "object",
        required: ["drive_id", "ring_no"],
        properties: {
          drive_id: { type: "string", format: "uuid" },
          ring_no: { type: "integer" },
          chainage: { type: "number" },
          shift: { type: "string", enum: ["day", "night", "A", "B", "C"] },
          ring_type: { type: "string" },
          advance_mm: { type: "integer" },
          grout_volume_m3: { type: "number" },
          grout_pressure_bar: { type: "number" }
        }
      },
      HseIncidentInput: {
        type: "object",
        required: ["project_id", "incident_type", "severity", "incident_date", "description"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          incident_type: { type: "string" },
          severity: { type: "string", enum: ["low", "medium", "high", "critical"] },
          incident_date: { type: "string", format: "date-time" },
          location: { type: "string" },
          description: { type: "string" },
          affected_employees: { type: "integer" },
          lost_days: { type: "integer" }
        }
      },
      NcrInput: {
        type: "object",
        required: ["project_id", "title", "description", "severity"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          title: { type: "string" },
          description: { type: "string" },
          severity: { type: "string", enum: ["minor", "major", "critical"] },
          location: { type: "string" },
          boq_item_id: { type: "string", format: "uuid" },
          ring_id: { type: "string", format: "uuid" }
        }
      },
      RiskInput: {
        type: "object",
        required: ["project_id", "risk_code", "name", "risk_category", "risk_type", "probability", "impact"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          risk_code: { type: "string" },
          name: { type: "string" },
          description: { type: "string" },
          risk_category: { type: "string" },
          risk_type: { type: "string" },
          probability: { type: "string" },
          impact: { type: "string" },
          potential_cost: { type: "number" },
          mitigation_plan: { type: "string" }
        }
      },
      ChangeOrderInput: {
        type: "object",
        required: ["project_id", "co_code", "co_name", "co_type"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          co_code: { type: "string" },
          co_name: { type: "string" },
          co_type: { type: "string", enum: ["variation", "change_directive", "claim_settlement", "compensation_event", "other"] },
          scope_change: { type: "string" },
          cost_change: { type: "number" },
          cost_currency: { type: "string" },
          schedule_change_days: { type: "integer" }
        }
      },
      PhysicalProgressInput: {
        type: "object",
        required: ["project_id", "contract_id", "measurement_date", "item_code", "unit", "contract_quantity", "current_qty", "unit_price"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          contract_id: { type: "string", format: "uuid" },
          boq_item_id: { type: "string", format: "uuid" },
          measurement_date: { type: "string", format: "date" },
          item_code: { type: "string" },
          description: { type: "string" },
          unit: { type: "string" },
          contract_quantity: { type: "number" },
          current_qty: { type: "number" },
          unit_price: { type: "number" }
        }
      },
      MilestoneInput: {
        type: "object",
        required: ["project_id", "milestone_code", "name", "milestone_type", "planned_date"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          milestone_code: { type: "string" },
          name: { type: "string" },
          description: { type: "string" },
          milestone_type: { type: "string" },
          planned_date: { type: "string", format: "date" },
          weight_pct: { type: "number" },
          is_gate: { type: "boolean" }
        }
      },
      DocumentInput: {
        type: "object",
        required: ["project_id", "doc_number", "title", "doc_type"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          doc_number: { type: "string" },
          title: { type: "string" },
          doc_type: { type: "string" },
          discipline: { type: "string" },
          originator: { type: "string" },
          revision: { type: "string" }
        }
      },
      EquipmentInput: {
        type: "object",
        required: ["project_id", "equipment_code", "equipment_name", "equipment_type"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          equipment_code: { type: "string" },
          equipment_name: { type: "string" },
          equipment_type: { type: "string" },
          manufacturer: { type: "string" },
          model: { type: "string" },
          status: { type: "string" },
          purchase_cost: { type: "number" }
        }
      },
      StakeholderInput: {
        type: "object",
        required: ["project_id", "stakeholder_type", "name"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          stakeholder_type: { type: "string" },
          name: { type: "string" },
          organization: { type: "string" },
          interest_level: { type: "string" },
          influence_level: { type: "string" }
        }
      },
      SettlementPointInput: {
        type: "object",
        required: ["project_id", "point_code", "chainage_m", "initial_level_m"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          point_code: { type: "string" },
          point_name: { type: "string" },
          point_type: { type: "string" },
          chainage_m: { type: "number" },
          offset_m: { type: "number" },
          initial_level_m: { type: "number" }
        }
      },
      TbmTelemetryInput: {
        type: "object",
        required: ["tbm_id", "recorded_at"],
        properties: {
          tbm_id: { type: "string", format: "uuid" },
          recorded_at: { type: "string", format: "date-time" },
          thrust_force_kn: { type: "number" },
          torque_knm: { type: "number" },
          advance_rate_mmmin: { type: "number" },
          face_pressure_bar: { type: "number" }
        }
      },
      CostTransactionInput: {
        type: "object",
        required: ["project_id", "boq_item_id", "cbs_chapter_id", "transaction_type", "amount", "period"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          boq_item_id: { type: "string", format: "uuid" },
          cbs_chapter_id: { type: "string", format: "uuid" },
          contract_id: { type: "string", format: "uuid" },
          transaction_type: { type: "string", enum: ["Plan", "Actual", "Forecast", "Variance", "Commitment"] },
          amount: { type: "number" },
          currency: { type: "string" },
          period: { type: "string", format: "date" },
          description: { type: "string" }
        }
      },
      BudgetVersionInput: {
        type: "object",
        required: ["project_id", "version_number", "version_name", "total_amount"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          version_number: { type: "integer" },
          version_name: { type: "string" },
          total_amount: { type: "number" },
          notes: { type: "string" }
        }
      },
      WbsItemInput: {
        type: "object",
        required: ["project_id", "wbs_code", "name", "wbs_level"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          wbs_code: { type: "string" },
          name: { type: "string" },
          wbs_level: { type: "integer" },
          is_leaf: { type: "boolean" },
          planned_start: { type: "string", format: "date" },
          planned_end: { type: "string", format: "date" },
          planned_cost: { type: "number" }
        }
      },
      LessonLearnedInput: {
        type: "object",
        required: ["project_id", "title", "description", "category"],
        properties: {
          project_id: { type: "string", format: "uuid" },
          title: { type: "string" },
          description: { type: "string" },
          category: { type: "string" },
          severity: { type: "string" },
          root_cause: { type: "string" },
          impact: { type: "string" },
          recommendation: { type: "string" }
        }
      },
      EmployeeInput: {
        type: "object",
        required: ["employee_code", "full_name", "position", "hire_date", "salary_currency"],
        properties: {
          employee_code: { type: "string" },
          full_name: { type: "string" },
          first_name: { type: "string" },
          last_name: { type: "string" },
          position: { type: "string" },
          department: { type: "string" },
          email: { type: "string", format: "email" },
          phone: { type: "string" },
          hire_date: { type: "string", format: "date" },
          salary_currency: { type: "string" }
        }
      },
      OrganizationInput: {
        type: "object",
        required: ["code", "name", "org_type"],
        properties: {
          code: { type: "string" },
          name: { type: "string" },
          org_type: { type: "string", enum: ["holding", "contractor", "client", "consultant", "subcontractor", "supplier", "bank"] },
          country: { type: "string" }
        }
      }
    }
  }
};

module.exports = swaggerSpec;
