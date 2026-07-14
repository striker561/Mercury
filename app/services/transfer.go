package services

import (
	"mercury/app/backend/transfer"
)

// Re-exported types so app.go doesn't import the backend package.
type FileOffer = transfer.Offer
type FileProgress = transfer.Progress

// TransferService wraps file offer/accept/reject/transfer logic.
// Backend: transfer.Manager.
type TransferService struct {
	mgr *transfer.Manager
}

// NewTransferService creates a transfer service using the given encryption key.
func NewTransferService(key []byte) *TransferService {
	return &TransferService{mgr: transfer.NewManager(key)}
}

// ChunkChan returns the channel where decrypted file chunks should be sent.
func (s *TransferService) ChunkChan() chan<- []byte {
	return s.mgr.ChunkChan()
}

// IncomingOffer registers a file offer from a peer.
func (s *TransferService) IncomingOffer(fileName string, fileSize int64, peerAddr string) FileOffer {
	return s.mgr.IncomingOffer(fileName, fileSize, peerAddr)
}

// PendingOffers returns all offers not yet acted on.
func (s *TransferService) PendingOffers() []FileOffer {
	return s.mgr.PendingOffers()
}

// AcceptOffer starts receiving a file. Returns a transfer ID.
func (s *TransferService) AcceptOffer(offerID, saveDir string) (string, error) {
	return s.mgr.AcceptOffer(offerID, saveDir)
}

// RejectOffer discards a pending offer.
func (s *TransferService) RejectOffer(offerID string) {
	s.mgr.RejectOffer(offerID)
}

// SendFile sends a file to a peer. Returns a transfer ID.
func (s *TransferService) SendFile(peerAddr, filePath string) (string, error) {
	return s.mgr.SendFile(peerAddr, filePath)
}

// AllProgress returns progress for active transfers.
func (s *TransferService) AllProgress() []FileProgress {
	return s.mgr.AllProgress()
}
