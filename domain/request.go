package domain

import (
	"strings"

	"github.com/mssola/user_agent"
)

type RequestInfo struct {
	ipAddress string
	userAgent string
	referer   string
}

func NewRequestInfo(ipAddress, userAgent, referer string) *RequestInfo {
	return &RequestInfo{
		ipAddress: ipAddress,
		userAgent: userAgent,
		referer:   referer,
	}
}

func (r *RequestInfo) IpAddress() string {
	var ip = r.ipAddress
	if strings.Contains(r.ipAddress, ",") {
		ip = strings.Split(r.ipAddress, ",")[0]
	}
	return ip
}
func (r *RequestInfo) UserAgent() string {
	return r.userAgent
}
func (r *RequestInfo) Referer() string {
	return r.referer
}
func (r *RequestInfo) Browser() Browser {
	ua := user_agent.New(r.userAgent)
	name, version := ua.Browser()
	platform := ua.Platform()
	isMobile := ua.Mobile()

	return Browser{
		Name:     name,
		Version:  version,
		IsMobile: isMobile,
		Platform: platform,
	}
}
