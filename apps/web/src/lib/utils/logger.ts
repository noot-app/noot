/**
 * Production-safe logger utility for authentication flows
 * Provides different log levels and can be disabled in production
 */

type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface LoggerConfig {
  enabled: boolean;
  level: LogLevel;
  prefix: string;
}

class Logger {
  private config: LoggerConfig;

  constructor(config: LoggerConfig) {
    this.config = config;
  }

  private shouldLog(level: LogLevel): boolean {
    if (!this.config.enabled) return false;
    
    const levels: Record<LogLevel, number> = {
      debug: 0,
      info: 1,
      warn: 2,
      error: 3
    };
    
    return levels[level] >= levels[this.config.level];
  }

  private formatMessage(level: LogLevel, message: string, ...args: unknown[]): [string, ...unknown[]] {
    const emoji = {
      debug: '🔍',
      info: 'ℹ️',
      warn: '⚠️',
      error: '❌'
    };
    
    return [`${emoji[level]} ${this.config.prefix} ${message}`, ...args];
  }

  debug(message: string, ...args: unknown[]): void {
    if (this.shouldLog('debug')) {
      console.debug(...this.formatMessage('debug', message, ...args));
    }
  }

  info(message: string, ...args: unknown[]): void {
    if (this.shouldLog('info')) {
      console.info(...this.formatMessage('info', message, ...args));
    }
  }

  warn(message: string, ...args: unknown[]): void {
    if (this.shouldLog('warn')) {
      console.warn(...this.formatMessage('warn', message, ...args));
    }
  }

  error(message: string, ...args: unknown[]): void {
    if (this.shouldLog('error')) {
      console.error(...this.formatMessage('error', message, ...args));
    }
  }
}

// Create auth logger with development/production awareness
export const authLogger = new Logger({
  enabled: typeof window !== 'undefined' && 
           (import.meta.env.DEV || import.meta.env.VITE_AUTH_DEBUG === 'true'),
  level: (import.meta.env.VITE_LOG_LEVEL as LogLevel) || 'debug',
  prefix: '[Auth]'
});

// Create API logger
export const apiLogger = new Logger({
  enabled: typeof window !== 'undefined' && 
           (import.meta.env.DEV || import.meta.env.VITE_API_DEBUG === 'true'),
  level: (import.meta.env.VITE_LOG_LEVEL as LogLevel) || 'info',
  prefix: '[API]'
});