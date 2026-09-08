package menu_sync

import (
	"github.com/google/uuid"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
)

type In struct {
	VenueID    uuid.UUID
	Categories []catalogrepo.SyncCategory
}

type Out struct {
	Result catalogrepo.SyncResult
}
