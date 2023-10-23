package utils

import (
	"errors"
	"math/big"
	"os"
)

const PullProgressFile = "progress.txt"

func WriteCurrentHeight(height uint64) error {

	content := big.NewInt(0).SetUint64(height)
	err := os.WriteFile(PullProgressFile, content.Bytes(), 0666)
	return err
}

func GetCurrentHeight() (uint64, error) {
	data, err := os.ReadFile(PullProgressFile)
	if err != nil {
		return 0, err
	}
	v := big.NewInt(0).SetBytes(data)
	return v.Uint64(), nil
}

const tokenProgressFile = "tokenProgress.txt"

func WriteTokenCurrentHeight(height uint64) error {
	str := big.NewInt(0).SetUint64(height).String()
	return os.WriteFile(tokenProgressFile, []byte(str), 0666)
}

func GetTokenCurrentHeight() (uint64, error) {
	data, err := os.ReadFile(tokenProgressFile)
	if err != nil {
		return 0, err
	}

	v, ok := big.NewInt(0).SetString(string(data), 10)
	if !ok {
		return 0, errors.New("get token current height err")
	}

	return v.Uint64(), nil
}
