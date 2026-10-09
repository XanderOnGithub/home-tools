// Mirrors internal/users.User in Go (JSON keys are snake_case, decision #14).
export type Profile = {
  id: string
  name: string
  color: ProfileColor
  units: 'metric' | 'imperial'
  birthday?: string // "YYYY-MM-DD"
  archived?: boolean
}

// Each value has a contrast-checked accent palette in tokens.css.
export type ProfileColor = 'green' | 'blue' | 'orange' | 'purple'
