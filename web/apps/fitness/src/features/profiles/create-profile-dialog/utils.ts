import type { ProfileColor } from '@/features/profiles/types'

export const COLORS: ProfileColor[] = ['green', 'blue', 'orange', 'purple']

/**
 * Turns a name into a profile ID: lowercase letters, digits and dashes
 * (the server rejects anything else, since IDs become file names).
 * Adds -2, -3, … if the ID is taken. "Missy B." → "missy-b".
 */
export function idFromName(name: string, taken: Set<string>): string {
  const base = name
    .toLowerCase()
    .normalize('NFD') // "é" → "e" + accent mark…
    .replace(/[̀-ͯ]/g, '') // …then drop the accent mark
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
  if (!base) return ''
  let id = base
  for (let n = 2; taken.has(id); n++) id = `${base}-${n}`
  return id
}

/** The first color nobody uses yet, so new profiles start out distinct. */
export function firstFreeColor(used: ProfileColor[]): ProfileColor {
  return COLORS.find((c) => !used.includes(c)) ?? COLORS[0]
}
