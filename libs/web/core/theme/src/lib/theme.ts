export const theme = {
  colors: {
    background: '#0f172a',
    foreground: '#f8fafc',
    primary: '#7c3aed',
    secondary: '#0ea5e9',
    accent: '#22c55e',
  },
  radius: {
    sm: '4px',
    md: '8px',
    lg: '12px',
  },
  spacing: {
    xs: '4px',
    sm: '8px',
    md: '16px',
    lg: '24px',
    xl: '32px',
  },
} as const;

export type Theme = typeof theme;
