export interface PersonalAccountUser {
  id: string;
  username: string;
  display_name: string;
  email?: string;
  groups: string[];
}

export interface PersonalAccountPermissions {
  functions: string[];
  contracts: string[];
}

export interface PersonalAccountPayload {
  user: PersonalAccountUser;
  permissions: PersonalAccountPermissions;
}

export interface PersonalAccountSession {
  id: string;
  ip_address: string;
  user_agent: string;
  last_activity_at: string;
  created_at: string;
  expires_at: string;
}

export interface PersonalMfaStatus {
  enrolled: boolean;
  enabled: boolean;
  recovery_codes_left: number;
}

export interface ChangePersonalPasswordInput {
  old_password: string;
  new_password: string;
}

export interface PersonalMfaEnrollResult {
  provisioning_uri: string;
  recovery_codes: string[];
}

export interface PersonalMfaTotpInput {
  totp_code: string;
}
