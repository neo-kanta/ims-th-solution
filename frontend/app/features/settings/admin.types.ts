export interface AdminUser {
  id: string;
  username: string;
  display_name: string;
  email: string;
  is_active: boolean;
  is_locked: boolean;
  locked_until?: string | null;
  failed_login_attempts: number;
  force_password_change: boolean;
  last_login_at?: string | null;
  password_changed_at?: string | null;
  groups: string[];
  created_at: string;
  updated_at: string;
}

export interface AdminUserListPayload {
  users: AdminUser[];
  total: number;
  offset: number;
  limit: number;
}

export interface AdminUserFilters {
  search?: string;
  is_active?: boolean;
  is_locked?: boolean;
  offset?: number;
  limit?: number;
}

export interface CreateAdminUserInput {
  username: string;
  display_name: string;
  email: string;
  password: string;
}

export interface AdminSession {
  id: string;
  ip_address: string;
  user_agent: string;
  last_activity_at: string;
  created_at: string;
  expires_at: string;
}

export type AdminUserStatusAction = "disable" | "enable" | "lock" | "unlock";
