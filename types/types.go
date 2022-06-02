package types

type Request struct {
	PubKey  string `json:"PubKey"`
	IPAddr  string `json:"IPAddr"`
	Comment string `json:"Comment"`
}
type ServerOpts struct {
	PersistentKeepAlive int
	PresharedKey        string
	Endpoint            string
}
