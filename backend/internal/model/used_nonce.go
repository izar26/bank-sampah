package model

import "time"

// UsedNonce prevents replay attacks by tracking consumed nonces
type UsedNonce struct {
	Nonce     string    `gorm:"primaryKey;size:100" json:"nonce"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
}

func (UsedNonce) TableName() string {
	return "used_nonces"
}
