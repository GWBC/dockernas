package models

type ParamItem struct {
	Prompt    string `json:"prompt"`
	Name      string `json:"name"`      // used for resource identification like local volume
	Key       string `json:"key"`       // resource in contariner
	Value     string `json:"value"`     // resource in host
	Reg       string `json:"reg"`       // Regular for input
	Hide      bool   `json:"hide"`      // is show on frontend
	Protocol  string `json:"protocol"`  // network protocol, used by port param
	MountFile bool   `json:"mountFile"` // mount file to container
}

type DockerTemplate struct {
	ImageUrl    string      `json:"imageUrl"`
	Version     string      `json:"version"`
	PortParams  []ParamItem `json:"portParams"`
	EnvParams   []ParamItem `json:"envParams"`
	LocalVolume []ParamItem `json:"localVolume"`
	DfsVolume   []ParamItem `json:"dfsVolume"`
	Privileged  bool        `json:"privileged"`
}

type InstanceParam struct {
	Name    string `json:"name"`
	AppName string `json:"appName"`
	Summary string `json:"summary"`
	IconUrl string `json:"iconUrl"`
	DockerTemplate
}
