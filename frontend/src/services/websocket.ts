import { ElMessage } from 'element-plus'

// WebSocket消息类型
export interface WebSocketMessage {
  type: string
  data: any
  timestamp: number
}

// WebSocket事件类型
export type WebSocketEventType = 
  | 'device_status_update'
  | 'temperature_data'
  | 'breaker_data'
  | 'server_data'
  | 'alarm_triggered'
  | 'ai_control_executed'

// WebSocket服务类
export class WebSocketService {
  private ws: WebSocket | null = null
  private url: string
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectInterval = 3000
  private heartbeatInterval: number | null = null
  private listeners: Map<WebSocketEventType, Function[]> = new Map()

  constructor(url?: string) {
    this.url = url || `ws://localhost:8080/ws`
  }

  // 连接WebSocket
  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url)

        this.ws.onopen = () => {
          console.log('WebSocket连接已建立')
          this.reconnectAttempts = 0
          this.startHeartbeat()
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data)
            this.handleMessage(message)
          } catch (error) {
            console.error('解析WebSocket消息失败:', error)
          }
        }

        this.ws.onclose = (event) => {
          console.log('WebSocket连接已关闭:', event.code, event.reason)
          this.stopHeartbeat()
          
          if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnect()
          }
        }

        this.ws.onerror = (error) => {
          console.error('WebSocket连接错误:', error)
          reject(error)
        }
      } catch (error) {
        reject(error)
      }
    })
  }

  // 断开连接
  disconnect() {
    if (this.ws) {
      this.ws.close(1000, '主动断开连接')
      this.ws = null
    }
    this.stopHeartbeat()
  }

  // 重连
  private reconnect() {
    this.reconnectAttempts++
    console.log(`尝试重连WebSocket (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)
    
    setTimeout(() => {
      this.connect().catch(() => {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
          ElMessage.error('WebSocket连接失败，请刷新页面重试')
        }
      })
    }, this.reconnectInterval)
  }

  // 发送消息
  send(type: string, data: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message: WebSocketMessage = {
        type,
        data,
        timestamp: Date.now()
      }
      this.ws.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket未连接，无法发送消息')
    }
  }

  // 处理接收到的消息
  private handleMessage(message: WebSocketMessage) {
    const { type, data } = message
    const listeners = this.listeners.get(type as WebSocketEventType)
    
    if (listeners) {
      listeners.forEach(listener => {
        try {
          listener(data)
        } catch (error) {
          console.error('WebSocket消息处理错误:', error)
        }
      })
    }
  }

  // 添加事件监听器
  on(eventType: WebSocketEventType, listener: Function) {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, [])
    }
    this.listeners.get(eventType)!.push(listener)
  }

  // 移除事件监听器
  off(eventType: WebSocketEventType, listener: Function) {
    const listeners = this.listeners.get(eventType)
    if (listeners) {
      const index = listeners.indexOf(listener)
      if (index > -1) {
        listeners.splice(index, 1)
      }
    }
  }

  // 开始心跳
  private startHeartbeat() {
    this.heartbeatInterval = window.setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.send('ping', { timestamp: Date.now() })
      }
    }, 30000) // 30秒心跳
  }

  // 停止心跳
  private stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval)
      this.heartbeatInterval = null
    }
  }

  // 获取连接状态
  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }
}

// 创建全局WebSocket实例
export const websocketService = new WebSocketService()

// 设备状态更新处理
export const handleDeviceStatusUpdate = (callback: (data: any) => void) => {
  websocketService.on('device_status_update', callback)
}

// 温度数据更新处理
export const handleTemperatureData = (callback: (data: any) => void) => {
  websocketService.on('temperature_data', callback)
}

// 断路器数据更新处理
export const handleBreakerData = (callback: (data: any) => void) => {
  websocketService.on('breaker_data', callback)
}

// 服务器数据更新处理
export const handleServerData = (callback: (data: any) => void) => {
  websocketService.on('server_data', callback)
}

// 告警触发处理
export const handleAlarmTriggered = (callback: (data: any) => void) => {
  websocketService.on('alarm_triggered', callback)
}

// AI控制执行处理
export const handleAIControlExecuted = (callback: (data: any) => void) => {
  websocketService.on('ai_control_executed', callback)
}

// 初始化WebSocket连接
export const initWebSocket = async () => {
  try {
    await websocketService.connect()
    console.log('WebSocket服务初始化成功')
  } catch (error) {
    console.error('WebSocket服务初始化失败:', error)
    // 不显示错误消息，因为在开发环境中WebSocket服务可能不可用
  }
}

// 清理WebSocket连接
export const cleanupWebSocket = () => {
  websocketService.disconnect()
  console.log('WebSocket服务已清理')
}
