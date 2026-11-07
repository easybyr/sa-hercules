"""服务容器：把业务服务装配到一处，供控制器依赖注入使用。"""

from __future__ import annotations

from app.pkg.token import TokenManager
from app.repository.store import Store
from app.service.access import AccessGuard
from app.service.auth import AuthService
from app.service.bootstrap import BootstrapService
from app.service.health import HealthService
from app.service.identifier import IdentifierGenerator
from app.service.organization import OrganizationService
from app.service.permission import PermissionService
from app.service.profile import ProfileService
from app.service.role import RoleService
from app.service.user import UserService


class Services:
    """业务服务集合。"""

    def __init__(self, store: Store, tokens: TokenManager) -> None:
        self.store = store
        self.tokens = tokens
        self.guard = AccessGuard(store)
        self.identifiers = IdentifierGenerator(store)
        self.users = UserService(store, self.guard, self.identifiers)
        self.auth = AuthService(store, tokens, self.users)
        self.profile = ProfileService(store, self.guard, self.users)
        self.organizations = OrganizationService(store, self.guard, self.identifiers)
        self.roles = RoleService(store, self.guard, self.identifiers)
        self.permissions = PermissionService(store, self.guard)
        self.health = HealthService(store)
        self.bootstrap = BootstrapService(store)
