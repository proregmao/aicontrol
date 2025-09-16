package shared

import (
	"smart-device-management/pkg/security"
)

// 全局JWT管理器实例，在整个应用中共享
var GlobalJWTManager = security.NewJWTManager("your-secret-key", "smart-device-management")
