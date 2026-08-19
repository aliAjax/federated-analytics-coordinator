package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Entry struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	TaskID       string    `json:"task_id,omitempty"`
	Action       string    `json:"action"`
	Actor        string    `json:"actor"`
	Detail       string    `json:"detail"`
	PreviousHash string    `json:"previous_hash"`
	Hash         string    `json:"hash"`
	At           time.Time `json:"at"`
}

func (e *Entry) Seal() {
	sum := sha256.Sum256([]byte(e.PreviousHash + "|" + e.TenantID + "|" + e.TaskID + "|" + e.Action + "|" + e.Actor + "|" + e.Detail + "|" + e.At.UTC().Format(time.RFC3339Nano)))
	e.Hash = hex.EncodeToString(sum[:])
}
