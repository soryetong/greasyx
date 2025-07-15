package ginahelper

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}

	for _, iface := range interfaces {
		// 忽略不活跃的接口、回环接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// 排除虚拟接口（如 docker、utun 等）
		if strings.HasPrefix(iface.Name, "docker") || strings.HasPrefix(iface.Name, "utun") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // 只处理 IPv4
			}

			return ip.String()
		}
	}

	return "unknown"
}

// GetClientRealIP 获取客户端真实 IP（适用于 Gin 框架）
func GetClientRealIP(c *gin.Context) string {
	// 先检查 X-Forwarded-For 头
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		// XFF 是逗号分隔的 IP 列表，第一个通常是客户端 IP
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// 再检查 X-Real-IP
	xri := c.GetHeader("X-Real-IP")
	if xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// 最后使用 RemoteAddr
	return getIPFromRemoteAddr(c.Request.RemoteAddr)
}

// getIPFromRemoteAddr 从 RemoteAddr 中提取 IP 地址（格式可能是 IP:PORT）
func getIPFromRemoteAddr(addr string) string {
	if strings.Contains(addr, ":") {
		host, _, err := net.SplitHostPort(strings.TrimSpace(addr))
		if err == nil && net.ParseIP(host) != nil {
			return host
		}
	}
	if net.ParseIP(addr) != nil {
		return addr
	}

	return ""
}
