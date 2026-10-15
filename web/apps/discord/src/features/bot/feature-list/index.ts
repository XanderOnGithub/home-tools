// Everything the bot does that has settings (decision #60): one list, so
// the overview's cards and the Features page's tabs never drift. Each
// entry knows its sub-page (/features/<id>) and how to sum itself up in
// a line for the overview.
import type { BotView, Config } from '@/features/bot'

export type FeatureId = 'game-statuses' | 'names' | 'polls' | 'status' | 'blob'

export type FeatureInfo = {
  id: FeatureId
  label: string
  icon: string // SVG path, 24×24 box, stroked
  /** On/off for features with a switch; null for ones that are always on. */
  on: (c: Config) => boolean | null
  /** One line for the overview card: what it's doing right now. */
  summary: (c: Config, b: BotView) => string
}

const plural = (n: number, one: string, many = `${one}s`) => `${n} ${n === 1 ? one : many}`

/** "#general", or "a channel" before the bot has seen it. */
function channelName(b: BotView, id: string): string {
  for (const g of b.guilds) {
    const c = g.channels.find((c) => c.id === id)
    if (c) return `#${c.name}`
  }
  return id ? 'a channel' : 'no channel yet'
}

export const FEATURES: FeatureInfo[] = [
  {
    id: 'game-statuses',
    label: 'Game statuses',
    icon: 'M4 4h16v7H4z M4 13h16v7H4z M8 7.5h.01 M8 16.5h.01', // the games tool's server rack
    on: () => null,
    summary: (c) =>
      c.status_boards.length ? `${plural(c.status_boards.length, 'live message')} in Discord` : 'None yet',
  },
  {
    id: 'names',
    label: 'Daily name',
    icon: 'M20 12a8 8 0 1 1-16 0 8 8 0 0 1 16 0z M12 7v5l3 2', // a clock: a new one every midnight
    on: () => null,
    summary: (c, b) => `${b.persona.name} today · ${plural(c.names.length, 'name')}`,
  },
  {
    id: 'polls',
    label: 'Scheduled polls',
    icon: 'M5 20V10 M12 20V4 M19 20v-7', // a bar chart
    on: (c) => c.features.poll.enabled,
    summary: (c, b) => {
      const p = c.features.poll
      if (!p.enabled) return 'Off'
      const every = p.every_days === 1 ? 'Every day' : `Every ${p.every_days} days`
      return `${every} in ${channelName(b, p.channel_id)}`
    },
  },
  {
    id: 'status',
    label: 'Bot status',
    icon: 'M4 6h16 M4 12h10 M4 18h7', // lines of text
    on: (c) => c.features.status.enabled,
    summary: (c, b) => (c.features.status.enabled ? (b.status ? `“${b.status}”` : 'A new line every hour') : 'Off'),
  },
  {
    id: 'blob',
    label: 'Blob maker',
    icon: 'M12 3c4 0 8 3 8 8s-3 10-8 10-8-4-8-9 4-9 8-9z M9.5 11h.01 M14.5 11h.01', // a blob with eyes
    on: (c) => c.features.blob.enabled,
    summary: (c) => (c.features.blob.enabled ? '/blob for anyone' : 'Off'),
  },
]

export const featureById = (id: string) => FEATURES.find((f) => f.id === id)
