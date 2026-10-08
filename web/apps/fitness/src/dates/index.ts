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
