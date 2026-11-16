// Release times are shown in UTC, formatted by hand so months match the "Sep" style regardless of the browser's locale data
const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

const pad = (n: number) => String(n).padStart(2, "0");

// e.g. "3 Mar 2026"
export function formatDate(ts: Date): string {
  return `${ts.getUTCDate()} ${MONTHS[ts.getUTCMonth()]} ${ts.getUTCFullYear()}`;
}

// e.g. "14:05"
export function formatTime(ts: Date): string {
  return `${pad(ts.getUTCHours())}:${pad(ts.getUTCMinutes())}`;
}

export function formatRelative(ts: Date, now: Date): string {
  const total = (now.getTime() - ts.getTime()) / 1000;
  const abs = Math.abs(total);
  if (abs < 60) return "just now";
  let unit: string;
  if (abs < 3600) unit = `${Math.floor(abs / 60)}m`;
  else if (abs < 86400) unit = `${Math.floor(abs / 3600)}h`;
  else unit = `${Math.floor(abs / 86400)}d`;
  return total >= 0 ? `${unit} ago` : `in ${unit}`;
}
