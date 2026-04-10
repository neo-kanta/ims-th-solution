export interface UserPermissions {
  functions: string[];
  contracts: string[];
}

export interface AuthUserPayload {
  id: string;
  username: string;
  display_name: string;
  email?: string;
  groups?: string[];
}

export interface AuthUser {
  id: string;
  username: string;
  displayName: string;
  groups: string[];
}

export interface AuthStoreResetState {
  token: string | null;
  expiresAt: string | null;
  user: AuthUser | null;
  permissions: UserPermissions;
  error: string | null;
}
