package dynamic

// +k8s:deepcopy-gen=true

// UDPMiddleware holds the UDPMiddleware configuration.
type UDPMiddleware struct {
	IPAllowList *UDPIPAllowList `json:"ipAllowList,omitempty" toml:"ipAllowList,omitempty" yaml:"ipAllowList,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// UDPIPAllowList holds the UDP IPAllowList middleware configuration.
type UDPIPAllowList struct {
	// SourceRange defines the allowed IPs (or ranges of allowed IPs by using CIDR notation).
	SourceRange []string `json:"sourceRange,omitempty" toml:"sourceRange,omitempty" yaml:"sourceRange,omitempty"`
}
