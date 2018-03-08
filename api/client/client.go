package client

import (
)


type CaravelaClient interface {
	Offer(destIP string, destGuid string, suppIP string, offerID int, amount int) error
}