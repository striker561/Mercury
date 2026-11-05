export interface Peer {
  id: string;
  hostname?: string;
  addr: string;
  reachable?: boolean;
  lastSeen?: string;
}

export interface FileOffer {
  id: string;
  file_name: string;
  file_size: number;
  peer_addr: string;
}

export interface FileProgress {
  id: string;
  file_name: string;
  file_size: number;
  received: number;
  speed: number;
  status: string;
}

export interface DashboardState {
  peers: Peer[];
  paused: boolean;
  hasPassphrase: boolean;
  offers: FileOffer[];
  transfers: FileProgress[];
  hint?: string;
  syncRunning?: boolean;
  syncError?: string;
  vpnActive?: boolean;
  gnomeTrayTip?: boolean;
}

export interface AppSettings {
  passphrase?: string;
  paused?: string;
  allow_files?: string;
  autostart?: string;
  auto_accept?: string;
  received_folder?: string;
  version?: string;
}

export type Tab = "home" | "settings";
