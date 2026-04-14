// Copyright 2016 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package msg

import (
	"net"
	"reflect"
)

const (
	TypeLogin              = 'L'
	TypeLoginResp          = 'l'
	TypeNewProxy           = 'P'
	TypeNewProxyResp       = 'p'
	TypeCloseProxy         = 'C'
	TypeNewWorkConn        = 'W'
	TypeReqWorkConn        = 'R'
	TypeStartWorkConn      = 'S'
	TypeNewVisitorConn     = 'V'
	TypeNewVisitorConnResp = 'v'
	TypePing               = 'H'
	TypePong               = 'h'
	TypeUDPPacket          = 'U'
	TypeNatHoleVisitor     = 'I'
	TypeNatHoleClient      = 'N'
	TypeNatHoleResp        = 'M'
	TypeNatHoleSid         = 'Y'
	TypeNatHoleReport      = 'Z'
)

var msgTypeMap = map[byte]any{
	TypeLogin:              Login{},
	TypeLoginResp:          LoginResp{},
	TypeNewProxy:           NewProxy{},
	TypeNewProxyResp:       NewProxyResp{},
	TypeCloseProxy:         CloseProxy{},
	TypeNewWorkConn:        NewWorkConn{},
	TypeReqWorkConn:        ReqWorkConn{},
	TypeStartWorkConn:      StartWorkConn{},
	TypeNewVisitorConn:     NewVisitorConn{},
	TypeNewVisitorConnResp: NewVisitorConnResp{},
	TypePing:               Ping{},
	TypePong:               Pong{},
	TypeUDPPacket:          UDPPacket{},
	TypeNatHoleVisitor:     NatHoleVisitor{},
	TypeNatHoleClient:      NatHoleClient{},
	TypeNatHoleResp:        NatHoleResp{},
	TypeNatHoleSid:         NatHoleSid{},
	TypeNatHoleReport:      NatHoleReport{},
}

var TypeNameNatHoleResp = reflect.TypeFor[NatHoleResp]().Name()

type ClientSpec struct {
	// Due to the support of VirtualClient, frps needs to know the client type in order to
	// differentiate the processing logic.
	// Optional values: ssh-tunnel
	Type string `json:"client_type,omitempty"`
	// If the value is true, the client will not require authentication.
	AlwaysAuthPass bool `json:"skip_auth,omitempty"`
}

// When frpc start, client send this message to login to server.
type Login struct {
	Version      string            `json:"build_version,omitempty"`
	Hostname     string            `json:"device_name,omitempty"`
	Os           string            `json:"platform_name,omitempty"`
	Arch         string            `json:"platform_arch,omitempty"`
	User         string            `json:"account_id,omitempty"`
	PrivilegeKey string            `json:"auth_token,omitempty"`
	Timestamp    int64             `json:"event_time,omitempty"`
	RunID        string            `json:"session_id,omitempty"`
	ClientID     string            `json:"terminal_id,omitempty"`
	Metas        map[string]string `json:"attributes,omitempty"`

	// Currently only effective for VirtualClient.
	ClientSpec ClientSpec `json:"client_profile,omitempty"`

	// Some global configures.
	PoolCount int `json:"worker_pool,omitempty"`
}

type LoginResp struct {
	Version string `json:"build_version,omitempty"`
	RunID   string `json:"session_id,omitempty"`
	Error   string `json:"message,omitempty"`
}

// When frpc login success, send this message to frps for running a new proxy.
type NewProxy struct {
	ProxyName          string            `json:"service_name,omitempty"`
	ProxyType          string            `json:"channel_type,omitempty"`
	UseEncryption      bool              `json:"security_enabled,omitempty"`
	UseCompression     bool              `json:"payload_compressed,omitempty"`
	BandwidthLimit     string            `json:"rate_limit,omitempty"`
	BandwidthLimitMode string            `json:"rate_plan,omitempty"`
	Group              string            `json:"tenant_group,omitempty"`
	GroupKey           string            `json:"tenant_key,omitempty"`
	Metas              map[string]string `json:"attributes,omitempty"`
	Annotations        map[string]string `json:"labels,omitempty"`

	// tcp and udp only
	RemotePort int `json:"access_port,omitempty"`

	// http and https only
	CustomDomains     []string          `json:"domain_list,omitempty"`
	SubDomain         string            `json:"domain_prefix,omitempty"`
	Locations         []string          `json:"route_paths,omitempty"`
	HTTPUser          string            `json:"auth_user,omitempty"`
	HTTPPwd           string            `json:"auth_secret,omitempty"`
	HostHeaderRewrite string            `json:"host_rewrite,omitempty"`
	Headers           map[string]string `json:"request_headers,omitempty"`
	ResponseHeaders   map[string]string `json:"response_headers,omitempty"`
	RouteByHTTPUser   string            `json:"route_user,omitempty"`

	// stcp, sudp, xtcp
	Sk         string   `json:"shared_secret,omitempty"`
	AllowUsers []string `json:"allow_accounts,omitempty"`

	// tcpmux
	Multiplexer string `json:"stream_mode,omitempty"`
}

type NewProxyResp struct {
	ProxyName  string `json:"service_name,omitempty"`
	RemoteAddr string `json:"endpoint,omitempty"`
	Error      string `json:"message,omitempty"`
}

type CloseProxy struct {
	ProxyName string `json:"service_name,omitempty"`
}

type NewWorkConn struct {
	RunID        string `json:"session_id,omitempty"`
	PrivilegeKey string `json:"auth_token,omitempty"`
	Timestamp    int64  `json:"event_time,omitempty"`
}

type ReqWorkConn struct{}

type StartWorkConn struct {
	ProxyName string `json:"service_name,omitempty"`
	SrcAddr   string `json:"client_addr,omitempty"`
	DstAddr   string `json:"server_addr,omitempty"`
	SrcPort   uint16 `json:"client_port,omitempty"`
	DstPort   uint16 `json:"server_port,omitempty"`
	Error     string `json:"message,omitempty"`
}

type NewVisitorConn struct {
	RunID          string `json:"session_id,omitempty"`
	ProxyName      string `json:"service_name,omitempty"`
	SignKey        string `json:"signature,omitempty"`
	Timestamp      int64  `json:"event_time,omitempty"`
	UseEncryption  bool   `json:"security_enabled,omitempty"`
	UseCompression bool   `json:"payload_compressed,omitempty"`
}

type NewVisitorConnResp struct {
	ProxyName string `json:"service_name,omitempty"`
	Error     string `json:"message,omitempty"`
}

type Ping struct {
	PrivilegeKey string `json:"auth_token,omitempty"`
	Timestamp    int64  `json:"event_time,omitempty"`
}

type Pong struct {
	Error string `json:"message,omitempty"`
}

type UDPPacket struct {
	Content    []byte       `json:"payload,omitempty"`
	LocalAddr  *net.UDPAddr `json:"local_endpoint,omitempty"`
	RemoteAddr *net.UDPAddr `json:"peer_endpoint,omitempty"`
}

type NatHoleVisitor struct {
	TransactionID string   `json:"transaction_id,omitempty"`
	ProxyName     string   `json:"service_name,omitempty"`
	PreCheck      bool     `json:"preflight,omitempty"`
	Protocol      string   `json:"network_protocol,omitempty"`
	SignKey       string   `json:"signature,omitempty"`
	Timestamp     int64    `json:"event_time,omitempty"`
	MappedAddrs   []string `json:"mapped_endpoints,omitempty"`
	AssistedAddrs []string `json:"relay_endpoints,omitempty"`
}

type NatHoleClient struct {
	TransactionID string   `json:"transaction_id,omitempty"`
	ProxyName     string   `json:"service_name,omitempty"`
	Sid           string   `json:"link_id,omitempty"`
	MappedAddrs   []string `json:"mapped_endpoints,omitempty"`
	AssistedAddrs []string `json:"relay_endpoints,omitempty"`
}

type PortsRange struct {
	From int `json:"start_port,omitempty"`
	To   int `json:"end_port,omitempty"`
}

type NatHoleDetectBehavior struct {
	Role              string       `json:"node_role,omitempty"` // sender or receiver
	Mode              int          `json:"strategy_mode,omitempty"` // 0, 1, 2...
	TTL               int          `json:"packet_ttl,omitempty"`
	SendDelayMs       int          `json:"send_delay_ms,omitempty"`
	ReadTimeoutMs     int          `json:"read_timeout_ms,omitempty"`
	CandidatePorts    []PortsRange `json:"candidate_ports,omitempty"`
	SendRandomPorts   int          `json:"random_send_ports,omitempty"`
	ListenRandomPorts int          `json:"random_listen_ports,omitempty"`
}

type NatHoleResp struct {
	TransactionID  string                `json:"transaction_id,omitempty"`
	Sid            string                `json:"link_id,omitempty"`
	Protocol       string                `json:"network_protocol,omitempty"`
	CandidateAddrs []string              `json:"candidate_endpoints,omitempty"`
	AssistedAddrs  []string              `json:"relay_endpoints,omitempty"`
	DetectBehavior NatHoleDetectBehavior `json:"probe_profile,omitempty"`
	Error          string                `json:"message,omitempty"`
}

type NatHoleSid struct {
	TransactionID string `json:"transaction_id,omitempty"`
	Sid           string `json:"link_id,omitempty"`
	Response      bool   `json:"is_response,omitempty"`
	Nonce         string `json:"challenge_code,omitempty"`
}

type NatHoleReport struct {
	Sid     string `json:"link_id,omitempty"`
	Success bool   `json:"success,omitempty"`
}
