package transfer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyeokx/Go_GAPTUI/internal/models"
)

type AddressEntry struct {
	Exchange         string `json:"exchange"`
	Coin             string `json:"coin"`
	Network          string `json:"network"`
	Address          string `json:"address"`
	SecondaryAddress string `json:"secondary_address,omitempty"`
	Label            string `json:"label,omitempty"`
}

type addressFile struct {
	Addresses []AddressEntry `json:"addresses"`
}

func LookupAddress(exchange models.Exchange, coin string, network string) (address string, memo string, found bool) {
	file, ok := loadAddressFile()
	if !ok {
		return "", "", false
	}

	exchangeName := exchange.String()
	for _, entry := range file.Addresses {
		if strings.EqualFold(entry.Exchange, exchangeName) &&
			strings.EqualFold(entry.Coin, coin) &&
			strings.EqualFold(entry.Network, network) {
			return entry.Address, entry.SecondaryAddress, true
		}
	}

	return "", "", false
}

func UpsertAddress(exchange models.Exchange, coin string, network string, address string, memo string) bool {
	file, ok := loadAddressFile()
	if !ok {
		file = addressFile{Addresses: make([]AddressEntry, 0)}
	}

	exchangeName := exchange.String()
	for i := range file.Addresses {
		entry := &file.Addresses[i]
		if strings.EqualFold(entry.Exchange, exchangeName) &&
			strings.EqualFold(entry.Coin, coin) &&
			strings.EqualFold(entry.Network, network) {
			entry.Address = address
			entry.SecondaryAddress = memo
			_ = saveAddressFile(file)
			return false
		}
	}

	file.Addresses = append(file.Addresses, AddressEntry{
		Exchange:         exchangeName,
		Coin:             strings.ToUpper(coin),
		Network:          network,
		Address:          address,
		SecondaryAddress: memo,
		Label:            exchangeName + " " + strings.ToUpper(coin),
	})

	_ = saveAddressFile(file)
	return true
}

func loadAddressFile() (addressFile, bool) {
	bytes, err := os.ReadFile(addressesPath())
	if err != nil {
		return addressFile{}, false
	}

	var file addressFile
	if err := json.Unmarshal(bytes, &file); err != nil {
		return addressFile{}, false
	}

	return file, true
}

func saveAddressFile(file addressFile) error {
	path := addressesPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}

	bytes = append(bytes, '\n')
	return os.WriteFile(path, bytes, 0o644)
}

func addressesPath() string {
	return filepath.Join("automation", "addresses.json")
}
