# Nodeloc OAuth2 登录测试指南

## 测试前准备

1. **配置Nodeloc OAuth应用**
   - 访问 https://conn.nodeloc.cc/apps
   - 创建新的OAuth应用
   - 设置回调URL为: `http://localhost:3000/oauth/nodeloc` (或你的实际域名)
   - 记录Client ID和Client Secret

2. **配置系统设置**
   - 启动应用后，访问系统设置页面
   - 找到"配置 Nodeloc OAuth"部分
   - 填入以下信息：
     - Nodeloc Client ID: `你的Client ID`
     - Nodeloc Client Secret: `你的Client Secret`
     - 重定向 URI: `http://localhost:3000/oauth/nodeloc` (可选，留空自动生成)
     - 授权端点: `https://conn.nodeloc.cc/oauth2/auth` (默认值)
     - 令牌端点: `https://conn.nodeloc.cc/oauth2/token` (默认值)
     - 用户信息端点: `https://conn.nodeloc.cc/oauth2/userinfo` (默认值)
   - 保存设置
   - 启用"允许通过 Nodeloc 账户登录 & 注册"选项

## 测试步骤

### 1. 测试登录界面显示
- 访问登录页面 `/login`
- 确认能看到"使用 Nodeloc 继续"按钮
- 按钮应该显示Nodeloc图标

### 2. 测试OAuth授权流程
- 点击"使用 Nodeloc 继续"按钮
- 应该跳转到Nodeloc授权页面
- URL应该包含正确的参数：
  - `client_id`: 你的Client ID
  - `redirect_uri`: 回调地址
  - `response_type`: code
  - `scope`: openid profile email
  - `state`: 随机状态值

### 3. 测试授权回调
- 在Nodeloc授权页面点击"授权"
- 应该跳转回你的应用
- 如果是新用户，应该自动注册并登录
- 如果是已存在用户，应该直接登录

### 4. 测试注册界面
- 访问注册页面 `/register`
- 确认也能看到"使用 Nodeloc 继续"按钮
- 功能应该与登录页面一致

## 预期结果

### 成功场景
1. **新用户注册**
   - 用户通过Nodeloc授权后自动创建账户
   - 用户名格式: `nodeloc_[username]` 或 `nodeloc_[user_id]`
   - 显示名使用Nodeloc的name字段
   - 邮箱使用Nodeloc的email字段
   - 自动登录并跳转到控制台

2. **已存在用户登录**
   - 已绑定Nodeloc账户的用户直接登录
   - 跳转到控制台页面

### 错误场景处理
1. **配置错误**
   - Client ID/Secret错误时显示相应错误信息
   - 回调地址不匹配时显示错误

2. **用户拒绝授权**
   - 显示"用户取消授权"或相关错误信息

3. **网络错误**
   - 显示网络连接错误信息

## 调试信息

### 后端日志检查点
- OAuth状态生成: `/api/oauth/state`
- 授权回调处理: `/api/oauth/nodeloc`
- 用户信息获取和处理
- 数据库用户创建/更新

### 前端检查点
- 状态信息获取: `/api/status`
- OAuth按钮显示逻辑
- 点击事件处理

### 数据库检查
- 用户表中的`nodeloc_id`字段
- 新用户记录的创建
- 用户状态和角色设置

## 常见问题排查

1. **按钮不显示**
   - 检查系统设置中是否启用了Nodeloc OAuth
   - 检查前端状态获取是否正常

2. **授权跳转失败**
   - 检查Client ID是否正确
   - 检查授权端点URL是否正确

3. **回调处理失败**
   - 检查回调URL配置
   - 检查Client Secret是否正确
   - 检查令牌端点和用户信息端点URL

4. **用户创建失败**
   - 检查数据库连接
   - 检查用户表结构
   - 检查必填字段是否正确设置