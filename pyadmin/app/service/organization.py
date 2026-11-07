"""组织业务逻辑：可见范围、层级校验、树形组装与增删改查。"""

from __future__ import annotations

from collections import defaultdict, deque
from typing import Dict, List, Optional, Sequence

from app.consts import identifiers as const
from app.consts.status import Status
from app.model import Organization
from app.pkg import apperror
from app.pkg.token import Claims
from app.repository.store import Store
from app.schema.organization import (
    OrganizationCreateRequest,
    OrganizationTreeNode,
    OrganizationUpdateRequest,
    OrganizationView,
)
from app.schema.pagination import PageRequest, PageResult
from app.service.access import AccessGuard
from app.service.identifier import IdentifierGenerator
from app.service.views import organization_tree_node, organization_view


class OrganizationService:
    """组织业务逻辑。"""

    def __init__(
        self,
        store: Store,
        guard: AccessGuard,
        identifiers: IdentifierGenerator,
    ) -> None:
        self._store = store
        self._guard = guard
        self._identifiers = identifiers

    async def search(self, keyword: str) -> List[OrganizationView]:
        """注册页精确搜索组织：完整名称或完整组织 ID 才返回。"""

        term = keyword.strip()
        if not term:
            return []
        items = await self._store.search_organizations_exact(term, 20)
        return [organization_view(item) for item in items]

    async def list(
        self,
        claims: Claims,
        keyword: str,
        page: PageRequest,
    ) -> PageResult[OrganizationView]:
        """分页查询当前用户可见的组织。"""

        restricted = not claims.is_super
        allowed_ids: List[int] = []
        if restricted:
            all_items = await self._store.list_organizations()
            allowed_ids = [
                item.id
                for item in visible_organizations(all_items, claims.org_id)
            ]

        items, total = await self._store.list_organizations_page(
            allowed_ids, restricted, keyword.strip(), page
        )
        return PageResult[OrganizationView](
            list=[organization_view(item) for item in items],
            total=total,
            page=page.page,
            page_size=page.page_size,
        )

    async def tree(
        self,
        claims: Claims,
        keyword: str,
    ) -> List[OrganizationTreeNode]:
        """获取组织树；普通用户仅能看到本组织及其上下级组织。"""

        items = await self._store.list_organizations()
        if not claims.is_super:
            items = visible_organizations(items, claims.org_id)

        known_ids = {item.id for item in items}
        children_map: Dict[int, List[Organization]] = defaultdict(list)
        for item in items:
            if item.parent_id is not None and item.parent_id in known_ids:
                children_map[item.parent_id].append(item)

        def build(item: Organization) -> OrganizationTreeNode:
            kids = [build(child) for child in children_map.get(item.id, [])]
            return organization_tree_node(item, kids)

        roots: List[OrganizationTreeNode] = [
            build(item)
            for item in items
            if item.parent_id is None or item.parent_id not in known_ids
        ]
        return filter_organization_tree(roots, keyword.strip())

    async def create(
        self,
        claims: Claims,
        payload: OrganizationCreateRequest,
    ) -> OrganizationView:
        """新建组织，仅超级管理员可调用。"""

        self._require_super_admin(claims)
        name = payload.name.strip()
        if not name:
            raise apperror.bad_request(
                "ORGANIZATION_NAME_REQUIRED", "组织名称不能为空"
            )
        await self._validate_parent(0, payload.parent_id)

        status = payload.status or int(Status.NORMAL)
        if not Status.is_valid(status):
            raise apperror.invalid_status()

        organization = Organization(
            org_id=await self._identifiers.generate_org_id(),
            name=name,
            parent_id=payload.parent_id,
            description=payload.description,
            status=int(status),
            version=1,
        )
        created = await self._store.create_organization(organization)
        return organization_view(created)

    async def update(
        self,
        claims: Claims,
        identifier: int,
        payload: OrganizationUpdateRequest,
    ) -> OrganizationView:
        """更新组织。"""

        self._require_super_admin(claims)
        item = await self._store.find_organization_by_id(identifier)
        if item is None:
            raise apperror.not_found("组织")
        if item.org_id == const.SUPER_ADMIN_ORG_ID:
            raise apperror.bad_request("PROTECTED_RESOURCE", "系统管理组织不能修改")

        name = payload.name.strip()
        if not name:
            raise apperror.bad_request(
                "ORGANIZATION_NAME_REQUIRED", "组织名称不能为空"
            )
        if not Status.is_valid(payload.status):
            raise apperror.invalid_status()
        await self._validate_parent(identifier, payload.parent_id)

        item.name = name
        item.parent_id = payload.parent_id
        item.description = payload.description
        item.status = int(payload.status)

        affected = await self._store.update_organization(item, payload.version)
        if affected == 0:
            raise apperror.version_conflict()
        item.version = payload.version + 1
        return organization_view(item)

    async def delete(self, claims: Claims, identifier: int) -> None:
        """删除空组织：存在子组织、用户、角色或权限时拒绝。"""

        self._require_super_admin(claims)
        item = await self._store.find_organization_by_id(identifier)
        if item is None:
            raise apperror.not_found("组织")
        if item.org_id == const.SUPER_ADMIN_ORG_ID:
            raise apperror.bad_request(
                "PROTECTED_RESOURCE", "不能删除超级管理员组织"
            )

        children = await self._store.count_children(identifier)
        users = await self._store.count_users(item.org_id)
        permissions = await self._store.count_permissions(item.org_id)
        roles = await self._store.count_roles(item.org_id)
        if children or users or permissions or roles:
            raise apperror.conflict(
                "ORGANIZATION_NOT_EMPTY", "组织包含子组织、用户、角色或权限，不能删除"
            )
        await self._store.delete_organization(identifier)

    def _require_super_admin(self, claims: Claims) -> None:
        """组织管理仅限超级管理员。"""

        if not claims.is_super:
            raise apperror.forbidden("SUPER_ADMIN_REQUIRED", "仅超级管理员可管理组织")

    async def _validate_parent(self, self_id: int, parent_id: Optional[int]) -> None:
        """校验上级组织存在且不形成环。"""

        if parent_id is None:
            return
        if parent_id == self_id:
            raise apperror.bad_request("INVALID_PARENT", "组织不能作为自己的上级")

        parent = await self._store.find_organization_by_id(parent_id)
        if parent is None:
            raise apperror.bad_request("INVALID_PARENT", "上级组织不存在")

        visited = {self_id}
        while parent is not None and parent.parent_id is not None:
            if parent.parent_id in visited:
                raise apperror.bad_request(
                    "ORGANIZATION_CYCLE", "组织层级不能形成循环"
                )
            visited.add(parent.id)
            parent = await self._store.find_organization_by_id(parent.parent_id)


def filter_organization_tree(
    items: Sequence[OrganizationTreeNode],
    keyword: str,
) -> List[OrganizationTreeNode]:
    """按关键词过滤组织树：命中自身或其子节点时保留。"""

    if not keyword:
        return list(items)

    term = keyword.lower()
    result: List[OrganizationTreeNode] = []
    for item in items:
        matched_children = filter_organization_tree(item.children or [], keyword)
        matched = (
            term in item.name.lower()
            or term in str(item.org_id)
            or term in (item.description or "").lower()
        )
        if not matched and not matched_children:
            continue
        result.append(item.model_copy(update={"children": matched_children or None}))
    return result


def visible_organizations(
    items: Sequence[Organization],
    org_id: int,
) -> List[Organization]:
    """计算普通用户可见组织：本组织 + 祖先链 + 全部后代。"""

    by_id = {item.id: item for item in items}
    children: Dict[int, List[int]] = defaultdict(list)
    current: Optional[Organization] = None
    for item in items:
        if item.org_id == org_id:
            current = item
        if item.parent_id is not None:
            children[item.parent_id].append(item.id)
    if current is None:
        return []

    visible = {current.id}
    parent_id = current.parent_id
    while parent_id is not None:
        parent = by_id.get(parent_id)
        if parent is None or parent.id in visible:
            break
        visible.add(parent.id)
        parent_id = parent.parent_id

    queue = deque([current.id])
    while queue:
        node_id = queue.popleft()
        for child_id in children.get(node_id, []):
            if child_id not in visible:
                visible.add(child_id)
                queue.append(child_id)

    return [item for item in items if item.id in visible]
