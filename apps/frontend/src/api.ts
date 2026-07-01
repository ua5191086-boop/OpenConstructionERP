// ============================================================
// API client for OpenConstructionERP Go backend
// ============================================================

const API_BASE = '/api';

async function fetchApi<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...options?.headers },
    ...options,
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`API ${res.status}: ${text}`);
  }
  const contentType = res.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    return res.json();
  }
  return (await res.text()) as unknown as T;
}

// --- BOQ ---
export const boqApi = {
  listSections: () => fetchApi<any[]>('/boq/sections'),
  listComplexes: () => fetchApi<any[]>('/boq/complexes'),
  listObjects: () => fetchApi<any[]>('/boq/objects'),
  listItems: (params?: { section_id?: string; cbs_code?: string }) => {
    const qs = new URLSearchParams();
    if (params?.section_id) qs.set('section_id', params.section_id);
    if (params?.cbs_code) qs.set('cbs_code', params.cbs_code);
    const q = qs.toString();
    return fetchApi<any[]>(`/boq/items${q ? '?' + q : ''}`);
  },
  listCBSChapters: () => fetchApi<any[]>('/boq/cbs-chapters'),
};

// --- Tenders ---
export const tendersApi = {
  list: (params?: { status?: string }) => {
    const qs = params?.status ? `?status=${params.status}` : '';
    return fetchApi<any[]>(`/tenders${qs}`);
  },
  get: (id: string) => fetchApi<any>(`/tenders/${id}`),
  listLots: (tenderId: string) => fetchApi<any[]>(`/tenders/${tenderId}/lots`),
  listBidders: (tenderId: string) => fetchApi<any[]>(`/tenders/${tenderId}/bidders`),
};

// --- Contracts ---
export const contractsApi = {
  list: (params?: { status?: string }) => {
    const qs = params?.status ? `?status=${params.status}` : '';
    return fetchApi<any[]>(`/contracts${qs}`);
  },
  get: (id: string) => fetchApi<any>(`/contracts/${id}`),
  listMilestones: (contractId: string) => fetchApi<any[]>(`/contracts/${contractId}/milestones`),
  listPayments: (contractId: string) => fetchApi<any[]>(`/contracts/${contractId}/payments`),
};

// --- HR ---
export const hrApi = {
  listEmployees: (params?: { status?: string; department?: string }) => {
    const qs = new URLSearchParams();
    if (params?.status) qs.set('status', params.status);
    if (params?.department) qs.set('department', params.department);
    const q = qs.toString();
    return fetchApi<any[]>(`/hr/employees${q ? '?' + q : ''}`);
  },
  listDepartments: () => fetchApi<any[]>('/hr/departments'),
  listAttendance: () => fetchApi<any[]>('/hr/attendance'),
  listLeaves: () => fetchApi<any[]>('/hr/leaves'),
};

// --- Finance ---
export const financeApi = {
  listBudgets: () => fetchApi<any[]>('/finance/budgets'),
  listBudgetItems: () => fetchApi<any[]>('/finance/budget-items'),
  listCashFlow: () => fetchApi<any[]>('/finance/cash-flow'),
  listInvoices: () => fetchApi<any[]>('/finance/invoices'),
};

// --- Procurement ---
export const procurementApi = {
  listRequests: (params?: { status?: string }) => {
    const qs = params?.status ? `?status=${params.status}` : '';
    return fetchApi<any[]>(`/procurement/requests${qs}`);
  },
  listPurchaseOrders: () => fetchApi<any[]>('/procurement/purchase-orders'),
  listInventory: () => fetchApi<any[]>('/procurement/inventory'),
};

// --- BIM ---
export const bimApi = {
  listModels: (params?: { discipline?: string }) => {
    const qs = params?.discipline ? `?discipline=${params.discipline}` : '';
    return fetchApi<any[]>(`/bim/models${qs}`);
  },
  listElements: () => fetchApi<any[]>('/bim/elements'),
  listClashes: () => fetchApi<any[]>('/bim/clashes'),
};

// --- AI ---
export const aiApi = {
  listAgents: () => fetchApi<any[]>('/ai/agents'),
  listTasks: () => fetchApi<any[]>('/ai/tasks'),
  listConversations: () => fetchApi<any[]>('/ai/conversations'),
};

// --- Project Management ---
export const pmApi = {
  listProjects: (params?: { status?: string; project_type?: string }) => {
    const qs = new URLSearchParams();
    if (params?.status) qs.set('status', params.status);
    if (params?.project_type) qs.set('project_type', params.project_type);
    const q = qs.toString();
    return fetchApi<any[]>(`/pm/projects${q ? '?' + q : ''}`);
  },
  getProject: (id: string) => fetchApi<any>(`/pm/projects/${id}`),
  listWBSItems: (projectId?: string) => {
    const qs = projectId ? `?project_id=${projectId}` : '';
    return fetchApi<any[]>(`/pm/wbs-items${qs}`);
  },
  listMilestones: (projectId?: string) => {
    const qs = projectId ? `?project_id=${projectId}` : '';
    return fetchApi<any[]>(`/pm/milestones${qs}`);
  },
  listPhases: (projectId?: string) => {
    const qs = projectId ? `?project_id=${projectId}` : '';
    return fetchApi<any[]>(`/pm/phases${qs}`);
  },
  listTeam: (projectId?: string) => {
    const qs = projectId ? `?project_id=${projectId}` : '';
    return fetchApi<any[]>(`/pm/team${qs}`);
  },
  listPortfolios: () => fetchApi<any[]>('/pm/portfolios'),
  listRisks: (projectId?: string) => {
    const qs = projectId ? `?project_id=${projectId}` : '';
    return fetchApi<any[]>(`/pm/risks${qs}`);
  },
  listChanges: (projectId?: string) => {
    const qs = projectId ? `?project_id=${projectId}` : '';
    return fetchApi<any[]>(`/pm/changes${qs}`);
  },
  listLessons: (projectId?: string) => {
    const qs = projectId ? `?project_id=${projectId}` : '';
    return fetchApi<any[]>(`/pm/lessons${qs}`);
  },
  getDashboard: () => fetchApi<any>('/pm/dashboard'),
};

// --- Schedule Management ---
export const scheduleApi = {
  listSchedules: (params?: { schedule_type?: string }) => {
    const qs = params?.schedule_type ? `?schedule_type=${params.schedule_type}` : '';
    return fetchApi<any[]>(`/api/v1/schedule/schedules${qs}`);
  },
  getSchedule: (id: string) => fetchApi<any>(`/api/v1/schedule/schedules/${id}`),
  listActivities: (scheduleId?: string) => {
    const qs = scheduleId ? `?schedule_id=${scheduleId}` : '';
    return fetchApi<any[]>(`/api/v1/schedule/activities${qs}`);
  },
  listRelationships: (scheduleId?: string) => {
    const qs = scheduleId ? `?schedule_id=${scheduleId}` : '';
    return fetchApi<any[]>(`/api/v1/schedule/relationships${qs}`);
  },
  listResources: (scheduleId?: string) => {
    const qs = scheduleId ? `?schedule_id=${scheduleId}` : '';
    return fetchApi<any[]>(`/api/v1/schedule/resources${qs}`);
  },
  listBaselines: (scheduleId?: string) => {
    const qs = scheduleId ? `?schedule_id=${scheduleId}` : '';
    return fetchApi<any[]>(`/api/v1/schedule/baselines${qs}`);
  },
  listChanges: (scheduleId?: string) => {
    const qs = scheduleId ? `?schedule_id=${scheduleId}` : '';
    return fetchApi<any[]>(`/api/v1/schedule/changes${qs}`);
  },
  getSummary: () => fetchApi<any>('/api/v1/schedule/summary'),
};

// --- Equipment Management ---
export const equipmentApi = {
  listCategories: () => fetchApi<any[]>('/api/v1/equipment/categories'),
  listItems: (params?: { status?: string; category_id?: string }) => {
    const qs = new URLSearchParams();
    if (params?.status) qs.set('status', params.status);
    if (params?.category_id) qs.set('category_id', params.category_id);
    const q = qs.toString();
    return fetchApi<any[]>(`/api/v1/equipment/items${q ? '?' + q : ''}`);
  },
  getItem: (id: string) => fetchApi<any>(`/api/v1/equipment/items/${id}`),
  listMaintenance: (equipmentId?: string) => {
    const qs = equipmentId ? `?equipment_id=${equipmentId}` : '';
    return fetchApi<any[]>(`/api/v1/equipment/maintenance${qs}`);
  },
  listMaintenanceSchedules: (equipmentId?: string) => {
    const qs = equipmentId ? `?equipment_id=${equipmentId}` : '';
    return fetchApi<any[]>(`/api/v1/equipment/maintenance-schedules${qs}`);
  },
  listTelemetry: (equipmentId?: string) => {
    const qs = equipmentId ? `?equipment_id=${equipmentId}` : '';
    return fetchApi<any[]>(`/api/v1/equipment/telemetry${qs}`);
  },
  listFuel: (equipmentId?: string) => {
    const qs = equipmentId ? `?equipment_id=${equipmentId}` : '';
    return fetchApi<any[]>(`/api/v1/equipment/fuel${qs}`);
  },
  listDowntime: (equipmentId?: string) => {
    const qs = equipmentId ? `?equipment_id=${equipmentId}` : '';
    return fetchApi<any[]>(`/api/v1/equipment/downtime${qs}`);
  },
  listSpareParts: (equipmentId?: string) => {
    const qs = equipmentId ? `?equipment_id=${equipmentId}` : '';
    return fetchApi<any[]>(`/api/v1/equipment/spare-parts${qs}`);
  },
  getSummary: () => fetchApi<any>('/api/v1/equipment/summary'),
};

// --- HSE ---
export const hseApi = {
  listIncidents: (params?: { status?: string; severity?: string }) => {
    const qs = new URLSearchParams();
    if (params?.status) qs.set('status', params.status);
    if (params?.severity) qs.set('severity', params.severity);
    const q = qs.toString();
    return fetchApi<any[]>(`/api/v1/hse/incidents${q ? '?' + q : ''}`);
  },
  getIncident: (id: string) => fetchApi<any>(`/api/v1/hse/incidents/${id}`),
  listPermits: (params?: { status?: string }) => {
    const qs = params?.status ? `?status=${params.status}` : '';
    return fetchApi<any[]>(`/api/v1/hse/permits${qs}`);
  },
  listAudits: (params?: { status?: string }) => {
    const qs = params?.status ? `?status=${params.status}` : '';
    return fetchApi<any[]>(`/api/v1/hse/audits${qs}`);
  },
  listInspections: (params?: { status?: string }) => {
    const qs = params?.status ? `?status=${params.status}` : '';
    return fetchApi<any[]>(`/api/v1/hse/inspections${qs}`);
  },
  listTraining: (params?: { status?: string }) => {
    const qs = params?.status ? `?status=${params.status}` : '';
    return fetchApi<any[]>(`/api/v1/hse/training${qs}`);
  },
  listPPE: () => fetchApi<any[]>('/api/v1/hse/ppe'),
  listDrills: () => fetchApi<any[]>('/api/v1/hse/drills'),
  listStatistics: () => fetchApi<any[]>('/api/v1/hse/statistics'),
  listEmergencyPlans: () => fetchApi<any[]>('/api/v1/hse/emergency-plans'),
  listChemicals: () => fetchApi<any[]>('/api/v1/hse/chemicals'),
};

// --- Quality Management ---
export const qualityApi = {
  listITPs: (params?: { status?: string }) => {
    const qs = params?.status ? `?status=${params.status}` : '';
    return fetchApi<any[]>(`/api/v1/quality/itps${qs}`);
  },
  listInspections: (params?: { project_id?: string; status?: string }) => {
    const qs = new URLSearchParams();
    if (params?.project_id) qs.set('project_id', params.project_id);
    if (params?.status) qs.set('status', params.status);
    const q = qs.toString();
    return fetchApi<any[]>(`/api/v1/quality/inspections${q ? '?' + q : ''}`);
  },
  listTestResults: () => fetchApi<any[]>('/api/v1/quality/test-results'),
  listNCRs: (params?: { status?: string }) => {
    const qs = params?.status ? `?status=${params.status}` : '';
    return fetchApi<any[]>(`/api/v1/quality/ncrs${qs}`);
  },
  listCorrectiveActions: () => fetchApi<any[]>('/api/v1/quality/corrective-actions'),
  listCalibration: () => fetchApi<any[]>('/api/v1/quality/calibration'),
  listQualityMetrics: () => fetchApi<any[]>('/api/v1/quality/quality-metrics'),
};

// --- GIS & Survey ---
export const gisApi = {
  listLayers: () => fetchApi<any[]>('/api/v1/gis/layers'),
  listFeatures: () => fetchApi<any[]>('/api/v1/gis/features'),
  listSurveyPoints: () => fetchApi<any[]>('/api/v1/gis/survey-points'),
  listSurveyRuns: () => fetchApi<any[]>('/api/v1/gis/survey-runs'),
  listStations: () => fetchApi<any[]>('/api/v1/gis/survey-stations'),
  listAlignments: () => fetchApi<any[]>('/api/v1/gis/alignments'),
  listCrossSections: () => fetchApi<any[]>('/api/v1/gis/cross-sections'),
  listDroneFlights: () => fetchApi<any[]>('/api/v1/gis/drone-flights'),
};

// --- Risk Management ---
export const riskApi = {
  listCategories: () => fetchApi<any[]>('/api/v1/risk/categories'),
  listRegisters: (params?: { project_id?: string; status?: string }) => {
    const qs = new URLSearchParams();
    if (params?.project_id) qs.set('project_id', params.project_id);
    if (params?.status) qs.set('status', params.status);
    const q = qs.toString();
    return fetchApi<any[]>(`/api/v1/risk/registers${q ? '?' + q : ''}`);
  },
  listMatrices: () => fetchApi<any[]>('/api/v1/risk/matrices'),
  listMonteCarlo: () => fetchApi<any[]>('/api/v1/risk/monte-carlo'),
  listScenarios: () => fetchApi<any[]>('/api/v1/risk/scenarios'),
  listMitigations: () => fetchApi<any[]>('/api/v1/risk/mitigations'),
  listEscalations: () => fetchApi<any[]>('/api/v1/risk/escalations'),
  listDashboard: () => fetchApi<any[]>('/api/v1/risk/dashboard'),
};

// --- Change Management ---
export const changeApi = {
  listRequests: (params?: { project_id?: string; status?: string }) => {
    const qs = new URLSearchParams();
    if (params?.project_id) qs.set('project_id', params.project_id);
    if (params?.status) qs.set('status', params.status);
    const q = qs.toString();
    return fetchApi<any[]>(`/api/v1/change/requests${q ? '?' + q : ''}`);
  },
  listOrders: () => fetchApi<any[]>('/api/v1/change/orders'),
  listImpactAnalysis: () => fetchApi<any[]>('/api/v1/change/impact-analysis'),
  listApprovalWorkflow: () => fetchApi<any[]>('/api/v1/change/approval-workflow'),
  listChangeLog: () => fetchApi<any[]>('/api/v1/change/change-log'),
};

// --- TBM Management ---
export const tbmApi = {
  listTelemetry: (params?: { tbm_id?: string; limit?: number }) => {
    const qs = new URLSearchParams();
    if (params?.tbm_id) qs.set('tbm_id', params.tbm_id);
    if (params?.limit) qs.set('limit', String(params.limit));
    const q = qs.toString();
    return fetchApi<any[]>(`/api/v1/tbm/telemetry${q ? '?' + q : ''}`);
  },
  listAlarms: (params?: { tbm_id?: string; active?: boolean }) => {
    const qs = new URLSearchParams();
    if (params?.tbm_id) qs.set('tbm_id', params.tbm_id);
    if (params?.active) qs.set('active', 'true');
    const q = qs.toString();
    return fetchApi<any[]>(`/api/v1/tbm/alarms${q ? '?' + q : ''}`);
  },
  listOperators: () => fetchApi<any[]>('/api/v1/tbm/operators'),
  listShifts: (tbm_id?: string) => {
    const qs = tbm_id ? `?tbm_id=${tbm_id}` : '';
    return fetchApi<any[]>(`/api/v1/tbm/shifts${qs}`);
  },
  listConsumables: (params?: { tbm_id?: string; consumable_type?: string }) => {
    const qs = new URLSearchParams();
    if (params?.tbm_id) qs.set('tbm_id', params.tbm_id);
    if (params?.consumable_type) qs.set('consumable_type', params.consumable_type);
    const q = qs.toString();
    return fetchApi<any[]>(`/api/v1/tbm/consumables${q ? '?' + q : ''}`);
  },
  listPerformance: (tbm_id?: string) => {
    const qs = tbm_id ? `?tbm_id=${tbm_id}` : '';
    return fetchApi<any[]>(`/api/v1/tbm/performance${qs}`);
  },
  getSummary: () => fetchApi<any>('/api/v1/tbm/summary'),
};

// --- Ring Builder ---
export const ringBuilderApi = {
  listDesigns: (params?: { project_id?: string }) => {
    const qs = params?.project_id ? `?project_id=${params.project_id}` : '';
    return fetchApi<any[]>(`/api/v1/ringbuilder/designs${qs}`);
  },
  listProduction: (params?: { project_id?: string; status?: string }) => {
    const qs = new URLSearchParams();
    if (params?.project_id) qs.set('project_id', params.project_id);
    if (params?.status) qs.set('status', params.status);
    const q = qs.toString();
    return fetchApi<any[]>(`/api/v1/ringbuilder/production${q ? '?' + q : ''}`);
  },
  listQC: (segment_id?: string) => {
    const qs = segment_id ? `?segment_id=${segment_id}` : '';
    return fetchApi<any[]>(`/api/v1/ringbuilder/qc${qs}`);
  },
  listInventory: () => fetchApi<any[]>('/api/v1/ringbuilder/inventory'),
  listMeasurements: (ring_id?: string) => {
    const qs = ring_id ? `?ring_id=${ring_id}` : '';
    return fetchApi<any[]>(`/api/v1/ringbuilder/measurements${qs}`);
  },
  getSummary: () => fetchApi<any>('/api/v1/ringbuilder/summary'),
};

// --- NATM & Microtunnelling ---
export const natmApi = {
  listExcavation: (drive_id?: string) => {
    const qs = drive_id ? `?drive_id=${drive_id}` : '';
    return fetchApi<any[]>(`/api/v1/natm/excavation${qs}`);
  },
  listShotcrete: (drive_id?: string) => {
    const qs = drive_id ? `?drive_id=${drive_id}` : '';
    return fetchApi<any[]>(`/api/v1/natm/shotcrete${qs}`);
  },
  listRockBolts: (drive_id?: string) => {
    const qs = drive_id ? `?drive_id=${drive_id}` : '';
    return fetchApi<any[]>(`/api/v1/natm/rock-bolts${qs}`);
  },
  listSteelSets: (drive_id?: string) => {
    const qs = drive_id ? `?drive_id=${drive_id}` : '';
    return fetchApi<any[]>(`/api/v1/natm/steel-sets${qs}`);
  },
  listConvergence: (drive_id?: string) => {
    const qs = drive_id ? `?drive_id=${drive_id}` : '';
    return fetchApi<any[]>(`/api/v1/natm/convergence${qs}`);
  },
  listFaceMapping: (drive_id?: string) => {
    const qs = drive_id ? `?drive_id=${drive_id}` : '';
    return fetchApi<any[]>(`/api/v1/natm/face-mapping${qs}`);
  },
  listMTBMDrives: () => fetchApi<any[]>('/api/v1/natm/mtbm-drives'),
  listMTBMThrust: (drive_id?: string) => {
    const qs = drive_id ? `?mtbm_drive_id=${drive_id}` : '';
    return fetchApi<any[]>(`/api/v1/natm/mtbm-thrust${qs ? '?' + qs : ''}`);
  },
  listMTBMLubrication: (drive_id?: string) => {
    const qs = drive_id ? `?mtbm_drive_id=${drive_id}` : '';
    return fetchApi<any[]>(`/api/v1/natm/mtbm-lubrication${qs ? '?' + qs : ''}`);
  },
  listMTBMSurvey: (drive_id?: string) => {
    const qs = drive_id ? `?mtbm_drive_id=${drive_id}` : '';
    return fetchApi<any[]>(`/api/v1/natm/mtbm-survey${qs ? '?' + qs : ''}`);
  },
  listShafts: () => fetchApi<any[]>('/api/v1/natm/shafts'),
  listShaftEquipment: (shaft_id?: string) => {
    const qs = shaft_id ? `?shaft_id=${shaft_id}` : '';
    return fetchApi<any[]>(`/api/v1/natm/shaft-equipment${qs}`);
  },
  listCrossPassages: () => fetchApi<any[]>('/api/v1/natm/cross-passages'),
  listGrouting: (project_id?: string) => {
    const qs = project_id ? `?project_id=${project_id}` : '';
    return fetchApi<any[]>(`/api/v1/natm/grouting${qs}`);
  },
  listSettlement: (project_id?: string) => {
    const qs = project_id ? `?project_id=${project_id}` : '';
    return fetchApi<any[]>(`/api/v1/natm/settlement${qs}`);
  },
  getSummary: () => fetchApi<any>('/api/v1/natm/summary'),
};
