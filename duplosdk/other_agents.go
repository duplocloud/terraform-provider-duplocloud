package duplosdk

type DuploDuploOtherAgentDisableInfo struct {
	IsUninstallPending bool `json:"IsUninstallPending,omitempty"`
	IsDisabled         bool `json:"IsDisabled,omitempty"`
	ExecutionCount     int  `json:"ExecutionCount,omitempty"`
}
type DuploDuploOtherAgent struct {
	AgentLinuxPackagePath        string                           `json:"AgentLinuxPackagePath,omitempty"`
	AgentName                    string                           `json:"AgentName,omitempty"`
	AgentWindowsPackagePath      string                           `json:"AgentWindowsPackagePath,omitempty"`
	LinuxAgentInstallStatusCmd   string                           `json:"LinuxAgentInstallStatusCmd,omitempty"`
	LinuxAgentServiceName        string                           `json:"LinuxAgentServiceName,omitempty"`
	LinuxAgentUninstallStatusCmd string                           `json:"LinuxAgentUninstallStatusCmd,omitempty"`
	LinuxInstallCmd              string                           `json:"LinuxInstallCmd,omitempty"`
	WindowsAgentServiceName      string                           `json:"WindowsAgentServiceName,omitempty"`
	UserRequestResetIsPending    bool                             `json:"UserRequestResetIsPending,omitempty"`
	ExecutionCount               int                              `json:"ExecutionCount,omitempty"`
	DisableInfo                  *DuploDuploOtherAgentDisableInfo `json:"DisableInfo,omitempty"`
}

type DuploDuploOtherAgentReq struct {
	AgentLinuxPackagePath        string
	AgentName                    string
	AgentWindowsPackagePath      string
	LinuxAgentInstallStatusCmd   string
	LinuxAgentServiceName        string
	LinuxAgentUninstallStatusCmd string
	LinuxInstallCmd              string
	WindowsAgentServiceName      string
}

func (c *Client) DuploOtherAgentCreate(rq *[]DuploDuploOtherAgentReq) ClientError {
	return c.postAPI(
		"DuploOtherAgentCreate",
		"compliance/UpdateOtherAgentConfig",
		&rq,
		nil,
	)
}

func (c *Client) DuploOtherAgentGet() (*[]DuploDuploOtherAgent, ClientError) {
	return c.DuploOtherAgentList()
}

func (c *Client) DuploOtherAgentList() (*[]DuploDuploOtherAgent, ClientError) {
	rp := []DuploDuploOtherAgent{}
	err := c.getAPI(
		"DuploOtherAgentList",
		"compliance/GetOtherAgentConfig",
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploOtherAgentExists() (bool, ClientError) {
	list, err := c.DuploOtherAgentList()
	if err != nil {
		return false, err
	}

	if list != nil && len(*list) > 0 {
		for _, element := range *list {
			if len(element.AgentName) > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) DuploOtherAgentDelete() ClientError {
	return c.postAPI(
		"DuploOtherAgentDelete",
		"compliance/UpdateOtherAgentConfig",
		&[]DuploDuploOtherAgent{},
		nil,
	)
}
