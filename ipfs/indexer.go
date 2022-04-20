package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Peltoche/ipfs-gh1000/metadata"
	shell "github.com/ipfs/go-ipfs-api"
)

type Indexer struct {
	shell      *shell.Shell
	indexName  string
	indexKeyID string
}

func NewIndexer(shell *shell.Shell, indexName string) (*Indexer, error) {
	keyList, err := shell.KeyList(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the key list: %w", err)
	}

	var indexKeyID string

	for _, key := range keyList {
		if key.Name == indexName {
			indexKeyID = key.Id
		}
	}

	if indexKeyID == "" {
		return nil, fmt.Errorf("key %q not found in the key list", indexName)
	}

	return &Indexer{shell, indexName, indexKeyID}, nil
}

func (i *Indexer) RetrieveIndex(ctx context.Context) (map[string]metadata.RepoMetadata, error) {
	indexCID, err := i.shell.Resolve(i.indexKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve the index id: %w", err)
	}

	raw, err := i.shell.Cat(indexCID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the raw index: %w", err)
	}

	data := map[string]metadata.RepoMetadata{}
	err = json.NewDecoder(raw).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode the index: %w", err)
	}

	return data, nil
}

func (i *Indexer) SaveIndex(ctx context.Context, index map[string]metadata.RepoMetadata) error {
	rawIndex, err := json.Marshal(index)
	if err != nil {
		return fmt.Errorf("failed to encode the index: %w", err)
	}

	indexCID, err := i.shell.Add(bytes.NewReader(rawIndex), shell.Pin(true))
	if err != nil {
		return fmt.Errorf("failed to save the new index: %s", err)
	}

	log.Println("start publishing the new index")
	lifetime, _ := time.ParseDuration("2400H") // 100 days
	ttl, _ := time.ParseDuration("1H")

	_, err = i.shell.PublishWithDetails(indexCID, i.indexName, lifetime, ttl, true)
	if err != nil {
		return fmt.Errorf("failed to publish the new index: %w", err)
	}
	log.Println("new index successfully published")

	return nil
}
