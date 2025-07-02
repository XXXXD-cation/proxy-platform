# Git工作流与开发规范

## 📋 概述

本文档定义了代理与隧道平台项目的Git工作流程、代码提交规范、代码评审标准和持续集成/持续部署(CI/CD)流程。遵循这些规范可以确保代码质量、团队协作效率和项目的可维护性。

**文档版本**: v1.0  
**最后更新**: 2025-07-02  
**维护人员**: ccnochch  

---

## 🌳 Git分支策略

### 分支模型

我们采用基于GitFlow的简化分支模型，适应项目的开发节奏和团队规模。

```
main (生产环境)
├── develop (开发主分支)
│   ├── feature/T001-dev-environment
│   ├── feature/T002-user-management
│   └── feature/T003-proxy-pool
├── hotfix/critical-security-fix
└── release/v1.0.0
```

### 分支说明

#### 🎯 主要分支

| 分支名称 | 用途 | 保护策略 | 合并方式 |
|---------|------|----------|----------|
| `main/master` | 生产环境代码，始终保持可发布状态 | 受保护，仅接受PR合并 | Merge Commit |
| `develop` | 开发主分支，用于功能集成和测试 | 受保护，仅接受PR合并 | Merge Commit |

#### 🔧 辅助分支

| 分支类型 | 命名规则 | 生命周期 | 合并目标 |
|---------|----------|----------|----------|
| `feature/` | `feature/T{任务编号}-{简短描述}` | 功能开发期间 | `develop` |
| `hotfix/` | `hotfix/{严重程度}-{简短描述}` | 紧急修复期间 | `main` + `develop` |
| `release/` | `release/v{版本号}` | 发布准备期间 | `main` |
| `bugfix/` | `bugfix/{bug编号}-{简短描述}` | Bug修复期间 | `develop` |

### 分支操作流程

#### 功能开发流程

```bash
# 1. 从develop创建功能分支
git checkout develop
git pull origin develop
git checkout -b feature/T005-api-gateway

# 2. 开发过程中定期同步develop
git fetch origin
git rebase origin/develop

# 3. 完成开发后推送分支
git push origin feature/T005-api-gateway

# 4. 创建Pull Request到develop分支
# 5. 代码评审通过后合并
# 6. 删除功能分支
git branch -d feature/T005-api-gateway
git push origin --delete feature/T005-api-gateway
```

#### 紧急修复流程

```bash
# 1. 从main创建hotfix分支
git checkout main
git pull origin main
git checkout -b hotfix/critical-security-fix

# 2. 快速修复问题
# 3. 同时合并到main和develop
git checkout main
git merge hotfix/critical-security-fix
git tag v1.0.1
git push origin main --tags

git checkout develop
git merge hotfix/critical-security-fix
git push origin develop
```

---

## 📝 提交规范

### 约定式提交(Conventional Commits)

我们采用约定式提交规范，格式如下：

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### 提交类型(Type)

| 类型 | 说明 | 示例场景 |
|------|------|----------|
| `feat` | 新功能 | 添加用户认证模块 |
| `fix` | 修复bug | 修复代理连接超时问题 |
| `docs` | 文档更新 | 更新API文档 |
| `style` | 代码格式调整(不影响功能) | 代码缩进、空格调整 |
| `refactor` | 代码重构(不改变功能) | 优化数据库连接池实现 |
| `perf` | 性能优化 | 优化代理池查询效率 |
| `test` | 测试相关 | 添加单元测试 |
| `chore` | 构建过程或辅助工具变动 | 更新依赖版本 |
| `ci` | CI/CD配置修改 | 修改GitHub Actions配置 |
| `build` | 构建系统修改 | 修改Docker配置 |

### 作用域(Scope)

推荐使用的作用域：

- `gateway` - API网关相关
- `proxy-pool` - 代理池服务
- `admin-api` - 管理API服务
- `free-crawler` - 免费代理爬虫
- `web` - 前端应用
- `db` - 数据库相关
- `config` - 配置相关
- `deploy` - 部署相关

### 提交示例

#### ✅ 好的提交示例

```bash
# 功能开发
feat(gateway): add HTTPS tunnel support

Implement CONNECT method handling for HTTPS proxy tunnels.
Support both HTTP and HTTPS target servers.
Add connection pooling for better performance.

Closes #123

# Bug修复
fix(proxy-pool): resolve connection timeout issue

Fix intermittent timeout errors when connecting to proxy servers.
Increase default timeout from 5s to 30s and add retry mechanism.

Fixes #456

# 文档更新
docs(api): update authentication endpoints documentation

Add examples for JWT token usage and refresh mechanism.
Update error response format documentation.

# 重构
refactor(admin-api): optimize user query performance

Replace N+1 queries with batch loading.
Reduce average response time from 200ms to 50ms.
```

#### ❌ 避免的提交示例

```bash
# 太简单，没有说明具体做了什么
fix: bug fix

# 混合多个不相关的修改
feat: add user auth and fix proxy timeout and update docs

# 没有遵循格式规范
Fixed the bug in gateway service
```

---

## 🔍 代码评审流程

### Pull Request检查清单

#### 📋 基础检查项

- [ ] **分支命名**: 符合命名规范
- [ ] **提交信息**: 遵循约定式提交格式
- [ ] **代码冲突**: 无合并冲突
- [ ] **CI状态**: 所有自动化检查通过

#### 🧪 代码质量检查

- [ ] **代码规范**: 通过golangci-lint和ESLint检查
- [ ] **单元测试**: 新增代码有对应测试，覆盖率≥80%
- [ ] **集成测试**: 相关集成测试通过
- [ ] **性能测试**: 无性能回归，关键路径测试通过

#### 🛡️ 安全与健壮性

- [ ] **安全检查**: 无已知安全漏洞
- [ ] **错误处理**: 完善的错误处理和日志记录
- [ ] **资源管理**: 正确的资源清理和内存管理
- [ ] **并发安全**: 并发操作的安全性检查

#### 📚 文档与可维护性

- [ ] **代码注释**: 关键逻辑有适当注释
- [ ] **API文档**: 新增API有对应文档
- [ ] **变更日志**: 重要变更记录在CHANGELOG
- [ ] **配置说明**: 新增配置项有说明文档

### 评审流程

1. **自检阶段**: 开发者提交PR前自行检查上述清单
2. **自动检查**: CI/CD系统执行自动化检查
3. **同行评审**: 至少一名同事进行代码评审
4. **技术负责人审核**: 重要功能需技术负责人最终审核
5. **合并**: 通过所有检查后合并到目标分支

### 评审意见分类

| 分类 | 说明 | 处理方式 |
|------|------|----------|
| 🚨 **Blocking** | 必须修复的问题 | 必须解决才能合并 |
| ⚠️ **Major** | 重要建议 | 建议在本PR中解决 |
| 💡 **Minor** | 优化建议 | 可在后续PR中处理 |
| 🤔 **Question** | 疑问或讨论 | 需要回复或解释 |
| 👍 **Praise** | 表扬好的实现 | 激励团队士气 |

---

## 🔄 CI/CD标准

### GitHub Actions工作流

#### 基础检查流程

```yaml
# .github/workflows/ci.yml
name: Continuous Integration

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  lint:
    name: Code Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Go Lint
        run: golangci-lint run
      - name: Frontend Lint
        run: |
          cd web
          npm ci
          npm run lint

  test:
    name: Unit Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run Tests
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  security:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Gosec Security Scanner
        uses: securecodewarrior/github-action-gosec@master
```

### 部署前检查清单

#### 🧪 测试要求

- [ ] **单元测试**: 覆盖率≥80%
- [ ] **集成测试**: 关键业务流程测试通过
- [ ] **性能测试**: 基准测试无回归
- [ ] **压力测试**: 并发测试达到指标

#### 🔒 安全要求

- [ ] **静态分析**: Gosec扫描无高危问题
- [ ] **依赖检查**: 无已知高危漏洞依赖
- [ ] **镜像扫描**: Docker镜像安全扫描通过
- [ ] **敏感信息**: 无硬编码密码或密钥

#### 📊 质量要求

- [ ] **代码覆盖率**: ≥80%
- [ ] **代码复杂度**: 圈复杂度<15
- [ ] **技术债务**: SonarQube质量门禁通过
- [ ] **文档完整性**: 重要变更有对应文档

---

## 🛠️ 开发工具配置

### Git配置建议

```bash
# 全局配置
git config --global user.name "你的姓名"
git config --global user.email "你的邮箱"
git config --global core.autocrlf input
git config --global core.editor "code --wait"

# 项目级配置
git config core.hooksPath .githooks
```

### Git Hooks

项目提供了预提交钩子，确保代码质量：

```bash
# 安装Git hooks
make install-hooks

# 手动运行检查
make pre-commit-check
```

### IDE配置

推荐的VSCode插件：
- `GitLens` - Git历史可视化
- `Conventional Commits` - 提交信息辅助
- `golangci-lint` - Go代码检查
- `ESLint` - JavaScript/TypeScript检查

---

## 📊 流程度量指标

### 开发效率指标

- **功能开发周期**: 从分支创建到合并的平均时间
- **代码评审时间**: PR创建到审核完成的平均时间
- **构建成功率**: CI/CD流水线成功率
- **部署频率**: 平均每周部署次数

### 质量指标

- **Bug修复时间**: 从发现到修复的平均时间
- **代码覆盖率趋势**: 测试覆盖率变化
- **技术债务**: 代码质量工具评分
- **安全漏洞**: 发现和修复的安全问题数量

---

## 🆘 常见问题与解决方案

### Q1: 忘记创建功能分支，直接在develop上开发了怎么办？

```bash
# 创建新的功能分支
git checkout -b feature/T006-fix-development

# 重置develop分支到远程状态
git checkout develop
git reset --hard origin/develop

# 切换回功能分支继续开发
git checkout feature/T006-fix-development
```

### Q2: 提交信息写错了如何修改？

```bash
# 修改最后一次提交信息
git commit --amend -m "feat(gateway): correct commit message"

# 修改历史提交信息（慎用）
git rebase -i HEAD~3
```

### Q3: PR被拒绝后如何处理？

1. 根据评审意见修改代码
2. 追加提交或修改历史提交
3. 推送更新到同一分支
4. 在PR中回复评审意见

### Q4: 合并冲突如何解决？

```bash
# 获取最新的目标分支
git fetch origin
git rebase origin/develop

# 解决冲突后
git add .
git rebase --continue
git push --force-with-lease origin feature/branch-name
```

---

## 📚 参考资料

1. [约定式提交规范](https://www.conventionalcommits.org/)
2. [GitFlow工作流](https://nvie.com/posts/a-successful-git-branching-model/)
3. [GitHub Flow](https://guides.github.com/introduction/flow/)
4. [代码评审最佳实践](https://google.github.io/eng-practices/review/)

---

**文档维护**: 本文档会根据团队实践和项目需求持续更新，如有建议请提Issue或PR。

**联系方式**: ccnochch  
**最后更新**: 2025-07-02 