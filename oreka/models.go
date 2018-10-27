package oreka

type OrkTape struct {
	ID              int         `json:"-"`
	Direction       int         `json:"-"`
	Duration        int         `json:"duration"`
	ExpiryTimestamp string      `json:"expiryTimestamp"`
	Filename        string      `json:"filename"`
	LocalEntryPoint string      `json:"localEntryPoint"`
	LocalParty      string      `json:"localParty"`
	PortName        string      `json:"portName"`
	RemoteParty     string      `json:"remoteParty"`
	Timestamp       string      `json:"timestamp"`
	NativeCallID    string      `json:"nativeCallId"`
	PortID          interface{} `json:"-"`
	ServiceID       int         `json:"-"`
}
