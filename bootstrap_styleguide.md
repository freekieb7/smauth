# Application Style Guide

## Branding Colors
**Primary:** #0B1962
**Secondary:** #24E362
**White:** #FFFFFF
**Light Gray:** #F2F4F8
**Dark:** #05103A

### Color Usage
- **Primary (#0B1962):** Navigation bars, headings, primary buttons
- **Secondary (#24E362):** Call-to-actions, success messages, accent buttons
- **White (#FFFFFF):** Backgrounds, text on dark surfaces
- **Light Gray (#F2F4F8):** Section backgrounds and cards
- **Dark (#05103A):** Text on light backgrounds, hover states

### Bootstrap SCSS Variables
```scss
$primary: #0B1962;
$secondary: #24E362;
$success: #24E362;
$light: #F2F4F8;
$dark: #05103A;

$body-bg: #FFFFFF;
$body-color: #0B1962;

$link-color: $secondary;
$link-hover-color: darken($secondary, 10%);
```

## Typography
- **Font Family:** `Inter`, `Roboto`, or `system-ui`
- **Headings:** Bold, dark color
- **Body text:** Regular, primary dark text (#05103A)
- **Line height:** 1.5x

## Buttons
### Primary Button
- Background: #0B1962
- Text: #FFFFFF
- Hover: darken(#0B1962, 10%)
- Border Radius: 6px

### Secondary Button
- Background: #24E362
- Text: #0B1962
- Hover: darken(#24E362, 10%)
- Border Radius: 6px

## Components
### Cards
- Background: #FFFFFF
- Border Radius: 8px
- Shadow: subtle, soft
- Padding: 1.25rem

### Navbar
- Background: #0B1962
- Link Color: #FFFFFF
- Hover: #24E362

## Spacing
- Use multiples of `4px` (4, 8, 12, 16, 24, 32)
- Section padding: 48px top & bottom

## Icons
- Use outline style
- Icon color: Primary or Gray-600

## Form Elements
- Border radius: 6px
- Focus outline: secondary green
- Background: white

## Shadows
- Soft layer shadow
- Example: `0 4px 12px rgba(0,0,0,0.08)`

## Copilot Prompt Examples
**Prompt: Create a primary button:**
> Generate a Bootstrap button using the style guide. Primary is #0B1962, rounded, white text.

**Prompt: Create a card design:**
> Build a card with white background, soft shadow, 16px padding, rounded corners, and title + text.

