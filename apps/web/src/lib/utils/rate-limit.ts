/**
 * Rate limiting utility for auth operations
 * Helps prevent excessive API calls and implements backoff strategies
 */

interface RateLimiter {
  canProceed(key: string): boolean;
  recordAttempt(key: string, success: boolean): void;
  getRetryDelay(key: string): number;
  reset(key: string): void;
}

class ExponentialBackoffRateLimiter implements RateLimiter {
  private attempts = new Map<string, { count: number; lastAttempt: number; failures: number }>();
  private readonly maxAttempts: number;
  private readonly windowMs: number;
  private readonly baseDelayMs: number;

  constructor(maxAttempts = 5, windowMs = 60000, baseDelayMs = 1000) {
    this.maxAttempts = maxAttempts;
    this.windowMs = windowMs;
    this.baseDelayMs = baseDelayMs;
  }

  canProceed(key: string): boolean {
    const now = Date.now();
    const attempt = this.attempts.get(key);

    if (!attempt) {
      return true;
    }

    // Reset if window has expired
    if (now - attempt.lastAttempt > this.windowMs) {
      this.attempts.delete(key);
      return true;
    }

    // Check if we've exceeded max attempts
    return attempt.count < this.maxAttempts;
  }

  recordAttempt(key: string, success: boolean): void {
    const now = Date.now();
    const existing = this.attempts.get(key) || { count: 0, lastAttempt: 0, failures: 0 };

    this.attempts.set(key, {
      count: existing.count + 1,
      lastAttempt: now,
      failures: success ? 0 : existing.failures + 1
    });
  }

  getRetryDelay(key: string): number {
    const attempt = this.attempts.get(key);
    if (!attempt || attempt.failures === 0) {
      return 0;
    }

    // Exponential backoff: baseDelay * (2 ^ failures)
    return this.baseDelayMs * Math.pow(2, Math.min(attempt.failures, 6));
  }

  reset(key: string): void {
    this.attempts.delete(key);
  }
}

// Rate limiter for token refresh operations
export const tokenRefreshLimiter = new ExponentialBackoffRateLimiter(3, 60000, 1000);

// Rate limiter for user profile fetches
export const profileFetchLimiter = new ExponentialBackoffRateLimiter(5, 30000, 500);

/**
 * Rate-limited async function wrapper
 */
export async function withRateLimit<T>(
  limiter: RateLimiter,
  key: string,
  fn: () => Promise<T>
): Promise<T> {
  if (!limiter.canProceed(key)) {
    const delay = limiter.getRetryDelay(key);
    throw new Error(`Rate limit exceeded. Retry after ${delay}ms`);
  }

  try {
    const result = await fn();
    limiter.recordAttempt(key, true);
    return result;
  } catch (error) {
    limiter.recordAttempt(key, false);
    throw error;
  }
}

/**
 * Retry wrapper with exponential backoff
 */
export async function withRetry<T>(
  fn: () => Promise<T>,
  maxRetries = 3,
  baseDelay = 1000,
  shouldRetry?: (error: unknown) => boolean
): Promise<T> {
  let lastError: unknown;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;
      
      // Don't retry if we've exceeded max retries
      if (attempt === maxRetries) {
        break;
      }
      
      // Don't retry if shouldRetry function says not to
      if (shouldRetry && !shouldRetry(error)) {
        break;
      }
      
      // Calculate delay with jitter
      const delay = baseDelay * Math.pow(2, attempt);
      const jitter = Math.random() * 0.1 * delay; // 10% jitter
      await new Promise(resolve => setTimeout(resolve, delay + jitter));
    }
  }

  throw lastError;
}