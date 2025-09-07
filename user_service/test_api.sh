#!/bin/bash

# User Service API 测试脚本
# 用于测试所有 User Service 的 HTTP API 接口

set -e

BASE_URL="http://localhost:8081"
API_BASE="$BASE_URL/api/v1"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印函数
print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_test() {
    echo -e "${YELLOW}测试: $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# 测试函数
test_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    print_test "$description"
    
    if [ -n "$data" ]; then
        response=$(curl -s -X "$method" "$API_BASE$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -X "$method" "$API_BASE$endpoint")
    fi
    
    echo "Response: $response"
    
    # 检查是否是有效的 JSON
    if echo "$response" | jq . >/dev/null 2>&1; then
        code=$(echo "$response" | jq -r '.code // empty')
        message=$(echo "$response" | jq -r '.message // empty')
        
        if [ "$code" = "200" ] || [ "$code" = "201" ]; then
            print_success "$description - 成功"
        elif [ "$code" = "404" ] || [ "$code" = "400" ]; then
            print_info "$description - 预期的错误响应: $message"
        else
            print_error "$description - 响应码: $code, 消息: $message"
        fi
    else
        print_error "$description - 无效的 JSON 响应"
    fi
    
    echo ""
}

# 主测试流程
main() {
    print_header "User Service API 测试"
    
    # 1. 健康检查
    print_test "健康检查"
    health_response=$(curl -s "$BASE_URL/health")
    echo "Health response: $health_response"
    if echo "$health_response" | jq . >/dev/null 2>&1; then
        status=$(echo "$health_response" | jq -r '.status')
        if [ "$status" = "ok" ]; then
            print_success "服务健康检查通过"
        else
            print_error "服务健康检查失败"
            exit 1
        fi
    else
        print_error "健康检查返回无效响应"
        exit 1
    fi
    echo ""
    
    # 2. 用户资料测试
    print_header "用户资料管理测试"
    
    # 获取不存在的用户资料
    test_api "GET" "/users/1/profile" "" "获取用户资料 (用户不存在)"
    
    # 创建用户资料
    profile_data='{
        "nickname": "测试用户",
        "first_name": "张",
        "last_name": "三",
        "bio": "这是一个测试用户的简介",
        "gender": "male",
        "language": "zh-CN",
        "timezone": "Asia/Shanghai"
    }'
    test_api "PUT" "/users/1/profile" "$profile_data" "创建/更新用户资料"
    
    # 再次获取用户资料
    test_api "GET" "/users/1/profile" "" "获取用户资料 (用户存在)"
    
    # 更新用户资料
    update_data='{
        "nickname": "更新的用户名",
        "bio": "更新后的简介"
    }'
    test_api "PUT" "/users/1/profile" "$update_data" "部分更新用户资料"
    
    # 3. 用户搜索测试
    print_header "用户搜索测试"
    
    # 创建更多测试用户
    for i in {2..5}; do
        user_data="{
            \"nickname\": \"用户$i\",
            \"first_name\": \"名字$i\",
            \"last_name\": \"姓氏$i\"
        }"
        test_api "PUT" "/users/$i/profile" "$user_data" "创建测试用户 $i"
    done
    
    # 搜索用户
    test_api "GET" "/users/search?keyword=用户&limit=10" "" "搜索用户 - 关键字'用户'"
    test_api "GET" "/users/search?keyword=测试&limit=5" "" "搜索用户 - 关键字'测试'"
    test_api "GET" "/users/search?keyword=不存在的用户" "" "搜索用户 - 无结果"
    
    # 4. 用户状态测试
    print_header "用户状态管理测试"
    
    status_data='{"status": "online"}'
    test_api "PUT" "/users/1/status" "$status_data" "更新用户状态为在线"
    
    status_data='{"status": "offline"}'
    test_api "PUT" "/users/1/status" "$status_data" "更新用户状态为离线"
    
    # 5. 用户设置测试
    print_header "用户设置管理测试"
    
    # 获取用户设置 (第一次会创建默认设置)
    test_api "GET" "/users/1/settings" "" "获取用户设置 (创建默认设置)"
    
    # 更新用户设置
    settings_data='{
        "allow_friend_requests": false,
        "show_online_status": false,
        "message_notifications": true
    }'
    test_api "PUT" "/users/1/settings" "$settings_data" "更新用户设置"
    
    # 再次获取设置验证更新
    test_api "GET" "/users/1/settings" "" "验证设置更新"
    
    # 6. 好友关系测试
    print_header "好友关系管理测试"
    
    # 发送好友请求
    friend_request_data='{
        "to_id": 2,
        "message": "你好，我想加你为好友"
    }'
    test_api "POST" "/users/1/friends/requests" "$friend_request_data" "用户1向用户2发送好友请求"
    
    # 再次发送相同请求 (应该失败)
    test_api "POST" "/users/1/friends/requests" "$friend_request_data" "重复发送好友请求 (应该失败)"
    
    # 用户3向用户1发送请求
    friend_request_data2='{
        "to_id": 1,
        "message": "Hi, 可以做朋友吗？"
    }'
    test_api "POST" "/users/3/friends/requests" "$friend_request_data2" "用户3向用户1发送好友请求"
    
    # 获取待处理的好友请求
    test_api "GET" "/users/1/friends/requests" "" "获取用户1的待处理好友请求"
    test_api "GET" "/users/2/friends/requests" "" "获取用户2的待处理好友请求"
    
    # 接受好友请求 (需要实际的request_id，这里用1尝试)
    test_api "PUT" "/users/2/friends/requests/1/accept" "" "用户2接受好友请求"
    
    # 获取好友列表
    test_api "GET" "/users/1/friends?page=1&page_size=10" "" "获取用户1的好友列表"
    test_api "GET" "/users/2/friends?page=1&page_size=10" "" "获取用户2的好友列表"
    
    # 获取共同好友
    test_api "GET" "/users/1/friends/mutual/2" "" "获取用户1和用户2的共同好友"
    
    # 删除好友
    test_api "DELETE" "/users/1/friends/2" "" "用户1删除用户2好友关系"
    
    # 7. 错误处理测试
    print_header "错误处理测试"
    
    # 无效的用户ID
    test_api "GET" "/users/abc/profile" "" "无效用户ID测试"
    
    # 无效的JSON数据
    test_api "PUT" "/users/1/profile" '{"invalid": json}' "无效JSON数据测试"
    
    # 缺少必要参数
    test_api "PUT" "/users/1/status" '{}' "缺少必要参数测试"
    
    print_header "测试完成"
    print_success "所有API测试执行完毕！"
    print_info "请检查上述结果，确认各个功能是否正常工作。"
}

# 检查服务是否运行
check_service() {
    if ! curl -s "$BASE_URL/health" >/dev/null 2>&1; then
        print_error "User Service 未运行或无法访问"
        print_info "请确保服务在 $BASE_URL 上运行"
        exit 1
    fi
}

# 运行测试
echo "检查 User Service 是否运行..."
check_service
echo ""

main
