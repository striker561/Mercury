package app

import (
	"fmt"
	"strings"

	"mercury/app/services"
)

// DashboardState is a single IPC payload for the frontend home view.
type DashboardState struct {
	Peers         []map[string]string     `json:"peers"`
	Paused        bool                    `json:"paused"`
	HasPassphrase bool                    `json:"hasPassphrase"`
	Offers        []services.FileOffer    `json:"offers"`
	Transfers     []services.FileProgress `json:"transfers"`
}

// GetDashboardState returns peers, sync status, offers, and transfers in one call.
func (m *MercuryApp) GetDashboardState() DashboardState {
	offers := m.GetPendingFileOffers()
	if offers == nil {
		offers = []services.FileOffer{}
	}

	transfers := m.GetTransferProgress()
	active := make([]services.FileProgress, 0, len(transfers))
	for _, t := range transfers {
		if t.Status != "done" {
			active = append(active, t)
		}
	}

	peers := m.GetPeers()
	if peers == nil {
		peers = []map[string]string{}
	}

	return DashboardState{
		Peers:         peers,
		Paused:        m.IsPaused(),
		HasPassphrase: m.GetSavedPassphrase() != "",
		Offers:        offers,
		Transfers:     active,
	}
}

// DashboardFingerprint returns a compact string used to detect UI-relevant changes.
func (m *MercuryApp) DashboardFingerprint() string {
	s := m.GetDashboardState()
	var b strings.Builder
	fmt.Fprintf(&b, "p=%d paused=%v pass=%v", len(s.Peers), s.Paused, s.HasPassphrase)
	for _, p := range s.Peers {
		fmt.Fprintf(&b, " %s@%s", p["id"], p["addr"])
	}
	for _, o := range s.Offers {
		fmt.Fprintf(&b, " o=%s", o.ID)
	}
	for _, t := range s.Transfers {
		fmt.Fprintf(&b, " t=%s:%s:%d", t.ID, t.Status, t.Received)
	}
	fmt.Fprintf(&b, " active=%v", m.trayActive())
	return b.String()
}

func (m *MercuryApp) notifyChange() {
	if m.emitChange != nil {
		m.emitChange()
	}
}
