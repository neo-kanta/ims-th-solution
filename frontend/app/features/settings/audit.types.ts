export interface AuditEvent {
  id: string;
  actor_id?: string | null;
  event_type: string;
  target_type: string;
  target_id: string;
  ip_address: string;
  user_agent: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface AuditListPayload {
  events: AuditEvent[];
  total: number;
  offset: number;
  limit: number;
}

export interface AuditFilters {
  actor_id?: string;
  event_type?: string;
  target_type?: string;
  target_id?: string;
  since?: string;
  until?: string;
  offset?: number;
  limit?: number;
}
