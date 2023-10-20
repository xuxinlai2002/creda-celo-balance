package utils

import (
	"math/big"
	"os"
)

const PullProgressFile = "progress.txt"

func WriteCurrentHeight(height uint64) error {

	content := big.NewInt(0).SetUint64(height)
	err := os.WriteFile(PullProgressFile, content.Bytes(), 0666)
	return err
}

func GetCurrentHeight(filePath string) (uint64, error) {
	data, err := os.ReadFile(PullProgressFile)
	if err != nil {
		return 0, err
	}
	v := big.NewInt(0).SetBytes(data)
	return v.Uint64(), nil
}
