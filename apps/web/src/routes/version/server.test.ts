import { describe, it, expect } from 'vitest'
import { GET } from './+server'

describe('/version API endpoint', () => {
  it('should return version information with correct structure', async () => {
    const response = await GET()
    const body = await response.json()

    expect(response.status).toBe(200)
    
    // Check that all required fields are present and are strings
    expect(body).toHaveProperty('commit')
    expect(body).toHaveProperty('commitShort')
    expect(body).toHaveProperty('buildTime')
    expect(body).toHaveProperty('tag')
    
    expect(typeof body.commit).toBe('string')
    expect(typeof body.commitShort).toBe('string')
    expect(typeof body.buildTime).toBe('string')
    expect(typeof body.tag).toBe('string')
  })

  it('should handle commit SHA shortening correctly', async () => {
    const response = await GET()
    const body = await response.json()
    
    // If commit is 'unknown', commitShort should also be 'unknown'
    if (body.commit === 'unknown') {
      expect(body.commitShort).toBe('unknown')
    } else if (body.commit.length >= 7) {
      // If commit is long enough, commitShort should be first 7 characters
      expect(body.commitShort).toBe(body.commit.slice(0, 7))
    } else {
      // If commit is shorter than 7 chars, commitShort should be the full commit
      expect(body.commitShort).toBe(body.commit)
    }
  })

  it('should handle unknown values gracefully', async () => {
    const response = await GET()
    const body = await response.json()

    // Should not throw and should have all required fields
    expect(typeof body.commit).toBe('string')
    expect(typeof body.commitShort).toBe('string')
    expect(typeof body.buildTime).toBe('string')
    expect(typeof body.tag).toBe('string')
    
    // Values should not be empty strings
    expect(body.commit.length).toBeGreaterThan(0)
    expect(body.commitShort.length).toBeGreaterThan(0)
    expect(body.buildTime.length).toBeGreaterThan(0)
    expect(body.tag.length).toBeGreaterThan(0)
  })
})