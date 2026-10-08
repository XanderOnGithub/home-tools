// Unit conversion for display and input. The API always stores metric
// (kg, m); the UI converts at the edges, per the profile's units (#13).

const KG_PER_LB = 0.45359237 // exact, by definition
const M_PER_IN = 0.0254 // exact, by definition

export const lbToKg = (lb: number) => lb * KG_PER_LB
export const kgToLb = (kg: number) => kg / KG_PER_LB

export const cmToM = (cm: number) => cm / 100
export const ftInToM = (ft: number, inches: number) => (ft * 12 + inches) * M_PER_IN

/** Height in whole feet and inches (e.g. 1.8 m → 5 ft 11 in). */
export function mToFtIn(m: number): { ft: number; inches: number } {
  const total = Math.round(m / M_PER_IN)
  return { ft: Math.floor(total / 12), inches: total % 12 }
}

/** Parses a typed number, accepting "80,5" as well as "80.5". NaN if empty or not a number. */
export function parseNumber(text: string): number {
  const t = text.trim().replace(',', '.')
  return t === '' ? NaN : Number(t)
}
