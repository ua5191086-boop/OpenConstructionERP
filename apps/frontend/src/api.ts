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
