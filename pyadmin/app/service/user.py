"""用户业务逻辑：组织成员管理、个人资料、角色/权限/组织分配。"""

from __future__ import annotations

from typing import List

from app.consts import identifiers as const
from app.consts.status import Status
from app.model import User
from app.pkg import apperror, idgen
from app.pkg.password import hash_password
from app.pkg.token import Claims
from app.repository.store import Store
from app.schema.access import UserDetailView
from app.schema.organization import OrganizationView
from app.schema.pagination import PageRequest, PageResult
from app.schema.user import UserCreateRequest, UserCreateResult, UserUpdateRequest, UserView
from app.service.access import AccessGuard
from app.service.helpers import ensure_personal_role
from app.service.identifier import IdentifierGenerator
from app.service.permission import validate_permission_codes
from app.service.views import (
    organization_view,
    permission_view,
    role_view,
    user_view,
)


class UserService:
    """用户业务逻辑。"""

    def __init__(
        self,
        store: Store,
        guard: AccessGuard,
        identifiers: IdentifierGenerator,
    ) -> None:
        self._store = store
        self._guard = guard
        self._identifiers = identifiers

    async def list(
        self,
        claims: Claims,
        requested_org_id: int,
        keyword: str,
        page: PageRequest,
    ) -> PageResult[UserView]:
        """查询当前可管理范围内的用户。"""

        org_id = self._guard.list_organization_id(claims, requested_org_id)
        items, total = await self._store.list_users(org_id, keyword.strip(), page)
        return PageResult[UserView](
            list=[user_view(item) for item in items],
            total=total,
            page=page.page,
            page_size=page.page_size,
        )

    async def detail(self, claims: Claims, uid: int) -> UserDetailView:
        """查询用户及其组织、角色、权限。"""

        user = await self._store.find_user_by_uid(uid)
        if user is None:
            raise apperror.not_found("用户")
        self._guard.require_org(claims, user.org_id)

        roles = await self._store.user_roles(uid, user.org_id)
        permissions = await self._store.user_permissions(uid, user.org_id)
        direct_codes = await self._store.role_permissions(idgen.personal_role_id(uid))
        organization = await self._store.find_organization_by_org_id(user.org_id)

        return UserDetailView(
            user=user_view(user),
            roles=[role_view(item) for item in roles],
            permissions=[permission_view(item) for item in permissions],
            direct_permission_codes=direct_codes,
            organization=organization_view(organization)
            if organization is not None
            else OrganizationView(
                id=0, org_id=user.org_id, name="", status="unknown", version=0
            ),
        )

    async def create_record(
        self,
        username: str,
        password: str,
        nickname: str,
        phone_num: str,
        email: str,
        org_id: int,
        status: int,
    ) -> User:
        """创建用户记录（注册与管理员创建共用）。"""

        if org_id <= 0 or org_id > const.ORG_ID_MAX_VALUE:
            raise apperror.bad_request(
                "INVALID_ORG_ID", "组织ID必须为不超过8位的正整数"
            )
        name = username.strip()
        if not name:
            raise apperror.bad_request("USERNAME_REQUIRED", "用户名不能为空")
        if await self._store.find_user_by_username(name) is not None:
            raise apperror.conflict("USERNAME_EXISTS", "用户名已存在")

        organization = await self._store.find_organization_by_org_id(org_id)
        if organization is None or int(organization.status) != int(Status.NORMAL):
            raise apperror.bad_request("INVALID_ORGANIZATION", "组织不存在或不可用")

        user = User(
            uid=await self._identifiers.generate_uid(),
            username=name,
            nickname=nickname.strip(),
            phone_num=phone_num,
            email=email,
            password=hash_password(password),
            org_id=org_id,
            status=int(status),
            version=1,
        )
        return await self._store.create_user(user)

    async def create(
        self,
        claims: Claims,
        payload: UserCreateRequest,
    ) -> UserCreateResult:
        """管理员创建用户；密码留空时生成一次性随机初始密码。"""

        if payload.status and not Status.is_valid(payload.status):
            raise apperror.invalid_status()
        self._guard.require_org(claims, payload.org_id)

        status = payload.status or int(Status.NORMAL)
        initial_password = payload.password
        if not initial_password:
            initial_password = idgen.random_password(const.INITIAL_PASSWORD_LENGTH)
        elif len(initial_password) < 6:
            raise apperror.bad_request("INVALID_PASSWORD", "密码至少需要6位")

        user = await self.create_record(
            username=payload.username,
            password=initial_password,
            nickname=payload.nickname,
            phone_num=payload.phone_num,
            email=payload.email,
            org_id=payload.org_id,
            status=int(status),
        )
        return UserCreateResult(user=user_view(user), initial_password=initial_password)

    async def update(
        self,
        claims: Claims,
        uid: int,
        payload: UserUpdateRequest,
    ) -> UserView:
        """管理员更新用户。"""

        user = await self._store.find_user_by_uid(uid)
        if user is None:
            raise apperror.not_found("用户")
        self._guard.require_org(claims, user.org_id)
        if user.uid == const.SUPER_ADMIN_UID and not claims.is_super:
            raise apperror.forbidden("PROTECTED_RESOURCE", "不能修改超级管理员")
        return await self.update_record(user, payload, allow_status=True)

    async def update_record(
        self,
        user: User,
        payload: UserUpdateRequest,
        allow_status: bool,
    ) -> UserView:
        """更新用户记录；allow_status 为假时忽略状态字段（个人资料场景）。"""

        name = payload.username.strip()
        existing = await self._store.find_user_by_username(name)
        if existing is not None and existing.uid != user.uid:
            raise apperror.conflict("USERNAME_EXISTS", "用户名已存在")

        status = int(user.status)
        if allow_status and payload.status:
            if not Status.is_valid(payload.status):
                raise apperror.invalid_status()
            status = int(payload.status)

        hashed = user.password
        if payload.password:
            hashed = hash_password(payload.password)

        item = User(
            id=user.id,
            uid=user.uid,
            username=name,
            nickname=payload.nickname,
            phone_num=payload.phone_num,
            email=payload.email,
            password=hashed,
            org_id=user.org_id,
            status=status,
            version=payload.version,
        )
        affected = await self._store.update_user(item, payload.version)
        if affected == 0:
            raise apperror.version_conflict()
        item.version = payload.version + 1
        return user_view(item)

    async def delete(self, claims: Claims, uid: int) -> None:
        """删除用户。"""

        if uid == const.SUPER_ADMIN_UID:
            raise apperror.bad_request("PROTECTED_RESOURCE", "不能删除超级管理员")
        user = await self._store.find_user_by_uid(uid)
        if user is None:
            raise apperror.not_found("用户")
        self._guard.require_org(claims, user.org_id)
        await self._store.delete_user(uid)

    async def assign_roles(self, claims: Claims, uid: int, role_ids: List[int]) -> None:
        """替换用户的普通角色集合，并清理旧的用户直接权限。"""

        user = await self._store.find_user_by_uid(uid)
        if user is None:
            raise apperror.not_found("用户")
        self._guard.require_org(claims, user.org_id)

        seen = set()
        for role_id in role_ids:
            if role_id <= 0 or role_id in seen:
                raise apperror.bad_request(
                    "INVALID_ROLE", "角色列表包含无效或重复角色"
                )
            seen.add(role_id)
            role = await self._store.find_role(user.org_id, role_id)
            if role is None:
                raise apperror.bad_request(
                    "INVALID_ROLE", "角色不存在或不属于用户组织"
                )
        await self._store.replace_user_roles(uid, role_ids)

    async def assign_permissions(
        self, claims: Claims, uid: int, codes: List[str]
    ) -> None:
        """通过隐藏的个人角色分配用户直接权限。"""

        user = await self._store.find_user_by_uid(uid)
        if user is None:
            raise apperror.not_found("用户")
        self._guard.require_org(claims, user.org_id)
        await validate_permission_codes(self._store, user.org_id, codes)

        role_id = await ensure_personal_role(self._store, user)
        await self._store.replace_role_permissions(role_id, codes)

    async def assign_organization(
        self, claims: Claims, uid: int, org_id: int
    ) -> None:
        """移动用户到其他组织，并清理原组织下的角色关联（仅超级管理员）。"""

        if not claims.is_super:
            raise apperror.forbidden(
                "SUPER_ADMIN_REQUIRED", "跨组织分配仅限超级管理员"
            )
        if uid == const.SUPER_ADMIN_UID:
            raise apperror.bad_request("PROTECTED_RESOURCE", "不能移动超级管理员")

        organization = await self._store.find_organization_by_org_id(org_id)
        if organization is None or int(organization.status) != int(Status.NORMAL):
            raise apperror.bad_request(
                "INVALID_ORGANIZATION", "目标组织不存在或不可用"
            )
        user = await self._store.find_user_by_uid(uid)
        if user is None:
            raise apperror.not_found("用户")
        await self._store.set_user_organization(uid, org_id)
