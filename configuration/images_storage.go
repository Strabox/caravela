package configuration

import (
	"fmt"
)

const (
	ImagesStorageDockerHub = "DockerHub"
	ImagesStorageIPFS      = "IPFS"
)

type imagesStorageBackend struct {
	Backend string
}

/*
Implementation of encoding.TextUnmarshal interface for the imagesStorageType.
*/
func (d *imagesStorageBackend) UnmarshalText(text []byte) error {
	temp := string(text)
	if temp == ImagesStorageDockerHub || temp == ImagesStorageIPFS {
		d.Backend = temp
		return nil
	} else {
		d.Backend = ""
		return fmt.Errorf("invalid image storage: %s", temp)
	}
}
