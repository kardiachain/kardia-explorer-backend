package types

type ServerStatus struct {
	Status        string `json:"status"`
	AppVersion    string `json:"appVersion"`
	ServerVersion string `json:"serverVersion"`
	DexStatus     string `json:"dexStatus"`
}
