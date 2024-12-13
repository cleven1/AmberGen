# Multi-Agent AmberGen

一个强大而灵活的多智能体协作框架，支持多种类型的智能体、智能体选择、多轮对话等特性。

## 特性

- 多种智能体类型：支持选择型和通用型智能体
- 模型自定义：支持指定不同的语言模型
- 智能选择：自动选择最合适的智能体处理任务
- 多轮对话：支持智能体之间的连续对话
- 执行模式：支持并行和串行执行
- 流式输出：实时响应支持

## 示例

框架提供了多个完整的示例，展示不同使用场景和功能：

### 1. [Graph 依赖示例 展示如何使用依赖图管理智能体之间的执行顺序](examples/graph_example.go)
### 2. [Group 示例 展示如何使用 Group 进行并行或串行执行](examples/group_example.go)
### 3. [选择器示例 展示如何使用选择器自动选择最合适的智能体](examples/selector_example.go)
### 4. [工具调用示例 展示如何使用工具扩展智能体的能力](examples/tool_usage.go)

## 智能体类型

### 1. 选择型智能体
专门用于特定场景的智能选择任务。

```go
// 默认模型的选择型智能体
selectorAgent := agent.NewSelectorAgent(
    "selector",           // 名称
    "selection expert",   // 专业领域
    "专门负责选择合适的智能体执行任务", // 描述
)

// 指定模型的选择型智能体
customModelAgent := agent.NewSelectorModelAgent(
    "custom_selector",    // 名称
    "selection expert",   // 专业领域
    "专门负责选择合适的智能体执行任务", // 描述
    "gpt-4",             // 指定模型
)
```

### 2. 通用型智能体
用于常规任务处理的智能体。

```go
// 默认模型的通用型智能体
generalAgent := agent.NewAgent(
    "general",           // 名称
    "general expert",    // 专业领域
    "通用型智能体，可以处理各种任务", // 描述
)

// 指定模型的通用型智能体
customAgent := agent.NewModelAgent(
    "custom_general",    // 名称
    "general expert",    // 专业领域
    "通用型智能体，可以处理各种任务", // 描述
    "gpt-4",            // 指定模型
)
```

## 快速开始

### 1. 创建智能体组

```go
// 创建回调处理器
callback := NewDefaultOutputCallback()

// 创建Agent组（参数：最大对话轮数，是否并行执行，回调处理器）
group := agent.NewGroup(3, true, callback)

// 创建不同类型的智能体
selectorAgent := agent.NewSelectorAgent(
    "selector",
    "selection",
    "负责选择合适的执行者",
)

generalAgent := agent.NewAgent(
    "general",
    "general tasks",
    "处理通用任务",
)

customModelAgent := agent.NewModelAgent(
    "custom",
    "specialized tasks",
    "使用特定模型处理专业任务",
    "gpt-4",
)

// 添加智能体到组
group.AddAgent(selectorAgent)
group.AddAgent(generalAgent)
group.AddAgent(customModelAgent)
```

### 2. 执行任务

```go
ctx := context.Background()
input := "分析这个问题..."

// 执行任务并获取结果
results, err := group.Execute(ctx, input)
if err != nil {
    log.Fatal("执行失败:", err)
}
```

## 高级使用

### 1. 选择器模式

```go
// 创建选择器智能体
selectorAgent := agent.NewSelectorModelAgent(
    "task_selector",
    "task selection",
    "负责任务分发的专家",
    "gpt-4",
)

// 创建执行智能体
executor1 := agent.NewAgent("executor1", "domain1", "专家1")
executor2 := agent.NewModelAgent("executor2", "domain2", "专家2", "gpt-4")

// 设置组
group := agent.NewGroup(1, false, callback)
group.AddAgent(selectorAgent)
group.AddAgent(executor1)
group.AddAgent(executor2)
```

### 2. 多轮对话

```go
// 创建3轮对话的组
group := agent.NewGroup(3, true, callback)

// 添加多个智能体
group.AddAgent(agent1)
group.AddAgent(agent2)

// 执行多轮对话
results, err := group.Execute(ctx, "让我们讨论这个话题...")
```

### 3. 流式输出设置

```go
// 启用流式输出
agent.SetStreamOutput(true)

// 设置回调处理流式输出
callback := &OutputCallback{
    OnContent: func(agentName, content string) {
        fmt.Printf("[%s]: %s", agentName, content)
    },
}
```

## Graph 依赖执行

框架支持通过依赖图（DependencyGraph）来管理智能体之间的执行依赖关系。

### 1. 创建依赖图

```go
// 创建依赖图（参数：最大对话轮数，回调处理器）
graph := agent.NewDependencyGraph(3, callback)

// 添加智能体
designerAgent := agent.NewAgent(
    "designer",
    "design",
    "负责设计工作",
)

developerAgent := agent.NewAgent(
    "developer",
    "development",
    "负责开发工作",
)

testerAgent := agent.NewAgent(
    "tester",
    "testing",
    "负责测试工作",
)

// 将智能体添加到依赖图
graph.AddAgent(designerAgent)
graph.AddAgent(developerAgent)
graph.AddAgent(testerAgent)
```

### 2. 设置依赖关系

```go
// 设置依赖关系
// developer 依赖 designer
graph.AddDependency("developer", "designer")
// tester 依赖 developer
graph.AddDependency("tester", "developer")

// 依赖链: designer -> developer -> tester
```

### 3. 执行依赖任务

```go
ctx := context.Background()
input := "请设计并实现一个新功能..."

// 执行任务
results, err := graph.Execute(ctx, input)
if err != nil {
    log.Fatal("执行失败:", err)
}
```

### 4. 依赖图特性

- **自动顺序执行**：
  - 基于依赖关系自动决定执行顺序
  - 确保依赖项在被依赖项之前执行
  - 支持多级依赖关系

- **灵活的依赖配置**：
  ```go
  // 可以设置多个依赖
  graph.AddDependency("agent3", "agent1")
  graph.AddDependency("agent3", "agent2")
  // agent3 将在 agent1 和 agent2 都执行完后执行
  ```

- **智能回退**：
  - 如果没有设置依赖关系，将基于专业度自动选择执行顺序
  - 支持部分依赖场景

### 5. 使用场景

1. **开发流程**:
```go
// 设计 -> 开发 -> 测试 -> 部署
graph.AddDependency("developer", "designer")
graph.AddDependency("tester", "developer")
graph.AddDependency("deployer", "tester")
```

2. **文档处理**:
```go
// 写作 -> 审核 -> 发布
graph.AddDependency("reviewer", "writer")
graph.AddDependency("publisher", "reviewer")
```

3. **数据处理**:
```go
// 数据收集 -> 数据清洗 -> 数据分析
graph.AddDependency("cleaner", "collector")
graph.AddDependency("analyzer", "cleaner")
```

### 6. 最佳实践

1. **依赖设计**:
   - 仔细规划依赖关系
   - 避免循环依赖
   - 保持依赖链清晰简单

2. **错误处理**:
```go
if err := graph.AddDependency(dependent, dependency); err != nil {
    switch err {
    case agent.ErrAgentNotFound:
        // 处理智能体不存在的情况
    case agent.ErrCircularDependency:
        // 处理循环依赖的情况
    default:
        // 处理其他错误
    }
}
```

3. **执行控制**:
   - 合理设置最大轮数
   - 利用回调监控执行过程
   - 及时处理错误情况

### 7. 与Group的区别

- **Group**: 适用于并行或简单串行执行场景
- **DependencyGraph**: 适用于有明确依赖关系的复杂执行场景

|特性|Group|DependencyGraph|
|---|---|---|
|执行顺序|并行或简单串行|基于依赖关系|
|适用场景|独立任务|有依赖的任务链|
|配置复杂度|简单|较复杂|
|灵活性|高|中等|
|执行效率|并行较高|取决于依赖链|


## 配置说明

### 1. 智能体配置
- 名称：唯一标识符
- 专业领域：描述专长
- 描述：详细能力说明
- 模型：可选的模型指定

### 2. 组配置
- 最大轮数：多轮对话的轮数
- 执行模式：并行/串行
- 选择器：可选的智能体选择器

### 3. 回调配置
```go
type OutputCallback interface {
    OnStart(agentName string)
    OnContent(agentName string, content string)
    OnComplete(agentName string)
    OnRoundComplete(round int, results map[string]string)
    OnAllComplete(allResults []map[string]string)
}
```

## 最佳实践

1. **智能体选择**:
   - 选择型智能体适用于需要智能分发任务的场景
   - 通用型智能体适用于具体任务执行

2. **模型选择**:
   - 简单任务使用默认模型
   - 复杂任务考虑使用高级模型（如gpt-4）

3. **执行模式**:
   - 并行模式适合多智能体协作
   - 串行模式适合有依赖的任务

4. **多轮对话**:
   - 设置合适的轮数
   - 利用回调监控对话进展

## 错误处理

```go
results, err := group.Execute(ctx, input)
if err != nil {
    switch err.(type) {
    case *agent.ExecutionError:
        // 处理执行错误
    case *agent.SelectionError:
        // 处理选择错误
    default:
        // 处理其他错误
    }
}
```

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

[MIT License](LICENSE)
