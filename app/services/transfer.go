// Package services — business logic layer between the Wails frontend
// and the backend engines.  TransferService wraps transfer.Manager.
package services

import (
	"mercury/app/backend/transfer"
)

// Re-exported types so app.go doesn't import the backend package.
type FileOffer = transfer.Offer
type FileProgress = transfer.Progress

// TransferService wraps file transfer: offers, accept/reject, streaming.
// Backend: transfer.Manager.
type TransferService struct {
	mgr *transfer.Manager
}

// NewTransferService creates a transfer service using the given encryption key.
func NewTransferService(key []byte) *TransferService {
	return &TransferService{mgr: transfer.NewManager(key)}
}

// OnWireMessage routes a decrypted file-transfer frame to the active receive.
func (s *TransferService) OnWireMessage(msgType byte, payload []byte) {
	s.mgr.OnWireMessage(msgType, payload)
}

// IncomingOffer registers a file offer from a peer.
func (s *TransferService) IncomingOffer(fileName string, fileSize int64, peerAddr string) FileOffer {
	return s.mgr.IncomingOffer(fileName, fileSize, peerAddr)
}

// IncomingOfferWithID registers a file offer using the sender's offer ID.
func (s *TransferService) IncomingOfferWithID(offerID, fileName string, fileSize int64, peerAddr string) FileOffer {
	return s.mgr.IncomingOfferWithID(offerID, fileName, fileSize, peerAddr)
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

// NewOfferID returns a unique offer identifier.
func (s *TransferService) NewOfferID() string {
	return s.mgr.NewOfferID()
}

// StoreOutgoing remembers an offer we broadcast so we can send the file when accepted.
func (s *TransferService) StoreOutgoing(offerID, filePath string) {
	s.mgr.StoreOutgoing(offerID, filePath)
}

// AcceptNotification is called when a remote peer accepts our offer — returns the file path.
func (s *TransferService) AcceptNotification(offerID string) string {
	return s.mgr.AcceptNotification(offerID)
}

// SendFile sends a file to a peer. Returns a transfer ID.
func (s *TransferService) SendFile(peerAddr, filePath string) (string, error) {
	return s.mgr.SendFile(peerAddr, filePath)
}

// SendFileForOffer sends a file after the peer accepted a specific offer.
func (s *TransferService) SendFileForOffer(offerID, peerAddr, filePath string) (string, error) {
	return s.mgr.SendFileForOffer(offerID, peerAddr, filePath)
}

// AllProgress returns progress for active transfers.
func (s *TransferService) AllProgress() []FileProgress {
	return s.mgr.AllProgress()
}

// CancelTransfer signals a running transfer to stop.
func (s *TransferService) CancelTransfer(tid string) {
	s.mgr.CancelTransfer(tid)
}
