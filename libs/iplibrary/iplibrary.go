package iplibrary

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/lionsoul2014/ip2region/binding/golang/ip2region"
)

var cidrs []*net.IPNet

func init() {

	lancidrs := []string{
		"127.0.0.1/8", "10.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12", "192.168.0.0/16", "::1/128", "fc00::/7",
	}

	cidrs = make([]*net.IPNet, len(lancidrs))

	for i, it := range lancidrs {
		_, cidrnet, err := net.ParseCIDR(it)
		if err != nil {
			log.Fatalf("ParseCIDR error: %v", err) // assuming I did it right above
		}

		cidrs[i] = cidrnet
	}

}

var Iplibrary *ip2region.Ip2Region

func InitIPLibrary(ipPath string) error {
	var err error
	Iplibrary, err = ip2region.New(ipPath)
	if err != nil {
		return err
	}
	return nil
}

func GetInfoByIp(ip string) (ip2region.IpInfo, error) {
	var ipInfo ip2region.IpInfo

	ipInfo, err := Iplibrary.MemorySearch(ip)
	if err != nil {
		return ipInfo, err
	}
	return ipInfo, nil
}

func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

func isLocalAddress(addr string) bool {
	for i := range cidrs {
		myaddr := net.ParseIP(addr)
		if cidrs[i].Contains(myaddr) {
			return true
		}
	}

	return false
}

func RealIP(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")

	if len(hdrForwardedFor) == 0 && len(hdrRealIP) == 0 {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}

	// X-Forwarded-For is potentially a list of addresses separated with ","
	for _, addr := range strings.Split(hdrForwardedFor, ",") {
		// return first non-local address
		addr = strings.TrimSpace(addr)
		if len(addr) > 0 && !isLocalAddress(addr) {
			return addr
		}
	}

	return hdrRealIP
}
