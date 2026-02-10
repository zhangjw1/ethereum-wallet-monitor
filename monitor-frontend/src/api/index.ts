// API 基础配置
const BASE_URL = '/api';

interface ApiResponse<T> {
  code: number;
  data: T;
  message?: string;
}

async function request<T>(url: string, params?: Record<string, string>): Promise<T> {
  const query = params ? '?' + new URLSearchParams(params).toString() : '';
  const res = await fetch(`${BASE_URL}${url}${query}`);
  if (!res.ok) {
    throw new Error(`请求失败: ${res.status}`);
  }
  const json: ApiResponse<T> = await res.json();
  if (json.code >= 400) {
    throw new Error(json.message || '请求失败');
  }
  return json.data;
}

// ==================== 交易流水 ====================

export interface TransferRecord {
  id: number;
  monitor_label: string;
  direction: string;
  from_address: string;
  to_address: string;
  amount: string;
  currency: string;
  tx_hash: string;
  block_number: number;
  notified: boolean;
  notify_status: string;
  created_at: string;
}

export function getTransferRecords(params?: {
  limit?: number;
  tx_hash?: string;
  address?: string;
  start?: string;
  end?: string;
}): Promise<TransferRecord[] | TransferRecord> {
  const query: Record<string, string> = {};
  if (params?.limit) query.limit = String(params.limit);
  if (params?.tx_hash) query.tx_hash = params.tx_hash;
  if (params?.address) query.address = params.address;
  if (params?.start) query.start = params.start;
  if (params?.end) query.end = params.end;
  return request('/transfer-records', query);
}

// ==================== 通知记录 ====================

export interface Notification {
  id: number;
  type: string;
  direction: string;
  from_address: string;
  to_address: string;
  amount: string;
  currency: string;
  tx_hash: string;
  block_num: number;
  mev_type: string;
  confidence: number;
  content: string;
  status: string;
  error_msg: string;
  publish_type: string;
  created_at: string;
  update_at: string;
}

export interface NotificationStats {
  total: number;
  success: number;
  failed: number;
  today: number;
  by_type: { Type: string; Count: number }[];
}

export function getNotifications(params?: {
  limit?: number;
  tx_hash?: string;
  type?: string;
  start?: string;
  end?: string;
  stats?: boolean;
}): Promise<Notification[] | Notification | NotificationStats> {
  const query: Record<string, string> = {};
  if (params?.limit) query.limit = String(params.limit);
  if (params?.tx_hash) query.tx_hash = params.tx_hash;
  if (params?.type) query.type = params.type;
  if (params?.start) query.start = params.start;
  if (params?.end) query.end = params.end;
  if (params?.stats) query.stats = '1';
  return request('/notifications', query);
}

// ==================== 代币分析 ====================

export interface TokenAnalysis {
  id: number;
  token_address: string;
  name: string;
  symbol: string;
  decimals: number;
  total_supply: string;
  has_liquidity: boolean;
  liquidity_usd: number;
  initial_market_cap: number;
  pair_address: string;
  is_verified: boolean;
  is_honeypot: boolean;
  honeypot_reason: string;
  buy_tax: number;
  sell_tax: number;
  holder_count: number;
  top10_holding_pct: number;
  owner_address: string;
  is_ownership_renounced: boolean;
  risk_score: number;
  risk_level: string;
  risk_flags: string;
  status: string;
  safety_status: string;
  pair_created_at: string;
  liquidity_added_at: string;
  last_check_at: string;
  website: string;
  twitter: string;
  telegram: string;
  analyzed_at: string;
  created_at: string;
}

export interface TokenDailyStats {
  total: number;
  honeypot_count: number;
  risk_distribution: { RiskLevel: string; Count: number }[];
}

export function getTokens(params?: {
  limit?: number;
  address?: string;
  status?: string;
  risk_level?: string;
  max_risk_score?: number;
  pending_liquidity?: boolean;
  date?: string;
}): Promise<TokenAnalysis[] | TokenAnalysis | TokenDailyStats> {
  const query: Record<string, string> = {};
  if (params?.limit) query.limit = String(params.limit);
  if (params?.address) query.address = params.address;
  if (params?.status) query.status = params.status;
  if (params?.risk_level) query.risk_level = params.risk_level;
  if (params?.max_risk_score !== undefined) query.max_risk_score = String(params.max_risk_score);
  if (params?.pending_liquidity) query.pending_liquidity = '1';
  if (params?.date) query.date = params.date;
  return request('/tokens', query);
}
