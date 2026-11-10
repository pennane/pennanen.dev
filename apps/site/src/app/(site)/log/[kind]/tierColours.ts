const TIER_COLOURS = ['#ff7f7f', '#ffbf7f', '#ffdf7f', '#ffff7f', '#bfff7f', '#7fff7f']

export const tierColour = (position: number) =>
  TIER_COLOURS[Math.min(position, TIER_COLOURS.length - 1)]
