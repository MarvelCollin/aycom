# MagniView CSS

MagniView is a lightweight utility-first CSS framework inspired by Tailwind CSS, but built as a vanilla CSS implementation. It provides a comprehensive set of utility classes to build modern websites without writing custom CSS.

## Features

- 🎨 Complete utility-based CSS framework
- 🔄 Responsive design utilities
- 📱 Mobile-first approach
- 🧩 Modular architecture
- 🚀 Lightweight and performant
- 💻 No build step required

## Getting Started

1. Link the MagniView CSS file in your HTML:

```html
<link rel="stylesheet" href="magniview.css">
```

2. Start using utility classes in your HTML:

```html
<div class="flex items-center justify-between p-4 bg-blue-500 text-white">
  <h1 class="text-2xl font-bold">MagniView Demo</h1>
  <button class="bg-white text-blue-500 px-4 py-2 rounded hover:bg-blue-50 transition">Click Me</button>
</div>
```

## Available Utilities

MagniView includes utilities for:

- Typography (font family, size, weight, etc.)
- Colors (text, background, border)
- Spacing (margin, padding)
- Sizing (width, height)
- Flexbox & Grid layouts
- Backgrounds & Borders
- Filters & Effects
- Transforms (scale, rotate, translate, skew)
- Transitions & Animations
- Interactive states (hover, focus, active)
- Responsive design

## Responsive Design

MagniView includes breakpoint prefixes for responsive design:

- `sm:` - Small screens (640px and up)
- `md:` - Medium screens (768px and up)
- `lg:` - Large screens (1024px and up)
- `xl:` - Extra large screens (1280px and up)
- `2xl:` - 2X large screens (1536px and up)

Example:

```html
<div class="w-full md:w-1/2 lg:w-1/3">
  Responsive column
</div>
```

## Interactive States

MagniView includes utility variants for interactive states:

- `hover:` - Styles applied on hover
- `focus:` - Styles applied on focus
- `active:` - Styles applied when active/pressed
- `group-hover:` - Styles applied to child elements when a parent with class "group" is hovered

Example:

```html
<button class="bg-blue-500 hover:bg-blue-700 text-white focus:ring-2">
  Hover and Focus Effects
</button>
```

## File Structure

The framework is organized into modular files:

```
MagniView/
├── magniview.css          # Main CSS file (imports all modules)
├── src/
│   ├── reset.css          # Modern CSS reset
│   ├── colors.css         # Color utilities
│   ├── typography.css     # Typography utilities
│   ├── spacing-sizing.css # Spacing and sizing utilities
│   ├── layout-flex.css    # Layout utilities
│   ├── flexbox.css        # Flexbox utilities
│   ├── grid.css           # Grid utilities
│   ├── background-border.css # Background and border utilities
│   ├── filter-effect.css  # Filter and effect utilities
│   ├── transform.css      # Transform utilities
│   ├── transition.css     # Transition and animation utilities
│   ├── interaction.css    # Interactive state utilities
│   ├── keyframes.css      # Animation keyframes
│   ├── utilities.css      # Common utilities
│   └── responsive.css     # Responsive utilities
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 
