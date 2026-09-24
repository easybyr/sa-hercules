import { useCallback, useEffect, useMemo, useState } from "react";
import type { FormEvent, ReactNode } from "react";
import {
  Activity,
  BadgeCheck,
  Building2,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  CircleUserRound,
  Copy,
  Eye,
  EyeOff,
  KeyRound,
  LayoutDashboard,
  LogOut,
  Menu,
  Pencil,
  Plus,
  RefreshCw,
  Search,
  ShieldCheck,
  Sparkles,
  Trash2,
  UserCog,
  Users,
  X,
} from "lucide-react";
import { api, session } from "./api";
import type {
  MeData,
  Organization,
  PageResult,
  Permission,
  Role,
  User,
  UserCreateResult,
  UserDetail,
} from "./types";

type Page =
  "overview" | "organizations" | "users" | "roles" | "permissions" | "profile";

const SUPER_ADMIN_ORG_ID = 10000000;
const DEFAULT_PAGE_SIZE = 10;

const pageMeta: Record<Page, { title: string; eyebrow: string }> = {
  overview: { title: "运营概览", eyebrow: "Overview" },
  organizations: { title: "组织架构", eyebrow: "Organization" },
  users: { title: "用户管理", eyebrow: "Identity" },
  roles: { title: "角色管理", eyebrow: "Roles" },
  permissions: { title: "权限管理", eyebrow: "Permissions" },
  profile: { title: "个人中心", eyebrow: "My account" },
};

function messageOf(error: unknown) {
  return error instanceof Error ? error.message : "操作失败，请稍后重试";
}

function useDebouncedValue<T>(value: T, delay = 280) {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const timeout = window.setTimeout(() => setDebounced(value), delay);
    return () => window.clearTimeout(timeout);
  }, [delay, value]);
  return debounced;
}

function pagePath(
  path: string,
  query: string,
  page: number,
  pageSize: number,
  organizationID = 0,
) {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  if (query) params.set("q", query);
  if (organizationID) params.set("org_id", String(organizationID));
  return `${path}?${params.toString()}`;
}

function App() {
  const [me, setMe] = useState<MeData | null>(null);
  const [checking, setChecking] = useState(Boolean(session.get()));
  const [page, setPage] = useState<Page>("overview");
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [toast, setToast] = useState("");

  const loadMe = useCallback(async () => {
    if (!session.get()) {
      setChecking(false);
      return;
    }
    try {
      setMe(await api.get<MeData>("/api/me"));
    } catch {
      session.clear();
      setMe(null);
    } finally {
      setChecking(false);
    }
  }, []);

  useEffect(() => {
    void loadMe();
  }, [loadMe]);

  useEffect(() => {
    if (!toast) return;
    const timeout = window.setTimeout(() => setToast(""), 3200);
    return () => window.clearTimeout(timeout);
  }, [toast]);

  if (checking) return <LoadingScreen />;
  if (!me) return <AuthScreen onAuthenticated={loadMe} />;

  const notify = (text: string) => setToast(text);
  const logout = () => {
    session.clear();
    setMe(null);
    setPage("overview");
  };

  return (
    <div className="app-shell">
      <Sidebar
        me={me}
        page={page}
        open={sidebarOpen}
        onSelect={(next) => {
          setPage(next);
          setSidebarOpen(false);
        }}
      />
      <main className="main-content">
        <header className="topbar">
          <button
            className="icon-button mobile-menu"
            onClick={() => setSidebarOpen(true)}
            aria-label="打开菜单"
          >
            <Menu size={20} />
          </button>
          <div>
            <div className="eyebrow">{pageMeta[page].eyebrow}</div>
            <h1>{pageMeta[page].title}</h1>
          </div>
          <div className="topbar-actions">
            <div className="live-pill">
              <span />
              系统运行正常
            </div>
            <div className="account-menu">
              <button
                className="avatar-button"
                onClick={() => setPage("profile")}
                aria-haspopup="menu"
              >
                <span>
                  {(me.user.nickname || me.user.username)
                    .slice(0, 1)
                    .toUpperCase()}
                </span>
                <div>
                  <strong>{me.user.nickname || me.user.username}</strong>
                  <small>
                    {me.is_super
                      ? "超级管理员"
                      : me.roles[0]?.name || "普通用户"}
                  </small>
                </div>
                <ChevronDown className="account-menu-chevron" size={15} />
              </button>
              <div className="account-dropdown" role="menu">
                <button
                  type="button"
                  role="menuitem"
                  onClick={() => setPage("profile")}
                >
                  <CircleUserRound size={17} />
                  个人中心
                </button>
                <button
                  type="button"
                  role="menuitem"
                  className="account-logout"
                  onClick={logout}
                >
                  <LogOut size={17} />
                  退出登录
                </button>
              </div>
            </div>
          </div>
        </header>

        <div className="page-content">
          {page === "overview" && <Overview me={me} onNavigate={setPage} />}
          {page === "organizations" && (
            <OrganizationsPage me={me} notify={notify} />
          )}
          {page === "users" && <UsersPage me={me} notify={notify} />}
          {page === "roles" && <RolesPage me={me} notify={notify} />}
          {page === "permissions" && (
            <PermissionsPage me={me} notify={notify} />
          )}
          {page === "profile" && (
            <ProfilePage me={me} refreshMe={loadMe} notify={notify} />
          )}
        </div>
      </main>
      {toast && (
        <div className="toast">
          <BadgeCheck size={18} />
          {toast}
        </div>
      )}
    </div>
  );
}

function LoadingScreen() {
  return (
    <div className="loading-screen">
      <div className="brand-mark">
        <Sparkles size={22} />
      </div>
      <strong>Hercules</strong>
      <div className="loading-line" />
    </div>
  );
}

function AuthScreen({
  onAuthenticated,
}: {
  onAuthenticated: () => Promise<void>;
}) {
  const [mode, setMode] = useState<"login" | "register">("login");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [orgQuery, setOrgQuery] = useState("");
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [selectedOrganization, setSelectedOrganization] =
    useState<Organization | null>(null);

  useEffect(() => {
    const keyword = orgQuery.trim();
    if (mode !== "register" || !keyword) {
      setOrganizations([]);
      setSelectedOrganization(null);
      return;
    }
    let active = true;
    const timeout = window.setTimeout(() => {
      void api
        .get<Organization[]>(
          `/api/public/organizations/search?q=${encodeURIComponent(keyword)}`,
        )
        .then((items) => {
          if (!active) return;
          setOrganizations(items);
          setSelectedOrganization(items.length === 1 ? items[0] : null);
        })
        .catch(() => {
          if (!active) return;
          setOrganizations([]);
          setSelectedOrganization(null);
        });
    }, 260);
    return () => {
      active = false;
      window.clearTimeout(timeout);
    };
  }, [mode, orgQuery]);

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setBusy(true);
    setError("");
    const data = new FormData(event.currentTarget);
    const username = String(data.get("username") || "").trim();
    if (!username) {
      setError("请输入用户名");
      setBusy(false);
      return;
    }
    if (mode === "register" && !selectedOrganization) {
      setError("请输入完整组织名称或组织 ID，并选择匹配的组织");
      setBusy(false);
      return;
    }
    try {
      if (mode === "login") {
        const result = await api.post<{ token: string }>("/api/auth/login", {
          username,
          password: data.get("password"),
        });
        session.set(result.token);
        await onAuthenticated();
      } else {
        await api.post("/api/auth/register", {
          username,
          password: data.get("password"),
          nickname: data.get("nickname"),
          phone_num: data.get("phone_num"),
          email: data.get("email"),
          org_id: selectedOrganization?.org_id,
        });
        setMode("login");
        setError("注册成功，请使用新账号登录");
      }
    } catch (submitError) {
      setError(messageOf(submitError));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="auth-page">
      <section className="auth-story">
        <div className="auth-brand">
          <div className="brand-mark">
            <Sparkles size={22} />
          </div>
          <span>Hercules</span>
        </div>
        <div className="story-copy">
          <div className="eyebrow">Enterprise access suite</div>
          <h1>
            让组织与权限，
            <br />
            清晰地协同生长。
          </h1>
          <p>
            面向现代团队的组织架构与身份权限工作台。边界明确、实时可见，也足够优雅。
          </p>
          <div className="story-stats">
            <div>
              <strong>01</strong>
              <span>组织隔离</span>
            </div>
            <div>
              <strong>24h</strong>
              <span>会话保护</span>
            </div>
            <div>
              <strong>RBAC</strong>
              <span>精细授权</span>
            </div>
          </div>
        </div>
        <div className="orb orb-one" />
        <div className="orb orb-two" />
      </section>
      <section className="auth-panel">
        <form className="auth-card" onSubmit={submit}>
          <div className="auth-card-head">
            <div className="eyebrow">
              {mode === "login" ? "Welcome back" : "Create account"}
            </div>
            <h2>{mode === "login" ? "登录管理后台" : "注册新用户"}</h2>
            <p>
              {mode === "login"
                ? "请输入你的账号信息以继续"
                : "搜索并选择所属组织后完成注册"}
            </p>
          </div>
          {error && (
            <div
              className={
                error.startsWith("注册成功") ? "form-success" : "form-error"
              }
            >
              {error}
            </div>
          )}
          {mode === "register" && (
            <Field
              label="昵称"
              name="nickname"
              placeholder="如何称呼你（选填）"
            />
          )}
          <Field
            label="用户名 *"
            name="username"
            placeholder="输入用户名"
            required
            autoComplete="username"
          />
          <Field
            label="密码"
            name="password"
            type="password"
            placeholder="至少6位"
            required
            autoComplete="current-password"
          />
          {mode === "register" && (
            <>
              <Field label="手机号" name="phone_num" placeholder="选填" />
              <Field
                label="邮箱"
                name="email"
                type="email"
                placeholder="name@company.com"
              />
              <div className="field">
                <span>所属组织 *</span>
                <div className="search-input">
                  <Search size={17} />
                  <input
                    value={orgQuery}
                    onChange={(event) => {
                      setOrgQuery(event.target.value);
                      setOrganizations([]);
                      setSelectedOrganization(null);
                    }}
                    placeholder="输入完整组织名称或组织 ID"
                    aria-label="所属组织"
                  />
                </div>
                {organizations.length > 0 && (
                  <div className="auth-organization-results">
                    {organizations.map((organization) => (
                      <button
                        type="button"
                        className={
                          selectedOrganization?.org_id === organization.org_id
                            ? "active"
                            : ""
                        }
                        key={organization.org_id}
                        onClick={() => setSelectedOrganization(organization)}
                      >
                        <Building2 size={17} />
                        <span>
                          <strong>{organization.name}</strong>
                          <small>ORG {organization.org_id}</small>
                        </span>
                        {selectedOrganization?.org_id === organization.org_id && (
                          <BadgeCheck size={17} />
                        )}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            </>
          )}
          <button className="primary-button auth-submit" disabled={busy}>
            {busy ? (
              <RefreshCw className="spin" size={17} />
            ) : mode === "login" ? (
              "进入工作台"
            ) : (
              "完成注册"
            )}
          </button>
          <button
            type="button"
            className="text-button"
            onClick={() => {
              setMode(mode === "login" ? "register" : "login");
              setError("");
              setOrgQuery("");
              setOrganizations([]);
              setSelectedOrganization(null);
            }}
          >
            {mode === "login" ? "没有账号？立即注册" : "已有账号？返回登录"}
          </button>
          {mode === "login" && (
            <div className="demo-tip">初始账号：admin · 密码：admin</div>
          )}
        </form>
      </section>
    </div>
  );
}

function Sidebar({
  me,
  page,
  open,
  onSelect,
}: {
  me: MeData;
  page: Page;
  open: boolean;
  onSelect: (page: Page) => void;
}) {
  const items: Array<{
    id: Page;
    label: string;
    icon: typeof LayoutDashboard;
    permission?: string;
  }> = [
    { id: "overview", label: "运营概览", icon: LayoutDashboard },
    {
      id: "organizations",
      label: "组织架构",
      icon: Building2,
      permission: "organization.read",
    },
    { id: "users", label: "用户管理", icon: Users, permission: "user.read" },
    { id: "roles", label: "角色管理", icon: UserCog, permission: "role.read" },
    {
      id: "permissions",
      label: "权限管理",
      icon: ShieldCheck,
      permission: "permission.read",
    },
    { id: "profile", label: "个人中心", icon: CircleUserRound },
  ];
  const allowed = (permission?: string) =>
    !permission ||
    me.is_super ||
    me.permissions.some((item) => item.code === permission);
  return (
    <>
      {open && (
        <button
          className="sidebar-backdrop"
          onClick={() => onSelect(page)}
          aria-label="关闭菜单"
        />
      )}
      <aside className={`sidebar ${open ? "open" : ""}`}>
        <div className="sidebar-brand">
          <div className="brand-mark">
            <Sparkles size={20} />
          </div>
          <div>
            <strong>Hercules</strong>
            <span>ADMIN CONSOLE</span>
          </div>
          <button
            className="icon-button sidebar-close"
            onClick={() => onSelect(page)}
          >
            <X size={18} />
          </button>
        </div>
        <nav>
          <div className="nav-label">工作空间</div>
          {items
            .filter((item) => allowed(item.permission))
            .map((item) => {
              const Icon = item.icon;
              return (
                <button
                  key={item.id}
                  className={page === item.id ? "active" : ""}
                  onClick={() => onSelect(item.id)}
                >
                  <Icon size={19} />
                  <span>{item.label}</span>
                  {page === item.id && <i />}
                </button>
              );
            })}
        </nav>
        <div className="sidebar-foot">
          <div className="org-chip">
            <Building2 size={17} />
            <div>
              <small>当前组织</small>
              <strong>{me.organization.name}</strong>
            </div>
          </div>
        </div>
      </aside>
    </>
  );
}

function Overview({
  me,
  onNavigate,
}: {
  me: MeData;
  onNavigate: (page: Page) => void;
}) {
  const weekday = new Intl.DateTimeFormat("zh-CN", {
    weekday: "long",
  }).format(new Date());
  const quickEntries = ([
    {
      target: "organizations",
      permission: "organization.read",
      icon: <Building2 />,
      description: "维护层级与组织状态",
    },
    {
      target: "users",
      permission: "user.read",
      icon: <Users />,
      description: "管理成员与归属关系",
    },
    {
      target: "roles",
      permission: "role.read",
      icon: <UserCog />,
      description: "配置角色访问能力",
    },
  ] as Array<{
    target: Page;
    permission: string;
    icon: ReactNode;
    description: string;
  }>).filter(
    (entry) =>
      me.is_super ||
      me.permissions.some((permission) => permission.code === entry.permission),
  );
  return (
    <div className="stack gap-large">
      <section className="welcome-card">
        <div>
          <h2>
            今天{weekday}，{me.user.nickname || me.user.username}
          </h2>
          <button
            className="light-button"
            onClick={() => onNavigate("profile")}
          >
            查看我的访问范围 <ChevronRight size={17} />
          </button>
        </div>
        <div className="welcome-graphic">
          <ShieldCheck size={54} />
          <span />
          <i />
        </div>
      </section>
      <section className="metric-grid">
        <Metric
          icon={<Building2 />}
          label="所属组织"
          value={me.organization.name}
          note={`ORG ${me.organization.org_id}`}
          tone="violet"
        />
        <Metric
          icon={<UserCog />}
          label="有效角色"
          value={String(me.roles.length)}
          note={me.roles[0]?.name || "暂未分配角色"}
          tone="blue"
        />
        <Metric
          icon={<KeyRound />}
          label="权限数量"
          value={String(me.permissions.length)}
          note="仅限当前组织"
          tone="mint"
        />
        <Metric
          icon={<Activity />}
          label="账户状态"
          value="正常"
          note="身份验证已启用"
          tone="orange"
        />
      </section>
      <section className={`split-grid ${quickEntries.length ? "" : "single"}`}>
        <div className="panel access-panel">
          <PanelHead
            title="我的访问能力"
            subtitle="当前组织下实时生效的角色与权限"
          />
          <div className="overview-access-groups">
            {me.role_accesses.map((access) => (
              <div className="overview-access-row" key={access.role.role_id}>
                <div className="permission-cloud" aria-label="角色">
                  <span className="permission-tag featured">
                    {access.role.name}
                  </span>
                </div>
                <div
                  className="permission-cloud"
                  aria-label={`${access.role.name}的权限列表`}
                >
                  {access.permissions.map((permission) => (
                    <span className="permission-tag" key={permission.code}>
                      {permission.name}
                    </span>
                  ))}
                  {!access.permissions.length && (
                    <span className="access-empty">该角色暂无权限</span>
                  )}
                </div>
              </div>
            ))}
            {!me.role_accesses.length && (
              <span className="access-empty">暂无角色及权限</span>
            )}
          </div>
        </div>
        {quickEntries.length > 0 && (
          <div className="panel quick-panel">
            <PanelHead title="快捷入口" subtitle="常用管理操作" />
            {quickEntries.map((entry) => (
              <button
                key={entry.target}
                onClick={() => onNavigate(entry.target)}
              >
                <span>
                  {entry.icon}
                </span>
                <div>
                  <strong>{pageMeta[entry.target].title}</strong>
                  <small>{entry.description}</small>
                </div>
                <ChevronRight size={17} />
              </button>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

function Metric({
  icon,
  label,
  value,
  note,
  tone,
}: {
  icon: ReactNode;
  label: string;
  value: string;
  note: string;
  tone: string;
}) {
  return (
    <div className="metric-card">
      <div className={`metric-icon ${tone}`}>{icon}</div>
      <div>
        <span>{label}</span>
        <strong title={value}>{value}</strong>
        <small>{note}</small>
      </div>
    </div>
  );
}

function OrganizationsPage({
  me,
  notify,
}: {
  me: MeData;
  notify: (text: string) => void;
}) {
  const [tree, setTree] = useState<Organization[]>([]);
  const [allOrganizations, setAllOrganizations] = useState<Organization[]>([]);
  const [query, setQuery] = useState("");
  const searchQuery = useDebouncedValue(query);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [total, setTotal] = useState(0);
  const [modal, setModal] = useState<"create" | "edit" | null>(null);
  const [selected, setSelected] = useState<Organization | null>(null);
  const [error, setError] = useState("");
  const load = useCallback(async () => {
    try {
      const [data, orgTree] = await Promise.all([
        api.get<PageResult<Organization>>(
          pagePath("/api/organizations", searchQuery, page, pageSize),
        ),
        api.get<Organization[]>("/api/organizations/tree"),
      ]);
      setTree(data.list);
      setTotal(data.total);
      setAllOrganizations(orgTree);
      const lastPage = Math.max(1, Math.ceil(data.total / pageSize));
      if (page > lastPage) setPage(lastPage);
    } catch (e) {
      setError(messageOf(e));
    }
  }, [page, pageSize, searchQuery]);
  useEffect(() => {
    void load();
  }, [load]);
  const allFlat = useMemo(
    () => flattenOrganizations(allOrganizations),
    [allOrganizations],
  );

  const save = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");
    const data = new FormData(event.currentTarget);
    const parent = data.get("parent_id");
    const name = String(data.get("name") || "").trim();
    if (!name) {
      setError("请输入组织名称");
      return;
    }
    try {
      if (modal === "create") {
        await api.post("/api/organizations", {
          name,
          parent_id: parent ? Number(parent) : null,
          description: data.get("description"),
          status: Number(data.get("status")),
        });
      } else if (selected) {
        await api.put(`/api/organizations/${selected.id}`, {
          name,
          parent_id: parent ? Number(parent) : null,
          description: data.get("description"),
          status: Number(data.get("status")),
          version: selected.version,
        });
      }
      setModal(null);
      setSelected(null);
      await load();
      notify("组织信息已保存");
    } catch (saveError) {
      setError(messageOf(saveError));
    }
  };
  const remove = async (item: Organization) => {
    if (!window.confirm(`确认删除组织「${item.name}」？`)) return;
    try {
      await api.delete(`/api/organizations/${item.id}`);
      await load();
      notify("组织已删除");
    } catch (e) {
      setError(messageOf(e));
    }
  };
  return (
    <div className="stack">
      <PageActions
        value={query}
        searchPlaceholder="搜索组织名称、ID 或描述"
        onSearch={(value) => {
          setQuery(value);
          setPage(1);
        }}
      >
        {me.is_super && (
          <button
            className="primary-button"
            onClick={() => {
              setError("");
              setSelected(null);
              setModal("create");
            }}
          >
            <Plus size={17} />
            新建组织
          </button>
        )}
      </PageActions>
      {error && <div className="form-error">{error}</div>}
      <div className="panel tree-panel">
        <PanelHead
          title="组织层级"
          subtitle={`共 ${total} 个组织 · 按更新时间倒序`}
        />
        <div className="tree-list">
          {tree.map((item) => (
            <OrganizationNode
              key={item.id}
              item={item}
              depth={organizationDepth(item, allFlat)}
              editable={me.is_super}
              onEdit={(org) => {
                setError("");
                setSelected(org);
                setModal("edit");
              }}
              onDelete={remove}
            />
          ))}
          {!tree.length && <Empty text="还没有组织数据" />}
        </div>
        <Pagination
          page={page}
          pageSize={pageSize}
          total={total}
          onPageChange={setPage}
          onPageSizeChange={(value) => {
            setPageSize(value);
            setPage(1);
          }}
        />
      </div>
      {modal && (
        <Modal
          title={modal === "create" ? "新建组织" : "编辑组织"}
          onClose={() => setModal(null)}
        >
          <form className="modal-form" onSubmit={save}>
            <Field
              label="组织名称 *"
              name="name"
              defaultValue={selected?.name}
              required
            />
            <label className="field">
              <span>直接上级</span>
              <select name="parent_id" defaultValue={selected?.parent_id ?? ""}>
                <option value="">无上级组织</option>
                {allFlat
                  .filter((item) => item.id !== selected?.id)
                  .map((item) => (
                    <option value={item.id} key={item.id}>
                      {item.name} · {item.org_id}
                    </option>
                  ))}
              </select>
            </label>
            <Field
              label="描述"
              name="description"
              defaultValue={selected?.description}
            />
            {error && <div className="form-error">{error}</div>}
            <StatusSelect
              defaultValue={selected?.status === "disabled" ? 2 : 1}
            />
            <ModalActions onCancel={() => setModal(null)} />
          </form>
        </Modal>
      )}
    </div>
  );
}

function OrganizationNode({
  item,
  depth,
  editable,
  onEdit,
  onDelete,
}: {
  item: Organization;
  depth: number;
  editable: boolean;
  onEdit: (item: Organization) => void;
  onDelete: (item: Organization) => void;
}) {
  return (
    <div className="tree-node-wrap">
      <div className="tree-node" style={{ marginLeft: depth * 28 }}>
        <div className="tree-rail" />
        <div className="org-symbol">
          <Building2 size={18} />
        </div>
        <div className="tree-main">
          <strong>{item.name}</strong>
          <span>
            ORG {item.org_id} · {item.description || "暂无描述"}
          </span>
        </div>
        <StatusBadge status={item.status} />
        {editable && item.org_id !== SUPER_ADMIN_ORG_ID && (
          <div className="row-actions">
            <button onClick={() => onEdit(item)}>
              <Pencil size={16} />
            </button>
            <button className="danger" onClick={() => onDelete(item)}>
              <Trash2 size={16} />
            </button>
          </div>
        )}
      </div>
      {item.children?.map((child) => (
        <OrganizationNode
          key={child.id}
          item={child}
          depth={depth + 1}
          editable={editable}
          onEdit={onEdit}
          onDelete={onDelete}
        />
      ))}
    </div>
  );
}

function UsersPage({
  me,
  notify,
}: {
  me: MeData;
  notify: (text: string) => void;
}) {
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [query, setQuery] = useState("");
  const [organizationID, setOrganizationID] = useState(0);
  const searchQuery = useDebouncedValue(query);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [total, setTotal] = useState(0);
  const [modal, setModal] = useState<"create" | "edit" | "access" | null>(null);
  const [selected, setSelected] = useState<UserDetail | null>(null);
  const [selectedRoleIDs, setSelectedRoleIDs] = useState<number[]>([]);
  const [rolePermissionCodes, setRolePermissionCodes] = useState<
    Record<number, string[]>
  >({});
  const [loadingRoleIDs, setLoadingRoleIDs] = useState<number[]>([]);
  const [error, setError] = useState("");
  const [createdUser, setCreatedUser] = useState<UserCreateResult | null>(null);
  const [showInitialPassword, setShowInitialPassword] = useState(false);
  const load = useCallback(async () => {
    try {
      const [userData, roleData, permissionData, orgTree] = await Promise.all([
        api.get<PageResult<User>>(
          pagePath("/api/users", searchQuery, page, pageSize, organizationID),
        ),
        api.get<PageResult<Role>>("/api/roles?page=1&page_size=1000"),
        api.get<PageResult<Permission>>(
          "/api/permissions?page=1&page_size=1000",
        ),
        api.get<Organization[]>("/api/organizations/tree"),
      ]);
      setUsers(userData.list);
      setTotal(userData.total);
      setRoles(roleData.list);
      setPermissions(permissionData.list);
      setOrganizations(flattenOrganizations(orgTree));
      const lastPage = Math.max(1, Math.ceil(userData.total / pageSize));
      if (page > lastPage) setPage(lastPage);
    } catch (e) {
      setError(messageOf(e));
    }
  }, [organizationID, page, pageSize, searchQuery]);
  useEffect(() => {
    void load();
  }, [load]);
  const openUser = async (user: User, next: "edit" | "access") => {
    try {
      const detail = await api.get<UserDetail>(`/api/users/${user.uid}`);
      setSelected(detail);
      if (next === "access") {
        const roleIDs = detail.roles.map((role) => role.role_id);
        setSelectedRoleIDs(roleIDs);
        setLoadingRoleIDs(roleIDs);
        const roleDetails = await Promise.all(
          detail.roles.map((role) =>
            api.get<{ permission_codes: string[] }>(
              `/api/roles/${role.org_id}/${role.role_id}`,
            ),
          ),
        );
        setRolePermissionCodes(
          Object.fromEntries(
            detail.roles.map((role, index) => [
              role.role_id,
              roleDetails[index].permission_codes,
            ]),
          ),
        );
        setLoadingRoleIDs([]);
      }
      setModal(next);
    } catch (e) {
      setLoadingRoleIDs([]);
      setError(messageOf(e));
    }
  };
  const toggleRole = async (role: Role, checked: boolean) => {
    setSelectedRoleIDs((current) =>
      checked
        ? current.includes(role.role_id)
          ? current
          : [...current, role.role_id]
        : current.filter((roleID) => roleID !== role.role_id),
    );
    if (!checked || rolePermissionCodes[role.role_id] !== undefined) return;
    setLoadingRoleIDs((current) => [...current, role.role_id]);
    try {
      const detail = await api.get<{ permission_codes: string[] }>(
        `/api/roles/${role.org_id}/${role.role_id}`,
      );
      setRolePermissionCodes((current) => ({
        ...current,
        [role.role_id]: detail.permission_codes,
      }));
    } catch (e) {
      setSelectedRoleIDs((current) =>
        current.filter((roleID) => roleID !== role.role_id),
      );
      setError(messageOf(e));
    } finally {
      setLoadingRoleIDs((current) =>
        current.filter((roleID) => roleID !== role.role_id),
      );
    }
  };
  const saveUser = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");
    const data = new FormData(event.currentTarget);
    const username = String(data.get("username") || "").trim();
    const orgID = Number(data.get("org_id"));
    if (!username) {
      setError("请输入用户名");
      return;
    }
    if (modal === "create" && !orgID) {
      setError("请选择所属组织");
      return;
    }
    const body = {
      username,
      nickname: data.get("nickname"),
      phone_num: data.get("phone_num"),
      email: data.get("email"),
      password: data.get("password"),
      org_id: orgID,
      status: Number(data.get("status")),
      version: selected?.user.version,
    };
    try {
      if (modal === "create") {
        const result = await api.post<UserCreateResult>("/api/users", body);
        setCreatedUser(result);
        setShowInitialPassword(false);
      } else if (selected) {
        await api.put(`/api/users/${selected.user.uid}`, body);
      }
      setModal(null);
      setSelected(null);
      await load();
      notify(modal === "create" ? "用户已创建" : "用户信息已保存");
    } catch (e) {
      setError(messageOf(e));
    }
  };
  const saveAccess = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!selected) return;
    const data = new FormData(event.currentTarget);
    try {
      const orgID = Number(data.get("org_id"));
      if (me.is_super && orgID !== selected.user.org_id) {
        await api.put(`/api/users/${selected.user.uid}/organization`, {
          org_id: orgID,
        });
        setModal(null);
        await load();
        notify("所属组织已更新，请重新打开用户以配置新组织的角色和权限");
        return;
      }
      await api.put(`/api/users/${selected.user.uid}/roles`, {
        role_ids: selectedRoleIDs,
      });
      setModal(null);
      await load();
      notify("用户访问范围已更新");
    } catch (e) {
      setError(messageOf(e));
    }
  };
  const remove = async (user: User) => {
    if (!window.confirm(`确认删除用户「${user.username}」？`)) return;
    try {
      await api.delete(`/api/users/${user.uid}`);
      await load();
      notify("用户已删除");
    } catch (e) {
      setError(messageOf(e));
    }
  };
  return (
    <div className="stack">
      <PageActions
        value={query}
        onSearch={(value) => {
          setQuery(value);
          setPage(1);
        }}
        searchPlaceholder="搜索用户名或昵称"
        filter={
          <OrganizationFilter
            organizations={organizations}
            value={organizationID}
            onChange={(value) => {
              setOrganizationID(value);
              setPage(1);
            }}
          />
        }
      >
        <button
          className="primary-button"
          onClick={() => {
            setError("");
            setSelected(null);
            setModal("create");
          }}
        >
          <Plus size={17} />
          新建用户
        </button>
      </PageActions>
      {error && <div className="form-error">{error}</div>}
      <div className="panel table-panel">
        <PanelHead
          title="成员列表"
          subtitle={`共 ${total} 位用户 · 按更新时间倒序`}
        />
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>用户</th>
                <th>UID</th>
                <th>组织</th>
                <th>状态</th>
                <th>更新时间</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.uid}>
                  <td>
                    <div className="person-cell">
                      <span>
                        {(user.nickname || user.username).slice(0, 1)}
                      </span>
                      <div>
                        <strong>{user.nickname || user.username}</strong>
                        <small>
                          @{user.username} · {user.email || "未填写邮箱"}
                        </small>
                      </div>
                    </div>
                  </td>
                  <td className="mono">{user.uid}</td>
                  <td className="mono">
                    {organizations.find((org) => org.org_id === user.org_id)
                      ?.name || user.org_id}
                  </td>
                  <td>
                    <StatusBadge status={user.status} />
                  </td>
                  <td>
                    {user.updated_at
                      ? new Date(user.updated_at).toLocaleDateString("zh-CN")
                      : "—"}
                  </td>
                  <td>
                    <div className="row-actions">
                      <button
                        title="权限配置"
                        onClick={() => void openUser(user, "access")}
                      >
                        <ShieldCheck size={16} />
                      </button>
                      <button onClick={() => void openUser(user, "edit")}>
                        <Pencil size={16} />
                      </button>
                      <button
                        className="danger"
                        onClick={() => void remove(user)}
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {!users.length && <Empty text="没有匹配的用户" />}
        </div>
        <Pagination
          page={page}
          pageSize={pageSize}
          total={total}
          onPageChange={setPage}
          onPageSizeChange={(value) => {
            setPageSize(value);
            setPage(1);
          }}
        />
      </div>
      {(modal === "create" || modal === "edit") && (
        <Modal
          title={modal === "create" ? "新建用户" : "编辑用户"}
          onClose={() => setModal(null)}
        >
          <form className="modal-form" onSubmit={saveUser}>
            <div className="field-grid two">
              <Field
                label="用户名 *"
                name="username"
                defaultValue={selected?.user.username}
                required
              />
              <Field
                label="昵称"
                name="nickname"
                defaultValue={selected?.user.nickname}
              />
            </div>
            <div className="field-grid two">
              <Field
                label="手机号"
                name="phone_num"
                defaultValue={selected?.user.phone_num}
              />
              <Field
                label="邮箱"
                name="email"
                type="email"
                defaultValue={selected?.user.email}
              />
            </div>
            <Field
              label={
                modal === "create"
                  ? "自定义初始密码（选填，留空自动生成）"
                  : "新密码（留空不修改）"
              }
              name="password"
              type="password"
              minLength={6}
            />
            <label className="field">
              <span>所属组织 *</span>
              <select
                name="org_id"
                disabled={modal === "edit"}
                defaultValue={modal === "edit" ? selected?.user.org_id : ""}
                required
              >
                {modal === "create" && (
                  <option value="" disabled>
                    请选择所属组织
                  </option>
                )}
                {organizations
                  .filter((org) => me.is_super || org.org_id === me.user.org_id)
                  .map((org) => (
                    <option key={org.org_id} value={org.org_id}>
                      {org.name}
                    </option>
                  ))}
              </select>
            </label>
            {error && <div className="form-error">{error}</div>}
            <StatusSelect
              defaultValue={selected?.user.status === "disabled" ? 2 : 1}
            />
            <ModalActions onCancel={() => setModal(null)} />
          </form>
        </Modal>
      )}
      {createdUser && (
        <Modal title="用户创建成功" onClose={() => setCreatedUser(null)}>
          <div className="modal-form">
            <div className="credential-note">
              初始密码只在此处展示一次，请立即妥善保存并交给用户。
            </div>
            <div className="credential-row">
              <div>
                <small>用户</small>
                <strong>
                  {createdUser.user.username} · UID {createdUser.user.uid}
                </strong>
              </div>
            </div>
            <div className="credential-row">
              <div>
                <small>初始密码</small>
                <code>
                  {showInitialPassword
                    ? createdUser.initial_password
                    : "*".repeat(createdUser.initial_password.length)}
                </code>
              </div>
              <button
                type="button"
                className="icon-button"
                onClick={() => setShowInitialPassword((current) => !current)}
                aria-label={showInitialPassword ? "隐藏密码" : "查看密码"}
              >
                {showInitialPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
            <div className="modal-actions">
              <button
                type="button"
                className="primary-button"
                onClick={() => setCreatedUser(null)}
              >
                我已保存
              </button>
            </div>
          </div>
        </Modal>
      )}
      {modal === "access" && selected && (
        <Modal
          title={`配置 ${selected.user.nickname || selected.user.username}`}
          wide
          onClose={() => setModal(null)}
        >
          <form className="modal-form" onSubmit={saveAccess}>
            <label className="field">
              <span>所属组织</span>
              <select
                name="org_id"
                disabled={!me.is_super}
                defaultValue={selected.user.org_id}
              >
                {organizations.map((org) => (
                  <option key={org.org_id} value={org.org_id}>
                    {org.name} · {org.org_id}
                  </option>
                ))}
              </select>
            </label>
            <RoleAccessSelector
              roles={roles.filter(
                (role) => role.org_id === selected.user.org_id,
              )}
              permissions={permissions.filter(
                (permission) => permission.org_id === selected.user.org_id,
              )}
              selectedRoleIDs={selectedRoleIDs}
              rolePermissionCodes={rolePermissionCodes}
              loadingRoleIDs={loadingRoleIDs}
              onToggle={toggleRole}
            />
            <ModalActions onCancel={() => setModal(null)} />
          </form>
        </Modal>
      )}
    </div>
  );
}

function RolesPage({
  me,
  notify,
}: {
  me: MeData;
  notify: (text: string) => void;
}) {
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [query, setQuery] = useState("");
  const [organizationID, setOrganizationID] = useState(0);
  const searchQuery = useDebouncedValue(query);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [total, setTotal] = useState(0);
  const [modal, setModal] = useState<"create" | "edit" | "access" | null>(null);
  const [selected, setSelected] = useState<Role | null>(null);
  const [selectedCodes, setSelectedCodes] = useState<string[]>([]);
  const [roleOrgID, setRoleOrgID] = useState(me.user.org_id);
  const [error, setError] = useState("");
  const load = useCallback(async () => {
    try {
      const [a, b, c] = await Promise.all([
        api.get<PageResult<Role>>(
          pagePath("/api/roles", searchQuery, page, pageSize, organizationID),
        ),
        api.get<PageResult<Permission>>(
          "/api/permissions?page=1&page_size=1000",
        ),
        api.get<Organization[]>("/api/organizations/tree"),
      ]);
      setRoles(a.list);
      setTotal(a.total);
      setPermissions(b.list);
      setOrganizations(flattenOrganizations(c));
      const lastPage = Math.max(1, Math.ceil(a.total / pageSize));
      if (page > lastPage) setPage(lastPage);
    } catch (e) {
      setError(messageOf(e));
    }
  }, [organizationID, page, pageSize, searchQuery]);
  useEffect(() => {
    void load();
  }, [load]);
  const openRoleModal = async (role: Role, next: "edit" | "access") => {
    try {
      const detail = await api.get<{ permission_codes: string[] }>(
        `/api/roles/${role.org_id}/${role.role_id}`,
      );
      setSelected(role);
      setSelectedCodes(detail.permission_codes);
      setRoleOrgID(role.org_id);
      setModal(next);
    } catch (e) {
      setError(messageOf(e));
    }
  };
  const save = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");
    const data = new FormData(event.currentTarget);
    const name = String(data.get("name") || "").trim();
    const permissionCodes = data.getAll("permissions").map(String);
    if (modal !== "access" && !name) {
      setError("请输入角色名称");
      return;
    }
    try {
      if (modal === "create")
        await api.post("/api/roles", {
          name,
          description: data.get("description"),
          org_id: Number(data.get("org_id")),
          permission_codes: permissionCodes,
        });
      else if (modal === "edit" && selected)
        await api.put(`/api/roles/${selected.org_id}/${selected.role_id}`, {
          name,
          description: data.get("description"),
          version: selected.version,
          permission_codes: permissionCodes,
        });
      else if (modal === "access" && selected)
        await api.put(
          `/api/roles/${selected.org_id}/${selected.role_id}/permissions`,
          { permission_codes: permissionCodes },
        );
      setModal(null);
      await load();
      notify("角色配置已保存");
    } catch (e) {
      setError(messageOf(e));
    }
  };
  const remove = async (role: Role) => {
    if (!window.confirm(`确认删除角色「${role.name}」？`)) return;
    try {
      await api.delete(`/api/roles/${role.org_id}/${role.role_id}`);
      await load();
      notify("角色已删除");
    } catch (e) {
      setError(messageOf(e));
    }
  };
  return (
    <div className="stack">
      <PageActions
        value={query}
        searchPlaceholder="按角色名称筛选"
        onSearch={(value) => {
          setQuery(value);
          setPage(1);
        }}
        filter={
          <OrganizationFilter
            organizations={organizations}
            value={organizationID}
            onChange={(value) => {
              setOrganizationID(value);
              setPage(1);
            }}
          />
        }
      >
        <button
          className="primary-button"
          onClick={() => {
            setError("");
            setSelected(null);
            setSelectedCodes([]);
            setRoleOrgID(me.user.org_id);
            setModal("create");
          }}
        >
          <Plus size={17} />
          新建角色
        </button>
      </PageActions>
      {error && <div className="form-error">{error}</div>}
      <div className="card-grid">
        {roles.map((role) => (
          <article className="role-card" key={`${role.org_id}-${role.role_id}`}>
            <div className="role-card-top">
              <span>
                <UserCog size={20} />
              </span>
              <StatusBadge status="normal" />
            </div>
            <h3>{role.name}</h3>
            <p>{role.description || "暂无角色描述"}</p>
            <div className="role-meta">
              <span>ROLE ID</span>
              <strong>{role.role_id}</strong>
              <span>ORG</span>
              <strong>
                {organizations.find((org) => org.org_id === role.org_id)
                  ?.name || role.org_id}
              </strong>
            </div>
            <div className="role-actions">
              <button onClick={() => void openRoleModal(role, "access")}>
                <ShieldCheck size={16} />
                配置权限
              </button>
              <button onClick={() => void openRoleModal(role, "edit")}>
                <Pencil size={16} />
              </button>
              <button className="danger" onClick={() => void remove(role)}>
                <Trash2 size={16} />
              </button>
            </div>
          </article>
        ))}
        {!roles.length && <Empty text="还没有角色" />}
      </div>
      <Pagination
        page={page}
        pageSize={pageSize}
        total={total}
        onPageChange={setPage}
        onPageSizeChange={(value) => {
          setPageSize(value);
          setPage(1);
        }}
      />
      {modal && (
        <Modal
          title={
            modal === "create"
              ? "新建角色"
              : modal === "edit"
                ? "编辑角色"
                : `配置 ${selected?.name} 的权限`
          }
          wide
          onClose={() => setModal(null)}
        >
          <form className="modal-form" onSubmit={save}>
            {modal === "create" && (
              <label className="field">
                <span>所属组织 *</span>
                <select
                  name="org_id"
                  value={roleOrgID}
                  onChange={(event) => {
                    setRoleOrgID(Number(event.target.value));
                    setSelectedCodes([]);
                  }}
                  required
                >
                  {organizations
                    .filter(
                      (org) => me.is_super || org.org_id === me.user.org_id,
                    )
                    .map((org) => (
                      <option key={org.org_id} value={org.org_id}>
                        {org.name} · {org.org_id}
                      </option>
                    ))}
                </select>
              </label>
            )}
            {modal !== "access" ? (
              <>
                <Field
                  label="角色名称 *"
                  name="name"
                  defaultValue={selected?.name}
                  required
                />
                <Field
                  label="角色描述"
                  name="description"
                  defaultValue={selected?.description}
                />
                <CheckGroup
                  key={`${modal}-${roleOrgID}`}
                  title="角色权限"
                  name="permissions"
                  items={permissions
                    .filter(
                      (permission) =>
                        permission.org_id ===
                        (modal === "create" ? roleOrgID : selected?.org_id),
                    )
                    .map((permission) => ({
                      key: permission.code,
                      label: permission.name,
                      description: permission.code,
                      checked: selectedCodes.includes(permission.code),
                    }))}
                />
              </>
            ) : (
              <CheckGroup
                title="可用权限"
                name="permissions"
                items={permissions
                  .filter(
                    (permission) => permission.org_id === selected?.org_id,
                  )
                  .map((permission) => ({
                    key: permission.code,
                    label: permission.name,
                    description: permission.code,
                    checked: selectedCodes.includes(permission.code),
                  }))}
              />
            )}
            {error && <div className="form-error">{error}</div>}
            <ModalActions onCancel={() => setModal(null)} />
          </form>
        </Modal>
      )}
    </div>
  );
}

function PermissionsPage({
  me,
  notify,
}: {
  me: MeData;
  notify: (text: string) => void;
}) {
  const [items, setItems] = useState<Permission[]>([]);
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [query, setQuery] = useState("");
  const [organizationID, setOrganizationID] = useState(0);
  const searchQuery = useDebouncedValue(query);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [total, setTotal] = useState(0);
  const [modal, setModal] = useState<"create" | "copy" | "edit" | null>(null);
  const [selected, setSelected] = useState<Permission | null>(null);
  const [error, setError] = useState("");
  const load = useCallback(async () => {
    try {
      const [permissionData, orgTree] = await Promise.all([
        api.get<PageResult<Permission>>(
          pagePath(
            "/api/permissions",
            searchQuery,
            page,
            pageSize,
            organizationID,
          ),
        ),
        api.get<Organization[]>("/api/organizations/tree"),
      ]);
      setItems(permissionData.list);
      setTotal(permissionData.total);
      setOrganizations(flattenOrganizations(orgTree));
      const lastPage = Math.max(1, Math.ceil(permissionData.total / pageSize));
      if (page > lastPage) setPage(lastPage);
    } catch (e) {
      setError(messageOf(e));
    }
  }, [organizationID, page, pageSize, searchQuery]);
  useEffect(() => {
    void load();
  }, [load]);
  const save = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");
    const data = new FormData(event.currentTarget);
    const code = String(data.get("code") || "").trim();
    const name = String(data.get("name") || "").trim();
    if (!name) {
      setError("请输入权限名称");
      return;
    }
    if (
      modal !== "edit" &&
      !/^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$/.test(code)
    ) {
      setError("权限编码需使用类似 report.read 的小写点分格式");
      return;
    }
    try {
      if (modal !== "edit")
        await api.post("/api/permissions", {
          code,
          name,
          description: data.get("description"),
          org_id: Number(data.get("org_id")),
        });
      else if (selected)
        await api.put(
          `/api/permissions/${selected.org_id}/${encodeURIComponent(selected.code)}`,
          {
            name,
            description: data.get("description"),
            version: selected.version,
          },
        );
      setModal(null);
      await load();
      notify(modal === "copy" ? "权限已复制到目标组织" : "权限信息已保存");
    } catch (e) {
      setError(messageOf(e));
    }
  };
  const remove = async (item: Permission) => {
    if (!window.confirm(`确认删除权限「${item.name}」？角色中的关联也会移除。`))
      return;
    try {
      await api.delete(
        `/api/permissions/${item.org_id}/${encodeURIComponent(item.code)}`,
      );
      await load();
      notify("权限已删除");
    } catch (e) {
      setError(messageOf(e));
    }
  };
  return (
    <div className="stack">
      <PageActions
        value={query}
        searchPlaceholder="搜索名称或权限编码"
        onSearch={(value) => {
          setQuery(value);
          setPage(1);
        }}
        filter={
          <OrganizationFilter
            organizations={organizations}
            value={organizationID}
            onChange={(value) => {
              setOrganizationID(value);
              setPage(1);
            }}
          />
        }
      >
        <button
          className="primary-button"
          onClick={() => {
            setError("");
            setSelected(null);
            setModal("create");
          }}
        >
          <Plus size={17} />
          新建权限
        </button>
      </PageActions>
      {error && <div className="form-error">{error}</div>}
      <div className="panel table-panel">
        <PanelHead
          title="权限目录"
          subtitle={`共 ${total} 项权限 · 按更新时间倒序`}
        />
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>权限名称</th>
                <th>编码</th>
                <th>组织</th>
                <th>描述</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={`${item.org_id}-${item.code}`}>
                  <td>
                    <div className="permission-name">
                      <span>
                        <KeyRound size={16} />
                      </span>
                      <strong>{item.name}</strong>
                    </div>
                  </td>
                  <td>
                    <code>{item.code}</code>
                  </td>
                  <td className="mono">{item.org_id}</td>
                  <td>{item.description || "—"}</td>
                  <td>
                    <div className="row-actions">
                      <button
                        disabled={!me.is_super || organizations.length < 2}
                        title={
                          me.is_super && organizations.length > 1
                            ? "复制权限到其他组织"
                            : "没有可复制到的其他组织"
                        }
                        aria-label={`复制权限 ${item.name}`}
                        onClick={() => {
                          setError("");
                          setSelected(item);
                          setModal("copy");
                        }}
                      >
                        <Copy size={16} />
                      </button>
                      <button
                        title="编辑权限"
                        aria-label={`编辑权限 ${item.name}`}
                        onClick={() => {
                          setError("");
                          setSelected(item);
                          setModal("edit");
                        }}
                      >
                        <Pencil size={16} />
                      </button>
                      <button
                        className="danger"
                        title="删除权限"
                        aria-label={`删除权限 ${item.name}`}
                        onClick={() => void remove(item)}
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {!items.length && <Empty text="没有匹配的权限" />}
        </div>
        <Pagination
          page={page}
          pageSize={pageSize}
          total={total}
          onPageChange={setPage}
          onPageSizeChange={(value) => {
            setPageSize(value);
            setPage(1);
          }}
        />
      </div>
      {modal && (
        <Modal
          title={
            modal === "create"
              ? "新建权限"
              : modal === "copy"
                ? "复制权限"
                : "编辑权限"
          }
          onClose={() => setModal(null)}
        >
          <form className="modal-form" onSubmit={save}>
            {modal !== "edit" && (
              <>
                <Field
                  label="权限编码 *"
                  name="code"
                  defaultValue={modal === "copy" ? selected?.code : undefined}
                  placeholder="例如 report.read"
                  pattern="[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+"
                  title="请使用类似 report.read 的小写点分格式"
                  readOnly={modal === "copy"}
                  required
                />
                <label className="field">
                  <span>{modal === "copy" ? "目标组织 *" : "所属组织 *"}</span>
                  <select
                    name="org_id"
                    defaultValue={modal === "copy" ? "" : me.user.org_id}
                    required
                  >
                    {modal === "copy" && (
                      <option value="" disabled>
                        请选择目标组织
                      </option>
                    )}
                    {organizations
                      .filter(
                        (org) =>
                          (me.is_super || org.org_id === me.user.org_id) &&
                          (modal !== "copy" || org.org_id !== selected?.org_id),
                      )
                      .map((org) => (
                        <option key={org.org_id} value={org.org_id}>
                          {org.name} · {org.org_id}
                        </option>
                      ))}
                  </select>
                </label>
              </>
            )}
            <Field
              label="权限名称 *"
              name="name"
              defaultValue={selected?.name}
              required
            />
            <Field
              label="权限描述"
              name="description"
              defaultValue={selected?.description}
            />
            {error && <div className="form-error">{error}</div>}
            <ModalActions onCancel={() => setModal(null)} />
          </form>
        </Modal>
      )}
    </div>
  );
}

function ProfilePage({
  me,
  refreshMe,
  notify,
}: {
  me: MeData;
  refreshMe: () => Promise<void>;
  notify: (text: string) => void;
}) {
  const [editing, setEditing] = useState(false);
  const [error, setError] = useState("");
  const save = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    try {
      await api.put("/api/me", {
        username: data.get("username"),
        nickname: data.get("nickname"),
        phone_num: data.get("phone_num"),
        email: data.get("email"),
        password: data.get("password"),
        version: me.user.version,
      });
      setEditing(false);
      await refreshMe();
      notify("个人信息已更新");
    } catch (e) {
      setError(messageOf(e));
    }
  };
  return (
    <div className="profile-grid">
      <section className="panel profile-card">
        <div className="profile-cover" />
        <div className="profile-avatar">
          {(me.user.nickname || me.user.username).slice(0, 1)}
        </div>
        <h2>{me.user.nickname || me.user.username}</h2>
        <p>@{me.user.username}</p>
        <StatusBadge status={me.user.status} />
        <div className="profile-facts">
          <div>
            <span>UID</span>
            <strong>{me.user.uid}</strong>
          </div>
          <div>
            <span>组织</span>
            <strong>{me.organization.name}</strong>
          </div>
        </div>
        <button className="secondary-button" onClick={() => setEditing(true)}>
          <Pencil size={16} />
          编辑个人资料
        </button>
      </section>
      <section className="panel profile-detail">
        <PanelHead title="访问档案" subtitle="你的角色与组织内有效权限" />
        <div className="detail-block">
          <span className="detail-label">角色</span>
          <div className="permission-cloud">
            {me.roles.map((role) => (
              <span className="permission-tag featured" key={role.role_id}>
                {role.name}
              </span>
            ))}
            {me.is_super && (
              <span className="permission-tag featured">超级管理员</span>
            )}
          </div>
        </div>
        <div className="detail-block">
          <span className="detail-label">权限</span>
          <div className="permission-cloud">
            {me.permissions.map((permission) => (
              <span className="permission-tag" key={permission.code}>
                {permission.name}
              </span>
            ))}
          </div>
        </div>
      </section>
      {editing && (
        <Modal title="编辑个人资料" onClose={() => setEditing(false)}>
          <form className="modal-form" onSubmit={save}>
            {error && <div className="form-error">{error}</div>}
            <Field
              label="用户名"
              name="username"
              defaultValue={me.user.username}
              required
            />
            <Field
              label="昵称"
              name="nickname"
              defaultValue={me.user.nickname}
            />
            <div className="field-grid two">
              <Field
                label="手机号"
                name="phone_num"
                defaultValue={me.user.phone_num}
              />
              <Field
                label="邮箱"
                name="email"
                type="email"
                defaultValue={me.user.email}
              />
            </div>
            <Field
              label="新密码（留空不修改）"
              name="password"
              type="password"
            />
            <ModalActions onCancel={() => setEditing(false)} />
          </form>
        </Modal>
      )}
    </div>
  );
}

function PageActions({
  value,
  searchPlaceholder,
  onSearch,
  filter,
  children,
}: {
  value?: string;
  searchPlaceholder: string;
  onSearch: (value: string) => void;
  filter?: ReactNode;
  children: ReactNode;
}) {
  return (
    <div className="page-actions">
      <div className="page-action-filters">
        <div className="search-input">
          <Search size={17} />
          <input
            value={value}
            onChange={(event) => onSearch(event.target.value)}
            placeholder={searchPlaceholder}
          />
        </div>
        {filter}
      </div>
      <div>{children}</div>
    </div>
  );
}

function OrganizationFilter({
  organizations,
  value,
  onChange,
}: {
  organizations: Organization[];
  value: number;
  onChange: (value: number) => void;
}) {
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);
  const selected = organizations.find((item) => item.org_id === value);
  const normalizedQuery = query.trim().toLocaleLowerCase();
  const matches = organizations.filter((item) => {
    if (!normalizedQuery) return true;
    return (
      item.name.toLocaleLowerCase().includes(normalizedQuery) ||
      String(item.org_id).includes(normalizedQuery)
    );
  });
  const displayValue = open
    ? query
    : selected
      ? `${selected.name} · ${selected.org_id}`
      : "全部组织";

  const selectOrganization = (orgID: number) => {
    onChange(orgID);
    setQuery("");
    setOpen(false);
  };

  return (
    <div className="organization-filter">
      <div className="search-input organization-filter-input">
        <Building2 size={17} />
        <input
          role="combobox"
          aria-label="按组织名称或组织 ID 筛选"
          aria-expanded={open}
          aria-controls="organization-filter-options"
          value={displayValue}
          onFocus={() => {
            setQuery("");
            setOpen(true);
          }}
          onBlur={() => window.setTimeout(() => setOpen(false), 120)}
          onChange={(event) => {
            setQuery(event.target.value);
            setOpen(true);
          }}
          placeholder="按组织名称或 ID 筛选"
        />
        {value ? (
          <button
            className="organization-filter-clear"
            type="button"
            aria-label="清除组织筛选"
            title="清除组织筛选"
            onMouseDown={(event) => event.preventDefault()}
            onClick={() => selectOrganization(0)}
          >
            <X size={15} />
          </button>
        ) : (
          <ChevronDown size={16} />
        )}
      </div>
      {open && (
        <div
          className="organization-filter-options"
          id="organization-filter-options"
          role="listbox"
        >
          <button
            className={value === 0 ? "active" : ""}
            type="button"
            role="option"
            aria-selected={value === 0}
            onMouseDown={(event) => event.preventDefault()}
            onClick={() => selectOrganization(0)}
          >
            <span>全部组织</span>
            <small>默认展示全部数据</small>
          </button>
          {matches.map((item) => (
            <button
              className={value === item.org_id ? "active" : ""}
              type="button"
              role="option"
              aria-selected={value === item.org_id}
              key={item.org_id}
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => selectOrganization(item.org_id)}
            >
              <span>{item.name}</span>
              <small>{item.org_id}</small>
            </button>
          ))}
          {!matches.length && (
            <div className="organization-filter-empty">未找到匹配组织</div>
          )}
        </div>
      )}
    </div>
  );
}

function Pagination({
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
}: {
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
}) {
  const pageCount = Math.max(1, Math.ceil(total / pageSize));
  if (!total) return null;
  return (
    <div className="pagination">
      <span>共 {total} 条</span>
      <div className="pagination-controls">
        <label>
          <span>每页</span>
          <select
            value={pageSize}
            onChange={(event) => onPageSizeChange(Number(event.target.value))}
          >
            {[10, 20, 50].map((size) => (
              <option value={size} key={size}>
                {size} 条
              </option>
            ))}
          </select>
        </label>
        <button
          type="button"
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
          aria-label="上一页"
        >
          <ChevronLeft size={16} />
        </button>
        <strong>
          {page} / {pageCount}
        </strong>
        <button
          type="button"
          disabled={page >= pageCount}
          onClick={() => onPageChange(page + 1)}
          aria-label="下一页"
        >
          <ChevronRight size={16} />
        </button>
      </div>
    </div>
  );
}

function PanelHead({ title, subtitle }: { title: string; subtitle: string }) {
  return (
    <div className="panel-head">
      <div>
        <h3>{title}</h3>
        <p>{subtitle}</p>
      </div>
    </div>
  );
}

function Field(props: {
  label: string;
  name: string;
  type?: string;
  placeholder?: string;
  required?: boolean;
  defaultValue?: string | number;
  autoComplete?: string;
  max?: string;
  minLength?: number;
  pattern?: string;
  title?: string;
  readOnly?: boolean;
}) {
  const { label, ...inputProps } = props;
  return (
    <label className="field">
      <span>{label}</span>
      <input {...inputProps} />
    </label>
  );
}

function StatusSelect({ defaultValue }: { defaultValue: number }) {
  return (
    <label className="field">
      <span>状态</span>
      <select name="status" defaultValue={defaultValue}>
        <option value={1}>正常</option>
        <option value={2}>禁用</option>
      </select>
    </label>
  );
}

function StatusBadge({ status }: { status: string }) {
  return (
    <span className={`status-badge ${status}`}>
      <i />
      {status === "normal" ? "正常" : status === "disabled" ? "禁用" : status}
    </span>
  );
}

function Modal({
  title,
  onClose,
  children,
  wide = false,
}: {
  title: string;
  onClose: () => void;
  children: ReactNode;
  wide?: boolean;
}) {
  return (
    <div className="modal-backdrop">
      <div className={`modal ${wide ? "wide" : ""}`}>
        <div className="modal-head">
          <div>
            <div className="eyebrow">Hercules admin</div>
            <h2>{title}</h2>
          </div>
          <button className="icon-button" onClick={onClose}>
            <X size={19} />
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

function ModalActions({ onCancel }: { onCancel: () => void }) {
  return (
    <div className="modal-actions">
      <button type="button" className="secondary-button" onClick={onCancel}>
        取消
      </button>
      <button className="primary-button">保存更改</button>
    </div>
  );
}

function RoleAccessSelector({
  roles,
  permissions,
  selectedRoleIDs,
  rolePermissionCodes,
  loadingRoleIDs,
  onToggle,
}: {
  roles: Role[];
  permissions: Permission[];
  selectedRoleIDs: number[];
  rolePermissionCodes: Record<number, string[]>;
  loadingRoleIDs: number[];
  onToggle: (role: Role, checked: boolean) => Promise<void>;
}) {
  const selectedRoles = roles.filter((role) =>
    selectedRoleIDs.includes(role.role_id),
  );
  const permissionByCode = new Map(
    permissions.map((permission) => [permission.code, permission]),
  );
  return (
    <>
      <div className="check-section">
        <div className="detail-label">分配角色</div>
        <div className="check-grid role-check-grid">
          {roles.map((role) => (
            <label className="check-card" key={role.role_id}>
              <input
                type="checkbox"
                name="roles"
                value={role.role_id}
                checked={selectedRoleIDs.includes(role.role_id)}
                onChange={(event) => void onToggle(role, event.target.checked)}
              />
              <span>
                <strong>{role.name}</strong>
                <small>{role.role_id}</small>
              </span>
            </label>
          ))}
          {!roles.length && <Empty text="当前组织暂无可选角色" />}
        </div>
      </div>
      <div className="role-permission-section">
        <div className="detail-label">已选角色对应权限</div>
        {!selectedRoles.length && (
          <div className="role-permission-empty">
            选择角色后，将在这里展示该角色的权限列表
          </div>
        )}
        <div className="role-permission-groups">
          {selectedRoles.map((role) => {
            const codes = rolePermissionCodes[role.role_id];
            const loading = loadingRoleIDs.includes(role.role_id);
            return (
              <section className="role-permission-group" key={role.role_id}>
                <div className="role-permission-head">
                  <strong>{role.name}</strong>
                  <span>ROLE {role.role_id}</span>
                </div>
                {loading || codes === undefined ? (
                  <span className="role-permission-loading">正在加载权限…</span>
                ) : codes.length ? (
                  <div className="permission-cloud">
                    {codes.map((code) => {
                      const permission = permissionByCode.get(code);
                      return (
                        <span className="permission-tag" key={code}>
                          {permission?.name || code}
                          <small>{code}</small>
                        </span>
                      );
                    })}
                  </div>
                ) : (
                  <span className="role-permission-empty">该角色暂无权限</span>
                )}
              </section>
            );
          })}
        </div>
      </div>
    </>
  );
}

function CheckGroup({
  title,
  name,
  items,
}: {
  title: string;
  name: string;
  items: Array<{
    key: string;
    label: string;
    description: string;
    checked: boolean;
  }>;
}) {
  return (
    <div className="check-section">
      <div className="detail-label">{title}</div>
      <div className="check-grid">
        {items.map((item) => (
          <label className="check-card" key={item.key}>
            <input
              type="checkbox"
              name={name}
              value={item.key}
              defaultChecked={item.checked}
            />
            <span>
              <strong>{item.label}</strong>
              <small>{item.description}</small>
            </span>
          </label>
        ))}
        {!items.length && <Empty text="当前组织暂无可选项" />}
      </div>
    </div>
  );
}

function Empty({ text }: { text: string }) {
  return (
    <div className="empty">
      <Sparkles size={20} />
      <span>{text}</span>
    </div>
  );
}

function flattenOrganizations(items: Organization[]): Organization[] {
  return items.flatMap((item) => [
    item,
    ...flattenOrganizations(item.children || []),
  ]);
}

function organizationDepth(item: Organization, items: Organization[]) {
  const byID = new Map(
    items.map((organization) => [organization.id, organization]),
  );
  const visited = new Set<number>();
  let parentID = item.parent_id;
  let depth = 0;
  while (parentID && !visited.has(parentID)) {
    visited.add(parentID);
    const parent = byID.get(parentID);
    if (!parent) break;
    depth += 1;
    parentID = parent.parent_id;
  }
  return depth;
}

export default App;
