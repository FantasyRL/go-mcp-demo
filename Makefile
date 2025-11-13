# 辅助工具安装列表（Windows 下建议在 PowerShell 中执行）
# go install github.com/cloudwego/hertz/cmd/hz@latest
# go install github.com/cloudwego/kitex/tool/cmd/kitex@latest
# go install github.com/hertz-contrib/swagger-generate/thrift-gen-http-swagger@latest

# 默认输出帮助信息
.DEFAULT_GOAL := help

# 纯 Windows 环境适配：使用 PowerShell 执行命令，处理路径和命令兼容性
# 项目 MODULE 名
MODULE = github.com/FantasyRL/go-mcp-demo
REMOTE_REPOSITORY ?= fantasyrl/go-mcp-demo

# 目录相关（Windows 路径使用反斜杠，通过 PowerShell 处理路径转换）
DIR = $(subst /,\,$(CURDIR))
CMD = $(DIR)\cmd
CONFIG_PATH = $(DIR)\config
IDL_PATH = $(DIR)\idl
OUTPUT_PATH = $(DIR)\output
API_PATH = $(DIR)\cmd\api

# Docker 网络名称
DOCKER_NET := docker_mcp_net
# Docker 镜像前缀和标签
IMAGE_PREFIX ?= hachimi
TAG          ?= $(shell powershell -Command "(git rev-parse --short HEAD 2>$null) -or 'dev'")

# 服务名
SERVICES := host mcp_local mcp_remote
service = $(word 1, $@)

# hertz HTTP 脚手架（Windows 下用 PowerShell 处理文件删除和命令执行）
.PHONY: hertz-gen-api
hertz-gen-api:
	hz update -idl "$(IDL_PATH)\api.thrift"
	powershell -Command "Remove-Item -Path '$(DIR)\swagger' -Recurse -Force -ErrorAction SilentlyContinue"
	thriftgo -g go -p http-swagger "$(IDL_PATH)\api.thrift"
	powershell -Command "Remove-Item -Path '$(DIR)\gen-go' -Recurse -Force -ErrorAction SilentlyContinue"

# 运行服务（Windows 下直接使用 go run）
.PHONY: $(SERVICES)
$(SERVICES):
	go run "$(CMD)\$(service)" -cfg "$(CONFIG_PATH)\config.yaml"

# 处理依赖（Windows 下 go 命令兼容）
.PHONY: vendor
vendor:
	@echo ">> go mod tidy && go mod vendor"
	go mod tidy
	go mod vendor

# 构建 Docker 镜像（Windows 下 Docker 路径兼容）
.PHONY: docker-build-%
docker-build-%: vendor
	@echo ">> Building image for service: $* (tag: $(TAG))"
	docker build ^
	  --build-arg SERVICE=$* ^
	  -f "docker\Dockerfile" ^
	  -t "$(IMAGE_PREFIX)\$*:$(TAG)" ^
	  .

# 拉取并运行 Docker 容器（Windows 专用逻辑）
.PHONY: pull-run-%
pull-run-%:
	@echo ">> Pulling and running docker (Windows): $*"
	docker pull "$(REMOTE_REPOSITORY):$*"
	powershell -NoProfile -ExecutionPolicy Bypass -File "$(DIR)\scripts\docker-run.ps1" -Service "$*" -Image "$(REMOTE_REPOSITORY):$*" -ConfigPath "$(CONFIG_PATH)\config.yaml"

# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  host                 - 运行 host 服务（使用 config.yaml）"
	@echo "  mcp_local            - 运行 mcp_local 服务（使用 config.yaml）"
	@echo "  vendor               - 整理并下载依赖到 vendor 目录"
	@echo "  docker-build-<svc>   - 构建指定服务的 Docker 镜像"
	@echo "  pull-run-<svc>       - 拉取并运行指定服务的 Docker 容器"
	@echo "  stdio                - 构建 mcp_local 并运行 host（使用 stdio 配置）"
	@echo "  push-<svc>           - 推送指定服务的镜像到远程仓库"
	@echo "  env                  - 启动基础环境（consul 等）"

# 本地调试（Windows 下生成 exe 可执行文件）
.PHONY: stdio
stdio:
	go build -o "bin\mcp_local.exe" "./cmd/mcp_local"
	go run "./cmd/host" -cfg "$(CONFIG_PATH)\config.stdio.yaml"

# 推送镜像到远程仓库（Windows 下适配 PowerShell 语法）
.PHONY: push-%
push-%:
	@powershell -Command "$ConfirmService = Read-Host 'Confirm service name to push (type ''$*'' to confirm)'; if ($ConfirmService -ne '$*') { Write-Error 'Confirmation failed'; exit 1 }"
	@powershell -Command "$Services = '$(SERVICES)'.Split(' '); if (-not $Services.Contains('$*')) { Write-Error 'Invalid service: $*'; exit 1 }"
	@if powershell -Command "[Environment]::Is64BitOperatingSystem -and [Environment]::Is64BitProcess"; then \
		echo "Building and pushing $* for amd64..."; \
		docker build --build-arg SERVICE=$* -t "$(REMOTE_REPOSITORY):$*" -f "docker\Dockerfile" .; \
		docker push "$(REMOTE_REPOSITORY):$*"; \
	else \
		echo "Building and pushing $* using buildx for amd64..."; \
		docker buildx build --platform linux/amd64 --build-arg SERVICE=$* -t "$(REMOTE_REPOSITORY):$*" -f "docker\Dockerfile" --push .; \
	fi

# 启动基础环境（Windows 下用 PowerShell 处理目录删除，先检查目录存在再删除）
.PHONY: env
env:
	powershell -Command "if (Test-Path '$(DIR)\docker\data\consul') { Remove-Item -Path '$(DIR)\docker\data\consul' -Recurse -Force -ErrorAction SilentlyContinue }"
	cd "$(DIR)\docker" && docker-compose up -d
	
# CI/CD 专用推送逻辑
.PHONY: push-cd-%
push-cd-%: vendor
	@powershell -Command "$Services = '$(SERVICES)'.Split(' '); if (-not $Services.Contains('$*')) { Write-Error 'Invalid service: $*'; exit 1 }"
	@if powershell -Command "[Environment]::Is64BitOperatingSystem -and [Environment]::Is64BitProcess"; then \
		echo "Building and pushing $* for amd64..."; \
		docker build --build-arg SERVICE=$* -t "$(REMOTE_REPOSITORY):$*" -f "docker\Dockerfile" .; \
		docker push "$(REMOTE_REPOSITORY):$*"; \
	else \
		echo "Building and pushing $* using buildx for amd64..."; \
		docker buildx build --platform linux/amd64 --build-arg SERVICE=$* -t "$(REMOTE_REPOSITORY):$*" -f "docker\Dockerfile" --push .; \
	fi