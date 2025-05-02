import { toastStore } from '../stores/toastStore';
import { writable } from 'svelte/store';

export enum LogLevel {
  TRACE = 0,
  DEBUG = 1,
  INFO = 2,
  WARN = 3,
  ERROR = 4,
  NONE = 5
}

export interface LogEntry {
  timestamp: string;
  level: LogLevel;
  levelName: string;
  message: string;
  data?: any;
  source: string;
}

export const logStore = writable<LogEntry[]>([]);

const MAX_LOGS = 1000;

const addLogEntry = (entry: LogEntry) => {
  logStore.update(logs => {
    const newLogs = [entry, ...logs];
    return newLogs.slice(0, MAX_LOGS);
  });
};

export const clearLogs = () => {
  logStore.set([]);
};

const DEFAULT_LOG_LEVEL = import.meta.env.DEV ? LogLevel.DEBUG : LogLevel.WARN;

const getLogLevel = (): LogLevel => {
  const storedLevel = localStorage.getItem('logLevel');
  if (storedLevel !== null && !isNaN(Number(storedLevel))) {
    return Number(storedLevel);
  }
  return DEFAULT_LOG_LEVEL;
};

const setLogLevel = (level: LogLevel): void => {
  localStorage.setItem('logLevel', level.toString());
};

let currentLogLevel = getLogLevel();

const LOG_STYLES = {
  [LogLevel.TRACE]: 'color: #6b7280',
  [LogLevel.DEBUG]: 'color: #3b82f6',
  [LogLevel.INFO]: 'color: #10b981',
  [LogLevel.WARN]: 'color: #f59e0b',
  [LogLevel.ERROR]: 'color: #ef4444',
};

// Map LogLevel to string names
const LOG_LEVEL_NAMES = {
  [LogLevel.TRACE]: 'TRACE',
  [LogLevel.DEBUG]: 'DEBUG',
  [LogLevel.INFO]: 'INFO',
  [LogLevel.WARN]: 'WARN',
  [LogLevel.ERROR]: 'ERROR',
  [LogLevel.NONE]: 'NONE',
};

type ToastOptions = {
  showToast?: boolean;
  timeout?: number;
}

interface Logger {
  trace(message: string, data?: any, options?: ToastOptions): void;
  debug(message: string, data?: any, options?: ToastOptions): void;
  info(message: string, data?: any, options?: ToastOptions): void;
  warn(message: string, data?: any, options?: ToastOptions): void;
  error(message: string, data?: any, options?: ToastOptions): void;
  getLevel(): LogLevel;
  setLevel(level: LogLevel): void;
}

const createLogger = (prefix: string): Logger => {
  const formatMessage = (message: string) => `[${prefix}] ${message}`;
  
  const log = (level: LogLevel, message: string, data?: any, options: ToastOptions = {}) => {
    // Always store logs regardless of level for Debug panel access
    const timestamp = new Date().toISOString();
    const logEntry: LogEntry = {
      timestamp,
      level,
      levelName: LOG_LEVEL_NAMES[level],
      message,
      data,
      source: prefix
    };
    
    // Add to store
    addLogEntry(logEntry);
    
    // Skip console output if level is below the current log level
    if (level < currentLogLevel) return;
    
    const formattedMessage = formatMessage(message);
    
    // Log to console with appropriate styling
    switch (level) {
      case LogLevel.TRACE:
        console.log(`%c[TRACE] ${timestamp} ${formattedMessage}`, LOG_STYLES[level], data || '');
        break;
      case LogLevel.DEBUG:
        console.log(`%c[DEBUG] ${timestamp} ${formattedMessage}`, LOG_STYLES[level], data || '');
        break;
      case LogLevel.INFO:
        console.log(`%c[INFO] ${timestamp} ${formattedMessage}`, LOG_STYLES[level], data || '');
        break;
      case LogLevel.WARN:
        console.warn(`%c[WARN] ${timestamp} ${formattedMessage}`, LOG_STYLES[level], data || '');
        break;
      case LogLevel.ERROR:
        console.error(`%c[ERROR] ${timestamp} ${formattedMessage}`, LOG_STYLES[level], data || '');
        break;
    }
    
    // Show toast notification if requested
    const { showToast = false, timeout } = options;
    if (showToast) {
      const toastType = level === LogLevel.ERROR ? 'error' 
        : level === LogLevel.WARN ? 'warning'
        : level === LogLevel.INFO ? 'info'
        : 'info';
      
      toastStore.showToast(message, toastType, timeout);
    }
  };
  
  return {
    trace: (message: string, data?: any, options?: ToastOptions) => 
      log(LogLevel.TRACE, message, data, options),
    debug: (message: string, data?: any, options?: ToastOptions) => 
      log(LogLevel.DEBUG, message, data, options),
    info: (message: string, data?: any, options?: ToastOptions) => 
      log(LogLevel.INFO, message, data, options),
    warn: (message: string, data?: any, options?: ToastOptions) => 
      log(LogLevel.WARN, message, data, options),
    error: (message: string, data?: any, options?: ToastOptions) => 
      log(LogLevel.ERROR, message, data, options),
    getLevel: () => currentLogLevel,
    setLevel: (level: LogLevel) => {
      currentLogLevel = level;
      setLogLevel(level);
    }
  };
};

// Export the logger creator
export const createLoggerWithPrefix = createLogger;

// Export a default logger without prefix
export const logger = createLogger('App');

// Add a global log level control function
export const setGlobalLogLevel = (level: LogLevel): void => {
  currentLogLevel = level;
  setLogLevel(level);
  logger.info(`Log level set to ${LogLevel[level]}`, null, { showToast: true });
};

// Expose logger to the window for browser console usage
if (typeof window !== 'undefined') {
  (window as any).logger = logger;
  (window as any).LogLevel = LogLevel;
  (window as any).setLogLevel = setGlobalLogLevel;
  (window as any).clearLogs = clearLogs;
} 