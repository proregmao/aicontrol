import { ref, onUnmounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'

export interface WebSocketMessage {
  type: string
  data: any
  timestamp: string
}

export interface WebSocketOptions {
  url: string
  protocols?: string | string[]
  autoReconnect?: boolean
  reconnectInterval?: number
  maxReconnectAttempts?: number
  heartbeatInterval?: number
  onOpen?: (event: Event) => void
  onMessage?: (message: WebSocketMessage) => void
  onError?: (event: Event) => void
  onClose?: (event: CloseEvent) => void
}

export function useWebSocket(options: WebSocketOptions) {
  const {
    url,
    protocols,
    autoReconnect = true,
    reconnectInterval = 3000,
    maxReconnectAttempts = 5,
    heartbeatInterval = 30000,
    onOpen,
    onMessage,
    onError,
    onClose
  } = options

  // 状态
  const isConnected = ref(false)
  const isConnecting = ref(false)
  const reconnectAttempts = ref(0)
  const lastMessage = ref<WebSocketMessage | null>(null)
  const connectionError = ref<string | null>(null)

  // WebSocket实例
  let ws: WebSocket | null = null
  let heartbeatTimer: NodeJS.Timeout | null = null
  let reconnectTimer: NodeJS.Timeout | null = null

  // 消息订阅
  const messageHandlers = new Map<string, ((data: any) => void)[]>()

  // 连接WebSocket
  const connect = () => {
    if (isConnecting.value || isConnected.value) {
      return
    }

    isConnecting.value = true
    connectionError.value = null

    try {
      ws = new WebSocket(url, protocols)

      ws.onopen = (event) => {
        isConnected.value = true
        isConnecting.value = false
        reconnectAttempts.value = 0
        connectionError.value = null
        
        console.log('WebSocket连接已建立:', url)
        
        // 启动心跳
        startHeartbeat()
        
        onOpen?.(event)
      }

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          lastMessage.value = message
          
          // 处理心跳响应
          if (message.type === 'pong') {
            return
          }
          
          // 触发全局消息处理器
          onMessage?.(message)
          
          // 触发特定类型的消息处理器
          const handlers = messageHandlers.get(message.type)
          if (handlers) {
            handlers.forEach(handler => {
              try {
                handler(message.data)
              } catch (error) {
                console.error('消息处理器执行失败:', error)
              }
            })
          }
        } catch (error) {
          console.error('解析WebSocket消息失败:', error)
        }
      }

      ws.onerror = (event) => {
        console.error('WebSocket连接错误:', event)
        connectionError.value = 'WebSocket连接错误'
        onError?.(event)
      }

      ws.onclose = (event) => {
        isConnected.value = false
        isConnecting.value = false
        
        // 停止心跳
        stopHeartbeat()
        
        console.log('WebSocket连接已关闭:', event.code, event.reason)
        
        onClose?.(event)
        
        // 自动重连
        if (autoReconnect && reconnectAttempts.value < maxReconnectAttempts) {
          scheduleReconnect()
        } else if (reconnectAttempts.value >= maxReconnectAttempts) {
          connectionError.value = '连接失败，已达到最大重试次数'
          ElMessage.error('WebSocket连接失败，请检查网络连接')
        }
      }
    } catch (error) {
      isConnecting.value = false
      connectionError.value = '创建WebSocket连接失败'
      console.error('创建WebSocket连接失败:', error)
    }
  }

  // 断开连接
  const disconnect = () => {
    if (ws) {
      ws.close(1000, '主动断开连接')
      ws = null
    }
    
    stopHeartbeat()
    stopReconnect()
    
    isConnected.value = false
    isConnecting.value = false
  }

  // 发送消息
  const send = (type: string, data: any) => {
    if (!isConnected.value || !ws) {
      console.warn('WebSocket未连接，无法发送消息')
      return false
    }

    try {
      const message: WebSocketMessage = {
        type,
        data,
        timestamp: new Date().toISOString()
      }
      
      ws.send(JSON.stringify(message))
      return true
    } catch (error) {
      console.error('发送WebSocket消息失败:', error)
      return false
    }
  }

  // 订阅特定类型的消息
  const subscribe = (messageType: string, handler: (data: any) => void) => {
    if (!messageHandlers.has(messageType)) {
      messageHandlers.set(messageType, [])
    }
    
    messageHandlers.get(messageType)!.push(handler)
    
    // 返回取消订阅函数
    return () => {
      const handlers = messageHandlers.get(messageType)
      if (handlers) {
        const index = handlers.indexOf(handler)
        if (index > -1) {
          handlers.splice(index, 1)
        }
        
        // 如果没有处理器了，删除该类型
        if (handlers.length === 0) {
          messageHandlers.delete(messageType)
        }
      }
    }
  }

  // 启动心跳
  const startHeartbeat = () => {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer)
    }
    
    heartbeatTimer = setInterval(() => {
      if (isConnected.value) {
        send('ping', { timestamp: Date.now() })
      }
    }, heartbeatInterval)
  }

  // 停止心跳
  const stopHeartbeat = () => {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer)
      heartbeatTimer = null
    }
  }

  // 计划重连
  const scheduleReconnect = () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
    }
    
    reconnectAttempts.value++
    
    console.log(`WebSocket重连中... (${reconnectAttempts.value}/${maxReconnectAttempts})`)
    
    reconnectTimer = setTimeout(() => {
      connect()
    }, reconnectInterval)
  }

  // 停止重连
  const stopReconnect = () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
  }

  // 手动重连
  const reconnect = () => {
    disconnect()
    reconnectAttempts.value = 0
    nextTick(() => {
      connect()
    })
  }

  // 获取连接状态文本
  const getStatusText = () => {
    if (isConnected.value) return '已连接'
    if (isConnecting.value) return '连接中...'
    if (reconnectAttempts.value > 0) return `重连中... (${reconnectAttempts.value}/${maxReconnectAttempts})`
    return '未连接'
  }

  // 获取连接状态颜色
  const getStatusColor = () => {
    if (isConnected.value) return '#67c23a'
    if (isConnecting.value || reconnectAttempts.value > 0) return '#e6a23c'
    return '#f56c6c'
  }

  // 清理资源
  const cleanup = () => {
    disconnect()
    messageHandlers.clear()
  }

  // 组件卸载时清理
  onUnmounted(() => {
    cleanup()
  })

  return {
    // 状态
    isConnected,
    isConnecting,
    reconnectAttempts,
    lastMessage,
    connectionError,
    
    // 方法
    connect,
    disconnect,
    send,
    subscribe,
    reconnect,
    getStatusText,
    getStatusColor,
    cleanup
  }
}

// 创建全局WebSocket实例
export function createGlobalWebSocket() {
  const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws'
  
  return useWebSocket({
    url: wsUrl,
    autoReconnect: true,
    reconnectInterval: 3000,
    maxReconnectAttempts: 5,
    heartbeatInterval: 30000,
    onOpen: () => {
      console.log('全局WebSocket连接已建立')
    },
    onError: (event) => {
      console.error('全局WebSocket连接错误:', event)
    },
    onClose: (event) => {
      console.log('全局WebSocket连接已关闭:', event.code, event.reason)
    }
  })
}
