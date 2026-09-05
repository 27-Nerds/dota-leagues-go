export const BASE_URL = process.env.baseUrl;

export const REGIONS = {
  0: "Unknown",
  1: "North America",
  2: "South America",
  3: "Europe",
  4: "CIS",
  5: "China",
  6: "South-East Asia"
};

export const TIERS = {
  0: "Unknown (0)",
  1: "Unknown (1)",
  2: "Non-DPC Events (2)",
  3: "DPC Minor Events (3)",
  4: "DPC Major Events (4)",
  5: "The International (5)"
};

// ELeagueStatus from Dota's dota_shared_enums.proto (approval/lifecycle state).
export const LEAGUE_STATUS = {
  0: "Not set",
  1: "Unsubmitted",
  2: "Submitted",
  3: "Accepted",
  4: "Rejected",
  5: "Concluded",
  6: "Deleted"
};

export function formatDate(unixSeconds) {
  const date = new Date(Number(unixSeconds) * 1000);
  if (!unixSeconds || !Number.isFinite(date.getTime())) return '—';
  return new Intl.DateTimeFormat('en-GB', { timeZone: 'UTC' }).format(date);
}

export function formatDateTime(unixSeconds) {
  const date = new Date(Number(unixSeconds) * 1000);
  if (!unixSeconds || !Number.isFinite(date.getTime())) return '—';
  return new Intl.DateTimeFormat('en-GB', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'UTC' }).format(date);
}

export function formatMoney(amount) {
  if (amount == null || amount === '') return '—';
  const value = Number(amount);
  return Number.isFinite(value) ? new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(value) : '—';
}

export function normalizeUrl(value) {
  if (typeof value !== 'string' || !value.trim() || value.trim() === '-') return null;
  try {
    const input = value.trim();
    const url = new URL(/^[a-z][a-z0-9+.-]*:/i.test(input) ? input : `https://${input}`);
    return ['http:', 'https:'].includes(url.protocol) ? url.href : null;
  } catch { return null; }
}

export function regionText(region) {
  return REGIONS[region] || "Unknown";
}

const countryNames = new Intl.DisplayNames(['en'], { type: 'region', fallback: 'none' });
export function countryText(code) {
  if (typeof code !== 'string' || !/^[a-z]{2}$/i.test(code.trim())) return '';
  return countryNames.of(code.trim().toUpperCase()) || '';
}

export function tierText(tier) {
  return TIERS[tier] || "Unknown";
}

export function leagueLogo(leagueId) {
  if (!leagueId) return null;
  return `${BASE_URL}/${leagueId}/logo.png`;
}

export function teamLogo(teamId) {
  if (!teamId) return null;
  return `${BASE_URL}/teams/${teamId}/logo.png`;
}

export function steamProfileUrl(accountId) {
  if (!/^[1-9][0-9]*$/.test(String(accountId))) return null;
  try {
    const id = BigInt(accountId);
    if (id > 4294967295n) return null;
    return `https://steamcommunity.com/profiles/${id + 76561197960265728n}/`;
  } catch { return null; }
}
