import { ShieldWarning, Tray, Warning } from "@phosphor-icons/react";

const ICONS = {
  vpn: ShieldWarning,
  tray: Tray,
  warn: Warning,
} as const;

export interface NoticeItem {
  icon?: keyof typeof ICONS;
  title?: string;
  text: string;
}

interface Props {
  items: NoticeItem[];
}

export default function HeaderNotice({ items }: Props) {
  if (!items.length) return null;

  const Icon = ICONS[items[0].icon ?? "warn"];

  return (
    <div className="header-notice">
      <button
        type="button"
        className="header-notice-btn"
        aria-label={items.length === 1 ? "1 notice" : `${items.length} notices`}
      >
        <Icon size={16} weight="fill" />
      </button>
      <div className="header-notice-pop" role="status">
        {items.map((item, i) => {
          const ItemIcon = ICONS[item.icon ?? "warn"];
          return (
            <div key={i} className="header-notice-item">
              <ItemIcon size={13} weight="fill" aria-hidden />
              <div>
                {item.title && (
                  <div className="header-notice-title">{item.title}</div>
                )}
                <div className="header-notice-text">{item.text}</div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
