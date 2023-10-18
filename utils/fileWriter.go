package utils

import (
	"math/big"
	"os"
)

const PullProgressFile = "progress.txt"

func WriteCurrentHeight(filePath string, height uint64) error {

	content := big.NewInt(0).SetUint64(height)
	name := filePath + PullProgressFile
	err := os.WriteFile(name, content.Bytes(), 0666)
	return err
}

func GetCurrentHeight(filePath string) (uint64, error) {
	name := filePath + PullProgressFile
	data, err := os.ReadFile(name)
	if err != nil {
		return 0, err
	}
	v := big.NewInt(0).SetBytes(data)
	return v.Uint64(), nil
}
