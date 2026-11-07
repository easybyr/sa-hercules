"""认证业务逻辑：登录与注册。"""

from __future__ import annotations

from app.consts import identifiers as const
from app.consts.status import Status
from app.pkg import apperror
from app.pkg.password import verify_password
from app.pkg.token import TokenManager
from app.repository.store import Store
from app.schema.auth import LoginRequest, LoginResult, RegisterRequest
from app.schema.user import UserView
from app.service.user import UserService
from app.service.views import user_view


class AuthService:
    """登录与注册。"""

    def __init__(self, store: Store, tokens: TokenManager, users: UserService) -> None:
        self._store = store
        self._tokens = tokens
        self._users = users

    async def login(self, payload: LoginRequest) -> LoginResult:
        """校验用户名密码并签发令牌。"""

        user = await self._store.find_user_by_username(payload.username.strip())
        if user is None or not verify_password(user.password, payload.password):
            raise apperror.unauthorized("INVALID_CREDENTIALS", "用户名或密码错误")
        if int(user.status) != int(Status.NORMAL):
            raise apperror.forbidden("USER_DISABLED", "用户已被禁用")

        organization = await self._store.find_organization_by_org_id(user.org_id)
        if organization is None or int(organization.status) != int(Status.NORMAL):
            raise apperror.forbidden("ORGANIZATION_DISABLED", "所属组织不可用")

        is_super = user.uid == const.SUPER_ADMIN_UID
        token = self._tokens.issue(user.uid, user.org_id, user.username, is_super)
        return LoginResult(
            token=token,
            token_type="Bearer",
            expires_in=self._tokens.expires_in,
            user=user_view(user),
        )

    async def register(self, payload: RegisterRequest) -> UserView:
        """在有效组织内注册新用户。"""

        user = await self._users.create_record(
            username=payload.username,
            password=payload.password,
            nickname=payload.nickname,
            phone_num=payload.phone_num,
            email=payload.email,
            org_id=payload.org_id,
            status=int(Status.NORMAL),
        )
        return user_view(user)
