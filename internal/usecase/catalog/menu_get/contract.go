package menu_get

import (
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
)

type In struct {
	VenueRef string
}

type Out struct {
	Menu *catalogrepo.Menu
	ETag string
}
