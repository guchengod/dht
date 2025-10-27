package main

//import (
//	"context"
//	"fmt"
//	"net"
//	"os"
//	"os/signal"
//	"time"
//
//	"github.com/anacrolix/dht/v2"
//	"github.com/anacrolix/log"
//	"github.com/anacrolix/publicip"
//	"github.com/anacrolix/torrent/metainfo"
//)
//
//func main() {
//	logger := log.Default.WithNames("main")
//	ctx := log.ContextWithLogger(context.Background(), logger)
//	ctx, stopSignalNotify := signal.NotifyContext(ctx, os.Interrupt)
//	defer stopSignalNotify()
//	var s *dht.Server
//	cfg := dht.NewDefaultServerConfig()
//	cfg.QueryResendDelay = func() time.Duration { return 10 * time.Second }
//	conn, err := net.ListenPacket("udp4", ":0")
//	if err != nil {
//		err = fmt.Errorf("listening: %w", err)
//		return
//	}
//	cfg.StartingNodes = func() ([]dht.Addr, error) {
//		// 使用默认的全局启动节点
//		resolvedAddrs, err := dht.ResolveHostPorts(dht.DefaultGlobalBootstrapHostPorts)
//		if err != nil {
//			return nil, err
//		}
//		var ipv4Addrs []dht.Addr
//		for _, addr := range resolvedAddrs {
//			// 只保留 IPv4 地址
//			if addr.IP().To4() != nil && addr.IP().To16() != nil && addr.IP().To4().Equal(addr.IP()) { // 确保是纯 IPv4
//				ipv4Addrs = append(ipv4Addrs, addr)
//			} else if addr.IP().To4() != nil { // 有些可能是IPv4映射的IPv6地址，也尝试保留
//				ipv4Addrs = append(ipv4Addrs, addr)
//			}
//		}
//		if len(ipv4Addrs) == 0 {
//			return nil, fmt.Errorf("no IPv4 starting nodes resolved")
//		}
//		log.Printf("Resolved %d IPv4 starting nodes", len(ipv4Addrs))
//		return ipv4Addrs, nil
//	}
//	cfg.Conn = conn
//	all, err := publicip.Get(context.TODO(), "udp4")
//	if err == nil {
//		cfg.PublicIP = all[0]
//		log.Printf("public ip: %q", cfg.PublicIP)
//		cfg.NoSecurity = false
//	}
//	cfg.OnAnnouncePeer = func(infoHash metainfo.Hash, ip net.IP, port int, portOk bool) {
//		log.Printf("announced peer: %s:%d  infohash: %s", ip, port, infoHash.String())
//	}
//	//cfg.StartingNodes = func() ([]dht.Addr, error) {
//	//	return dht.ResolveHostPorts(serverArgs.BootstrapAddr)
//	//}
//	s, err = dht.NewServer(cfg)
//	if err != nil {
//		log.Printf("error: %v", err)
//	} else {
//		log.Printf("dht server on %s with id %x", s.Addr(), s.ID())
//	}
//	defer s.Close()
//	go func() {
//		qlimt := dht.QueryRateLimiting{}
//		ticker := time.NewTicker(5 * time.Second)
//		for {
//			select {
//			case <-ticker.C:
//				startingNodes, err := cfg.StartingNodes()
//				if err != nil {
//					log.Printf("error: %v", err)
//				}
//				for _, addr := range startingNodes {
//					res := s.FindNode(addr, s.Id(), qlimt)
//					if res.Err != nil {
//						log.Printf("error: %v", res.Err)
//					} else {
//						log.Printf("find node: %v", res.Reply)
//					}
//
//				}
//			case <-ctx.Done():
//				ticker.Stop()
//				return
//			}
//		}
//	}()
//	//go func() {
//	//	log.Println("Starting DHT table maintainer...")
//	//	s.TableMaintainer() //
//	//	log.Println("DHT table maintainer stopped.")
//	//}()
//	<-ctx.Done()
//}
