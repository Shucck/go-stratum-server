package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"time"
)

// BlockHeader represents the block header structure.
type BlockHeader struct {
	Version        uint32
	PrevBlockHash  string
	MerkleRootHash string
	Timestamp      uint32
	Nonce          uint32
}

// GenerateBlockHeader generates a new block header for mining.
func GenerateBlockHeader(prevBlockHash string) BlockHeader {
	return BlockHeader{
		Version:        1,
		PrevBlockHash:  prevBlockHash,
		MerkleRootHash: "dummy-merkle-root-hash",
		Timestamp:      uint32(time.Now().Unix()),
		Nonce:          0,
	}
}

// NonceIsValid checks if the given nonce produces a valid block hash based on the target difficulty.
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

// intToHex converts an integer to a hexadecimal string.
func intToHex(n uint32) string {
	return hex.EncodeToString([]byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)})
}
