# The Complete Dark Mode Guide
tags: tutorial, dark-mode, design

Dark mode isn't just inverting colors. It's about creating a comfortable
reading experience that respects user preferences.

## Principles

- Use ~~pure black~~ near-black backgrounds like `bg-gray-950`
- Reduce contrast slightly — *very bright* white on black causes eye strain
- Use **muted** accent colors that don't overwhelm in dark environments
- Maintain __consistent__ hierarchy between light and dark modes

## Implementation

Every element needs both a light and dark variant. For example:

- `bg-white dark:bg-gray-900` for surfaces
- `text-gray-900 dark:text-gray-100` for primary text
- `border-gray-200 dark:border-gray-800` for borders

## Testing

Always test your dark mode with:

1. System preference toggling
2. Manual toggle button
3. Different screen brightnesses
