/**
 * API module exports
 * 
 * This barrel file exports all the API functions for easy imports.
 */

// Re-export all API modules
export * from './ai';
export * from './admin';
export * from './auth';

// Handle conflicting exports with explicit re-exports
export { getThreadCategories } from './categories';
export * from './chat';
export { getCommunityCategories } from './community';

export * from './notifications';
export * from './passwordReset';
export * from './suggestions';
export * from './thread';
export * from './trends';
export * from './user';

/**
 * NOTE: user_block.ts is deprecated.
 * All functionality is now available in user.ts
 * user_block.ts will be removed in a future update
 */
// We don't export user_block.ts directly to discourage its use 