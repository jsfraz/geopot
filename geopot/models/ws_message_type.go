package models

type WSMessageType string

const (
	WSMessageTypeAttackerInfo WSMessageType = "attacker"
	WSMessageTypeServeInfo    WSMessageType = "server_info"
)
