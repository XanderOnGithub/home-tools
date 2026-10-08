// Dates as plain "YYYY-MM-DD" strings in the user's local time zone (what a
// person means by "today"), matching the API's date-only fields.

/** Formats d as "YYYY-MM-DD" in local time. */
export function isoDate(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

/** Today's local date, "YYYY-MM-DD". */
export function today(): string {
  return isoDate(new Date())
}

/** Parses "YYYY-MM-DD" as a local date (midnight), not UTC. */
export function parseDate(iso: string): Date {
  const [y, m, d] = iso.split('-').map(Number)
  return new Date(y, m - 1, d)
}

/** Monday 00:00 (local) of the week containing d. Weeks start Monday. */
export function startOfWeek(d: Date): Date {
  const start = new Date(d.getFullYear(), d.getMonth(), d.getDate())
  const daysSinceMonday = (start.getDay() + 6) % 7 // getDay: 0 = Sunday
  start.setDate(start.getDate() - daysSinceMonday)
  return start
}

/** The 7 days (Monday first) of the week containing d. */
export function weekDays(d: Date): Date[] {
  const start = startOfWeek(d)
  return Array.from({ length: 7 }, (_, i) => new Date(start.getFullYear(), start.getMonth(), start.getDate() + i))
}

/**
 * ISO 8601 week of d, e.g. "2026-W41" (weeks start Monday; week 1 is the
 * week with the year's first Thursday). Matches the API's skipped-week field.
 */
export function isoWeek(d: Date): string {
  // The Thursday of d's week decides which year the week belongs to.
  const thursday = new Date(d.getFullYear(), d.getMonth(), d.getDate() + 3 - ((d.getDay() + 6) % 7))
  const jan1 = new Date(thursday.getFullYear(), 0, 1)
  // Round to whole days first: across a daylight-saving change a "day"
  // is 23 or 25 hours, and dividing raw milliseconds would be off by one.
  const days = Math.round((thursday.getTime() - jan1.getTime()) / 86_400_000)
  const week = 1 + Math.floor(days / 7)
  return `${thursday.getFullYear()}-W${String(week).padStart(2, '0')}`
}

const WEEKDAY_KEYS = ['sunday', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'] as const

/** "monday" … "sunday" for d (the API's schedule keys). */
export function weekdayKey(d: Date): (typeof WEEKDAY_KEYS)[number] {
  return WEEKDAY_KEYS[d.getDay()]
}
