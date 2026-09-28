在附件中的project包含前后端服务，前端服务是admin-web，使用vite+react开发；后端服务是admin，使用golang+mysql数据库。此project提供了docker-compose.yaml部署方式，通过使用docker统一打包部署前端服务+后端服务+mysql数据库，scripts/run.sh 是启动所有服务的脚本命令。请帮我将sa-hercules project部署到atoms云上，前端能够调用后端接口并将数据保存在mysql数据库，并提供对外访问能力。


atoms对外访问入口：hercules.pub.atoms.world
