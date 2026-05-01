const API_BASE_URL = "https://api-bank-sampah.smakniscyjr.sch.id";

// ── Token Management ──
export const getAccessToken = (): string | null =>
  localStorage.getItem("access_token");

export const getRefreshToken = (): string | null =>
  localStorage.getItem("refresh_token");

export const setTokens = (access: string, refresh: string) => {
  localStorage.setItem("access_token", access);
  localStorage.setItem("refresh_token", refresh);
};

export const clearTokens = () => {
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
};

export const isAuthenticated = (): boolean => !!getAccessToken();

// ── API Fetch Wrapper ──
interface APIResponse<T = unknown> {
  success: boolean;
  message?: string;
  data?: T;
  meta?: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
}

async function apiFetch<T = unknown>(
  endpoint: string,
  options: RequestInit = {}
): Promise<APIResponse<T>> {
  const token = getAccessToken();

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  // Handle 401 — try refresh token
  if (response.status === 401 && getRefreshToken()) {
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      headers["Authorization"] = `Bearer ${getAccessToken()}`;
      const retryResponse = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers,
      });
      return retryResponse.json();
    } else {
      clearTokens();
      window.location.href = "/signin";
      throw new Error("Session expired");
    }
  }

  return response.json();
}

async function refreshAccessToken(): Promise<boolean> {
  try {
    const refreshToken = getRefreshToken();
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) return false;

    const data = await response.json();
    if (data.success && data.data) {
      setTokens(data.data.access_token, data.data.refresh_token);
      return true;
    }
    return false;
  } catch {
    return false;
  }
}

// ── Auth API ──
export const authAPI = {
  login: (username: string, password: string) =>
    apiFetch<{ access_token: string; refresh_token: string; expires_at: number }>(
      "/auth/login",
      {
        method: "POST",
        body: JSON.stringify({ username, password }),
      }
    ),

  me: () => apiFetch<{ admin_id: string; username: string }>("/auth/me"),

  logout: () => {
    clearTokens();
    window.location.href = "/signin";
  },
};

// ── Dashboard API ──
export const dashboardAPI = {
  getStats: () =>
    apiFetch<{
      total_schools: number;
      total_si: number;
      pending_si: number;
      processing_si: number;
      verified_si: number;
      approved_si: number;
      disbursed_si: number;
      rejected_si: number;
      total_disbursed: number;
      total_nasabah: number;
    }>("/dashboard"),
};

// ── SI API ──
export interface SIDocument {
  id: string;
  school_id: string;
  si_number: string;
  batch_id: string;
  status: "PENDING" | "PROCESSING" | "VERIFIED" | "APPROVED" | "DISBURSED" | "REJECTED";
  total_items: number;
  total_amount: number;
  notes: string;
  verified_at: string | null;
  approved_at: string | null;
  disbursed_at: string | null;
  created_at: string;
  school?: { id: string; name: string; code: string };
  items?: SIItem[];
  verifier?: { id: string; username: string };
  approver?: { id: string; username: string };
}

export interface SIItem {
  id: string;
  nasabah_name: string;
  nasabah_identifier: string;
  nasabah_type: string;
  amount: number;
  status: string;
  failure_reason: string;
  processed_at: string | null;
}

export const siAPI = {
  list: (page = 1, perPage = 15, status = "", schoolId = "") => {
    const params = new URLSearchParams({
      page: String(page),
      per_page: String(perPage),
    });
    if (status) params.set("status", status);
    if (schoolId) params.set("school_id", schoolId);
    return apiFetch<SIDocument[]>(`/si?${params}`);
  },

  getDetail: (id: string) => apiFetch<SIDocument>(`/si/${id}`),

  verify: (id: string) =>
    apiFetch(`/si/${id}/verify`, { method: "PUT" }),

  approve: (id: string) =>
    apiFetch(`/si/${id}/approve`, { method: "PUT" }),

  disburse: (id: string) =>
    apiFetch(`/si/${id}/disburse`, { method: "PUT" }),

  reject: (id: string, reason: string) =>
    apiFetch(`/si/${id}/reject`, {
      method: "PUT",
      body: JSON.stringify({ reason }),
    }),
};

// ── School API ──
export interface School {
  id: string;
  name: string;
  code: string;
  api_key: string;
  api_secret?: string;
  callback_url: string;
  is_active: boolean;
  created_at: string;
}

export const schoolAPI = {
  list: () => apiFetch<School[]>("/schools"),

  getDetail: (id: string) => apiFetch<School>(`/schools/${id}`),

  create: (data: { name: string; code: string; callback_url: string }) =>
    apiFetch<School & { api_secret: string }>("/schools", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: string, data: Partial<{ name: string; callback_url: string; is_active: boolean }>) =>
    apiFetch<School>(`/schools/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  regenerate: (id: string) =>
    apiFetch<{ id: string; api_key: string; api_secret: string }>(
      `/schools/${id}/regenerate`,
      { method: "POST" }
    ),

  delete: (id: string) =>
    apiFetch(`/schools/${id}`, { method: "DELETE" }),
};

// ── Audit API ──
export interface AuditLog {
  id: string;
  admin_id: string;
  si_document_id: string;
  action: string;
  old_data: string;
  new_data: string;
  ip_address: string;
  user_agent: string;
  created_at: string;
  admin?: { id: string; username: string };
}

export const auditAPI = {
  list: (page = 1, perPage = 20) =>
    apiFetch<AuditLog[]>(`/audit-logs?page=${page}&per_page=${perPage}`),
};
