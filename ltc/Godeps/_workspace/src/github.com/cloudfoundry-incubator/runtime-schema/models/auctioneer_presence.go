package models

type AuctioneerPresence struct {
	AuctioneerID      string `json:"auctioneer_id"`
	AuctioneerAddress string `json:"auctioneer_address"`
}

func NewAuctioneerPresence(id, address string) AuctioneerPresence {
	return AuctioneerPresence{
		AuctioneerID:      id,
		AuctioneerAddress: address,
	}
}

func (a AuctioneerPresence) Validate() error {
	var validationError ValidationError

	if a.AuctioneerID == "" {
		validationError = validationError.Append(ErrInvalidField{"auctioneer_id"})
	}

	if a.AuctioneerAddress == "" {
		validationError = validationError.Append(ErrInvalidField{"auctioneer_address"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}
