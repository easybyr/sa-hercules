export type Status = "normal" | "disabled" | "unknown";

export interface Tenant {
  id: number;
  tenant_id: number;
  code: string;
  name: string;
  parent_code: string | null;
  description: string;
  status: Status;
  version: number;
  created_at?: string;
  updated_at?: string;
  children?: Tenant[];
}

export interface User {
  id: number;
  uid: number;
  username: string;
  nickname: string;
  phone_num: string;
  email: string;
  tenant_ids: number[];
  status: Status;
  version: number;
  created_at?: string;
  updated_at?: string;
}

export interface Permission {
  id: number;
  code: string;
  node_type: "MENU" | "BUTTON" | "API" | "DATA";
  apis: string[];
  name: string;
  description: string;
  tenant_id: number;
  version: number;
  created_at?: string;
  updated_at?: string;
}

export interface APIPermission {
  id: number;
  method: string;
  path: string;
  service: string;
  permission_code: string;
}

export interface Role {
  id: number;
  role_id: number;
  name: string;
  description: string;
  tenant_id: number;
  is_system: number;
  version: number;
  created_at?: string;
  updated_at?: string;
}

export type PermissionNodeType = Permission["node_type"];

export interface PageResult<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface MeData {
  user: User;
  roles: Role[];
  permissions: Permission[];
  role_accesses: RoleAccess[];
  tenant: Tenant;
  is_super: boolean;
}

export interface RoleAccess {
  role: Role;
  permissions: Permission[];
}

export interface UserDetail {
  user: User;
  roles: Role[];
  permissions: Permission[];
  direct_permission_codes: string[];
  tenant: Tenant;
  active_tenant_id: number;
}

export interface UserCreateResult {
  user: User;
  initial_password: string;
}

export interface ApiEnvelope<T> {
  code: number;
  message: string;
  data: T;
}
