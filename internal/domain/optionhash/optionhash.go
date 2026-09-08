package optionhash

import (
	"crypto/md5" //nolint:gosec // MD5 is used as grouping key per spec, not for security
	"encoding/hex"
	"sort"
	"strings"

	"github.com/google/uuid"
)

func Hash(optionIDs []uuid.UUID) string {
	if len(optionIDs) == 0 {
		sum := md5.Sum(nil) //nolint:gosec
		return hex.EncodeToString(sum[:])
	}

	seen := make(map[uuid.UUID]struct{}, len(optionIDs))
	unique := make([]string, 0, len(optionIDs))
	for _, id := range optionIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, strings.ToLower(id.String()))
	}
	sort.Strings(unique)
	payload := strings.Join(unique, ",")
	sum := md5.Sum([]byte(payload)) //nolint:gosec
	return hex.EncodeToString(sum[:])
}
