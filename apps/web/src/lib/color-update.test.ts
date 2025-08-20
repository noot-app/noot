import { describe, it, expect } from 'vitest'

describe('Color definitions', () => {
  it('should have correct color values in CSS', async () => {
    // Read the app.css file and verify color definitions
    const fs = await import('fs')
    const path = await import('path')
    
    const cssPath = path.resolve(__dirname, '../app.css')
    const cssContent = fs.readFileSync(cssPath, 'utf-8')
    
    // Verify new error color
    expect(cssContent).toContain('--color-error: #b94e48')
    expect(cssContent).toContain('--color-error-content: #ffffff')
    
    // Verify new dark color  
    expect(cssContent).toContain('--color-dark: #2b2621')
    expect(cssContent).toContain('--color-dark-content: #fdf8ee')
    
    // Verify old error color is no longer there
    expect(cssContent).not.toContain('--color-error: #2b2621')
  })
})