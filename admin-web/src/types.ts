export type Status = "normal" | "disabled" | "unknown";

export interface Organization {
  id: number;
  org_id: number;
  name: string;
  parent_id: number | null;
  description: string;
  status: Status;
  version: number;
  created_at?: string;
  updated_at?: string;
  children?: Organization[];
}

export interface User {
  id: number;
  uid: number;
  username: string;
  nickname: string;
  phone_num: string;
  email: string;
  org_id: number;
  status: Status;
  version: number;
  created_at?: string;
  updated_at?: string;
}

export interface Permission {
  id: number;
  code: string;
  name: string;
  description: string;
  org_id: number;
  version: number;
  created_at?: string;
  updated_at?: string;
}

export interface Role {
  id: number;
  role_id: number;
  name: string;
  description: string;
  org_id: number;
  version: number;
  created_at?: string;
  updated_at?: string;
}

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
  organization: Organization;
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
  organization: Organization;
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
