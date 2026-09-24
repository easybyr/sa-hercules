## 枚举类型定义规范
1. 枚举类型一般定义成int类型的值，数据库对应字段存储为int类型，接口返回对应的string；如果枚举值是币种/符号，则将枚举值定义成string类型；枚举类型或者常量不要使用iota，请统一定义枚举值为固定值；
2. 字段值只有少数限定值，请将该字段定义成枚举类型；
3. 枚举类型定义在consts目录下，不要定义在model目录下；
4. 枚举值要添加简洁的中文注释，比如枚举类型值SendEventStatusUnspecified则添加注释“未指定”；

## SQL建表规范
1. 不要在sql表中校验字段限定值，改为在代码中校验。
2. 在新建的sql表中，对于限定值的字段去掉CONSTRAINT，改为代码中的int类型的枚举类型，对应数据库的字段也为int，但是接口返回为对应string;
3. sql表中的字段名/表名添加comment。
4. 每张表要保证存在ID字段，字段类型为BIGSERIAL/BIGINT，PRIMARY KEY，同时需要另外新增一个唯一的业务ID字段，可以通过固定前缀+去掉连字符的uuid来实现。
5. sql表中去掉CONSTRAINT...CHECK... 的限制，改为在代码中实现对应字段值的校验。
6. 表字段必须包含id、extra、version，created_at，updated_at字段，其中id为BIGSERIAL/BIGINT类型且为auto_increment、primary key，extra为TEXT/JSONB类型，version为int类型。如果是mysql数据库，created_at，updated_at为datetime(3)类型，并且created_at为数据库记录的创建时间，updated_at则跟随记录更新动态更新。
7. 数据库表设计不要使用FOREIGN KEY。
8. 数据库表status相关字段不要使用tinyint，直接使用int类型。
9. 数据库表中的字段名/表名添加comment。

## 开发规则
1. 注意格式化代码，struct多个字段不要写在同一行，函数之间要有空行。
2. 一行最多120个字符。
3. 金额，汇率不要使用math/big，统一使用decimal.Decimal类型，decimal对应包路径：github.com/shopspring/decimal。
4. g.Meta 的comment中先写path，再写method，最后包含tag、summary、description等其他属性。
5. 服务间的调用优先使用grpc接口，然后才是配置http接口调用。
6. 不要使用InsertIgnore，而是先查询再Insert。数据insert/update 不要使用g.Map，而是使用struct类型。
7. repository相关操作放在repository目录下，不要有service逻辑，不要有业务逻辑，只负责数据库操作。
8. service相关操作放在service目录下。
9. controller相关操作放在controller目录下。
10. controller中不要有业务逻辑，只负责接收请求，调用service，返回结果。
11. service中不要有controller逻辑，只负责业务逻辑，调用repository。
12. config配置文件按服务部署环境进行区分，包含config.yaml、config.dev.yaml、config.test.yaml、config.prod.yaml等，其中config.yaml为默认配置文件，config.dev.yaml为开发环境配置文件，config.test.yaml为测试环境配置文件，config.prod.yaml为生产环境配置文件。默认按照dev环境启动。
13. controller目录下的接口文件命名参考格式：user_v1_create_org_member.go，其中user表示user的接口集合，v1表示版本号，create_org_member表示接口名称，在这个接口文件实现了接口：CreateOrgMember。其他目录下的代码也按照功能拆分到不同文件。
14. API返回统一的接口响应Response，包含code、message、data三个字段，其中code为int类型，message为string类型，data为any类型，其中data为业务数据。



