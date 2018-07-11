package api

type Server interface {
	Start(node LocalNode) error
	Stop()
}
