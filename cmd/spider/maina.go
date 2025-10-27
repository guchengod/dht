package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync/atomic" // 用于计数
	"time"

	"github.com/anacrolix/dht/v2"
	"github.com/anacrolix/log"
	"github.com/anacrolix/publicip"
	"github.com/anacrolix/torrent/metainfo"
)

// 用于统计收到的 announce 数量
var announceCounter uint64

func main() {
	// --- 初始化 ---
	logger := log.Default.WithNames("sniffer") // 使用不同的名称
	ctx := log.ContextWithLogger(context.Background(), logger)
	ctx, stopSignalNotify := signal.NotifyContext(ctx, os.Interrupt)
	defer stopSignalNotify()

	logger.Print("启动 DHT 嗅探器...")

	// --- 配置 DHT 服务器 ---
	cfg := dht.NewDefaultServerConfig()

	// 1. 设置网络 (强制 IPv4 避免之前的网络问题)
	conn, err := net.ListenPacket("udp4", ":0") // 使用 udp4 监听随机 IPv4 端口
	if err != nil {
		logger.Printf("监听 UDP 端口失败: %v", err)
		return
	}
	cfg.Conn = conn
	logger.Printf("DHT 节点正在监听: %s", conn.LocalAddr())

	// 2. 获取并设置公网 IP (用于 BEP 42)
	publicIPv4, err := publicip.Get(context.TODO(), "udp4") // 获取 IPv4 公网地址
	if err == nil && len(publicIPv4) > 0 {
		cfg.PublicIP = publicIPv4[0]
		cfg.NoSecurity = false // 启用 BEP 42 安全扩展
		logger.Printf("获取到公网 IPv4: %q，已启用安全扩展", cfg.PublicIP)
	} else {
		cfg.NoSecurity = true // 如果无法获取公网 IP，则禁用安全扩展
		logger.Printf("无法获取公网 IPv4 地址 (%v)，将禁用 DHT 安全扩展。节点可能难以被发现。", err)
	}

	// 3. 设置 OnAnnouncePeer 回调函数 (核心)
	cfg.OnAnnouncePeer = func(infoHash metainfo.Hash, ip net.IP, port int, portOk bool) {
		// 增加计数器
		atomic.AddUint64(&announceCounter, 1)
		// 打印收到的信息
		// 注意：在高流量下，这里打印日志可能会成为性能瓶颈，实际应用中可能需要写入文件或数据库
		logger.Printf("收到 Announce: Infohash=%s, Peer=%s:%d", infoHash.HexString(), ip.String(), port)
		// 在这里可以添加你自己的处理逻辑，例如将 infohash 存入数据库等
	}

	// 4. 配置启动节点 (过滤 IPv6)
	cfg.StartingNodes = func() ([]dht.Addr, error) {
		hostPorts := dht.DefaultGlobalBootstrapHostPorts
		logger.Printf("正在解析启动节点: %v", hostPorts)
		resolvedAddrs, err := dht.ResolveHostPorts(hostPorts)
		// 即使解析部分失败，也尝试使用成功的
		if err != nil {
			logger.Printf("解析部分启动节点时出错: %v", err)
		}

		var ipv4Addrs []dht.Addr
		for _, addr := range resolvedAddrs {
			if addr.IP().To4() != nil {
				ipv4Addrs = append(ipv4Addrs, addr)
			}
		}

		if len(ipv4Addrs) == 0 {
			logger.Printf("未能解析到任何 IPv4 启动节点，无法启动 bootstrap")
			return nil, fmt.Errorf("no IPv4 starting nodes resolved")
		}
		logger.Printf("已解析到 %d 个 IPv4 启动节点", len(ipv4Addrs))
		return ipv4Addrs, nil
	}

	// 5. 可选：调整其他参数
	cfg.QueryResendDelay = func() time.Duration { return 3 * time.Second } // 稍微增加超时
	// cfg.DefaultWant = []krpc.Want{krpc.WantNodes} // 如果确定只处理 IPv4

	// --- 创建并启动 DHT 服务器 ---
	s, err := dht.NewServer(cfg)
	if err != nil {
		logger.Printf("创建 DHT 服务器失败: %v", err)
		return
	}
	logger.Printf("DHT 服务器已创建，节点 ID: %s", s.ID())
	defer s.Close()

	// --- 启动路由表维护 ---
	// 这是让节点保持活跃、被发现并接收 announce 的关键
	go func() {
		logger.Printf("启动 DHT 路由表维护程序...")
		s.TableMaintainer()              //
		logger.Printf("DHT 路由表维护程序已停止。") // 只有在 Server 关闭时才会到这里
	}()

	// --- 定期打印统计信息 (可选) ---
	go func() {
		ticker := time.NewTicker(10 * time.Second) // 每 30 秒打印一次
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				stats := s.Stats()
				currentCount := atomic.LoadUint64(&announceCounter)
				logger.Printf("状态: GoodNodes=%d, TotalNodes=%d, OutstandingTx=%d, ReceivedAnnounces=%d",
					stats.GoodNodes, stats.Nodes, stats.OutstandingTransactions, currentCount)
			case <-ctx.Done():
				return
			}
		}
	}()

	// --- 等待程序退出信号 ---
	logger.Printf("嗅探器正在运行... 按 Ctrl+C 退出。")
	<-ctx.Done() // 阻塞直到接收到中断信号

	// --- 清理 ---
	logger.Printf("正在关闭 DHT 服务器...")
	// defer s.Close() 会在这里执行
	logger.Printf("嗅探器已停止。")
	finalCount := atomic.LoadUint64(&announceCounter)
	logger.Printf("总共收到 %d 个 AnnouncePeer 消息。", finalCount)
}
