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
	Hint          string                  `json:"hint"`
	VpnActive     bool                    `json:"vpnActive"`
	GnomeTrayTip  bool                    `json:"gnomeTrayTip"`
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
		Hint:          m.dashboardHint(len(peers)),
		VpnActive:     vpnActive(),
		GnomeTrayTip:  m.gnomeTray,
	}
}

func (m *MercuryApp) dashboardHint(peerCount int) string {
	if peerCount == 0 {
		return ""
	}
	if m.syncSvc != nil && m.syncSvc.DecryptFailCount() >= 3 {
		return "Your passphrase does not match another device. Same secret on every machine, or I cannot deliver."
	}
	return ""
}

// DashboardFingerprint returns a compact string used to detect UI-relevant changes.
func (m *MercuryApp) DashboardFingerprint() string {
	var b strings.Builder
	peerCount := 0
	if m.syncSvc != nil {
		peers := m.syncSvc.GetPeers()
		peerCount = len(peers)
		fmt.Fprintf(&b, "p=%d", peerCount)
		for _, p := range peers {
			fmt.Fprintf(&b, " %s@%s", p["id"], p["addr"])
		}
	} else {
		b.WriteString("p=0")
	}

	paused := m.IsPaused()
	hasPass := m.GetSavedPassphrase() != ""
	fmt.Fprintf(&b, " paused=%v pass=%v", paused, hasPass)

	offers := m.GetPendingFileOffers()
	if offers == nil {
		offers = []services.FileOffer{}
	}
	for _, o := range offers {
		fmt.Fprintf(&b, " o=%s", o.ID)
	}

	transfers := m.GetTransferProgress()
	for _, t := range transfers {
		if t.Status != "done" {
			fmt.Fprintf(&b, " t=%s:%s:%d", t.ID, t.Status, t.Received)
		}
	}

	hint := m.dashboardHint(peerCount)
	fmt.Fprintf(&b, " hint=%q vpn=%v gnome=%v active=%v", hint, vpnActive(), m.gnomeTray, m.trayActive())
	return b.String()
}

func (m *MercuryApp) notifyChange() {
	m.syncClipboardWatch()
	if m.emitChange != nil {
		m.emitChange()
	}
}
