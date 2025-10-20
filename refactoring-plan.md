# Page Refactoring Checklist

## Common Patterns to Replace with CSS Classes:

### Alert Messages
- Replace inline styles for success/error/warning alerts with:
  - `.alert-success-custom`
  - `.alert-danger-custom` 
  - `.alert-warning-custom`
  - `.alert-heading-custom`
  - `.alert-icon-large` / `.alert-icon-small`

### Page Headers
- Replace inline header styles with:
  - `.admin-page-header` (for the container)
  - `.admin-page-title` (for h1 titles)
  - `.admin-page-subtitle` (for subtitle paragraphs)

### Cards
- Replace inline card styles with:
  - `.admin-card` (for main card container)
  - `.admin-card-header` (for card headers)

### Buttons
- Replace inline button styles with:
  - `.btn-primary-gradient` (for primary action buttons)
  - `.btn-secondary-outline` (for secondary buttons)
  - `.btn-sm-custom` (for small buttons)

### Forms
- Replace inline form styles with:
  - `.form-label-custom`
  - `.form-control-custom`
  - `.form-text-custom`

### ID Display Components
- Replace inline ID display styles with:
  - `.id-display-container`
  - `.id-display-value`
  - `.id-copy-button`

### Code/Monospace Display
- Replace inline monospace styles with:
  - `.code-display`
  - `.text-monospace`

### Modals
- Replace inline modal styles with:
  - `.modal-header-primary` / `.modal-header-danger`
  - `.modal-body-custom`
  - `.modal-footer-custom`

### Scope Components (for resource server pages)
- Use:
  - `.scope-badge`
  - `.scope-card`
  - `.scope-card-title`
  - `.scope-card-description`
  - `.scope-card-description-empty`

## Pages to Refactor:

1. ✅ **users.templ** - Partially completed
2. 🔄 **clients.templ** - In progress
3. ❌ **resource_servers.templ** - Needs refactoring
4. ❌ **resource_server_edit.templ** - Needs refactoring  
5. ❌ **audit_logs.templ** - Needs refactoring
6. ❌ **home.templ** - Check if needs refactoring
7. ❌ **login.templ** - Check if needs refactoring
8. ❌ **authorize.templ** - Check if needs refactoring

## Benefits of This Refactoring:

1. **Consistency**: All pages will have the same look and feel
2. **Maintainability**: Changes to styling can be made in one place (CSS file)
3. **Performance**: Reduced HTML size due to fewer inline styles
4. **Accessibility**: Better semantic structure with consistent classes
5. **Responsive**: CSS classes can include responsive breakpoints
6. **Theme Support**: Easy to create theme variations by modifying CSS variables

## Testing Checklist:

- [ ] All pages render correctly
- [ ] Interactive elements (buttons, forms, modals) work properly
- [ ] Copy functionality works on all ID displays
- [ ] Responsive design works on mobile devices
- [ ] Colors and spacing are consistent across all pages
- [ ] Accessibility is maintained or improved