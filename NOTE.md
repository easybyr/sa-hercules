当前project是一个go-frame 微服务构成的后台管理系统，包含后端服务admin和前端服务admin-web，请构建微服务框架，并实现以下功能：
- 组织（Organization）管理：包括组织架构的增删改查，如果是组织（Organization）的子组织（sub-Organization），则子组织只保存直接上级组织的id，如果没有父级组织，父级组织字段（parent_id）则为空。组织和子组织都保存在organization表。
- 用户（User）管理：包括用户的增删改查，用户属于哪个组织，以及在该组织下的权限管理。用户保存在user表。
- 权限（Permission）管理：包括权限的增删改查，权限保存在permission表。
- 角色（Role）管理：包括角色的增删改查。角色保存在role表。
- 用户角色管理：包括用户与角色的关联。用户角色保存在user_role表。
- 角色权限管理：包括角色与权限的关联。角色权限保存在role_permission表。
- 登录功能：用户登录时，验证用户名和密码，并返回token。仅在系统内user表存在的有效用户才可以登陆。
- 注册功能：用户注册时，验证用户名是否已存在，如果不存在，则创建新用户，并且需要关联对应组织，组织以搜索方式进行匹配，不直接展示全部组织树形结构。
- 权限验证：用户访问受保护资源时，验证用户是否有权限访问该资源。
- 用户信息修改：用户可以修改自己的信息，包括用户名、昵称、手机号、邮箱、密码等。
- 用户状态管理：用户可以查看自己的状态，包括正常、禁用等状态。
- 用户权限管理：用户可以查看自己的权限，包括查看、编辑、删除等权限。
- 用户角色管理：用户可以查看自己的角色，包括管理员、普通用户等角色。
- 用户组织管理：用户可以查看自己的组织，包括所属组织、上级组织、下级组织等。
- 用户权限分配：管理员可以给用户分配权限，包括查看、编辑、删除等权限。
- 用户角色分配：管理员可以给用户分配角色，包括管理员、普通用户等角色。
- 用户组织分配：管理员可以给用户分配组织，包括所属组织、上级组织、下级组织等。
- 用户组织架构管理：管理员可以查看整个组织的架构，包括组织树形结构、组织层级关系等。
- 用户权限验证：用户访问受保护资源时，验证用户是否有权限访问该资源。
- 用户角色验证：用户访问受保护资源时，验证用户是否有角色访问该资源。
- 用户组织验证：用户访问受保护资源时，验证用户是否有组织访问该资源。
- 用户uid及组织org_id为数字类型，并且用户uid为全局唯一，组织org_id为组织内唯一。用户uid数字长度限制为10位，组织org_id数字长度限制为8位。
- 权限管理要按照组织进行区分。组织内的用户只会拥有当前组织下的权限，不会拥有其他组织的权限，也不会拥有上级组织或子组织的权限。
- 新建一个超级管理员，用户名：admin，密码：admin，uid：1000000000，org_id：10000000，该用户拥有所有权限，并且可以创建其他用户、组织、权限、角色等。


请使用Go语言和go-frame框架实现以上后端功能，并使用MySQL数据库存储数据。
前端代码使用react框架，前端UI使用浅色/亮色调，注意UI美观性。
开发代码前请参考本project下的AGENTS.local.md文件，了解项目的基本开发规范。
后端服务运行时，mysql使用docker启动，请使用docker-compose.yml文件配置mysql的启动参数。
前端服务，后端服务，mysql都配置在docker-compose.yaml中，使用脚本命令统一启动。


请提供以下内容：
- 项目结构
- 数据库表结构
- 代码实现
- docker-compose.yml文件
- README.md文件，包含项目介绍、使用说明、注意事项等。
项目结构：
```
app
├── admin-web
│   ├── public
│   ├── src
│   ├── package.json
│   ├── README.md
├── admin
│   ├── cmd
│   ├── config
│   ├── internal
│   ├── pkg
│   ├── README.md
├── docker-compose.yml
├── README.md
```

数据库表结构：
```
organization
  - id
  - org_id
  - name
  - parent_id
  - description
  - status
  - extra
  - version
  - created_at
  - updated_at

user
  - id
  - uid
  - username
  - nickname
  - phone_num
  - email
  - password
  - org_id
  - status
  - extra
  - version
  - created_at
  - updated_at


permission
  - id
  - code
  - name
  - description
  - org_id
  - created_at
  - updated_at

role
  - id
  - role_id
  - name
  - description
  - org_id
  - created_at
  - updated_at

user_role
  - id
  - uid
  - role_id
  - created_at
  - updated_at

role_permission
  - id
  - role_id
  - permission_code // 对应permission表中的code
  - created_at
  - updated_at
```


## 补充
- 前端新建组织不用传入组织ID，由后端自动生成；
- 前端新建组织时，需要传入父级组织ID，如果没有父级组织，父级组织ID为空；组织名称带上*号标记，表示必传，如果没有传入组织名称，前端也需要校验拦截；组织描述非必传。
- 前端新建用户不用传入用户ID，由后端自动生成；
- 新建用户时，必须要传入组织ID，如果没有组织，组织ID为空，前端校验拦截；用户名必传，如果没有传入用户名，前端也需要校验拦截；昵称、手机号、邮箱、密码非必传。初始密码生成一段随机字符串，前端展示时，密码用*号代替，前端可点击查看密码。
- 新建角色不用传入角色ID，由后端生成，角色名称为必传，后端校验同一组织下的角色是否重复，如果重复，前端需要提示用户重新输入角色名称。
- 新建权限不用传入权限ID，由后端生成，权限编码和权限名称为必传，后端校验同一组织下的权限是否重复，如果重复，前端需要提示用户重新输入权限名称。权限编码的格式为report.read，后端校验权限编码格式；同时前端新建权限的权限编码的hint提示也需要同步调整。
