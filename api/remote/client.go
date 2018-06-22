package remote

import "github.com/strabox/caravela/configuration"

/*
Client for remote CARAVELA's nodes in order to trade/coordinate messages with each other.
*/
type Caravela interface {
	// =============================== Discovery ===============================

	/* Sends a create offer message to a trader from a supplier that wants to offer its resources. */
	CreateOffer(fromSupplierIP string, fromSupplierGUID string, toTraderIP string, toTraderGUID string, offerID int64,
		amount int, cpus int, ram int) error

	/* Sends a refresh message from a trader to a supplier. It is used to mutually know that both are alive. */
	RefreshOffer(toSupplierIP string, fromTraderGUID string, offerID int64) (error, bool)

	/*
		Sends a remove offer message from a supplier to a trader.
		It means the supplier does not handle the offer anymore.
	*/
	RemoveOffer(fromSupplierIP string, fromSupplierGUID, toTraderIP string, toTraderGUID string, offerID int64) error

	/* Sends a get message to obtain all the available offers in a trader. */
	GetOffers(toTraderIP string, toTraderGUID string) (error, []Offer)

	// =============================== Scheduling ===============================

	/* Sends a launch container message to a supplier in order to deploy the container */
	LaunchContainer(toSupplierIP string, fromBuyerIP string, offerID int64, containerImageKey string,
		containerArgs []string, cpus int, ram int) error

	// ============================== Configuration ==============================

	/*
		Sends a message to obtain the system configurations of an existing node. Used by joining nodes to know what are
		the system configuration parameters and the respective values.
	*/
	ObtainConfiguration(systemsNodeIP string) (*configuration.Configuration, error)
}

/*
Client's offer struct/DAO.
*/
type Offer struct {
	ID         int64
	SupplierIP string
}
