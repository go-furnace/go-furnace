package cfg

// Configuration is a Configuration object.
type Configuration struct {
  LogLevel string
  UploadPath string
}

type Monitoring struct {
  Enable string `json:"enabled"`
}

type EC2Config struct {
  DryRun string `json:"dry_run"`
  ImageID string `json:"image_id"`
  KeyName string `json:"key_name"`
  MinCount int `json:"min_count"`
  MaxCount int `json:"max_count"`
  InstanceType string `json:"instance_type"`
  Monitoring Monitoring `json:"monitoring"`
}

// IPRange ip range
type IPRange struct {
  // CidrIP cidr_ip
  CidrIP string `json:"cidr_ip"`
}

// IPPermission ip permission
type IPPermission struct {
  IPProtocol string `json:"ip_protocol"`
  FromPort string `json:"from_port"`
  ToPort string `json:"to_port"`
  IPRanges []IPRange `json:"ip_ranges"`
}

// SecurityGroup Security Group
type SecurityGroup struct {
  IPPermissions []IPPermission `json:"ip_permission"`
}

// LoadConfiguration Loads the configuration file.
func LoadConfiguration() {

}
