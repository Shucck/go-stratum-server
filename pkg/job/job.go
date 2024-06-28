package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type BlockHeader struct {
	Version        uint32 `json:"version"`
	PrevBlockHash  string `json:"prev_block_hash"`
	MerkleRootHash string `json:"merkle_root_hash"`
	Timestamp      uint32 `json:"timestamp"`
	Nonce          uint32 `json:"nonce"`
}

func NewJob(header BlockHeader) *Job {
	data, err := json.Marshal(header)
	if err != nil {
		log.Printf("Error encoding block header: %v", err)
		return nil
	}
	hexData := hex.EncodeToString(data)

	return &Job{
		ID:   generateJobID(),
		Data: hexData,
	}
}

func GenerateBlockHeader(prevBlockHash string) BlockHeader {
	return BlockHeader{
		Version:        1,
		PrevBlockHash:  prevBlockHash,
		MerkleRootHash: "dummy-merkle-root-hash",
		Timestamp:      uint32(time.Now().Unix()),
		Nonce:          0,
	}
}


func NonceIsValid(header BlockHeader, targetDifficulty uint32) bool {
	data, err := json.Marshal(header)
	if err != nil {
		log.Printf("Error encoding block header: %v", err)
		return false
	}
	hexData := hex.EncodeToString(data)

	headerWithNonce := hexData + intToHex(header.Nonce)
	hash := sha256.Sum256([]byte(headerWithNonce))
	hashString := hex.EncodeToString(hash[:])

	return hashString[:targetDifficulty/4] == "0000"
}

func intToHex(n uint32) string {
	return hex.EncodeToString([]byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)})
}

//to generaes a unique job ID.
func generateJobID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("job-%d", timestamp)
}

