import { vi, beforeAll, afterAll } from 'vitest'

// Store original console methods and process.stderr.write
const originalConsole = {
  log: console.log,
  debug: console.debug,
  info: console.info,
  warn: console.warn,
  error: console.error,
}

const originalStderrWrite = process.stderr.write

beforeAll(() => {
  // Mock console methods to suppress output during tests
  // Only suppress debug, info, and log - keep warn and error for important issues
  console.debug = vi.fn()
  console.info = vi.fn()
  console.log = vi.fn()
  
  // Optionally suppress warn for specific expected warnings
  console.warn = vi.fn()
  
  // Keep error visible but could be mocked if needed
  // console.error = vi.fn()

  // Suppress stderr output during tests (this will suppress stack traces from intentional errors)
  process.stderr.write = vi.fn().mockReturnValue(true)
})

afterAll(() => {
  // Restore original console methods and stderr after all tests
  console.log = originalConsole.log
  console.debug = originalConsole.debug
  console.info = originalConsole.info
  console.warn = originalConsole.warn
  console.error = originalConsole.error
  process.stderr.write = originalStderrWrite
})
