package nodes

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/utils"
)

var LocalTVS []net.IP
var LocalTVSMu sync.RWMutex

var LocalRedis []string
var LocalRedisMu sync.RWMutex

var LocalObjectdb []string
var LocalObjectdbMu sync.RWMutex

var Regions []string
var RegionsMu sync.RWMutex

func GetAllLocalRedisNodes() []string {
	var ip []string
	LocalRedisMu.RLock()
	if len(LocalRedis) > 0 {
		ip = append(ip, LocalRedis...)
	}
	LocalRedisMu.RUnlock()
	return ip
}

func GetAllLocalObjectdbNodes() []string {
	var ip []string
	LocalObjectdbMu.RLock()
	if len(LocalObjectdb) > 0 {
		ip = append(ip, LocalObjectdb...)
	}
	LocalObjectdbMu.RUnlock()
	return ip
}

func startDNSNodeLookupRoutine() {
	getNodeLoopIter()

	go func() {
		for {
			time.Sleep(5 * time.Minute)
			getNodeLoopIter()
		}
	}()
}

func getNodeLoopIter() {
	regs, nodes, err := getAllNodesDNS()
	if err != nil {
		fmt.Println("Discovery loop:", err)
		time.Sleep(1 * time.Minute)
		return
	}

	RegionsMu.Lock()
	Regions = regs
	RegionsMu.Unlock()

	lr, ok := nodes[config.Region]
	if ok {
		tmp, ok := lr[config.TVSRole]
		if ok {
			lt := parseIPs(tmp)
			if len(lt) > 0 {
				LocalTVSMu.Lock()
				LocalTVS = lt
				LocalTVSMu.Unlock()
			}
		}

		tmp, ok = lr[config.RedisRole]
		if ok {
			lts := checkIPPort(tmp)
			if len(lts) > 0 {
				LocalRedisMu.Lock()
				LocalRedis = lts
				LocalRedisMu.Unlock()
			}
		}

		tmp, ok = lr[config.ObjectdbRole]
		if ok {
			lts := checkIPPort(tmp)
			if len(lts) > 0 {
				LocalObjectdbMu.Lock()
				LocalObjectdb = lts
				LocalObjectdbMu.Unlock()
			}
		}
	}
}

func checkIPPort(ips []string) []string {
	var lt []string
	for i := 0; i < len(ips); i++ {
		spl := strings.Split(ips[i], ":")
		if len(spl) == 1 {
			ip := net.ParseIP(spl[0])
			if ip == nil {
				utils.Warn("msg", "Discovery loop:", "invalid ip "+ips[i])
				continue
			}
			ip = ip.To4()
			if ip == nil {
				utils.Warn("msg", "Discovery loop:", "invalid ip "+ips[i])
				continue
			}
			lt = append(lt, spl[0])
		} else if len(spl) == 2 {
			ip := net.ParseIP(spl[0])
			if ip == nil {
				utils.Warn("msg", "Discovery loop:", "invalid ip "+ips[i])
				continue
			}
			ip = ip.To4()
			if ip == nil {
				utils.Warn("msg", "Discovery loop:", "invalid ip "+ips[i])
				continue
			}
			_, err := strconv.Atoi(spl[1])
			if err != nil {
				utils.Warn("msg", "Discovery loop:", "invalid ip "+ips[i])
				continue
			}
			lt = append(lt, ips[i])
		}
	}
	return lt
}

func parseIPs(ips []string) []net.IP {
	var lt []net.IP
	for i := 0; i < len(ips); i++ {
		ip := net.ParseIP(ips[i])
		if ip == nil {
			utils.Warn("msg", "Discovery loop:", "invalid ip "+ips[i])
			continue
		}
		ip = ip.To4()
		if ip == nil {
			utils.Warn("msg", "Discovery loop:", "invalid ip "+ips[i])
			continue
		}
		lt = append(lt, ip)
	}
	return lt
}

func getAllNodesDNS() ([]string, map[string]map[string][]string, error) {
	ret := map[string]map[string][]string{}

	regions, err := utils.R.LookupTXT(context.Background(), config.RegionTxtRecord())
	if err != nil {
		return nil, nil, err
	}

	for i := 0; i < len(regions); i++ {
		ret[regions[i]] = map[string][]string{}

		for j := 0; j < len(config.PossibleRoles); j++ {
			ips, err := utils.R.LookupTXT(context.Background(), config.RegionNodeTxtRecord(regions[i], config.PossibleRoles[j]))
			if err == nil {
				ret[regions[i]][config.PossibleRoles[j]] = ips
			}
		}
	}

	return regions, ret, nil
}
