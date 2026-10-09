import type { ProfileColor } from '../types'

export { idFromName } from '../../ids'

export const COLORS: ProfileColor[] = ['green', 'blue', 'orange', 'purple']

/** The first color nobody uses yet, so new profiles start out distinct. */
export function firstFreeColor(used: ProfileColor[]): ProfileColor {
  return COLORS.find((c) => !used.includes(c)) ?? COLORS[0]
}
